#!/usr/bin/env python3
"""
face_guard.py — PassQuantum Face Recognition Guard
===================================================
Runs as a child process launched by the Go app.
Communicates with Go via a persistent TCP connection on localhost:9876.

Protocol (Python → Go):
  FRAME:<base64-encoded JPEG>       — live camera frame (training mode only)
  PROGRESS:<current>/<total>        — training progress
  TRAINING_DONE                     — all samples captured and saved
  FACE_OK                           — known face reappeared after FACE_LOST
  FACE_LOST                         — known face absent for GRACE_SECONDS

Protocol (Go → Python):
  START_TRAINING                    — begin capturing face samples
  START_MONITOR                     — enter continuous monitor loop
"""

import base64
import os
import pickle
import socket
import sys
import time
from datetime import datetime
from io import BytesIO
from typing import List, Optional, Tuple

import cv2
import face_recognition
import numpy as np

# ==============================
# Constants
# ==============================

HOST = "127.0.0.1"
PORT = 9876
TOLERANCE = 0.45
GRACE_SECONDS = 3.0
CAPTURE_SAMPLES = 20
CONNECT_RETRIES = 10
CONNECT_RETRY_DELAY = 0.5   # seconds
MONITOR_INTERVAL = 0.1      # seconds (100 ms)
FACE_DATA_FILE = "face_data.pkl"
FRAME_QUALITY = 60          # JPEG quality for FRAME messages
FRAME_WIDTH = 320
FRAME_HEIGHT = 240
DETECT_SCALE = 0.5          # resize factor for HOG detection (faster)

# ==============================
# Logging Helpers
# ==============================

def log(message: str) -> None:
    """Print a timestamped log line to stderr (stdout is reserved for protocol)."""
    ts = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    print(f"[{ts}] {message}", file=sys.stderr, flush=True)


# ==============================
# Connection
# ==============================

def connect_to_go() -> socket.socket:
    """
    Connect to Go's TCP server on HOST:PORT.
    Retries up to CONNECT_RETRIES times with CONNECT_RETRY_DELAY between each.
    Exits the process if all attempts fail.
    """
    for attempt in range(1, CONNECT_RETRIES + 1):
        try:
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            sock.connect((HOST, PORT))
            log(f"Connected to Go server on attempt {attempt}")
            return sock
        except ConnectionRefusedError:
            log(f"Connection attempt {attempt}/{CONNECT_RETRIES} failed — retrying in {CONNECT_RETRY_DELAY}s")
            time.sleep(CONNECT_RETRY_DELAY)
    log("ERROR: Could not connect to Go server after all retries. Exiting.")
    sys.exit(1)


def send_line(sock: socket.socket, message: str) -> None:
    """Send a newline-terminated message over the socket."""
    sock.sendall((message + "\n").encode("utf-8"))


# ==============================
# Camera Utilities
# ==============================

def open_camera() -> cv2.VideoCapture:
    """Open the default webcam and return a VideoCapture handle."""
    cap = cv2.VideoCapture(0)
    if not cap.isOpened():
        log("ERROR: Could not open webcam (index 0).")
        sys.exit(1)
    log("Webcam opened successfully.")
    return cap


def read_frame(cap: cv2.VideoCapture) -> Optional[np.ndarray]:
    """
    Read one frame from the capture device.
    Returns the BGR frame or None on failure.
    """
    ok, frame = cap.read()
    if not ok or frame is None:
        return None
    return frame


# ==============================
# Face Detection
# ==============================

def detect_faces_in_frame(
    frame_bgr: np.ndarray,
) -> Tuple[List[Tuple[int, int, int, int]], List[np.ndarray]]:
    """
    Detect faces using HOG model at DETECT_SCALE resolution.
    Returns (locations_full_scale, encodings).
    Locations are scaled back to the original frame dimensions.
    """
    small = cv2.resize(frame_bgr, (0, 0), fx=DETECT_SCALE, fy=DETECT_SCALE)
    rgb_small = cv2.cvtColor(small, cv2.COLOR_BGR2RGB)

    locs_small = face_recognition.face_locations(rgb_small, model="hog")
    encodings = face_recognition.face_encodings(rgb_small, locs_small)

    # Scale bounding boxes back to original frame size
    inv = 1.0 / DETECT_SCALE
    locs_full = [
        (
            int(top * inv),
            int(right * inv),
            int(bottom * inv),
            int(left * inv),
        )
        for top, right, bottom, left in locs_small
    ]
    return locs_full, encodings


# ==============================
# Frame Encoding
# ==============================

def encode_frame_b64(frame_bgr: np.ndarray) -> str:
    """
    Resize frame to FRAME_WIDTH x FRAME_HEIGHT, encode as JPEG at FRAME_QUALITY,
    and return as a base64 string (no newlines).
    """
    resized = cv2.resize(frame_bgr, (FRAME_WIDTH, FRAME_HEIGHT))
    success, buf = cv2.imencode(
        ".jpg", resized, [cv2.IMWRITE_JPEG_QUALITY, FRAME_QUALITY]
    )
    if not success:
        return ""
    return base64.b64encode(buf.tobytes()).decode("ascii")


# ==============================
# Persistence
# ==============================

def save_encodings(encodings: List[np.ndarray]) -> None:
    """Persist face encodings to FACE_DATA_FILE using pickle."""
    with open(FACE_DATA_FILE, "wb") as f:
        pickle.dump(encodings, f)
    log(f"Saved {len(encodings)} face encodings to {FACE_DATA_FILE}.")


def load_encodings() -> List[np.ndarray]:
    """Load face encodings from FACE_DATA_FILE."""
    with open(FACE_DATA_FILE, "rb") as f:
        data = pickle.load(f)
    log(f"Loaded {len(data)} face encodings from {FACE_DATA_FILE}.")
    return data


# ==============================
# Training Mode
# ==============================

def run_training(sock: socket.socket, reader) -> None:
    """
    Training mode:
      1. Wait for START_TRAINING command from Go.
      2. Open webcam, capture CAPTURE_SAMPLES face samples.
      3. For every frame (face or not), send FRAME:<b64>.
      4. Each time a sample is saved, send PROGRESS:N/TOTAL.
      5. After all samples collected, save to face_data.pkl, send TRAINING_DONE.
      6. Wait for START_MONITOR before returning.
    """
    log("Waiting for START_TRAINING command...")
    for line in reader:
        cmd = line.strip()
        if cmd == "START_TRAINING":
            break
        log(f"Ignoring unexpected command while waiting for START_TRAINING: {cmd!r}")

    log("Training mode started.")
    cap = open_camera()
    known_encodings: List[np.ndarray] = []

    try:
        while len(known_encodings) < CAPTURE_SAMPLES:
            frame = read_frame(cap)
            if frame is None:
                log("WARNING: Failed to read frame during training — skipping.")
                time.sleep(0.05)
                continue

            locs, encodings = detect_faces_in_frame(frame)

            # Draw a green rectangle around the first detected face
            display_frame = frame.copy()
            for top, right, bottom, left in locs:
                cv2.rectangle(display_frame, (left, top), (right, bottom), (0, 255, 0), 2)

            # Always send the frame (with rectangle if face found)
            b64 = encode_frame_b64(display_frame)
            if b64:
                send_line(sock, f"FRAME:{b64}")

            # Save the first face encoding in this frame
            if encodings:
                known_encodings.append(encodings[0])
                log(f"Sample {len(known_encodings)}/{CAPTURE_SAMPLES} captured.")
                send_line(sock, f"PROGRESS:{len(known_encodings)}/{CAPTURE_SAMPLES}")
    finally:
        cap.release()

    save_encodings(known_encodings)
    send_line(sock, "TRAINING_DONE")
    log("Training complete. Waiting for START_MONITOR...")

    for line in reader:
        cmd = line.strip()
        if cmd == "START_MONITOR":
            break
        log(f"Ignoring unexpected command while waiting for START_MONITOR: {cmd!r}")


# ==============================
# Monitor Mode
# ==============================

def run_monitor(sock: socket.socket, reader, known_encodings: List[np.ndarray]) -> None:
    """
    Monitor mode:
      - Loop every MONITOR_INTERVAL seconds.
      - Detect faces; compare against known_encodings with TOLERANCE.
      - If no known face seen for GRACE_SECONDS, send FACE_LOST (once per absence).
      - When known face returns after a FACE_LOST event, send FACE_OK.
      - Does NOT send FRAME messages.
    """
    log("Waiting for START_MONITOR command...")
    for line in reader:
        cmd = line.strip()
        if cmd == "START_MONITOR":
            break
        log(f"Ignoring unexpected command while waiting for START_MONITOR: {cmd!r}")

    log("Monitor mode started.")
    cap = open_camera()
    last_seen: float = time.time()
    face_lost_sent: bool = False

    try:
        while True:
            frame = read_frame(cap)
            if frame is None:
                log("WARNING: Failed to read frame during monitoring.")
                time.sleep(MONITOR_INTERVAL)
                continue

            _, encodings = detect_faces_in_frame(frame)

            face_found = False
            for encoding in encodings:
                distances = face_recognition.face_distance(known_encodings, encoding)
                if len(distances) > 0 and np.min(distances) <= TOLERANCE:
                    face_found = True
                    break

            now = time.time()
            if face_found:
                last_seen = now
                if face_lost_sent:
                    # Known face returned after being lost
                    send_line(sock, "FACE_OK")
                    log("FACE_OK sent.")
                    face_lost_sent = False
            else:
                elapsed = now - last_seen
                if elapsed >= GRACE_SECONDS and not face_lost_sent:
                    send_line(sock, "FACE_LOST")
                    log("FACE_LOST sent.")
                    face_lost_sent = True

            time.sleep(MONITOR_INTERVAL)
    finally:
        cap.release()


# ==============================
# Main Entry
# ==============================

def main() -> None:
    log("face_guard.py starting up.")

    sock = connect_to_go()
    reader = sock.makefile("r", encoding="utf-8")

    has_face_data = os.path.isfile(FACE_DATA_FILE)

    if not has_face_data:
        # First-time setup: train then monitor
        log("No face data found — entering training mode.")
        run_training(sock, reader)
        known_encodings = load_encodings()
        run_monitor(sock, reader, known_encodings)
    else:
        # Subsequent launches: load existing encodings and monitor
        log("Face data found — entering monitor mode directly.")
        known_encodings = load_encodings()
        run_monitor(sock, reader, known_encodings)


if __name__ == "__main__":
    main()
