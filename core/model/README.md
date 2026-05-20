# core/model/

Typed data model for vault entries and their binary serialization.

| File | Description |
|---|---|
| `vault_entry.go` | Defines `VaultEntry` (the in-memory representation of a single stored credential) and the `EntryType` enum (`Password`, `Note`, `Card`). Implements v1/v2 binary serialization and legacy format decode so older vault files can be read transparently. |
