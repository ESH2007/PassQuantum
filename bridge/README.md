# bridge

Package `bridge` manages the Python face-recognition sidecar process and companion-app kill list.

## Contents

- **face_guard.go** — `FaceGuard` type: launches and communicates with `python/face_guard.py` over stdin/stdout
- **face_guard_apps.go** — kill-list helpers: `LoadKillApps`, `SaveKillApps`, `ListRunningProcesses`, `KillProcessesByName`

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
