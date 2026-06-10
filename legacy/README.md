# legacy/

This directory holds archived files that are **no longer part of the active application**.
They are kept for historical reference only. Do not import or build from these files.

> Note: the `new-passquantum/` paths below refer to a former project subdirectory
> that has since become the repository root.

| File | Original location | Why archived |
|---|---|---|
| `main.py` | repo root | Early Python prototype. Stored passwords in a plaintext `passwords.txt` file. Replaced by the Go application in `new-passquantum/`. |
| `face_debug_cli.py` | repo root | Standalone face-recognition debug/prototyping tool. Not imported or launched by any part of the active application. |
| `auth_server.py` | `new-passquantum/` | Experimental standalone HTTP server that wrapped `face_authenticator.py`. Never integrated into the Go binary; not referenced by any build path. |
| `build-native.sh` | `new-passquantum/` | 20-line shell wrapper that ran `go build -o build/linux/PassQuantum ./ui`. Fully superseded by `build.sh linux`. |
| `main.go.backup` | `new-passquantum/` | Backup snapshot of an older Go entry point. No references anywhere in the codebase. |
