# PassQuantum User Guide

This guide is the practical companion to `USER_EXPERIENCE.md`. It focuses on how to install, run, and use the app in its current state.

## 1. Before you start

PassQuantum is a local desktop application. Your important runtime data is stored next to the app in files such as:

- `public.key`
- `private.key`
- `app-security.pqmeta`
- `vaults/*.pqdb`
- `face_data.npy`

If you lose the private key, the app-security profile, or the vault files together, recovery becomes difficult or impossible.

## 2. Running the app

### Option A: simple developer build

From the repository root:

```bash
go build -o build/PassQuantum ./ui
./build/PassQuantum
```

This is the easiest local build for development (use `build\PassQuantum.exe` and
`.\ui` on Windows PowerShell).

### Option B: self-contained Windows build

For a Windows binary that embeds the Python face-guard bundle:

```powershell
.\Build-FaceBundle.ps1
.\build\windows\PassQuantum.exe
```

Notes:

- The script builds `ui\face_guard_bundle.exe` first.
- It then embeds that bundle into `PassQuantum.exe`.
- It prefers MSYS2 MinGW-w64 GCC from `C:\msys64\mingw64\bin\gcc.exe`.

### Option C: Linux build script

```bash
./build.sh linux
./build/linux/PassQuantum
```

This builds the PyInstaller face bundle when possible and links it into the
binary; otherwise it copies the Python sources and `models/` alongside the app.

### Option D: macOS native build

```bash
./build.sh mac        # or ./build-mac-native.sh
open build/mac/PassQuantum.app
```

Produces a self-contained, ad-hoc-signed `.app`/`.dmg` with the face bundle
embedded.

## 3. First launch

On first launch:

1. the app generates `public.key` and `private.key` if they do not already exist
2. you create a **global** master password
3. the app creates a default vault if none exists
4. if the face guard is available, the app may ask you to complete facial registration

Important: the master password protects the whole app session first. Vaults are then opened using keys derived from that unlocked password.

## 4. Unlocking the app

On later launches:

1. enter the global master password
2. the app verifies it against `app-security.pqmeta`
3. face monitoring starts in the background if the face guard is active
4. the vault-selection screen opens

If you enter the wrong password, the app stays locked.

## 5. Creating and opening vaults

### Create a vault

1. Open the `Vaults` view
2. Click `+ CREATE VAULT`
3. Enter a vault name
4. Confirm creation

The app uses the already unlocked global master password automatically.

### Open a vault

1. In the `Vaults` view, find the vault card
2. Click `OPEN`
3. The main sidebar shell opens with that vault active

## 6. Adding items

Open the `Add item` view and choose the item type:

- `Password`
- `Cyphered Note`
- `Card`
- `TOTP`

(Encrypted files are added from the `Files` view — see §8.)

### Save a password

Fill in:

- service name
- username or email
- password

Then click `SAVE ITEM`.

### Save a note

Fill in:

- note title
- note content

Then click `SAVE ITEM`.

### Save a card

Fill in:

- card type
- nickname
- holder
- number
- expiry
- CVV

Then click `SAVE ITEM`.

### Save a TOTP code

Add a 2FA entry in one of three ways:

- paste an `otpauth://totp/...` URI,
- import a QR code image, or
- import a Google Authenticator export (`otpauth-migration://...`).

The entry then appears in the `Authenticator` view with a live code.

## 7. Viewing, editing, and deleting items

Open the `Items` view.

Available actions:

- **passwords**: show, copy, edit, delete
- **notes**: view, copy, delete
- **cards**: show, copy, delete
- **TOTP**: live code with countdown, copy (auto-clears from clipboard)

Deleting an item is permanent.

## 8. Storing files

Open the `Files` view to keep encrypted files inside the current vault:

- **Add a file** — pick a file; it is encrypted and copied into the vault's file
  store.
- **Open a file** — it is decrypted to a temporary location for viewing, then
  securely deleted afterward.
- **Retrieve** — save a decrypted copy to a path you choose.
- **Delete** — permanently removes the stored file.

## 9. Importing from another password manager

Open the `Import` view and follow the wizard:

1. **Pick** the source (or let auto-detection identify it from the file).
2. **Parse** the export.
3. **Preview** the entries that will be imported.
4. **Import** — entries are encrypted into the current vault, skipping duplicates.

Supported sources: 1Password (1PUX), Bitwarden (CSV/JSON), KeePass/KeePassXC,
LastPass, Dashlane, NordPass, Proton Pass, Kaspersky (TXT), Chrome/Brave/Edge/
Opera/Vivaldi, Firefox, and a generic CSV fallback.

Export the data from your old manager first, then point the wizard at the file.

## 10. Pairing the browser extension

To autofill in your browser:

1. Load the extension from the `extension/` folder (Chrome/Edge: *Load unpacked*;
   Firefox: *Load Temporary Add-on*).
2. In the app, open the pairing dialog to display a one-time token.
3. Enter the token in the extension popup.

Once paired and with a vault unlocked, the extension autofills matching logins and
offers to save new ones. It only talks to the app over `127.0.0.1:8765`; locking
the vault cuts off access.

## 11. Using the password generator

Open the `Generate` view.

You can choose:

- password length
- uppercase letters
- lowercase letters
- numbers
- special characters
- whether to exclude ambiguous characters

Then:

- click `GENERATE`
- optionally `COPY`
- or `SAVE TO VAULT`

## 12. Using the password checker

Open the `Analyze` view.

Type any password to see:

- score
- strength label
- crack-time estimate
- detected weaknesses

The checker also compares against passwords already stored in your current vault.

## 13. Changing the master password

1. Go to `Settings`
2. Open `Security`
3. Click `CHANGE MASTER PASSWORD`
4. Enter:
   - current password
   - new password
   - confirmation

On success the app:

- updates `app-security.pqmeta`
- re-encrypts every vault with keys derived from the new password

## 14. Face-guard usage

If the face guard is available, PassQuantum uses a Python subprocess to monitor the webcam.

### Training

During registration:

- stay in front of the camera
- let the app collect samples
- blink when prompted by the liveness logic

### Monitoring

After unlock:

- the app starts monitoring automatically
- if your recognized face disappears for 5 seconds, the app locks

If you configured monitored apps in `Settings -> Security`, those apps will be force-closed at that moment.

## 15. Settings you can rely on today

### Fully or mostly implemented

- change master password
- monitored-app selection
- manual color personalization
- image-driven palette extraction
- app icon replacement
- palette reset

- import from other password managers (the `Import` view — §9)
- TOTP / authenticator codes (§6)
- encrypted file storage (§8)
- browser-extension autofill (§10)

### Present but mostly placeholder

- compact vault
- the Settings → Vaults export/backup/restore buttons
- docs button
- updates button

Treat the second group as UI placeholders, not full backup features. (The
Settings → Vaults `Import` button is also a placeholder — the working importer is
the dedicated `Import` view.)

## 16. Backup guidance

Back up these together:

- `public.key`
- `private.key`
- `app-security.pqmeta`
- `vaults/` (including any encrypted file store and manifest kept alongside it)
- `face_data.npy` if you want to preserve face training

If you move the app to another machine or folder, keep those files together.
Losing `private.key` or `app-security.pqmeta` makes the vaults unrecoverable.

## 17. Common problems

### The app keeps asking me to create a master password

Possible causes:

- `app-security.pqmeta` is missing
- `private.key` changed and no longer matches the stored profile

### A vault will not open

Possible causes:

- wrong global master password before unlock
- corrupted vault file
- mismatched migrated files

### Windows self-contained build fails

Use:

```powershell
.\Build-FaceBundle.ps1
```

And make sure:

- Python is installed
- MSYS2 MinGW-w64 GCC is available at `C:\msys64\mingw64\bin\gcc.exe`
- the script can build `ui\face_guard_bundle.exe`

### The face guard does not start

Possible causes:

- missing Python dependencies for source-based runs
- missing embedded bundle for bundle-based runs
- webcam access failure
- missing `models\face_landmarker.task`

## 18. Safe usage recommendations

- Choose a strong global master password
- Back up keys and vaults before changing systems
- Use full-disk encryption on the host machine
- Only add apps to the monitored kill list if you are comfortable losing unsaved work
- Remember that the app is local-first and does not provide cloud sync or remote recovery
