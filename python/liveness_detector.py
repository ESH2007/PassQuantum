#!/usr/bin/env python3
"""
liveness_detector.py — Eye Aspect Ratio blink detector for anti-spoofing
=========================================================================
Detects genuine blinks by computing the Eye Aspect Ratio (EAR) from
MediaPipe FaceLandmarker landmarks (Tasks API).  A blink is counted when
EAR drops below EAR_THRESHOLD for at least CONSEC_FRAMES consecutive frames
and then rises above it again.

Usage:
    from liveness_detector import LivenessDetector

    detector = LivenessDetector()
    while capturing:
        detector.update(frame_bgr)
    print(detector.blink_count)
    detector.close()
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

# ──────────────────────────────────────────────────────────────────────────────
# FaceLandmarker landmark indices for each eye (6-point EAR model)
# These indices are stable across MediaPipe face mesh variants.
# P0/P3 = horizontal corners, P1/P5 and P2/P4 = upper/lower vertical pairs
# ──────────────────────────────────────────────────────────────────────────────
_LEFT_EYE_IDX  = [362, 385, 387, 263, 373, 380]
_RIGHT_EYE_IDX = [33,  160, 158, 133, 153, 144]

EAR_THRESHOLD = 0.25   # below this value counts as "eye closed"
CONSEC_FRAMES = 2      # minimum consecutive closed frames per blink


def _ear(landmarks: list, indices: list, img_w: int, img_h: int) -> float:
    """Compute Eye Aspect Ratio for one eye."""
    pts = np.array(
        [[landmarks[i].x * img_w, landmarks[i].y * img_h] for i in indices],
        dtype=np.float64,
    )
    A = np.linalg.norm(pts[1] - pts[5])
    B = np.linalg.norm(pts[2] - pts[4])
    C = np.linalg.norm(pts[0] - pts[3])
    return (A + B) / (2.0 * C) if C > 1e-9 else 0.0 # type: ignore


class LivenessDetector:
    """
    Stateful per-session blink counter using the Tasks API in VIDEO mode,
    which maintains tracking state across frames.

    Create one instance per enrollment/verification session.
    Call ``update(frame_bgr)`` for each captured frame (pass monotonically
    increasing timestamps or leave at default).
    Read ``blink_count`` to check how many blinks have been detected.
    Call ``close()`` when done.
    """

    def __init__(
        self,
        model_path: str  = _MODEL_PATH,
        ear_threshold: float = EAR_THRESHOLD,
        consec_frames: int   = CONSEC_FRAMES,
    ) -> None:
        self.ear_threshold = ear_threshold
        self.consec_frames = consec_frames
        self.blink_count: int = 0
        self._below_count: int = 0
        self._timestamp_ms: int = 0

        # Last detected face landmarks (NormalizedLandmark list) and EAR, exposed
        # so callers such as the Security-settings visualizer can draw the points
        # without running a second FaceLandmarker. None when no face was found.
        self.last_landmarks = None
        self.last_ear: Optional[float] = None

        options = _mp_vision.FaceLandmarkerOptions(
            base_options=_mp_python.BaseOptions(model_asset_path=model_path),
            running_mode=_mp_vision.RunningMode.VIDEO,
            num_faces=1,
            min_face_detection_confidence=0.5,
            min_tracking_confidence=0.5,
        )
        self._landmarker = _mp_vision.FaceLandmarker.create_from_options(options)

    def update(self, frame_bgr: np.ndarray) -> Optional[float]:
        """
        Process one BGR frame.  Returns the average EAR or None if no face.
        Updates ``blink_count`` in place.
        """
        h, w = frame_bgr.shape[:2]
        rgb = cv2.cvtColor(frame_bgr, cv2.COLOR_BGR2RGB)
        mp_image = mp.Image(image_format=mp.ImageFormat.SRGB, data=rgb)

        self._timestamp_ms += 33   # ~30 fps synthetic clock
        result = self._landmarker.detect_for_video(mp_image, self._timestamp_ms)

        if not result.face_landmarks:
            self.last_landmarks = None
            self.last_ear = None
            return None

        lm = result.face_landmarks[0]
        left_ear  = _ear(lm, _LEFT_EYE_IDX,  w, h)
        right_ear = _ear(lm, _RIGHT_EYE_IDX, w, h)
        avg_ear   = (left_ear + right_ear) / 2.0

        self.last_landmarks = lm
        self.last_ear = avg_ear

        if avg_ear < self.ear_threshold:
            self._below_count += 1
        else:
            if self._below_count >= self.consec_frames:
                self.blink_count += 1
            self._below_count = 0

        return avg_ear

    def reset(self) -> None:
        """Reset the blink counter and EAR state without recreating the landmarker.

        _timestamp_ms is intentionally NOT reset — MediaPipe VIDEO mode requires
        strictly monotonically increasing timestamps across the lifetime of the
        landmarker instance.
        """
        self.blink_count = 0
        self._below_count = 0

    def close(self) -> None:
        """Release the underlying FaceLandmarker."""
        self._landmarker.close()
