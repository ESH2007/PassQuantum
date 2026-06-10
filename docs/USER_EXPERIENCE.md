# PassQuantum User Experience

This document describes the current user-facing behavior of the app as implemented in `ui/`, `strength/`, and the face-guard bridge.

## 1. First impression

PassQuantum opens as a desktop Fyne application with:

- a dark visual style with a teal accent (`#2dd4bf` on a near-black blue-grey background)
- a subtle animated particle background
- custom cards, dividers, pills, and buttons from the `theme` design system
- a single-window flow that changes content rather than opening many independent windows

The app title is:

```text
PassQuantum - Post-Quantum Safe Password Manager
```

## 2. Startup experience

### 2.1 First-time setup

If no app-security profile exists, the user sees a **Create Master Password** screen.

That flow:

1. asks for a new global master password
2. asks for confirmation
3. creates `app-security.pqmeta`
4. creates a default vault if no vaults exist
5. optionally opens face registration if the face guard is available and no face profile is detected
6. continues to vault selection

### 2.2 Returning user

If the app-security profile exists and matches the current private key, the user sees an **Unlock PassQuantum** screen.

Unlocking:

1. verifies the global master password
2. starts face monitoring if the face guard is running
3. opens the vault-selection view

### 2.3 Private-key mismatch case

If the stored profile belongs to a different `private.key`, the app returns to setup mode and shows a warning that the existing vaults may need manual migration.

## 3. Face registration and monitoring

### 3.1 Registration experience

The in-app training screen currently includes:

- title: `FACIAL REGISTRATION`
- live camera preview
- progress bar
- status text
- `START REGISTRATION` button

Training behavior:

- captures many samples from the webcam
- requires at least one blink
- updates progress inside the Fyne window
- automatically starts monitoring after completion

### 3.2 Monitoring experience

After unlock, the face guard monitors continuously in the background.

If the recognized face disappears for 5 seconds:

- the app locks
- sensitive session state is cleared
- any monitored apps selected by the user are force-closed

This is a strong UX behavior and should be treated carefully because monitored apps are killed without a save prompt.

## 4. Main navigation model

After a vault is opened, the app uses a grouped left sidebar:

- **Vault:** Vaults, Add item, Items, Authenticator (TOTP), Files, Import
- **Tools:** Generate, Analyze
- **System:** Settings, Collapse/Expand sidebar, Lock vault

The content area on the right changes in place.

## 5. Vault-selection experience

The vaults view presents:

- a list of existing vault cards
- vault names
- vault file location hints
- `OPEN` and `DELETE` actions
- `+ CREATE VAULT` action

Creating a vault:

- asks only for the vault name
- uses the already unlocked global master password automatically

Deleting a vault:

- requires confirmation
- removes the `.pqdb` file permanently

## 6. Add item view

The `Add item` screen is the main vault-item entry form.

### 6.1 Supported item types

Users can add:

- `Password`
- `Cyphered Note`
- `Card`
- `TOTP`

The form changes dynamically based on the selected item type. (Encrypted files
are added from the dedicated Files view rather than this form.)

### 6.2 Password item UX

For password items the user sees:

- service name
- username/email
- password field
- live strength analysis

### 6.3 Note item UX

For note items the user sees:

- note title
- multi-line note content

Internally the note is stored as encrypted JSON payload.

### 6.4 Card item UX

For card items the user sees:

- card type
- card nickname
- card holder
- card number
- expiry
- CVV

Internally the card is also stored as encrypted JSON payload.

### 6.5 Save flow

On save:

1. the current vault is read and decrypted
2. the item payload is encrypted
3. the entry is appended
4. the vault is rewritten
5. the form is cleared
6. a success dialog is shown

### 6.6 View-all flow

`VIEW ALL ITEMS` opens the vault-item list.

That list supports:

- password show/hide
- password copy
- password edit
- password delete
- note view/copy/delete
- card show/copy/delete

## 7. Authenticator (TOTP) view

The Authenticator view manages 2FA codes. Users can add a TOTP entry by:

- scanning/importing a QR code image,
- pasting an `otpauth://totp/...` URI, or
- importing a Google Authenticator export (`otpauth-migration://...`).

Each stored entry shows a live 6/8-digit code with a countdown bar that refreshes
as the period rolls over. Copying a code places it on the clipboard and clears it
automatically after a short delay.

## 8. Files view

The Files view stores arbitrary files inside the current vault, each encrypted
individually. Users can:

- add a file (it is encrypted and copied into the vault's file store),
- open a file (decrypted to a temporary location for viewing, then securely
  deleted afterward),
- retrieve/save a decrypted copy to a chosen path, and
- delete a stored file.

A file-type icon and size are shown per item.

## 9. Import view

The Import view is a wizard that migrates data from other password managers:

1. **Pick** — choose the source manager (or let auto-detection pick from the file).
2. **Parse** — the export is read and parsed into a normalized preview.
3. **Preview** — review the entries that will be imported.
4. **Import** — entries are encrypted into the current vault, de-duplicating
   against existing items.

Supported sources include 1Password, Bitwarden, KeePass/KeePassXC, LastPass,
Dashlane, NordPass, Proton Pass, Kaspersky, Chromium browsers, Firefox, and a
generic CSV fallback.

## 10. Browser extension pairing

The app can pair with the PassQuantum browser extension for autofill. The pairing
dialog displays a short-lived token; the user enters it in the extension's popup.
Once paired and with a vault unlocked, the extension autofills matching logins and
offers to save new ones over a loopback-only connection. Per-site "never save"
choices are remembered.

## 11. Password generator experience

The generator screen lets the user configure:

- length
- uppercase
- lowercase
- numbers
- special characters
- exclude ambiguous characters

Actions:

- `GENERATE`
- `COPY`
- `SAVE TO VAULT`

The screen is integrated into the main app instead of being a separate tool window.

## 12. Password checker experience

The checker screen gives:

- live score label
- strength bar
- crack-time estimate
- issue list

The analysis considers:

- repeated characters
- keyboard walks
- leet patterns
- date-like patterns
- common names and passwords
- missing character classes
- similarity to passwords already stored in the vault

There is also an easter-egg mode when the input includes `neal.fun`.

## 13. Settings experience

The settings area uses four custom tab-like sections:

- Security
- Vaults
- Visuals
- About

### 13.1 Security section

Currently implemented:

- change master password
- monitored-app selection from the current running-process list
- warnings about force-kill behavior
- refresh of the process list

### 13.2 Vaults section

Currently shown:

- current vault label
- total vault count
- maintenance and backup buttons

Current implementation status:

- `COMPACT VAULT` -> informational dialog
- `EXPORT VAULT` -> informational dialog
- `IMPORT VAULT` -> informational dialog
- `BACKUP NOW` -> informational dialog
- `RESTORE` -> confirmation + informational dialog

So this section is present in the UX, but most of its actions are placeholders today.
Note that the `IMPORT` button here is **not** the real importer — actual import
from other password managers is the dedicated **Import** sidebar view (§9), which
is fully implemented.

### 13.3 Visuals section

This section is much more complete.

Implemented features:

- theme selector UI
- font-size selector UI
- delete-confirmation toggle UI
- OS-native image pickers for palette/icon selection
- manual palette personalization dialog
- upload image to analyze and extract top colors
- reset palette to defaults
- change app icon from image file
- reset app icon

### 13.4 About section

The About screen currently shows:

- static app name
- static version text `Version 1.0.0`
- feature bullets
- docs button
- updates button

Current implementation status:

- Docs button -> informational dialog
- Updates button -> informational dialog

## 14. Locking and exit behavior

The sidebar's System group includes `Lock vault`.

Current behavior:

- clears sensitive state
- quits the app

Face-loss locking returns the app to the unlock flow without quitting.

## 15. What is real vs. what is aspirational

### Real today

- global master-password gate
- multi-vault workflow
- typed vault items (password, note, card, TOTP, file)
- TOTP / authenticator codes (manual, QR, Google Authenticator import)
- encrypted file storage
- import from 11 other password managers (the Import view)
- browser-extension autofill with pairing
- password generator
- password strength analyzer
- theme/palette/icon customization
- face training and face monitoring
- monitored-app kill list
- master-password rotation with vault re-encryption

### Present in UI but mostly placeholder

- compact vault
- the Settings → Vaults export/import/backup/restore buttons (the real importer is the Import view above)
- docs link behavior
- update check behavior

## 16. Files a user will notice

Depending on the workflow, users may notice:

- `public.key`
- `private.key`
- `app-security.pqmeta`
- `vaults/*.pqdb`
- `face_data.npy`

For packaged builds they may also encounter:

- `PassQuantum.exe`
- `face_guard_bundle.exe` during packaging workflows

## 17. Practical UX summary

The current product experience is:

- unlock once with a global master password
- choose a vault
- manage encrypted items (passwords, notes, cards, TOTP, files) from a sidebar shell
- import existing credentials from other password managers
- use generator/analyzer tools in the same app
- autofill in the browser via the paired extension
- optionally rely on face monitoring to auto-lock when away
- personalize the look and icon locally

That is the current implemented experience; the docs should not present the
Settings → Vaults backup/export/restore buttons or online support flows as fully
shipped features.
