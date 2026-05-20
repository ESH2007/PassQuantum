# core/storage/

Vault and security-metadata persistence. Sits above `internal/storage` (which
handles raw file I/O and OS permissions) and uses `core/crypto` for all
encryption/decryption operations.

| File | Description |
|---|---|
| `storage.go` | `ReadVault` and `WriteVault`: top-level entry points for loading and saving a vault file. Handles automatic format migration from legacy formats. |
| `vault_format.go` | `PQV2` container format: plaintext header with magic bytes, version, and salt wrapping the encrypted payload. Defines format constants and encode/decode helpers. |
| `security_metadata.go` | `SaveAppSecurityProfile` / `LoadAppSecurityProfile`: persists and loads the app-level master-password verifier to/from `app-security.pqmeta`. |
| `security_metadata_test.go` | Tests for security-profile round-trip (save → load, verify fields). |
| `vault_migration_test.go` | Tests for vault write/read round-trip, typed-entry round-trip (Password, Note, Card), re-encryption/key-rotation, and legacy format rejection. |
