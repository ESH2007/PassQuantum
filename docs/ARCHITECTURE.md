# PassQuantum Architecture

This document describes the current implementation as shipped by the Go, Python,
and build files in the repository.

## 1. System overview

PassQuantum is a desktop password manager with these major runtime concerns:

1. **App access control** — global master-password unlock, private-key
   fingerprint binding, session-key management.
2. **Vault storage** — multiple encrypted vault files holding typed entries
   (passwords, notes, cards, TOTP secrets, encrypted files).
3. **Desktop UX** — Fyne windowing, navigation, generator, strength analyzer,
   import wizard, and visual customization.
4. **Face guard** — an optional Python sidecar that monitors the webcam after
   unlock and locks the app when the recognized face disappears.
5. **Browser bridge** — a localhost-only HTTP server that lets the browser
   extension autofill and save credentials from an unlocked vault.

## 2. Module map

```text
app/                       application lifecycle
  state.go                 AppState + helper methods
  access.go                startup access, profile create/verify, master-pw rotation
  helpers.go               vault CRUD + crypto wrappers + password validation
  import.go                import glue between core/migration and the UI

bridge/                    face-guard sidecar manager
  face_guard.go            TCP server (127.0.0.1:9876) + process lifecycle + IPC
  face_guard_apps.go       companion-app kill list

core/crypto/
  kdf.go                   Argon2id + domain-separated keys
  vault.go                 legacy vault container (AES-GCM + HMAC) — read-compat
  vault_pq.go              PQ vault container "PQVT" (Kyber768 + Dilithium3 + AES-GCM)
  kyber.go                 Kyber768 keypair / encapsulation
  aes.go                   AES-256-GCM helpers
  app_security.go          global master-password verifier profile

core/model/vault_entry.go  typed entry model (Password, Note, Card, TOTP, File)
core/storage/              vault + security-metadata persistence, key rotation
core/filevault/            encrypted per-file storage + manifest
core/migration/            import framework + 11 parsers
core/totp/                 TOTP generation, otpauth:// parsing, QR decoding

internal/storage/          secure file I/O, OS keyring, Windows DPAPI
internal/browser/          localhost autofill server (127.0.0.1:8765)

ui/main.go                 app startup, key init, sidecar wiring, first screen
ui/python_bundle*.go       embedded face-bundle extraction
ui/screens/                all screens (login, vaults, main shell, passwords,
                           totp, files, import, settings, training, pairing, …)
ui/widgets/                shared dialogs + native file pickers

theme/                     design tokens, fonts, icons, widget factories
strength/                  password analysis pipeline
palette/                   image color sampling + k-means

build.sh                   Linux/cross build pipeline
build-mac-native.sh        native macOS .app/.dmg build
Build-FaceBundle.ps1       Windows self-contained build pipeline
```

## 3. Runtime architecture

### 3.1 Startup flow

`ui/main.go` performs startup roughly in this order:

1. Normalize locale for Fyne
2. Create the Fyne app and main window
3. Restore icon/theme preferences
4. Load or generate `public.key` and `private.key`
5. Start the face-guard bridge if possible
6. Register the global lock callback
7. Show the master-password screen

### 3.2 Access control flow

`app/access.go` drives startup access via `ResolveStartupAccessState`:

- If `app-security.pqmeta` does not exist → setup is required.
- If the stored profile fingerprint does not match the loaded `private.key` →
  setup is required again and a warning is shown.
- Otherwise → the user is prompted to unlock the app.

Once unlocked, app-level session keys are stored in `app.AppState`, the global
master password remains in memory for vault opening and rotation, and continuous
face monitoring is started.

### 3.3 Vault flow

Every vault is stored as `vaults/<name>.pqdb`. Opening a vault reads the file,
deserializes its KDF params, derives vault keys from the unlocked global master
password and the vault salt, verifies/decrypts, and caches the state in
`AppState`. Creating a vault generates a fresh salt, derives keys, and writes an
empty encrypted vault.

Changing the master password (`ChangeMasterPassword` in `app/access.go`):

1. Verify the current password against `app-security.pqmeta`
2. Re-encrypt every vault with keys derived from the new password
3. Stage `.tmp` files for all vaults and the metadata
4. Activate them with atomic renames, rolling back on any failure

## 4. Data layout

### 4.1 App security profile

`app-security.pqmeta` is JSON persisted by `core/storage/security_metadata.go`.
It stores `format_version`, `private_key_fingerprint`, `kdf_params`, and a
`verifier` — never the master password itself.

### 4.2 Vault file formats

Two on-disk container formats coexist:

- **`PQVT` (current, `core/crypto/vault_pq.go`):** Argon2id (iter 2, 64 MB,
  parallelism 4, 32-byte salt) derives a master key; HKDF expands it into a
  Kyber768 seed and a Dilithium3 seed. The payload is AES-256-GCM encrypted and a
  Dilithium3 signature authenticates the header + payload.
- **Legacy (`core/crypto/vault.go`):** AES-256-GCM payload with an HMAC-SHA256 over
  version + KDF params + ciphertext. Kept for backward-compatible reads of older
  vaults; `core/storage` migrates them transparently.

### 4.3 Vault plaintext format

Inside the decrypted payload, `core/storage/vault_format.go` stores a magic
header, an entry count, and the typed entries length-prefixed. Legacy entry
decoding is retained for older payloads.

### 4.4 Typed entry model

`core/model/vault_entry.go` supports `EntryTypePassword`, `EntryTypeNote`,
`EntryTypeCard`, `EntryTypeTOTP`, and `EntryTypeFile`. Each entry carries a random
`ID`, `Type`, optional `CardSubtype`, `Service`, `Username`, and the per-entry
`KyberCiphertext` + `Nonce` + `Ciphertext` wrapping its payload.

## 5. Cryptographic layering

### 5.1 App-level unlock

`core/crypto/app_security.go` derives keys with Argon2id and computes a verifier
from a fixed app label, the private-key fingerprint, and the verification key.
This makes the unlock profile unusable with a different `private.key`.

### 5.2 Vault-level encryption

See §4.2. New vaults use the `PQVT` pipeline (Kyber768 + Dilithium3 +
AES-256-GCM); legacy AES-GCM + HMAC vaults remain readable.

### 5.3 Item-level secret wrapping

Each vault item is also protected individually: Kyber768 encapsulation produces a
shared secret, AES-256-GCM encrypts the item payload with it, and the entry
stores the Kyber ciphertext plus the AES nonce/ciphertext — independent of the
outer vault layer. Encrypted files (`core/filevault`) use the same scheme,
streamed chunk-by-chunk so large files are never fully buffered.

## 6. Face-guard subsystem

### 6.1 Components

| File | Role |
| --- | --- |
| `bridge/face_guard.go` | opens the TCP listener, launches Python, parses IPC |
| `bridge/face_guard_apps.go` | companion-app kill list |
| `ui/python_bundle*.go` | embed/extract the PyInstaller bundle |
| `python/face_guard.py` | training + monitoring entry point |
| `python/geometric_encoder.py` | MediaPipe landmark encoder |
| `python/liveness_detector.py` | blink-based liveness gate |
| `python/face_authenticator.py` | enroll/verify helper (build-time module) |
| `models/face_landmarker.task` | required MediaPipe model asset |

(The old `auth_server.py` experiment now lives under `legacy/`.)

### 6.2 Wire protocol

Go opens a TCP listener on `127.0.0.1:9876` and the Python process connects back.
Python sends `FRAME:<base64 jpeg>`, `PROGRESS:<n>/<total>`, `TRAINING_DONE`,
`TRAINING_FAILED`, `FACE_OK`, and `FACE_LOST`; Go sends `START_TRAINING`,
`STOP_TRAINING`, `START_MONITOR`, and `STOP_MONITOR`.

### 6.3 Behavior

Training captures face samples, requires at least one blink, and saves encodings
to `face_data.npy`. Monitoring waits for a recognized live face, then continuously
checks for it and sends `FACE_LOST` after ~5 seconds of absence (`FACE_OK` when it
returns). On `FACE_LOST`, Go locks the app, clears sensitive state, and kills any
user-selected companion processes.

## 7. UI architecture

### 7.1 Flow

```text
Create/Unlock
  -> Face Registration (first setup when needed)
  -> Vault Selection
  -> Main Sidebar Shell
```

### 7.2 Sidebar (`ui/screens/main_screen.go`)

`NavigationState` hosts a grouped sidebar:

- **Vault:** Vaults, Add item, Items, Authenticator (TOTP), Files, Import
- **Tools:** Generate, Analyze
- **System:** Settings, Collapse/Expand, Lock vault

The "Add item" form supports Password, Cyphered Note, Card, and TOTP types.

### 7.3 Settings (`ui/screens/settings.go`)

Four sections: **Security** (change master password, monitored-app kill list),
**Vaults** (status + mostly-placeholder maintenance/backup actions), **Visuals**
(theme/palette/icon customization), and **About** (static product info).

## 8. TOTP subsystem

`core/totp` generates time-based codes, parses `otpauth://totp/...` URIs, decodes
QR images, and decodes Google Authenticator export payloads. A TOTP entry stores
its serialized parameters as an encrypted vault payload; the live code is computed
on demand by `ui/screens/totp.go`, which also handles clipboard auto-clear.

## 9. File-vault subsystem

`core/filevault` stores arbitrary files inside a vault, each encrypted with
Kyber768 + AES-256-GCM and tracked by a JSON manifest. `ui/screens/filevault.go`
drives store/retrieve/open/delete, decrypting to tracked temp files that are
securely deleted afterward.

## 10. Import / migration subsystem

`core/migration` auto-detects an export file's format, parses it into a normalized
model, then maps and encrypts entries into vault entries (de-duplicating against
existing items). Eleven parsers are registered; `ui/screens/import_wizard.go`
walks the user through pick → parse → preview → import.

## 11. Browser bridge

`internal/browser` runs a loopback-only HTTP server on `127.0.0.1:8765`, gated by
a pairing token, exposing `/vault/pair`, `/vault/status`, `/vault/exists`,
`/vault/save`, `/vault/update/`, and `/vault/never-save`. A `DomainMap` associates
domains with entry IDs. The extension client lives in `extension/`; the desktop
side surfaces pairing through `ui/screens/pairing_dialog.go`.

## 12. Theme and palette subsystem

The UI uses the `theme` package's design tokens — a teal accent (`#2dd4bf`) on a
dark blue-grey background (`#0a0d12`) — plus an animated particle background.
Visual customization includes image-driven palette extraction (`palette/`),
manual color personalization, and app-icon replacement, all stored in Fyne
preferences.

## 13. Build architecture

- **Plain Go build:** `go build ./ui` compiles the Fyne app without embedding the
  PyInstaller face bundle (source-based Python fallback path).
- **Linux (`build.sh linux`):** builds the Python bundle with PyInstaller when
  possible and links it via `-tags with_face_bundle`; otherwise copies Python
  sources + `models/`.
- **Windows (`Build-FaceBundle.ps1`):** builds `ui\face_guard_bundle.exe`,
  generates `ui\rsrc.syso` from `ui\app.manifest`, and builds
  `build\windows\PassQuantum.exe` with `CGO_ENABLED=1`, preferring MSYS2
  MinGW-w64 GCC.
- **macOS (`build-mac-native.sh` / `build.sh mac`):** builds a self-contained,
  ad-hoc-signed `.app`/`.dmg` with the face bundle inside `Contents/Resources`.
- **Embedded bundle extraction:** when built with `with_face_bundle`, the Python
  bundle is embedded via `//go:embed` and extracted to a temp dir, exposed to the
  launcher through `PASSQUANTUM_FACE_GUARD_BUNDLE`.

## 14. Generated and third-party areas

`build/` (local outputs), `fyne-cross/` (cross-packaging outputs, created on
demand), and `.venv-faceguard/` (build-time Python env) are generated. `legacy/`
holds archived prototypes that are not part of the active app.

## 15. Current implementation caveats

- The app is gated by a global app-security profile first, not by opening a vault
  directly from the login screen.
- Some Settings → Vaults actions (compaction/export/backup/restore) are still
  placeholder dialogs. Import, by contrast, is fully implemented as its own view.
- The About page still contains static version/support copy from the UI layer.
- The Windows self-contained build path depends on the PowerShell script, `rsrc`,
  and MSYS2 GCC.
