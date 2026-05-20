# app

Package `app` manages the application lifecycle state, vault operations, and access control for PassQuantum.

## Contents

- **state.go** — `AppState` struct with all exported fields + helper methods
- **access.go** — startup access state resolution, master-password profile creation and rotation
- **helpers.go** — vault CRUD helpers (`CreateNewVault`, `UnlockVault`, `OpenVault`, `ReadVault`, `WriteVault`), crypto wrappers, and password validation

## AppState lifecycle

```
InitializeApp() → AppState{PublicKey, PrivateKey}
        ↓
UnlockVault(appState, password) → appState.IsUnlocked = true
        ↓
OpenVault(appState, vaultName, onSuccess) → appState.CurrentVault set, onSuccess() called
        ↓
ClearSensitiveState() → fields zeroed, IsUnlocked = false
```

## Callback pattern

`OpenVault` accepts an `onSuccess func()` callback instead of calling UI functions directly. This prevents an `app → ui/screens → app` import cycle. Callers in `ui/screens` pass:

```go
app.OpenVault(state, vaultName, func() { screens.ShowMainScreen(w, a, state) })
```
