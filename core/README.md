# core/

Core business logic for PassQuantum. Split into six subpackages — no UI
dependency, no external I/O beyond what the storage layer requires.

| Package | Description |
|---|---|
| [`crypto/`](crypto/README.md) | Cryptographic primitives: KDF, AES-GCM, Kyber768/Dilithium, app-security profile, vault encryption pipelines. |
| [`model/`](model/README.md) | Typed vault entry model and binary serialization (v1/v2 format, legacy decode). |
| [`storage/`](storage/README.md) | Vault file persistence, format versioning, security-metadata persistence, and key-rotation helpers. |
| [`filevault/`](filevault/README.md) | Encrypted per-file storage: store, retrieve, and open arbitrary files protected with Kyber768 + AES-256-GCM, tracked by a manifest. |
| [`migration/`](migration/README.md) | Import framework: format auto-detection and parsers for 11 password managers, normalized into vault entries. |
| [`totp/`](totp/README.md) | TOTP/2FA code generation, `otpauth://` URI parsing, and QR helpers (built on `pquerna/otp`). |
