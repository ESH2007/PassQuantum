# ZimPass (PassQuantum_C)

This folder is initialized as the C implementation workspace derived from:
- `new-passquantum/GO_APP_SPECIFICATION.md`

## Purpose
ZimPass is the C-language counterpart of PassQuantum Go, preserving the same architecture and security boundaries:
- `src/crypto`: KDF, AES-GCM, KEM integration, app security profile logic.
- `src/storage`: vault file persistence and metadata persistence.
- `src/model`: password entry data structures and binary serialization.
- `src/biometric`: face template extraction, matching, camera/runtime adapters.
- `src/ui`: desktop UI and workflow orchestration.
- `apps`: command/demo binaries.
- `tests`: unit/integration tests.

## Folder Layout
- `include/` public headers
- `src/` implementation by module
- `apps/` executable entrypoints
- `tests/` verification suite
- `docs/` implementation and migration notes

## Next Implementation Steps
1. Define C data contracts that match Go file formats (`pqdb`, `pqmeta`).
2. Implement crypto module with strict compatibility tests against Go outputs.
3. Implement storage parser/serializer and vault lifecycle operations.
4. Implement app-state and access-control state machine.
5. Integrate UI and optional biometric runtime under feature flags.
