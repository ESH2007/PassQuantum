#!/usr/bin/env python3
"""
auth_server.py — PassQuantum Face Auth IPC Server (JSON over Unix socket)
==========================================================================
Alternative IPC bridge for callers that prefer JSON over the text-line
protocol used by face_guard.py.  Listens on a Unix domain socket and handles
newline-delimited JSON requests.

Socket path
-----------
  Linux/macOS : /tmp/passquantum_face.sock
  Windows     : \\\\.\\pipe\\passquantum_face   (not yet implemented)

Protocol (client → server, one JSON object per line)
------------------------------------------------------
  Ping:   {"cmd": "ping"}
  Enroll: {"cmd": "enroll",  "camera": 0}
  Verify: {"cmd": "verify",  "camera": 0,
           "template_b64": "<base64>", "timeout": 10}

Responses (server → client, one JSON object per line)
------------------------------------------------------
  {"ok": true}
  {"ok": true,  "template_b64": "<base64>"}
  {"ok": true,  "status": "OK", "distance": 0.05, "message": "Face matched"}
  {"ok": false, "error": "<reason>"}

Running standalone
------------------
  python auth_server.py
"""

import base64
import json
import os
import socket
import sys
from typing import Any

import numpy as np

from face_authenticator import FaceAuthenticator, AuthStatus

# ──────────────────────────────────────────────────────────────────────────────
# Configuration
# ──────────────────────────────────────────────────────────────────────────────
SOCK_PATH = "/tmp/passquantum_face.sock"
BACKLOG   = 1   # one client at a time (the Go app)


# ──────────────────────────────────────────────────────────────────────────────
# Helpers
# ──────────────────────────────────────────────────────────────────────────────
def _send(conn: socket.socket, payload: dict[str, Any]) -> None:
    """Send a JSON response terminated with a newline."""
    conn.sendall((json.dumps(payload) + "\n").encode("utf-8"))


def _handle(conn: socket.socket, auth: FaceAuthenticator) -> None:
    """Read newline-delimited JSON requests from one client connection."""
    with conn.makefile("rb") as fh:
        for raw_line in fh:
            raw_line = raw_line.strip()
            if not raw_line:
                continue

            # ── Parse ──────────────────────────────────────────────────────
            try:
                req = json.loads(raw_line)
            except json.JSONDecodeError as exc:
                _send(conn, {"ok": False, "error": f"JSON parse error: {exc}"})
                continue

            cmd = req.get("cmd", "")

            # ── Dispatch ───────────────────────────────────────────────────
            if cmd == "ping":
                _send(conn, {"ok": True})

            elif cmd == "enroll":
                camera = int(req.get("camera", 0))
                try:
                    template = auth.enroll(camera_index=camera)
                    b64 = base64.b64encode(template.tobytes()).decode("ascii")
                    _send(conn, {"ok": True, "template_b64": b64})
                except Exception as exc:
                    _send(conn, {"ok": False, "error": str(exc)})

            elif cmd == "verify":
                camera  = int(req.get("camera", 0))
                timeout = float(req.get("timeout", 10.0))
                b64     = req.get("template_b64", "")
                try:
                    raw_bytes = base64.b64decode(b64)
                    template  = np.frombuffer(raw_bytes, dtype=np.float64)
                    result    = auth.verify(
                        camera_index=camera,
                        template=template,
                        timeout_s=timeout,
                    )
                    _send(conn, {
                        "ok":       result.success,
                        "status":   result.status.name,
                        "distance": result.distance,
                        "message":  result.message,
                    })
                except Exception as exc:
                    _send(conn, {"ok": False, "error": str(exc)})

            else:
                _send(conn, {"ok": False, "error": f"Unknown command: {cmd!r}"})


# ──────────────────────────────────────────────────────────────────────────────
# Entry point
# ──────────────────────────────────────────────────────────────────────────────
def main() -> None:
    # Remove stale socket file from a previous run
    if os.path.exists(SOCK_PATH):
        os.unlink(SOCK_PATH)

    auth   = FaceAuthenticator()
    server = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
    server.bind(SOCK_PATH)
    server.listen(BACKLOG)

    print(f"[auth_server] Listening on {SOCK_PATH}", flush=True)

    try:
        while True:
            conn, _ = server.accept()
            try:
                _handle(conn, auth)
            finally:
                conn.close()
    except KeyboardInterrupt:
        print("\n[auth_server] Shutting down.", flush=True)
    finally:
        server.close()
        if os.path.exists(SOCK_PATH):
            os.unlink(SOCK_PATH)


if __name__ == "__main__":
    main()
