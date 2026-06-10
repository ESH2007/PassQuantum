# ui/widgets

Small, screen-agnostic UI helpers shared across `ui/screens`. Depends only on
`theme/` and Fyne.

| File | Description |
|---|---|
| `dialogs.go` | Consistent app dialogs: `ShowAppInformation`, `ShowAppWarning`, `ShowAppError`, `ShowAppConfirm`, and `ShowAppConfirmWithRemember` (a confirm dialog with a "remember this choice" checkbox). |
| `nativefile.go` | OS-native file pickers: `PickImageFile`, `PickAnyFile`, and `PickSaveFile`, each taking `onPicked`/`onErr` callbacks. Wraps `ncruces/zenity` so the native dialog is used instead of the in-window Fyne picker. |
