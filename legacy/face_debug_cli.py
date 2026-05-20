#!/usr/bin/env python3

# ==============================
# Face Recognition Debug CLI
# ==============================

import argparse
import os
import pickle
import time
from datetime import datetime
from typing import List, Sequence, Tuple

import cv2
import face_recognition
import numpy as np


# ==============================
# Constants
# ==============================

DATA_FILE = "face_data.pkl"
DEFAULT_TOLERANCE = 0.45
DEFAULT_TRAIN_SAMPLES = 50
RESIZE_SCALE = 0.35
MONITOR_INTERVAL_SECONDS = 0.5
FACE_LOST_SIGNAL_SECONDS = 3.0
READ_FAILURE_REOPEN_THRESHOLD = 8
READ_FAILURE_LOG_INTERVAL = 20


# ==============================
# Logging Helpers
# ==============================

def log(message: str) -> None:
    ts = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    print(f"[{ts}] {message}", flush=True)


# ==============================
# CLI Parsing
# ==============================

def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Standalone terminal CLI to train and monitor face recognition behavior."
    )
    parser.add_argument(
        "--reset",
        action="store_true",
        help="Delete face_data.pkl and retrain before monitoring.",
    )
    parser.add_argument(
        "--tolerance",
        type=float,
        default=DEFAULT_TOLERANCE,
        help="Face match tolerance override (default: 0.45). Lower is stricter.",
    )
    parser.add_argument(
        "--samples",
        type=int,
        default=DEFAULT_TRAIN_SAMPLES,
        help="Number of training samples to capture (default: 20).",
    )
    parser.add_argument(
        "--show-dist",
        action="store_true",
        help="Print the face distance for every detected face each monitor cycle.",
    )
    parser.add_argument(
        "--no-preview",
        action="store_true",
        help="Run without cv2.imshow windows (headless debug mode).",
    )
    parser.add_argument(
        "--camera-index",
        type=int,
        default=0,
        help="Camera index to open (default: 0). Try 1/2 if 0 is busy.",
    )
    parser.add_argument(
        "--camera-backend",
        choices=["auto", "dshow", "msmf", "any"],
        default="auto",
        help=(
            "Camera backend to use. auto tries DSHOW then MSMF then ANY on Windows; "
            "default: auto."
        ),
    )
    args = parser.parse_args()

    if args.tolerance <= 0:
        parser.error("--tolerance must be > 0")
    if args.samples <= 0:
        parser.error("--samples must be > 0")
    if args.camera_index < 0:
        parser.error("--camera-index must be >= 0")

    return args


# ==============================
# Data Persistence
# ==============================

def save_encodings(file_path: str, encodings: Sequence[np.ndarray]) -> None:
    payload = {
        "encodings": [np.asarray(enc, dtype=np.float64).tolist() for enc in encodings],
        "count": len(encodings),
        "saved_at": datetime.now().isoformat(timespec="seconds"),
    }
    with open(file_path, "wb") as f:
        pickle.dump(payload, f)


def load_encodings(file_path: str) -> List[np.ndarray]:
    with open(file_path, "rb") as f:
        payload = pickle.load(f)

    if isinstance(payload, dict) and "encodings" in payload:
        raw = payload["encodings"]
    elif isinstance(payload, list):
        raw = payload
    else:
        raise ValueError("Unsupported face_data.pkl format")

    encodings: List[np.ndarray] = []
    for row in raw:
        arr = np.asarray(row, dtype=np.float64)
        if arr.shape == (128,):
            encodings.append(arr)
    if not encodings:
        raise ValueError("No valid face encodings found in face_data.pkl")
    return encodings


# ==============================
# Camera + Face Detection
# ==============================

def _camera_backends_for_selection(selection: str) -> List[Tuple[str, int]]:
    if selection == "any":
        return [("any", cv2.CAP_ANY)]
    if selection == "dshow":
        return [("dshow", cv2.CAP_DSHOW)]
    if selection == "msmf":
        return [("msmf", cv2.CAP_MSMF)]

    if os.name == "nt":
        return [
            ("dshow", cv2.CAP_DSHOW),
            ("msmf", cv2.CAP_MSMF),
            ("any", cv2.CAP_ANY),
        ]
    return [("any", cv2.CAP_ANY)]


def open_camera(camera_index: int, backend_selection: str) -> Tuple[cv2.VideoCapture, str]:
    backends = _camera_backends_for_selection(backend_selection)
    errors: List[str] = []

    for backend_name, backend_id in backends:
        cap = cv2.VideoCapture(camera_index, backend_id)
        if not cap.isOpened():
            errors.append(f"{backend_name}: open failed")
            cap.release()
            continue

        cap.set(cv2.CAP_PROP_BUFFERSIZE, 1)

        # Some Windows camera stacks need a handful of grabs before frames stabilize.
        warmup_ok = False
        for _ in range(5):
            ok, _ = cap.read()
            if ok:
                warmup_ok = True
                break

        if warmup_ok:
            log(f"[CAMERA] Opened index={camera_index} backend={backend_name}")
            return cap, backend_name

        errors.append(f"{backend_name}: opened but failed warmup reads")
        cap.release()

    raise RuntimeError(
        "Unable to open webcam. "
        f"index={camera_index}, backend_selection={backend_selection}, attempts={'; '.join(errors)}"
    )


def reopen_camera(args: argparse.Namespace, last_backend: str) -> Tuple[cv2.VideoCapture, str]:
    preferred = args.camera_backend if args.camera_backend != "auto" else last_backend
    try:
        return open_camera(args.camera_index, preferred)
    except RuntimeError:
        return open_camera(args.camera_index, args.camera_backend)


def detect_faces(frame_bgr: np.ndarray) -> Tuple[List[Tuple[int, int, int, int]], List[np.ndarray]]:
    small = cv2.resize(frame_bgr, (0, 0), fx=RESIZE_SCALE, fy=RESIZE_SCALE)
    rgb_small = cv2.cvtColor(small, cv2.COLOR_BGR2RGB)

    small_locations = face_recognition.face_locations(rgb_small, model="hog")
    encodings = face_recognition.face_encodings(rgb_small, small_locations)

    scale_up = int(round(1 / RESIZE_SCALE))
    full_locations = [
        (top * scale_up, right * scale_up, bottom * scale_up, left * scale_up)
        for (top, right, bottom, left) in small_locations
    ]
    return full_locations, encodings


def largest_face_index(locations: Sequence[Tuple[int, int, int, int]]) -> int:
    if not locations:
        return -1
    largest_idx = 0
    largest_area = -1
    for i, (top, right, bottom, left) in enumerate(locations):
        area = max(0, bottom - top) * max(0, right - left)
        if area > largest_area:
            largest_area = area
            largest_idx = i
    return largest_idx


def draw_boxes(
    frame: np.ndarray,
    locations: Sequence[Tuple[int, int, int, int]],
    colors: Sequence[Tuple[int, int, int]],
) -> None:
    for (top, right, bottom, left), color in zip(locations, colors):
        cv2.rectangle(frame, (left, top), (right, bottom), color, 2)


# ==============================
# Training Mode
# ==============================

def run_training_mode(args: argparse.Namespace) -> bool:
    log("[MODE] TRAINING MODE")
    log(f"[TRAIN] Target samples: {args.samples}")
    log("[TRAIN] Press Q to quit early.")

    collected: List[np.ndarray] = []
    start = time.perf_counter()
    cap, backend_name = open_camera(args.camera_index, args.camera_backend)
    consecutive_read_failures = 0

    try:
        while len(collected) < args.samples:
            ok, frame = cap.read()
            if not ok:
                consecutive_read_failures += 1
                if consecutive_read_failures % READ_FAILURE_LOG_INTERVAL == 1:
                    log(
                        "[TRAIN] Failed to read frame from webcam "
                        f"({consecutive_read_failures} consecutive failures)."
                    )
                if consecutive_read_failures >= READ_FAILURE_REOPEN_THRESHOLD:
                    log("[TRAIN] Reopening camera after repeated frame read failures.")
                    cap.release()
                    cap, backend_name = reopen_camera(args, backend_name)
                    consecutive_read_failures = 0
                continue
            consecutive_read_failures = 0

            locations, encodings = detect_faces(frame)
            colors = [(0, 255, 0)] * len(locations)
            draw_boxes(frame, locations, colors)

            if encodings:
                idx = largest_face_index(locations)
                if idx >= 0:
                    collected.append(encodings[idx])
                    log(f"[TRAIN] Sample {len(collected)}/{args.samples}")

            if not args.no_preview:
                cv2.putText(
                    frame,
                    f"TRAINING {len(collected)}/{args.samples}",
                    (10, 30),
                    cv2.FONT_HERSHEY_SIMPLEX,
                    0.8,
                    (0, 255, 0),
                    2,
                )
                cv2.imshow("Face Debug - Training", frame)
                if cv2.waitKey(1) & 0xFF == ord("q"):
                    log("[TRAIN] Training interrupted by user (Q).")
                    break
    except KeyboardInterrupt:
        log("[TRAIN] Training interrupted by user (Ctrl+C).")
    finally:
        cap.release()
        if not args.no_preview:
            cv2.destroyAllWindows()

    if not collected:
        log("[TRAIN] No face samples captured. Nothing was saved.")
        return False

    save_encodings(DATA_FILE, collected)
    elapsed = time.perf_counter() - start
    log(f"[TRAIN] Saved {len(collected)} samples to {DATA_FILE}")
    log(f"[TRAIN] Total time taken: {elapsed:.2f}s")
    return len(collected) >= args.samples


# ==============================
# Monitor Mode
# ==============================

def run_monitor_mode(args: argparse.Namespace, known_encodings: Sequence[np.ndarray]) -> None:
    log("[MODE] MONITOR MODE")
    log(f"[MONITOR] Loaded known encodings: {len(known_encodings)}")
    if args.no_preview:
        log("[MONITOR] Preview disabled (--no-preview). Use Ctrl+C to quit.")
    else:
        log("[MONITOR] Press Q to quit.")

    cap, backend_name = open_camera(args.camera_index, args.camera_backend)
    consecutive_read_failures = 0

    last_check = 0.0
    last_known_seen = time.monotonic()
    lost_signal_sent = False

    latest_locations: List[Tuple[int, int, int, int]] = []
    latest_colors: List[Tuple[int, int, int]] = []
    latest_overlay = "FACE NOT RECOGNIZED"
    latest_overlay_color = (0, 0, 255)

    try:
        while True:
            ok, frame = cap.read()
            if not ok:
                consecutive_read_failures += 1
                if consecutive_read_failures % READ_FAILURE_LOG_INTERVAL == 1:
                    log(
                        "[MONITOR] Failed to read frame from webcam "
                        f"({consecutive_read_failures} consecutive failures)."
                    )
                if consecutive_read_failures >= READ_FAILURE_REOPEN_THRESHOLD:
                    log("[MONITOR] Reopening camera after repeated frame read failures.")
                    cap.release()
                    cap, backend_name = reopen_camera(args, backend_name)
                    consecutive_read_failures = 0
                continue
            consecutive_read_failures = 0

            now = time.monotonic()
            if now - last_check >= MONITOR_INTERVAL_SECONDS:
                locations, encodings = detect_faces(frame)

                known_present = False
                min_distance_global = float("inf")
                colors: List[Tuple[int, int, int]] = []

                for idx, enc in enumerate(encodings):
                    distances = face_recognition.face_distance(known_encodings, enc)
                    min_distance = float(np.min(distances)) if len(distances) else float("inf")
                    min_distance_global = min(min_distance_global, min_distance)
                    is_match = min_distance <= args.tolerance
                    colors.append((0, 255, 0) if is_match else (0, 0, 255))
                    known_present = known_present or is_match

                    if args.show_dist:
                        if np.isfinite(min_distance):
                            log(f"[MONITOR] FACE_DIST face={idx} distance={min_distance:.3f}")
                        else:
                            log(f"[MONITOR] FACE_DIST face={idx} distance=inf")

                if known_present:
                    if np.isfinite(min_distance_global):
                        log(f"[MONITOR] FACE_OK — min_distance={min_distance_global:.3f}")
                    else:
                        log("[MONITOR] FACE_OK — min_distance=inf")
                    latest_overlay = "AUTHORIZED"
                    latest_overlay_color = (0, 255, 0)
                    last_known_seen = now
                    lost_signal_sent = False
                else:
                    if np.isfinite(min_distance_global):
                        log(f"[MONITOR] FACE_LOST — no match (min_distance={min_distance_global:.3f})")
                    else:
                        log("[MONITOR] FACE_LOST — no face detected (min_distance=inf)")
                    latest_overlay = "FACE NOT RECOGNIZED"
                    latest_overlay_color = (0, 0, 255)

                if now - last_known_seen >= FACE_LOST_SIGNAL_SECONDS and not lost_signal_sent:
                    log("[SIGNAL] FACE_LOST — would lock app")
                    lost_signal_sent = True

                latest_locations = locations
                if len(colors) < len(locations):
                    colors.extend([(0, 0, 255)] * (len(locations) - len(colors)))
                latest_colors = colors[: len(locations)]
                last_check = now

            draw_boxes(frame, latest_locations, latest_colors)
            cv2.putText(
                frame,
                latest_overlay,
                (10, 30),
                cv2.FONT_HERSHEY_SIMPLEX,
                0.8,
                latest_overlay_color,
                2,
            )

            if not args.no_preview:
                cv2.imshow("Face Debug - Monitor", frame)
                if cv2.waitKey(1) & 0xFF == ord("q"):
                    break
    except KeyboardInterrupt:
        log("[MONITOR] Monitor interrupted by user (Ctrl+C).")
    finally:
        cap.release()
        if not args.no_preview:
            cv2.destroyAllWindows()


# ==============================
# Main Entry
# ==============================

def main() -> None:
    args = parse_args()

    if args.reset and os.path.exists(DATA_FILE):
        os.remove(DATA_FILE)
        log(f"[INIT] --reset active. Deleted {DATA_FILE}")

    if not os.path.exists(DATA_FILE):
        completed = run_training_mode(args)
        if not completed:
            log("[INIT] Training ended before reaching target samples.")
        log("[INIT] Training phase finished. Run again to start monitor mode.")
        return

    try:
        known_encodings = load_encodings(DATA_FILE)
    except Exception as exc:
        log(f"[ERROR] Failed to load {DATA_FILE}: {exc}")
        log("[HINT] Run again with --reset to retrain.")
        return

    run_monitor_mode(args, known_encodings)


if __name__ == "__main__":
    main()
