# internal/storage/

Low-level, OS-aware file I/O and keyring integration. Used by `core/storage`
and the UI layer. Not intended to be imported from outside this module.

| File | Description |
|---|---|
| `storage.go` | `WriteVaultFile` / `ReadVaultFile`: atomic write and read of vault files from the user's config directory. Manages the vault directory path. |
| `permissions.go` | Unix permission hardening: sets `0600` on vault and key files to prevent other users from reading them. |
| `permissions_windows.go` | Windows stub for permission hardening (no-op; DPAPI handles access control at the OS level on Windows). |
| `keyring.go` | OS keyring helpers: store and retrieve the master password via the system keyring. Falls back to in-memory storage when the keyring is unavailable. Wraps DPAPI on Windows. |
| `dpapi_windows.go` | Windows DPAPI encrypt/decrypt helpers used by `keyring.go` to protect keyring entries at rest via the current user's Windows session key. |
