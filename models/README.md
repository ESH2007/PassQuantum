# models/

Model assets required by the Python face-guard sidecar at runtime.

| File | Description |
|---|---|
| `face_landmarker.task` | MediaPipe Face Landmarker task bundle. Required for facial landmark detection used by `python/geometric_encoder.py` and `python/liveness_detector.py`. Downloaded from the [MediaPipe model repository](https://developers.google.com/mediapipe/solutions/vision/face_landmarker). |
| `README.txt` | Legacy note about earlier ONNX model files (superseded by the MediaPipe task-based approach). |

## Notes

- `face_landmarker.task` is included via `--add-data` in all PyInstaller builds so it is available inside the bundled executable at runtime.
- The file is tracked with Git LFS (see `.gitattributes`).
