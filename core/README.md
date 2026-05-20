# core/

Core business logic for PassQuantum. Split into three subpackages — no UI
dependency, no external I/O beyond what the storage layer requires.

| Package | Description |
|---|---|
| [`crypto/`](crypto/README.md) | Cryptographic primitives: KDF, AES-GCM, Kyber768/Dilithium, app-security profile, vault encryption pipelines. |
| [`model/`](model/README.md) | Typed vault entry model and binary serialization (v1/v2 format, legacy decode). |
| [`storage/`](storage/README.md) | Vault file persistence, format versioning, security-metadata persistence, and key-rotation helpers. |
