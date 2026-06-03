#!/usr/bin/env python3
"""
face_guard.py — PassQuantum Face Recognition Guard (MediaPipe edition)
=======================================================================
Runs as a child process launched by the Go app.
Communicates with Go via a persistent TCP connection on localhost:9876.

Internals now use MediaPipe Tasks FaceLandmarker + geometric encoding
instead of dlib / face_recognition.  The wire protocol to Go is unchanged.

Protocol (Python → Go):
  FRAME:<base64-encoded JPEG>       — live camera frame (training / demo modes)
  PROGRESS:<current>/<total>        — training progress
  TRAINING_DONE                     — all samples captured and saved
  FACE_OK                           — known face reappeared after FACE_LOST
  FACE_LOST                         — known face absent for GRACE_SECONDS

Protocol (Go → Python):
  START_TRAINING                    — begin capturing face samples
  START_MONITOR                     — enter continuous monitor loop
  START_DEMO                        — pause monitoring; stream annotated frames
                                      (478 landmarks + blink HUD) for the
                                      Security-settings visualizer
  STOP_DEMO                         — leave demo mode and resume monitoring
"""

import base64
import os
import socket
import sys
import threading
import time
from datetime import datetime
from typing import List, Optional

import cv2
import numpy as np

from geometric_encoder import Encoder
from liveness_detector import (
    LivenessDetector,
    _LEFT_EYE_IDX,
    _RIGHT_EYE_IDX,
    EAR_THRESHOLD,
)

# ==============================
# Constants
# ==============================

HOST = "127.0.0.1"
PORT = 9876
SIMILARITY_THRESHOLD = 0.92  # cosine similarity; replaces old L2 TOLERANCE
GRACE_SECONDS = 5.0
CAPTURE_SAMPLES = 100
CONNECT_RETRIES = 10
CONNECT_RETRY_DELAY = 0.5  # seconds
MONITOR_INTERVAL = 0.1  # seconds (100 ms)
FACE_DATA_FILE = "face_data.npy"  # numpy binary; replaces face_data.pkl
FRAME_QUALITY = 60  # JPEG quality for FRAME messages
FRAME_WIDTH = 320
FRAME_HEIGHT = 240
LIVENESS_BLINKS_REQUIRED = 1    # blinks needed to pass anti-spoofing check
LIVENESS_WINDOW_SECONDS  = 10.0 # seconds allowed per liveness gate attempt

# Detection visualizer (Security-settings demo): larger frame than the monitor
# preview so all 478 landmark dots stay legible, and a higher frame rate.
DEMO_FRAME_WIDTH  = 480
DEMO_FRAME_HEIGHT = 360
DEMO_INTERVAL     = 0.04        # seconds between demo frames (~25 fps)

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
            log(
                f"Connection attempt {attempt}/{CONNECT_RETRIES} failed — retrying in {CONNECT_RETRY_DELAY}s"
            )
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
# Cosine similarity
# ==============================


def _cosine_similarity(a: np.ndarray, b: np.ndarray) -> float:
    norm_a = np.linalg.norm(a)
    norm_b = np.linalg.norm(b)
    if norm_a < 1e-9 or norm_b < 1e-9:
        return 0.0
    return float(np.dot(a, b) / (norm_a * norm_b))


# ==============================
# Frame Encoding
# ==============================


def encode_frame_b64(
    frame_bgr: np.ndarray,
    width: int = FRAME_WIDTH,
    height: int = FRAME_HEIGHT,
) -> str:
    """
    Resize frame to width x height, encode as JPEG at FRAME_QUALITY,
    and return as a base64 string (no newlines).
    """
    resized = cv2.resize(frame_bgr, (width, height))
    success, buf = cv2.imencode(
        ".jpg", resized, [cv2.IMWRITE_JPEG_QUALITY, FRAME_QUALITY]
    )
    if not success:
        return ""
    return base64.b64encode(buf.tobytes()).decode("ascii")


# ==============================
# Persistence  (numpy instead of pickle)
# ==============================


def save_encodings(encodings: List[np.ndarray]) -> None:
    """Persist face encodings to FACE_DATA_FILE as a numpy array."""
    np.save(FACE_DATA_FILE, np.stack(encodings))
    log(f"Saved {len(encodings)} face encodings to {FACE_DATA_FILE}.")


def load_encodings() -> List[np.ndarray]:
    """Load face encodings from FACE_DATA_FILE."""
    arr = np.load(FACE_DATA_FILE)
    encodings = [arr[i] for i in range(len(arr))]
    log(f"Loaded {len(encodings)} face encodings from {FACE_DATA_FILE}.")
    return encodings


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
      5. Continue until BOTH all samples are captured AND liveness is confirmed
         (at least LIVENESS_BLINKS_REQUIRED blinks detected).
      6. Save to face_data.npy, send TRAINING_DONE.
      7. Wait for START_MONITOR before returning.
    """
    log("Waiting for START_TRAINING command...")
    for line in reader:
        cmd = line.strip()
        if cmd == "START_TRAINING":
            break
        log(f"Ignoring unexpected command while waiting for START_TRAINING: {cmd!r}")

    log("Training mode started.")
    cap = open_camera()
    encoder = Encoder()
    liveness = LivenessDetector()
    known_encodings: List[np.ndarray] = []
    _prev_blink_count = 0

    try:
        while (
            len(known_encodings) < CAPTURE_SAMPLES
            or liveness.blink_count < LIVENESS_BLINKS_REQUIRED
        ):
            frame = read_frame(cap)
            if frame is None:
                log("WARNING: Failed to read frame during training — skipping.")
                time.sleep(0.05)
                continue

            # Feed frame to liveness detector; log when a new blink is confirmed
            liveness.update(frame)
            if liveness.blink_count != _prev_blink_count:
                log(
                    f"Liveness: blink detected "
                    f"({liveness.blink_count}/{LIVENESS_BLINKS_REQUIRED})"
                )
                _prev_blink_count = liveness.blink_count

            vec  = encoder.encode(frame)
            bbox = encoder.bounding_box(frame)

            # Rectangle colour: green once liveness is confirmed, orange while pending
            display_frame = frame.copy()
            if bbox is not None:
                color = (
                    (0, 255, 0)
                    if liveness.blink_count >= LIVENESS_BLINKS_REQUIRED
                    else (0, 165, 255)
                )
                x1, y1, x2, y2 = bbox
                cv2.rectangle(display_frame, (x1, y1), (x2, y2), color, 2)

            # Always send the frame (with rectangle if face found)
            b64 = encode_frame_b64(display_frame)
            if b64:
                send_line(sock, f"FRAME:{b64}")

            # Save the encoding when a face is detected and we still need samples
            if vec is not None and len(known_encodings) < CAPTURE_SAMPLES:
                known_encodings.append(vec)
                log(f"Sample {len(known_encodings)}/{CAPTURE_SAMPLES} captured.")
                send_line(sock, f"PROGRESS:{len(known_encodings)}/{CAPTURE_SAMPLES}")
    finally:
        liveness.close()
        encoder.close()
        cap.release()

    log(
        f"Training liveness confirmed: {liveness.blink_count} blink(s) detected."
    )

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


def _liveness_gate(
    cap: cv2.VideoCapture,
    encoder: "Encoder",
    known_encodings: List[np.ndarray],
) -> None:
    """Block until a *recognized* face also passes the liveness check.

    A blink from the recognized face must be detected within
    LIVENESS_WINDOW_SECONDS.  If the window expires the gate resets and
    tries again, so the function never returns without confirmation.
    An unrecognized or absent face also resets the window.
    """
    log(
        f"Liveness gate: please blink within {LIVENESS_WINDOW_SECONDS}s "
        "to confirm you are live."
    )
    liveness = LivenessDetector()
    gate_start = time.time()
    try:
        while True:
            frame = read_frame(cap)
            if frame is None:
                time.sleep(0.05)
                continue

            vec = encoder.encode(frame)
            face_known = False
            if vec is not None:
                for known in known_encodings:
                    if _cosine_similarity(vec, known) >= SIMILARITY_THRESHOLD:
                        face_known = True
                        break

            if face_known:
                liveness.update(frame)
                if liveness.blink_count >= LIVENESS_BLINKS_REQUIRED:
                    log("Liveness gate passed — face confirmed as live.")
                    return
                # Timeout: recognized face present but no blink in time
                if time.time() - gate_start > LIVENESS_WINDOW_SECONDS:
                    log(
                        f"WARNING: Liveness gate timed out after "
                        f"{LIVENESS_WINDOW_SECONDS}s — restarting gate."
                    )
                    liveness.reset()
                    gate_start = time.time()
            else:
                # Unknown or absent face: reset the window
                if vec is not None:
                    log("Liveness gate: unrecognized face — resetting gate.")
                liveness.reset()
                gate_start = time.time()
    finally:
        liveness.close()


# ==============================
# Detection Visualizer (Demo Mode)
# ==============================


class _MonitorCommands:
    """Thread-safe flags set by the background command listener during monitoring.

    Once monitoring begins a single daemon thread owns the socket reader and
    translates incoming commands into these flags; the monitor loop polls them.
    """

    def __init__(self) -> None:
        self._lock = threading.Lock()
        self._demo_active = False

    def set_demo(self, active: bool) -> None:
        with self._lock:
            self._demo_active = active

    def demo_active(self) -> bool:
        with self._lock:
            return self._demo_active


def _command_listener(reader, commands: "_MonitorCommands") -> None:
    """Read commands from Go during monitor mode and update shared flags.

    Runs as a daemon thread.  This is the only socket reader once monitoring has
    started, so there is no contention with the monitor loop.  Exits when the
    connection closes (the ``for line in reader`` iterator ends).
    """
    for line in reader:
        cmd = line.strip()
        if cmd == "START_DEMO":
            commands.set_demo(True)
            log("Demo mode requested — pausing monitor.")
        elif cmd == "STOP_DEMO":
            commands.set_demo(False)
            log("Demo mode stopped — resuming monitor.")
        elif cmd in ("START_MONITOR", "START_TRAINING"):
            # Benign duplicates: the UI dispatches START_MONITOR from several
            # places (unlock, main screen) and we already entered monitor mode.
            pass
        elif cmd:
            log(f"Ignoring unexpected command during monitor: {cmd!r}")
    log("Command listener: connection closed.")


def run_demo(
    sock: socket.socket,
    cap: cv2.VideoCapture,
    commands: "_MonitorCommands",
) -> None:
    """Detection visualizer sub-loop for the Security-settings dialog.

    Streams annotated FRAME messages — all 478 MediaPipe landmarks drawn as
    dots (eye landmarks highlighted) plus a HUD showing blink count, live EAR,
    and eyes open/closed — until STOP_DEMO clears the demo flag.  Reuses the
    monitor's already-open camera handle so only one process owns the webcam.
    """
    log("Demo mode started.")
    detector = LivenessDetector()
    try:
        while commands.demo_active():
            frame = read_frame(cap)
            if frame is None:
                time.sleep(0.05)
                continue

            # Work at demo resolution so the dots stay legible after encoding.
            frame = cv2.resize(frame, (DEMO_FRAME_WIDTH, DEMO_FRAME_HEIGHT))
            avg_ear = detector.update(frame)
            lm = detector.last_landmarks
            h, w = frame.shape[:2]

            if lm is not None:
                # All 478 landmarks as small soft-yellow dots.
                for p in lm:
                    cv2.circle(frame, (int(p.x * w), int(p.y * h)), 1, (80, 220, 220), -1)
                # Eye landmarks used for the EAR blink test, highlighted larger.
                for idx in (*_LEFT_EYE_IDX, *_RIGHT_EYE_IDX):
                    p = lm[idx]
                    cv2.circle(frame, (int(p.x * w), int(p.y * h)), 3, (0, 255, 0), -1)

            # HUD: blink count, live EAR, and eyes open/closed.
            eyes_closed = avg_ear is not None and avg_ear < EAR_THRESHOLD
            ear_text = f"{avg_ear:.2f}" if avg_ear is not None else "--"
            eyes_text = "no face" if avg_ear is None else ("CLOSED" if eyes_closed else "OPEN")
            hud_color = (0, 165, 255) if eyes_closed else (0, 255, 0)
            cv2.putText(frame, f"Blinks: {detector.blink_count}", (10, 24),
                        cv2.FONT_HERSHEY_SIMPLEX, 0.6, (255, 255, 255), 2, cv2.LINE_AA)
            cv2.putText(frame, f"EAR: {ear_text}", (10, 48),
                        cv2.FONT_HERSHEY_SIMPLEX, 0.6, (255, 255, 255), 2, cv2.LINE_AA)
            cv2.putText(frame, f"Eyes: {eyes_text}", (10, 72),
                        cv2.FONT_HERSHEY_SIMPLEX, 0.6, hud_color, 2, cv2.LINE_AA)

            b64 = encode_frame_b64(frame, DEMO_FRAME_WIDTH, DEMO_FRAME_HEIGHT)
            if b64:
                send_line(sock, f"FRAME:{b64}")
            time.sleep(DEMO_INTERVAL)
    finally:
        detector.close()
    log("Demo mode ended.")


def run_monitor(sock: socket.socket, reader, known_encodings: List[np.ndarray]) -> None:
    """
    Monitor mode:
      - Run a liveness gate first (require a blink from the recognized face).
      - Then loop every MONITOR_INTERVAL seconds.
      - Detect faces; compare against known_encodings using cosine similarity.
      - If no known face seen for GRACE_SECONDS, send FACE_LOST (once per absence).
      - When known face returns after a FACE_LOST event, send FACE_OK.
      - Does NOT send FRAME messages (except while a demo session is active).
      - On START_DEMO, pause monitoring and stream annotated frames via
        run_demo() until STOP_DEMO; then resume with a fresh grace window.
    """
    log("Waiting for START_MONITOR command...")
    for line in reader:
        cmd = line.strip()
        if cmd == "START_MONITOR":
            break
        log(f"Ignoring unexpected command while waiting for START_MONITOR: {cmd!r}")

    log("Monitor mode started.")
    cap = open_camera()
    encoder = Encoder()

    # Anti-spoofing: confirm a live, recognized face before entering the main loop
    _liveness_gate(cap, encoder, known_encodings)

    # From here on, a background daemon thread is the sole socket reader: it turns
    # incoming START_DEMO / STOP_DEMO commands into flags the loop below polls.
    commands = _MonitorCommands()
    threading.Thread(
        target=_command_listener, args=(reader, commands), daemon=True
    ).start()

    last_seen: float = time.time()
    face_lost_sent: bool = False

    try:
        while True:
            # Detection visualizer requested — hand the camera to the demo loop,
            # then resume monitoring with a fresh grace window so we don't fire
            # FACE_LOST immediately after the dialog closes.
            if commands.demo_active():
                run_demo(sock, cap, commands)
                last_seen = time.time()
                face_lost_sent = False
                continue

            frame = read_frame(cap)
            if frame is None:
                log("WARNING: Failed to read frame during monitoring.")
                time.sleep(MONITOR_INTERVAL)
                continue

            vec = encoder.encode(frame)

            face_found = False
            if vec is not None:
                for known in known_encodings:
                    if _cosine_similarity(vec, known) >= SIMILARITY_THRESHOLD:
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
        encoder.close()
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
