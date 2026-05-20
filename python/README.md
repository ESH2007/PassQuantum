# python/ — Face Guard Sidecar

This directory contains the Python face-recognition sidecar that PassQuantum
launches as a child process. The Go binary communicates with it over a localhost
TCP socket using a simple line-oriented protocol.

## Architecture overview

```
Go binary  ──TCP (127.0.0.1:9876)──  face_guard.py
                                          │
                                          ├── geometric_encoder.py  (landmark encoding)
                                          ├── liveness_detector.py  (blink / anti-spoof)
                                          └── face_authenticator.py (enroll / verify API)
```

The Go binary starts `face_guard.py` as a subprocess. Once the Python process
connects back over TCP, the Go side sends commands (`START_TRAINING`,
`START_MONITOR`) and receives events (`FACE_OK`, `FACE_LOST`, `FRAME:<base64>`,
`PROGRESS:<n>/<total>`, `TRAINING_DONE`).

In production builds the entire Python layer is bundled into a single
self-contained executable (`face_guard_bundle` / `face_guard_bundle.exe`) using
PyInstaller, so no Python installation is required on the target machine.

## Files

| File | Description |
|---|---|
| `face_guard.py` | **Entry point.** Manages the webcam loop, face training, and continuous monitoring. Connects back to Go over TCP and sends protocol messages. Imports `geometric_encoder` and `liveness_detector`. |
| `geometric_encoder.py` | Encodes face landmarks produced by MediaPipe into a compact numeric representation used for identity matching. |
| `liveness_detector.py` | Implements an Eye Aspect Ratio (EAR) blink detector to distinguish live faces from photos/replays (anti-spoofing). |
| `face_authenticator.py` | High-level enroll/verify API that wraps `geometric_encoder` and `liveness_detector`. Used by the build pipeline for module resolution; not called directly from Go. |
| `requirements.txt` | Python dependencies: `mediapipe`, `opencv-python`, `numpy`. |

## Setup (development)

```bash
python3 -m venv .venv
.venv/bin/pip install -r python/requirements.txt
```

For PyInstaller bundling the build scripts handle dependency installation
automatically in an isolated environment.

## Protocol reference

Commands sent **from Go to Python:**

| Command | Effect |
|---|---|
| `START_TRAINING` | Begin capturing face samples |
| `START_MONITOR` | Begin continuous identity monitoring |

Messages sent **from Python to Go:**

| Message | Meaning |
|---|---|
| `FRAME:<base64 JPEG>` | Live camera frame (training UI only) |
| `PROGRESS:<n>/<total>` | Training sample progress |
| `TRAINING_DONE` | All face samples saved successfully |
| `FACE_OK` | Recognized face reappeared after a `FACE_LOST` event |
| `FACE_LOST` | Recognized face absent for the grace period |
