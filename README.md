# PassQuantum

PassQuantum is a Fyne desktop password manager built around post-quantum vault items, a global master-password gate, multiple encrypted vaults, password analysis tools, and a Python face-guard sidecar that can lock the app when your face disappears.

This README reflects the current code in `new-passquantum/`, including generated build assets and the current Windows self-contained build flow.

## What the app currently does

- Uses a global master password verified through `app-security.pqmeta`
- Binds that verifier to the current `private.key` fingerprint
- Stores multiple encrypted vaults in `vaults/*.pqdb`
- Supports three vault item types:
  - Passwords
  - "Cyphered Note" items
  - Card items
- Includes a password generator and a password checker
- Starts a face-guard subprocess that can:
  - train a local face profile
  - monitor the webcam continuously after unlock
  - lock the app after 5 seconds without a recognized face
  - optionally kill user-selected companion apps
- Lets the user customize palette colors and the app icon

## Core runtime files

| File or folder | Role |
| --- | --- |
| `public.key` | Kyber768 public key |
| `private.key` | Kyber768 private key |
| `app-security.pqmeta` | Global master-password verifier profile |
| `vaults/*.pqdb` | Encrypted vault files |
| `face_data.npy` | Stored face encodings for the Python face guard |
| `ui/face_guard_bundle.exe` | PyInstaller output used for self-contained Windows builds |
| `build/windows/PassQuantum.exe` | Windows build output from `Build-FaceBundle.ps1` |
| `build/linux/PassQuantum` | Native Linux build output |

## Repository layout

| Path | Purpose |
| --- | --- |
| `core/crypto/` | KDF, vault encryption, Kyber, app-security profile logic |
| `core/model/` | Vault entry types and binary serialization |
| `core/storage/` | Vault persistence, metadata persistence, migration helpers |
| `ui/` | Fyne UI, navigation, dialogs, face-guard bridge, embedded bundle support |
| `strength/` | Password analysis, scoring, issue detection, generator helpers |
| `palette/` | Image-based palette extraction helpers |
| `python/` | Face-guard Python sidecar (face_guard.py, face_authenticator.py, liveness_detector.py, geometric_encoder.py) |
| `docs/` | Architecture, security, UX, and specification documents |
| `models/` | Face-landmarker model asset documentation and required task file |
| `cmd/test-vault/` | Manual vault test utility |
| `build/` | Local build outputs |
| `fyne-cross/` | Cross-build outputs and temporary packaging artifacts |
| `.venv-faceguard/` | Isolated Windows build-time Python environment |
| `vendor/` | Vendored Go dependencies |

## Security model summary

PassQuantum uses two distinct layers of protection:

1. **App-level unlock**
   - The master password is verified through `app-security.pqmeta`
   - The verifier is derived with Argon2id
   - The profile is tied to the fingerprint of the current `private.key`

2. **Vault-level encryption**
   - Each vault has its own salt and derived encryption/verification keys
   - Vault payloads are encrypted with AES-256-GCM
   - Vault files are authenticated with HMAC-SHA256
   - Each stored item also uses Kyber768 encapsulation plus AES-GCM for its own payload

See `docs/SECURITY_ARCHITECTURE.md` for the full design.

## Build paths

### Fast local Go build

This is the simplest developer build:

```powershell
go build -o build\PassQuantum.exe .\ui
```

That build keeps the regular source-based Python fallback path. It does **not** produce the Windows self-contained embedded face bundle.

### Windows self-contained build

Use the PowerShell build script:

```powershell
.\Build-FaceBundle.ps1
```

Current behavior:

1. Creates or reuses `.venv-faceguard`
2. Installs `pyinstaller`, `mediapipe`, `opencv-python`, and `numpy<2`
3. Builds `ui\face_guard_bundle.exe`
4. Generates `ui\rsrc.syso` from `ui\app.manifest`
5. Builds `build\windows\PassQuantum.exe` with `-tags with_face_bundle`
6. Prefers `C:\msys64\mingw64\bin\gcc.exe` because MSYS2 MinGW-w64 works correctly for the CGO/Fyne Windows build

The embedded bundle is extracted at runtime to:

```text
%TEMP%\passquantum-face-guard\face_guard.exe
```

### Cross-platform script

`build.sh` provides:

- `./build.sh linux`
- `./build.sh windows`
- `./build.sh mac`
- `./build.sh all`

Important notes:

- Linux can embed the PyInstaller face bundle into the Go binary.
- Windows cross-builds can create a Windows PyInstaller bundle through Docker + Wine.
- macOS currently ships Python sources and models alongside the app rather than an embedded PyInstaller bundle.

## Current UI surface

After unlock, the app uses a sidebar with these views:

- Vaults
- Passwords
- Generate
- Check Password
- Settings
- Lock & Exit

The Settings area currently has four sections:

- Security
- Vaults
- Visuals
- About

Some settings actions are real today, and some are placeholders that only show informational dialogs. The UX docs below call that out explicitly.

## Known implementation realities to document accurately

- The app is protected by a **global** master-password profile first, then vaults are opened with keys derived from that unlocked password.
- The face guard is a separate Python process managed from Go over localhost TCP.
- The monitored-app kill list is stored in Fyne preferences and uses process names.
- Theme/image/icon personalization is implemented.
- Backup/export/import/restore actions in the Vault settings screen are mostly UI placeholders right now.
- The About screen still shows static version and support text from the UI layer.

## Main documentation

- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) — code and runtime architecture
- [`docs/SECURITY_ARCHITECTURE.md`](docs/SECURITY_ARCHITECTURE.md) — security model and threat boundaries
- [`docs/USER_EXPERIENCE.md`](docs/USER_EXPERIENCE.md) — screen-by-screen product behavior
- [`docs/USER_GUIDE.md`](docs/USER_GUIDE.md) — practical setup and usage instructions
- [`docs/GO_APP_SPECIFICATION.md`](docs/GO_APP_SPECIFICATION.md) — additional implementation notes and file coverage

## Development notes

- The project vendors Go dependencies in `vendor/`.
- The face-guard Python tooling lives in `python/`. See [`python/README.md`](python/README.md) for setup instructions.
- `.venv-faceguard/`, `build/`, and `fyne-cross/` contain generated or packaging artifacts.
- `models/face_landmarker.task` is required for the MediaPipe-based face workflow.

## Support expectations

This repository currently documents the code as it exists. If you are building or packaging the app, prefer the source files and build scripts over older assumptions in historical docs.
