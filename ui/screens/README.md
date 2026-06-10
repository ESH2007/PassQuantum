# ui/screens

All application screens and view-builders. Depends on `app/`, `bridge/`,
`theme/`, `ui/widgets`, `strength`, `palette`, and the `core/*` packages. The
main shell (`main_screen.go`) defines a `NavigationState` whose per-view builder
methods are split across several files in this package (TOTP, files, import live
in their own files but are methods on `NavigationState`).

| File | Description |
|---|---|
| `login.go` | `PromptMasterPassword` — the create / unlock master-password screen shown at startup. |
| `vault_selection.go` | `ShowVaultSelection` — list, create, open, and delete vaults. |
| `main_screen.go` | `ShowMainScreen` and `NavigationState` — the sidebar shell and navigation state machine that hosts every in-app view. |
| `passwords.go` | `ShowPasswordsView`, the vault-item cards (password/note/card), edit/delete dialogs, and the `GeneratePassword` engine + `PasswordGeneratorSettings`. |
| `generator_panel.go` | `newGeneratorControls` — the reusable password-generator control panel (length, character classes, etc.) embedded by other views. |
| `settings.go` | `ShowSettingsScreen` and the Security / Vaults / Visuals / About sections, including change-master-password, palette extraction, and reset actions. |
| `totp.go` | `NavigationState.createTOTPView`, the add-TOTP dialog (manual + QR import), live code cards, and the ticker that refreshes codes and clipboard auto-clear. |
| `filevault.go` | `NavigationState.createFilesView` — store/retrieve/open/delete encrypted files via `core/filevault`, with file-type icons and source-file cleanup. |
| `import_wizard.go` | `NavigationState.createImportView` — the multi-step import wizard (pick source → parse → preview → map) driving `core/migration`. |
| `training.go` | `ShowTrainingScreen` — the face-guard enrollment screen (camera preview, progress, blink prompt). |
| `pairing_dialog.go` | `ShowPairingDialog` — displays the browser-extension pairing token and pairing status. |
| `theme_picker.go` | `ShowThemePicker` and `RestoreThemeOnLaunch` — built-in theme selection and persistence. |
| `color_picker.go` | `ShowColorPersonalizationDialog` — manual HSV color personalization of the palette. |
| `strength_widgets.go` | `NewStrengthBar` / `BindStrengthBar` / `NewIssuesList` and the easter-egg panel — the live password-strength UI built on `strength`. |
| `app_icon.go` | `SetApplicationIcon` — sets and restores the app/window icon. |
