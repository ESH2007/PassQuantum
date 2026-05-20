# PassQuantum User Experience

This document describes the current user-facing behavior of the app as implemented in `ui/`, `strength/`, and the face-guard bridge.

## 1. First impression

PassQuantum opens as a desktop Fyne application with:

- a neon/dark visual style
- a particle animated background
- custom cards, dividers, and buttons
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

After a vault is opened, the app uses a left sidebar with these destinations:

- Vaults
- Passwords
- Generate
- Check Password
- Settings
- Lock & Exit

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

## 6. Passwords view

The main passwords screen is really an **Add Vault Item** screen.

### 6.1 Supported item types

Users can add:

- `Password`
- `Cyphered Note`
- `Card`

The form changes dynamically based on the selected item type.

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

## 7. Password generator experience

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

## 8. Password checker experience

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

## 9. Settings experience

The settings area uses four custom tab-like sections:

- Security
- Vaults
- Visuals
- About

### 9.1 Security section

Currently implemented:

- change master password
- monitored-app selection from the current running-process list
- warnings about force-kill behavior
- refresh of the process list

### 9.2 Vaults section

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

### 9.3 Visuals section

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

### 9.4 About section

The About screen currently shows:

- static app name
- static version text `Version 1.0.0`
- feature bullets
- docs button
- updates button

Current implementation status:

- Docs button -> informational dialog
- Updates button -> informational dialog

## 10. Locking and exit behavior

The sidebar includes `Lock & Exit`.

Current behavior:

- clears sensitive state
- quits the app

Face-loss locking returns the app to the unlock flow without quitting.

## 11. What is real vs. what is aspirational

### Real today

- global master-password gate
- multi-vault workflow
- typed vault items
- password generator
- password checker
- theme/palette/icon customization
- face training and face monitoring
- monitored-app kill list
- master-password rotation with vault re-encryption

### Present in UI but mostly placeholder

- compact vault
- export/import
- backup/restore
- docs link behavior
- update check behavior

## 12. Files a user will notice

Depending on the workflow, users may notice:

- `public.key`
- `private.key`
- `app-security.pqmeta`
- `vaults/*.pqdb`
- `face_data.npy`

For packaged builds they may also encounter:

- `PassQuantum.exe`
- `face_guard_bundle.exe` during packaging workflows

## 13. Practical UX summary

The current product experience is:

- unlock once with a global master password
- choose a vault
- manage encrypted items from a sidebar shell
- use generator/checker tools in the same app
- optionally rely on face monitoring to auto-lock when away
- personalize the look and icon locally

That is the current implemented experience; the docs should not present backup/export/import or online support flows as fully shipped features.
