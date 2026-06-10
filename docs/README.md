# docs/

Design and specification documents for PassQuantum. These are intended as
reference material for contributors and security reviewers.

| File | Description |
|---|---|
| [`ARCHITECTURE.md`](ARCHITECTURE.md) | Code and runtime architecture: module layout, data flow, vault pipeline, face-guard protocol, and build paths. Start here if you want to understand how the pieces fit together. |
| [`SECURITY_ARCHITECTURE.md`](SECURITY_ARCHITECTURE.md) | Full security model: threat boundaries, cryptographic layering (Argon2id, Kyber768, Dilithium, AES-256-GCM), key-derivation domains, and HMAC authentication. |
| [`USER_EXPERIENCE.md`](USER_EXPERIENCE.md) | Screen-by-screen product behavior specification: what each UI surface does, how the face-guard interacts with the app state, and which settings actions are currently implemented vs placeholders. |
| [`USER_GUIDE.md`](USER_GUIDE.md) | Practical end-user setup and usage instructions: first-run flow, vault creation, TOTP/file/import usage, browser-extension pairing, face-guard training, and build prerequisites. |

For lower-level, package-by-package implementation detail, see the `README.md` in
each Go package (`app/`, `bridge/`, `core/*`, `internal/*`, `ui/*`, `theme/`, …).
