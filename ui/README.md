# ui

The `ui` package is the entry point for the PassQuantum application.

## Contents

- **main.go** — `main()`: initialises the Fyne app, loads the crypto keypair, starts the `bridge.FaceGuard` sidecar, wires the lock callback, and launches the first screen
- **python_bundle.go / python_bundle_windows.go** — embed the pre-built Python sidecar (`face_guard_bundle` / `face_guard_bundle.exe`) and set `PASSQUANTUM_FACE_GUARD_BUNDLE` at init time
- **app.manifest** — Windows application manifest

## Subpackages

| Package | Import path | Description |
|---|---|---|
| `ui/screens` | `passquantum/ui/screens` | All application screens (login, vault selection, main, passwords, settings, training, color/theme pickers, strength widgets) |
| `ui/widgets` | `passquantum/ui/widgets` | Shared dialog helpers (`ShowAppError`, `ShowAppConfirm`, …) and native file picker |

## Dependency graph

```
bridge/   → fyne (prefs), stdlib
app/      → bridge/, core/*, internal/storage
theme/    → fyne only
ui/widgets → theme/, fyne
ui/screens → app/, bridge/, theme/, ui/widgets, strength, palette
ui/main   → app/, bridge/, ui/screens, core/crypto, internal/storage, fyne
```

## Build

The Python sidecar must be bundled before building on non-Linux platforms. See `build.sh` and `Build-FaceBundle.ps1` at the repo root.

