# Place ONNX model files here before building with biometric support.
#
# Required files:
#   blazeface.onnx   — BlazeFace short-range face detector (128×128 input)
#   face_mesh.onnx   — MediaPipe Face Mesh (468 landmarks, 192×192 input)
#
# Source: Convert or download from the MediaPipe/ONNX model repositories.
# CPU-only inference via gocv (OpenCV DNN module) — no GPU required.
