# PassQuantum

PassQuantum is a Fyne desktop password manager built around post-quantum vault items, a global master-password gate, multiple encrypted vaults, password analysis tools, a browser-extension autofill bridge, and a Python face-guard sidecar that can lock the app when your face disappears.

This README reflects the current code in the repository, including generated build assets and the current Windows self-contained build flow.

## What the app currently does

- Uses a global master password verified through `app-security.pqmeta`
- Binds that verifier to the current `private.key` fingerprint
- Stores multiple encrypted vaults in `vaults/*.pqdb`
- Supports five vault item types:
  - Passwords
  - "Cyphered Note" items
  - Card items
  - TOTP / 2FA codes (manual entry, QR import, Google Authenticator export)
  - Encrypted files (arbitrary documents stored inside a vault)
- Imports from 11 other password managers (1Password, Bitwarden, KeePass, LastPass, Dashlane, NordPass, Proton Pass, Kaspersky, Chrome/Brave/Edge, Firefox, and generic CSV)
- Includes a password generator and a password strength analyzer
- Ships a browser extension that autofills and saves credentials by talking to a localhost-only server (`127.0.0.1:8765`) gated by a pairing token
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
| `build/mac/PassQuantum.app` | macOS application bundle |
| `build/mac/PassQuantum.dmg` | macOS distributable disk image |
| `build-mac-native.sh` | Native macOS build script |

## Repository layout

| Path | Purpose |
| --- | --- |
| `app/` | Application lifecycle: `AppState`, startup access control, vault CRUD, master-password rotation |
| `bridge/` | Face-guard sidecar manager (TCP IPC) and companion-app kill list |
| `core/crypto/` | KDF, vault encryption (legacy + PQ), Kyber768/Dilithium3, app-security profile logic |
| `core/model/` | Vault entry types (Password, Note, Card, TOTP, File) and binary serialization |
| `core/storage/` | Vault persistence, metadata persistence, key-rotation helpers |
| `core/filevault/` | Encrypted per-file storage and manifest |
| `core/migration/` | Import framework and parsers for 11 password managers |
| `core/totp/` | TOTP generation, `otpauth://` parsing, QR/Google-Authenticator decoding |
| `internal/storage/` | Low-level secure file I/O, OS keyring, Windows DPAPI |
| `internal/browser/` | Localhost autofill server for the browser extension |
| `ui/` | Fyne app entry point and embedded-bundle support |
| `ui/screens/` | All application screens and view-builders |
| `ui/widgets/` | Shared dialogs and native file pickers |
| `theme/` | Design tokens, fonts, icons, and widget factories |
| `strength/` | Password analysis, scoring, issue detection, generator helpers |
| `palette/` | Image-based palette extraction helpers |
| `python/` | Face-guard Python sidecar (face_guard.py, face_authenticator.py, liveness_detector.py, geometric_encoder.py) |
| `extension/` | Browser extension (Manifest V3) source |
| `landing page/` | Marketing website (React/Vite), deployed to GitHub Pages |
| `docs/` | Architecture, security, and UX documents |
| `models/` | Face-landmarker model asset and required task file |
| `legacy/` | Archived prototypes, kept for reference only |
| `cmd/test-vault/` | Manual vault test utility |
| `build/` | Local build outputs |
| `.venv-faceguard/` | Isolated build-time Python environment (created on demand by the build scripts) |

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

### macOS native build (Apple Silicon / Intel)

Build a fully self-contained, ad-hoc-signed `.app` and `.dmg` directly on a Mac:

```bash
./build-mac-native.sh
# or, equivalently:
./build.sh mac
```

Prerequisites:

- Go 1.22+
- Python 3.11+
- Xcode Command Line Tools (`xcode-select --install`)

What it does:

1. Creates or reuses `.venv-faceguard` (isolated from `.venv`)
2. Installs `pyinstaller`, `mediapipe`, `opencv-python-headless`, and `numpy<2`
3. Builds the native `face_guard_bundle` with PyInstaller `--onedir`
4. Packages `build/mac/PassQuantum.app` with `fyne package -os darwin`
5. Copies the helper **inside** the bundle at `Contents/Resources/faceguard/`
6. Injects `NSCameraUsageDescription` into the app's `Info.plist`
7. Ad-hoc code signs the `.app` (including the in-bundle helper), then creates and signs `build/mac/PassQuantum.dmg`

Outputs:

- `build/mac/PassQuantum.app` — application bundle (embedded face guard, camera permission declared)
- `build/mac/PassQuantum.dmg` — distributable disk image

The architecture is detected automatically (`arm64` → Apple Silicon, `x86_64` → Intel).
Because the DMG is **ad-hoc signed** (not notarized), users may need to right-click → Open
on first launch — see [macOS: First launch](#macos-first-launch) below. Flags
`--skip-bundle`, `--skip-sign`, and `--skip-dmg` are available; run with `--help` for
details. Unlike Linux/Windows (which embed the helper via `go:embed` and extract it to a
temp dir), macOS ships the face-guard helper **inside** the signed `.app` and runs it in
place — this is required so the camera permission (TCC) attributes to PassQuantum. If the
camera was previously denied during testing, reset its decision with
`tccutil reset Camera com.passquantum.app`.

### Cross-platform script

`build.sh` provides:

- `./build.sh linux`
- `./build.sh windows`
- `./build.sh mac`
- `./build.sh all`

Important notes:

- Linux can embed the PyInstaller face bundle into the Go binary.
- Windows cross-builds can create a Windows PyInstaller bundle through Docker + Wine.
- On a Mac, `./build.sh mac` delegates to `build-mac-native.sh`, producing a self-contained
  `.app`/`.dmg` with the PyInstaller face bundle embedded (see the macOS native build above).
- On Linux, `./build.sh mac` still cross-compiles via `fyne-cross` and ships Python sources
  alongside the app (an external macOS SDK is required).

## Installation notes

### macOS: First launch

macOS may show a warning because the app is not notarized by Apple. To open it:

**Option A:** Right-click the app → Open → "Open" in the dialog

**Option B (terminal):**
```bash
xattr -rd com.apple.quarantine /Applications/PassQuantum.app
```

This only needs to be done once.

## Current UI surface

After a vault is opened, the app uses a grouped sidebar:

- **Vault** — Vaults, Add item, Items, Authenticator (TOTP), Files, Import
- **Tools** — Generate, Analyze (password strength)
- **System** — Settings, Collapse/Expand sidebar, Lock vault

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
- The About screen still shows static version and support text from the UI layer.

## Main documentation

- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) — code and runtime architecture
- [`docs/SECURITY_ARCHITECTURE.md`](docs/SECURITY_ARCHITECTURE.md) — security model and threat boundaries
- [`docs/USER_EXPERIENCE.md`](docs/USER_EXPERIENCE.md) — screen-by-screen product behavior
- [`docs/USER_GUIDE.md`](docs/USER_GUIDE.md) — practical setup and usage instructions

Most packages also carry their own `README.md` with a file-by-file overview.

## Development notes

- Go dependencies are resolved via the module cache (no committed `vendor/` tree); run `go mod download` if needed.
- The face-guard Python tooling lives in `python/`. See [`python/README.md`](python/README.md) for setup instructions.
- `.venv-faceguard/` and `build/` contain generated or packaging artifacts created by the build scripts.
- `models/face_landmarker.task` is required for the MediaPipe-based face workflow.

## Support expectations

This repository currently documents the code as it exists. If you are building or packaging the app, prefer the source files and build scripts over older assumptions in historical docs.
