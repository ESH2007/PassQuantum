# core/crypto/

Cryptographic building blocks used throughout the vault pipeline and app-security layer.
No UI dependency. All functions are pure or operate only on in-memory data.

| File | Description |
|---|---|
| `kdf.go` | Argon2id-based key derivation. Derives domain-separated encryption and verification keys from a master password and per-vault salt. |
| `aes.go` | AES-256-GCM helpers: encrypt and decrypt with authentication. Used by both the legacy and PQ vault pipelines. |
| `kyber.go` | Kyber768 keypair generation, encapsulation, and decapsulation wrappers (via `cloudflare/circl`). Provides per-entry key exchange. |
| `vault.go` | Legacy vault encryption pipeline: AES-GCM + HMAC-SHA256. Kept for backward-compatible decryption of older vault files. |
| `vault_pq.go` | Post-quantum vault encryption pipeline: Kyber768 encapsulation + Dilithium signing + AES-256-GCM. All new vault entries use this path. |
| `app_security.go` | App-level master-password security profile: creates and verifies the Argon2id-derived verifier that is bound to the private-key fingerprint. |
| `app_security_test.go` | Unit tests for `CreateAppSecurityProfile` and `VerifyAppSecurityProfile` (correct password, wrong password, mismatched keypair). |
