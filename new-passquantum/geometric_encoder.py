#!/usr/bin/env python3
"""
geometric_encoder.py — MediaPipe Face Landmarker geometric feature extractor
=============================================================================
Extracts a translation- and scale-invariant float64 feature vector from a
single BGR frame using the MediaPipe Tasks FaceLandmarker (478 landmarks ×
3 coordinates = 1434-dimensional vector when using default model).

Usage:
    from geometric_encoder import Encoder
    enc = Encoder()           # creates a single-use IMAGE-mode landmarker
    vec = enc.encode(frame)   # np.ndarray shape (1434,) or None
    enc.close()
"""

import os
import sys
import cv2
import mediapipe as mp
import numpy as np
from typing import Optional

from mediapipe.tasks import python as _mp_python
from mediapipe.tasks.python import vision as _mp_vision


def _base_dir() -> str:
    """Return the directory that contains the models/ folder.

    * Inside a PyInstaller --onefile bundle, data files land in sys._MEIPASS.
    * Otherwise use the directory of this source file.
    """
    if getattr(sys, "frozen", False) and hasattr(sys, "_MEIPASS"):
        return sys._MEIPASS  # type: ignore[attr-defined]
    return os.path.dirname(os.path.abspath(__file__))


_MODEL_PATH = os.path.join(_base_dir(), "models", "face_landmarker.task")


class Encoder:
    """
    Wraps a MediaPipe FaceLandmarker for single-frame (IMAGE mode) encoding.
    Re-use one instance per session to avoid repeated model-load overhead.
    """

    def __init__(self, model_path: str = _MODEL_PATH) -> None:
        options = _mp_vision.FaceLandmarkerOptions(
            base_options=_mp_python.BaseOptions(model_asset_path=model_path),
            running_mode=_mp_vision.RunningMode.IMAGE,
            num_faces=1,
            min_face_detection_confidence=0.5,
        )
        self._landmarker = _mp_vision.FaceLandmarker.create_from_options(options)

    def encode(self, frame_bgr: np.ndarray) -> Optional[np.ndarray]:
        """
        Extract a normalized geometric feature vector from a BGR frame.

        Returns a float64 ndarray of shape (N*3,) or None if no face is detected.
        """
        h, w = frame_bgr.shape[:2]
        rgb = cv2.cvtColor(frame_bgr, cv2.COLOR_BGR2RGB)
        mp_image = mp.Image(image_format=mp.ImageFormat.SRGB, data=rgb)
        result = self._landmarker.detect(mp_image)

        if not result.face_landmarks:
            return None

        lm = result.face_landmarks[0]  # list of NormalizedLandmark
        pts = np.array(
            [[p.x * w, p.y * h, p.z * w] for p in lm], dtype=np.float64
        )

        # Normalise: subtract centroid, divide by RMS distance
        centroid = pts.mean(axis=0)
        pts -= centroid
        scale = np.sqrt((pts ** 2).sum(axis=1).mean())
        if scale < 1e-9:
            return None
        pts /= scale

        return pts.flatten()

    def bounding_box(self, frame_bgr: np.ndarray) -> Optional[tuple]:
        """
        Return (x1, y1, x2, y2) bounding box of the first detected face,
        or None if no face is found.
        """
        h, w = frame_bgr.shape[:2]
        rgb = cv2.cvtColor(frame_bgr, cv2.COLOR_BGR2RGB)
        mp_image = mp.Image(image_format=mp.ImageFormat.SRGB, data=rgb)
        result = self._landmarker.detect(mp_image)

        if not result.face_landmarks:
            return None

        lm = result.face_landmarks[0]
        xs = [p.x * w for p in lm]
        ys = [p.y * h for p in lm]
        return (
            max(0, int(min(xs))),
            max(0, int(min(ys))),
            min(w, int(max(xs))),
            min(h, int(max(ys))),
        )

    def close(self) -> None:
        self._landmarker.close()


# ── Convenience top-level function (creates a throw-away encoder) ─────────────
def extract_encoding(frame_bgr: np.ndarray) -> Optional[np.ndarray]:
    """
    One-shot helper: create a temporary Encoder, encode one frame, close it.
    Use Encoder() directly in tight loops to avoid repeated model loads.
    """
    enc = Encoder()
    try:
        return enc.encode(frame_bgr)
    finally:
        enc.close()


def bounding_box_from_landmarks(frame_bgr: np.ndarray) -> Optional[tuple]:
    """
    One-shot helper: return the face bounding box or None.
    """
    enc = Encoder()
    try:
        return enc.bounding_box(frame_bgr)
    finally:
        enc.close()
