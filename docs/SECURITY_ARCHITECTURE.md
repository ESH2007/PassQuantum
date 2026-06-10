# PassQuantum Security Architecture

This document describes the current security design implemented in the
repository.

## 1. Security goals

PassQuantum currently tries to protect:

- the application's global unlock state
- the confidentiality and integrity of each vault file
- the confidentiality of each stored item payload (passwords, notes, cards, TOTP
  secrets, and encrypted files)
- local use of the unlocked app through optional face monitoring

It does **not** attempt to solve:

- malware on the host
- keyloggers or screen capture on an already-compromised machine
- general network security beyond the loopback-only browser bridge described in
  §10 (the product is otherwise local-first and offline)
- secure multi-device sync

## 2. Primary security assets

| Asset | Protection |
| --- | --- |
| Global master password | never stored directly |
| `private.key` | local file, fingerprint-bound to app profile |
| `app-security.pqmeta` | verifier profile, not plaintext password |
| `vaults/*.pqdb` | authenticated encryption at rest |
| Vault item payloads | Kyber-wrapped shared secret + AES-GCM |
| Encrypted file blobs | Kyber-wrapped shared secret + streamed AES-GCM (`core/filevault`) |
| Face profile | stored as local NumPy encodings in `face_data.npy` |
| Browser autofill access | loopback-only HTTP server gated by a pairing token (§10) |

## 3. App-level security profile

### 3.1 Purpose

The app does not unlock directly against a vault file first. Instead, it verifies a global master-password profile stored in:

```text
app-security.pqmeta
```

This is the top-level gate for entering the application.

### 3.2 Stored fields

`core/crypto/app_security.go` defines:

- `FormatVersion`
- `PrivateKeyFingerprint`
- `KDFParams`
- `Verifier`

### 3.3 How verification works

1. Marshal the current Kyber private key
2. Hash it with SHA-256 to produce a fingerprint
3. Derive keys from the provided master password with Argon2id
4. Compute the verifier from:
   - label `passquantum-app-verifier-v1`
   - private-key fingerprint
   - derived verification key
5. Compare against stored verifier

Result:

- the correct password is required
- the stored profile must also match the current `private.key`

If the fingerprint does not match, startup falls back to setup mode and shows a warning.

## 4. Key derivation

Two derivation profiles exist, one per vault format:

- **PQ vaults (`core/crypto/vault_pq.go`, current):** Argon2id with a 32-byte
  salt, 64 MB memory, 2 iterations, parallelism 4 → a 32-byte master key. HKDF
  then expands that master key into domain-separated seeds for the Kyber768 and
  Dilithium3 keys, keeping each purpose's key independent.
- **Legacy vaults (`core/crypto/kdf.go`):** Argon2id with a 16-byte salt, 64 MB,
  1 iteration, parallelism 4 → 64 bytes split by domain separation into an
  encryption key and a verification key.

Domain separation in both cases prevents key reuse across purposes.

## 5. Vault security

### 5.1 One vault, one salt

Every vault stores its own KDF parameters. Even if the same global master password is used, vault keys differ because each vault has its own salt.

### 5.2 Vault file protection

New vaults use the **PQ pipeline** (`core/crypto/vault_pq.go`, magic `PQVT`):

1. Serialize all entries into plaintext
2. Encrypt the plaintext with AES-256-GCM under the derived AES key
3. Sign the header + ciphertext with Dilithium3 (post-quantum signature)

Older **legacy vaults** (`core/crypto/vault.go`) are still readable: AES-256-GCM
payload plus an HMAC-SHA256 over version + serialized KDF params + encrypted data.
`core/storage` detects the format and migrates legacy vaults on write.

Both formats give confidentiality through AES-GCM and whole-file integrity
(Dilithium3 signature or HMAC), so tampered or wrongly decrypted vaults are
rejected.

### 5.3 On-disk vault structure

```text
PQ format (PQVT):           Legacy format:
Magic "PQVT"                Version
Version                     KDF params length
KDF params (salt, …)        KDF params
Dilithium3 signature        HMAC
Encrypted data              Encrypted data length
  (nonce + AES-GCM ct)      Encrypted data (nonce + ciphertext)
```

## 6. Entry-level secret protection

PassQuantum also protects each item payload separately.

For each entry:

1. `Kyber768.Encapsulate` creates:
   - Kyber ciphertext
   - shared secret
2. AES-256-GCM encrypts the actual item payload with that shared secret
3. The entry stores:
   - entry metadata
   - Kyber ciphertext
   - AES nonce
   - AES ciphertext

This means the outer vault encryption and the inner item encryption are separate layers.

## 7. Typed vault payloads

`core/storage/vault_format.go` and `core/model/vault_entry.go` currently support:

- Password items
- Note items
- Card items
- TOTP / 2FA items (serialized `otpauth` parameters; live codes are never stored)
- Encrypted file items (the file blob is encrypted separately via `core/filevault`)

The vault plaintext starts with a magic header followed by typed entries. Legacy
payload parsing is still supported for older data.

## 8. Master-password rotation

Changing the master password is handled by `ChangeMasterPassword` in
`app/access.go`.

Security properties of that flow:

1. The current password is re-verified first
2. Every vault is decrypted with current keys
3. Every vault is re-encrypted with new salt-derived keys
4. Temporary files are staged before replacement
5. The app-security metadata is staged and swapped as part of the same process

This reduces the chance of partially migrated state.

## 9. Face-guard security layer

The face guard is not the primary unlock mechanism. It is a **post-unlock runtime control**.

### 9.1 What it does

- trains a local face template from webcam frames
- stores encodings in `face_data.npy`
- requires a liveness blink during training
- requires a recognized live face before entering monitor mode
- sends `FACE_LOST` after 5 seconds without a recognized face

### 9.2 What the app does on face loss

When Go receives `FACE_LOST`:

- it clears sensitive app state
- returns the UI to the login flow
- force-kills any user-selected monitored apps

The monitored app list is stored in Fyne preferences as JSON.

### 9.3 Security boundary

This feature improves local shoulder-surfing and walk-away resistance, but it does not replace:

- the master password
- OS session locking
- full-disk encryption

## 10. Browser autofill bridge

`internal/browser` runs an HTTP server bound to `127.0.0.1:8765` so the browser
extension can autofill and save credentials. Its security properties:

- **Loopback only.** Requests whose host is not localhost are rejected, so the
  server is not reachable from the network.
- **Pairing required.** The desktop app shows a short-lived token; the extension
  must submit it to pair, and authenticates with the paired secret thereafter.
  Unpaired requests are refused.
- **Unlock-gated.** Credentials are only ever served for an already-unlocked
  vault; locking the app cuts off the extension.
- **Rate limited.** A dependency-free token-bucket limiter throttles requests.
- **Per-site opt-out.** A persisted "never save" list suppresses save prompts for
  chosen domains.

This still widens the local attack surface: any process able to bind/connect on
loopback and complete pairing could query the unlocked vault. It is a
convenience feature, not a hardened boundary.

## 11. Local file expectations

| File | Meaning |
| --- | --- |
| `private.key` | must remain private |
| `public.key` | public half of the Kyber pair |
| `app-security.pqmeta` | app-level verifier metadata |
| `vaults/*.pqdb` | encrypted vault files |
| `face_data.npy` | local face profile encodings |

Current write modes in code:

- vault files: `0600`
- `private.key`: `0600`
- `public.key`: `0644`
- app-security metadata: `0600`

## 12. Threat model summary

| Threat | Current mitigation |
| --- | --- |
| Offline guessing of vault password material | Argon2id |
| Vault tampering | HMAC-SHA256 |
| Wrong-password unlock of vault file | KDF mismatch + HMAC/AES failure |
| Reuse of master-password output for multiple purposes | domain-separated derived keys |
| Local walk-away exposure | face guard + app lock |
| Access with wrong private key | fingerprint-bound app profile |

## 13. Current limitations

These are important to understand when evaluating the present implementation:

- An unlocked session still keeps sensitive material in process memory until lock/exit.
- TOTP code copies are auto-cleared from the clipboard after a delay; other copy paths may not be.
- Backup/export/restore actions shown in the Settings → Vaults UI are mostly placeholders right now (import, however, is fully implemented).
- A compromised host can still capture keystrokes, screenshots, or decrypted content.
- The face-guard process is local and practical, but it is not a hardened biometric enclave.
- The browser bridge widens the local attack surface (see §10).

## 14. Operational recommendations

- Use a strong global master password
- Back up `private.key`, `public.key`, `app-security.pqmeta`, and `vaults/`
- Protect the host with full-disk encryption
- Treat `private.key` as highly sensitive
- Enable monitored-app kill behavior only for apps you can safely lose unsaved work in
- Use the Windows self-contained build process when distributing to systems without Python
