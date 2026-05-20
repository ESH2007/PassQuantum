# PassQuantum Architecture

This document describes the current implementation in `new-passquantum/` as shipped by the Go, Python, and build files in the repository.

## 1. System overview

PassQuantum is a desktop password manager with four major runtime concerns:

1. **App access control**
   - Global master-password unlock
   - Private-key fingerprint binding
   - Session key management

2. **Vault storage**
   - Multiple vault files under `vaults/`
   - Authenticated vault encryption
   - Typed vault entries for passwords, notes, and cards

3. **Desktop UX**
   - Fyne windowing and navigation
   - Password generator and checker
   - Visual customization

4. **Face guard**
   - Python sidecar process
   - Local face training and monitoring
   - Lock-on-face-loss behavior
   - Optional process kill list

## 2. High-level module map

```text
ui/
  main.go                 app startup, window, AppState, face-guard wiring
  access_control.go       app unlock, vault open, master-password rotation
  login_screen.go         create/unlock flow
  vault_selection.go      vault list/create/delete/open
  main_screen.go          main sidebar and content views
  passwords_view.go       decrypt/display/edit/copy/delete items
  settings_screen.go      Security, Vaults, Visuals, About
  nativefile.go           OS-native image picker integration
  face_guard.go           Go <-> Python bridge
  python_bundle*.go       embedded face bundle extraction
  app.manifest            Windows compatibility and DPI manifest

core/crypto/
  kdf.go                  Argon2id + domain-separated keys
  vault.go                vault container encryption and HMAC
  kyber.go                Kyber768 key operations
  aes.go                  AES-GCM item payload helpers
  app_security.go         global master-password verifier profile

core/storage/
  storage.go              encrypted vault read/write
  vault_format.go         typed vault payload format and legacy decode
  security_metadata.go    app-security metadata save/load and vault rotation

core/model/
  vault_entry.go          typed entry model and serialization

strength/
  analyzer.go             password analysis pipeline
  *.go                    entropy, similarity, patterns, easter egg rules

palette/
  extractor.go            image color sampling and k-means clustering

top-level helpers
  auth_server.py          alternate JSON IPC server for face auth experiments
  Build-FaceBundle.ps1    Windows self-contained build pipeline
  build.sh                Linux/cross-platform build pipeline
  build-native.sh         simple native Linux build
```

## 3. Runtime architecture

### 3.1 Startup flow

`ui/main.go` performs startup in this order:

1. Normalize locale for Fyne
2. Create the Fyne app and main window
3. Restore icon/theme preferences
4. Load or generate `public.key` and `private.key`
5. Start the face-guard bridge if possible
6. Register the global lock callback
7. Show the master-password screen

### 3.2 Access control flow

`ui/access_control.go` drives startup access:

- If `app-security.pqmeta` does not exist:
  - setup is required
- If the stored profile fingerprint does not match the loaded `private.key`:
  - setup is required again and a warning is shown
- Otherwise:
  - the user is prompted to unlock the app

Once unlocked:

- app-level session keys are stored in `AppState`
- the global master password remains available in memory for vault opening and rotation
- continuous face monitoring is started

### 3.3 Vault flow

Every vault is stored as `vaults/<name>.pqdb`.

Opening a vault:

1. Read the vault file
2. Deserialize its KDF params
3. Derive vault keys from the unlocked global master password and the vault salt
4. Verify and decrypt the vault
5. Cache the current vault state in `AppState`

Creating a vault:

1. Generate new vault salt
2. Derive fresh vault keys from the unlocked global password
3. Write an empty encrypted vault file

Changing the master password:

1. Verify the current password against `app-security.pqmeta`
2. Re-encrypt every vault with keys derived from the new password
3. Write staged `.tmp` files
4. Replace vault files and metadata atomically

## 4. Data layout

### 4.1 App security profile

`app-security.pqmeta` is JSON persisted by `core/storage/security_metadata.go`.

It stores:

- `format_version`
- `private_key_fingerprint`
- `kdf_params`
- `verifier`

It does **not** store the master password itself.

### 4.2 Vault file format

`core/crypto/vault.go` stores:

```text
Version (1 byte)
KDF params length (1 byte)
KDF params (26 bytes)
HMAC-SHA256 (32 bytes)
Encrypted data length (4 bytes)
Encrypted data (nonce + AES-GCM ciphertext)
```

### 4.3 Vault plaintext format

Inside the encrypted vault payload, `core/storage/vault_format.go` uses:

```text
PQV2
EntryCount (uint32)
Repeated:
  EntryLength (uint32)
  TypedEntry
```

The code still supports legacy entry decoding for older vault payloads.

### 4.4 Typed entry model

`core/model/vault_entry.go` supports:

- `EntryTypePassword`
- `EntryTypeNote`
- `EntryTypeCard`

Each entry stores:

- random `ID`
- `Type`
- `CardSubtype`
- `Service`
- `Username`
- `KyberCiphertext`
- `Nonce`
- `Ciphertext`

## 5. Cryptographic layering

### 5.1 App-level unlock

`core/crypto/app_security.go`:

- derives keys with Argon2id
- computes a verifier from:
  - a fixed app label
  - the private-key fingerprint
  - the verification key

This makes the app-level unlock profile unusable with a different `private.key`.

### 5.2 Vault-level encryption

`core/crypto/kdf.go` and `core/crypto/vault.go` provide:

- Argon2id with defaults:
  - 64 MB memory
  - 1 iteration
  - parallelism 4
- domain-separated encryption and verification keys
- AES-256-GCM for vault payload encryption
- HMAC-SHA256 for vault integrity

### 5.3 Item-level secret wrapping

Each vault item is also protected individually:

1. Kyber768 encapsulation generates a shared secret
2. AES-256-GCM encrypts the item payload with that shared secret
3. The entry stores the Kyber ciphertext plus the AES nonce/ciphertext

This is separate from the outer vault encryption layer.

## 6. Face-guard subsystem

### 6.1 Components

| File | Role |
| --- | --- |
| `ui/face_guard.go` | launches and talks to Python |
| `ui/python_bundle.go` | embeds non-Windows bundle |
| `ui/python_bundle_windows.go` | embeds Windows bundle |
| `face_guard.py` | training + monitoring process |
| `auth_server.py` | alternative JSON-over-socket face auth IPC server |
| `geometric_encoder.py` | MediaPipe face-landmarker encoder |
| `liveness_detector.py` | blink-based liveness gate |
| `face_authenticator.py` | higher-level enroll/verify helper |
| `models/face_landmarker.task` | required MediaPipe model asset |

### 6.2 Wire protocol

Python sends:

- `FRAME:<base64 jpeg>`
- `PROGRESS:<n>/<total>`
- `TRAINING_DONE`
- `FACE_OK`
- `FACE_LOST`

Go sends:

- `START_TRAINING`
- `START_MONITOR`

### 6.3 Behavior

Training:

- captures 100 face samples
- requires at least one blink
- saves encodings to `face_data.npy`

Monitoring:

- waits for a recognized live face first
- then continuously checks for a recognized face
- sends `FACE_LOST` after 5 seconds of absence
- sends `FACE_OK` when the face returns

Go reacts to `FACE_LOST` by:

- locking the app
- clearing sensitive state
- killing monitored companion processes selected by the user

## 7. UI architecture

### 7.1 Main screens

The user-visible flow is:

```text
Create/Unlock
  -> Face Registration (first setup when needed)
  -> Vault Selection
  -> Main Sidebar Shell
       -> Passwords
       -> Generate
       -> Check Password
       -> Settings
```

### 7.2 Sidebar views

`ui/main_screen.go` defines five primary views:

- Vaults
- Passwords
- Generate
- Check Password
- Settings

The main password-entry view supports:

- password items
- "Cyphered Note" items
- card items

### 7.3 Settings structure

`ui/settings_screen.go` currently exposes four sections:

- **Security**
  - change master password
  - monitored app kill list
- **Vaults**
  - status labels and placeholder maintenance/backup actions
- **Visuals**
  - theme/image palette/icon customization
- **About**
  - static product information and placeholder docs/update actions

## 8. Password intelligence subsystem

The `strength/` package powers both:

- inline strength feedback on password entry
- the dedicated password checker screen

The analyzer combines:

- repeated-character checks
- keyboard-pattern checks
- leet-speak detection
- date detection
- common words and names
- missing character classes
- similarity to already stored vault passwords
- entropy and crack-time estimation

There is also an easter-egg mode triggered by passwords containing `neal.fun`.

## 9. Theme and palette subsystem

The UI uses a custom neon palette and animated particle background from `ui/ui_theme.go`.

Visual customization currently includes:

- image-driven palette extraction using `palette/extractor.go`
- manual color personalization dialogs
- app icon replacement stored in preferences

## 10. Build architecture

### 10.1 Plain Go build

`go build .\ui` compiles the Fyne app without embedding the PyInstaller face bundle.

### 10.2 Linux bundle-aware build

`build.sh linux`:

1. builds the Python bundle with PyInstaller when possible
2. builds Go with `-tags with_face_bundle`
3. otherwise falls back to copying Python sources and `models/`

### 10.3 Windows self-contained build

`Build-FaceBundle.ps1`:

1. creates `.venv-faceguard`
2. produces `ui\face_guard_bundle.exe`
3. generates `ui\rsrc.syso` from `ui\app.manifest`
4. sets `CGO_ENABLED=1`
5. prefers MSYS2 MinGW-w64 GCC
6. builds `build\windows\PassQuantum.exe`
7. removes the temporary `rsrc.syso`

### 10.4 Embedded bundle extraction

When built with `with_face_bundle`:

- the Python bundle is embedded with `//go:embed`
- extracted into `%TEMP%/passquantum-face-guard`
- exposed through environment variables to the Go launcher

## 11. Generated and third-party areas

These paths are present in the folder but are not primary product source:

- `.venv-faceguard/` - build-time Python environment
- `build/` - local build outputs
- `fyne-cross/` - cross-packaging outputs
- `vendor/` - vendored Go dependencies

They still matter operationally because the current repository includes packaged artifacts and build outputs alongside the source.

## 12. Current implementation caveats

The docs should reflect these realities:

- The app is gated by a global app-security profile first, not by opening a vault directly from the login screen.
- Vault settings for compaction/export/import/backup/restore are mostly placeholder dialogs right now.
- The About page still contains static version/support copy from the UI layer.
- The Windows self-contained build path depends on the PowerShell script, `rsrc`, and MSYS2 GCC.
