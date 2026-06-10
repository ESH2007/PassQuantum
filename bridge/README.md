# bridge

Package `bridge` manages the Python face-recognition sidecar process and companion-app kill list.

## Contents

- **face_guard.go** — `FaceGuard` type: opens a TCP listener on `127.0.0.1:9876`, launches `python/face_guard.py`, and exchanges line-oriented messages once the sidecar connects back
- **face_guard_apps.go** — kill-list helpers: `LoadKillApps`, `SaveKillApps`, `ListRunningProcesses`, `KillProcessesByName`

## Transport

The Go side opens a TCP listener on `127.0.0.1:9876` *before* launching the
sidecar, then the Python process connects back to that address. All commands and
events below travel over that loopback connection as newline-delimited messages.

## FaceGuard IPC protocol

| Command (→ Python) | Event (← Python) |
|---|---|
| `START_MONITOR` | `FACE_OK` |
| `STOP_MONITOR` | `FACE_LOST` |
| `START_TRAINING` | `TRAINING_DONE` |
| `STOP_TRAINING` | `TRAINING_FAILED` |

The sidecar process path is resolved via the `PASSQUANTUM_FACE_GUARD_BUNDLE` env var (set by `ui/python_bundle*.go` at startup) or falls back to `python/face_guard.py`.

## Kill list

`KillProcessesByName` is called on `FACE_LOST` to terminate user-configured companion applications (e.g. a password-filled browser). The list is persisted with Fyne preferences.
