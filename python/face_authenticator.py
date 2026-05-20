#!/usr/bin/env python3
"""
face_authenticator.py — High-level enroll / verify API
=======================================================
Wraps geometric_encoder.Encoder and LivenessDetector into two operations:

    enroll(camera_index) → np.ndarray          (the face template)
    verify(camera_index, template, timeout_s)  → AuthResult

Templates are raw float64 numpy arrays serialised as:
    base64.b64encode(template.tobytes())                          # → bytes
    np.frombuffer(base64.b64decode(b64), dtype=np.float64)        # ←

This module is intentionally free of any TCP/socket logic.
"""

import time
from dataclasses import dataclass
from enum import Enum, auto
from typing import Optional

import cv2
import numpy as np

from geometric_encoder import Encoder
from liveness_detector import LivenessDetector

# ──────────────────────────────────────────────────────────────────────────────
SIMILARITY_THRESHOLD = 0.92
REQUIRED_BLINKS      = 1
ENROLL_SAMPLES       = 20


# ──────────────────────────────────────────────────────────────────────────────
class AuthStatus(Enum):
    OK             = auto()
    FACE_NOT_FOUND = auto()
    NO_LIVENESS    = auto()
    MISMATCH       = auto()
    TIMEOUT        = auto()
    CAMERA_ERROR   = auto()


@dataclass
class AuthResult:
    status:   AuthStatus
    distance: float = 0.0
    message:  str   = ""

    @property
    def success(self) -> bool:
        return self.status == AuthStatus.OK


# ──────────────────────────────────────────────────────────────────────────────
def _cosine_similarity(a: np.ndarray, b: np.ndarray) -> float:
    norm_a = np.linalg.norm(a)
    norm_b = np.linalg.norm(b)
    if norm_a < 1e-9 or norm_b < 1e-9:
        return 0.0
    return float(np.dot(a, b) / (norm_a * norm_b))


# ──────────────────────────────────────────────────────────────────────────────
class FaceAuthenticator:
    """
    Stateless authenticator.  Each call creates its own camera and encoder
    instances and releases them on return.
    """

    def enroll(
        self,
        camera_index:    int = 0,
        required_blinks: int = REQUIRED_BLINKS,
        target_samples:  int = ENROLL_SAMPLES,
    ) -> np.ndarray:
        """
        Capture face encodings from the webcam.

        Raises RuntimeError if the camera fails, no face is detected,
        or liveness check fails.
        Returns the mean encoding as a float64 ndarray.
        """
        cap = cv2.VideoCapture(camera_index)
        if not cap.isOpened():
            raise RuntimeError(f"Cannot open camera {camera_index}")

        encoder  = Encoder()
        detector = LivenessDetector()
        encodings: list[np.ndarray] = []

        try:
            while len(encodings) < target_samples:
                ok, frame = cap.read()
                if not ok or frame is None:
                    continue
                detector.update(frame)
                vec = encoder.encode(frame)
                if vec is not None:
                    encodings.append(vec)
        finally:
            encoder.close()
            detector.close()
            cap.release()

        if not encodings:
            raise RuntimeError("No face detected during enrollment")
        if detector.blink_count < required_blinks:
            raise RuntimeError(
                f"Liveness check failed: {detector.blink_count}/{required_blinks} blinks"
            )

        return np.mean(np.stack(encodings), axis=0).astype(np.float64)

    def verify(
        self,
        camera_index: int        = 0,
        template:     np.ndarray = None,
        timeout_s:    float      = 10.0,
    ) -> AuthResult:
        """
        Continuously grab frames until a matching face is seen or timeout.
        """
        if template is None:
            return AuthResult(AuthStatus.FACE_NOT_FOUND, message="No template provided")

        cap = cv2.VideoCapture(camera_index)
        if not cap.isOpened():
            return AuthResult(
                AuthStatus.CAMERA_ERROR,
                message=f"Cannot open camera {camera_index}",
            )

        encoder  = Encoder()
        deadline = time.time() + timeout_s
        best_sim = 0.0

        try:
            while time.time() < deadline:
                ok, frame = cap.read()
                if not ok or frame is None:
                    continue
                vec = encoder.encode(frame)
                if vec is None:
                    continue
                sim = _cosine_similarity(vec, template)
                if sim > best_sim:
                    best_sim = sim
                if sim >= SIMILARITY_THRESHOLD:
                    return AuthResult(
                        AuthStatus.OK,
                        distance=1.0 - sim,
                        message="Face matched",
                    )
        finally:
            encoder.close()
            cap.release()

        if best_sim == 0.0:
            return AuthResult(
                AuthStatus.FACE_NOT_FOUND,
                distance=1.0,
                message="No face detected",
            )
        return AuthResult(
            AuthStatus.MISMATCH,
            distance=1.0 - best_sim,
            message=f"Best similarity {best_sim:.3f} < {SIMILARITY_THRESHOLD}",
        )

# ──────────────────────────────────────────────────────────────────────────────
# Tuning knobs
# ──────────────────────────────────────────────────────────────────────────────
SIMILARITY_THRESHOLD = 0.92   # cosine similarity; tune up (stricter) if needed
REQUIRED_BLINKS      = 1      # minimum blinks required for liveness during enroll
ENROLL_SAMPLES       = 20     # number of valid face frames to average


# ──────────────────────────────────────────────────────────────────────────────
# Result types
# ──────────────────────────────────────────────────────────────────────────────
class AuthStatus(Enum):
    OK             = auto()
    FACE_NOT_FOUND = auto()
    NO_LIVENESS    = auto()
    MISMATCH       = auto()
    TIMEOUT        = auto()
    CAMERA_ERROR   = auto()


@dataclass
class AuthResult:
    status:   AuthStatus
    distance: float = 0.0      # 1 − cosine_similarity; 0 = perfect match
    message:  str   = ""

    @property
    def success(self) -> bool:
        return self.status == AuthStatus.OK


# ──────────────────────────────────────────────────────────────────────────────
# Internal helpers
# ──────────────────────────────────────────────────────────────────────────────
def _cosine_similarity(a: np.ndarray, b: np.ndarray) -> float:
    norm_a = np.linalg.norm(a)
    norm_b = np.linalg.norm(b)
    if norm_a < 1e-9 or norm_b < 1e-9:
        return 0.0
    return float(np.dot(a, b) / (norm_a * norm_b))


def _open_camera(camera_index: int) -> Optional[cv2.VideoCapture]:
    cap = cv2.VideoCapture(camera_index)
    return cap if cap.isOpened() else None


# ──────────────────────────────────────────────────────────────────────────────
# FaceAuthenticator
# ──────────────────────────────────────────────────────────────────────────────
class FaceAuthenticator:
    """
    Stateless authenticator — create one instance and call enroll/verify as
    many times as needed.  Thread-safety: each call opens/closes its own camera
    and MediaPipe session; do not share a single call across threads.
    """

    def enroll(
        self,
        camera_index:    int = 0,
        required_blinks: int = REQUIRED_BLINKS,
        target_samples:  int = ENROLL_SAMPLES,
    ) -> np.ndarray:
        """
        Capture `target_samples` face encodings from the webcam.

        Requires at least `required_blinks` blinks to pass the liveness check.
        Returns the mean encoding as a float64 ndarray.

        Raises
        ------
        RuntimeError
            If the camera cannot be opened, no face is detected, or the
            liveness check fails.
        """
        cap = _open_camera(camera_index)
        if cap is None:
            raise RuntimeError(f"Cannot open camera {camera_index}")

        detector = LivenessDetector()
        mesh     = mp.solutions.face_mesh.FaceMesh(
            static_image_mode=False,
            max_num_faces=1,
            refine_landmarks=False,
            min_detection_confidence=0.5,
            min_tracking_confidence=0.5,
        )
        encodings: list[np.ndarray] = []

        try:
            while len(encodings) < target_samples:
                ok, frame = cap.read()
                if not ok or frame is None:
                    continue
                detector.update(frame)
                enc = extract_encoding(frame, mesh=mesh)
                if enc is not None:
                    encodings.append(enc)
        finally:
            mesh.close()
            detector.close()
            cap.release()

        if not encodings:
            raise RuntimeError("No face detected during enrollment")

        if detector.blink_count < required_blinks:
            raise RuntimeError(
                f"Liveness check failed: {detector.blink_count}/{required_blinks} blinks detected"
            )

        return np.mean(np.stack(encodings), axis=0).astype(np.float64)

    def verify(
        self,
        camera_index: int          = 0,
        template:     np.ndarray   = None,
        timeout_s:    float        = 10.0,
    ) -> AuthResult:
        """
        Continuously grab frames until a matching face is seen or timeout.

        Parameters
        ----------
        camera_index : int
            Webcam index (0 = default).
        template : np.ndarray
            The enrolled face template (output of enroll()).
        timeout_s : float
            How long to try before giving up.

        Returns
        -------
        AuthResult with .success == True on a match.
        """
        if template is None:
            return AuthResult(AuthStatus.FACE_NOT_FOUND, message="No template provided")

        cap = _open_camera(camera_index)
        if cap is None:
            return AuthResult(
                AuthStatus.CAMERA_ERROR,
                message=f"Cannot open camera {camera_index}",
            )

        mesh     = mp.solutions.face_mesh.FaceMesh(
            static_image_mode=False,
            max_num_faces=1,
            refine_landmarks=False,
            min_detection_confidence=0.5,
            min_tracking_confidence=0.5,
        )
        deadline = time.time() + timeout_s
        best_sim = 0.0

        try:
            while time.time() < deadline:
                ok, frame = cap.read()
                if not ok or frame is None:
                    continue
                enc = extract_encoding(frame, mesh=mesh)
                if enc is None:
                    continue
                sim = _cosine_similarity(enc, template)
                if sim > best_sim:
                    best_sim = sim
                if sim >= SIMILARITY_THRESHOLD:
                    return AuthResult(
                        AuthStatus.OK,
                        distance=1.0 - sim,
                        message="Face matched",
                    )
        finally:
            mesh.close()
            cap.release()

        if best_sim == 0.0:
            return AuthResult(
                AuthStatus.FACE_NOT_FOUND,
                distance=1.0,
                message="No face detected during verification window",
            )
        return AuthResult(
            AuthStatus.MISMATCH,
            distance=1.0 - best_sim,
            message=f"Best similarity {best_sim:.3f} < threshold {SIMILARITY_THRESHOLD}",
        )
