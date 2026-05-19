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

From `new-passquantum/`:

```powershell
go build -o build\PassQuantum.exe .\ui
.\build\PassQuantum.exe
```

This is the easiest local build for development.

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

### Option C: Linux helper script

```bash
./build-native.sh
./build/linux/PassQuantum
```

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

Open the `Passwords` view and choose the item type:

- `Password`
- `Cyphered Note`
- `Card`

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

## 7. Viewing, editing, and deleting items

Click `VIEW ALL ITEMS` from the Passwords screen.

Available actions:

- **passwords**: show, copy, edit, delete
- **notes**: view, copy, delete
- **cards**: show, copy, delete

Deleting an item is permanent.

## 8. Using the password generator

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

## 9. Using the password checker

Open the `Check Password` view.

Type any password to see:

- score
- strength label
- crack-time estimate
- detected weaknesses

The checker also compares against passwords already stored in your current vault.

## 10. Changing the master password

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

## 11. Face-guard usage

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

## 12. Settings you can rely on today

### Fully or mostly implemented

- change master password
- monitored-app selection
- manual color personalization
- image-driven palette extraction
- app icon replacement
- palette reset

### Present but mostly placeholder

- compact vault
- export/import
- backup now
- restore
- docs button
- updates button

Treat the second group as UI placeholders, not full backup features.

## 13. Backup guidance

Back up these together:

- `public.key`
- `private.key`
- `app-security.pqmeta`
- `vaults/`
- `face_data.npy` if you want to preserve face training

If you move the app to another machine or folder, keep those files together.

## 14. Common problems

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

## 15. Safe usage recommendations

- Choose a strong global master password
- Back up keys and vaults before changing systems
- Use full-disk encryption on the host machine
- Only add apps to the monitored kill list if you are comfortable losing unsaved work
- Remember that the app is local-first and does not provide cloud sync or remote recovery
