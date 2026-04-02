# PassQuantum - Secure Storage Architecture

## Overview

PassQuantum now implements a cryptographically sound, production-ready storage system suitable for open-source applications. All passwords are encrypted at rest with strong authentication and integrity verification.

## Security Architecture

### 1. Master Password & Key Derivation

**Argon2id KDF** is used to derive encryption and verification keys from the user's master password.

**Parameters (defaults):**
- Memory: 64 MB (resistant to GPU/ASIC attacks)
- Iterations: 1 (optimized for interactive use)
- Parallelism: 4 (multi-core)
- Salt: 16 bytes (cryptographically random)

**Key derivation flow:**
```
Master Password
    ↓
[Argon2id with stored salt]
    ↓
Master Key (64 bytes)
    ↓
┌─────────────────────────────────┐
│   Domain Separation (SHA-256)   │
└─────────────────────────────────┘
    ↙                           ↘
Encryption Key (32 bytes)   Verification Key (32 bytes)
    ↓                           ↓
AES-256-GCM                 HMAC-SHA256
```

### 2. Encrypted Vault File Format

All passwords are stored in a single binary file: `vault.pqdb`

**File Structure:**
```
┌─────────────────────────────────────────────────┐
│ Version (1 byte)                                │
├─────────────────────────────────────────────────┤
│ KDF Params Length (1 byte)                      │
├─────────────────────────────────────────────────┤
│ KDF Parameters (26 bytes)                       │
│  - Salt (16 bytes)                              │
│  - Memory (4 bytes, big-endian)                 │
│  - Iterations (4 bytes, big-endian)             │
│  - Parallelism (1 byte)                         │
│  - Version (1 byte)                             │
├─────────────────────────────────────────────────┤
│ HMAC-SHA256 (32 bytes)                          │
│ [computed over Version + KDF Params + CipherText]│
├─────────────────────────────────────────────────┤
│ Encrypted Data Length (4 bytes, big-endian)    │
├─────────────────────────────────────────────────┤
│ Encrypted Data (variable)                       │
│ ┌───────────────────────────────────────────┐   │
│ │ Nonce (12 bytes, per encryption)          │   │
│ ├───────────────────────────────────────────┤   │
│ │ [Encrypted Entries...]                    │   │
│ │ All AES-256-GCM encrypted together        │   │
│ └───────────────────────────────────────────┘   │
└─────────────────────────────────────────────────┘
```

### 3. Password Entry Structure

Each encrypted password entry within the vault:

```
Entry Format:
┌─────────────────────────────┐
│ Entry ID (8 bytes)          │
├─────────────────────────────┤
│ Kyber Ciphertext Length (2) │
├─────────────────────────────┤
│ Kyber Ciphertext (~1088 B)  │
├─────────────────────────────┤
│ AES-GCM Nonce (12 bytes)    │
├─────────────────────────────┤
│ Ciphertext Length (2 bytes) │
├─────────────────────────────┤
│ AES-GCM Ciphertext (var)    │
└─────────────────────────────┘
```

**Storage characteristics:**
- Each entry has a unique 64-bit ID
- Hybrid encryption: Kyber768 + AES-256-GCM
- No plaintext data ever stored or transmitted
- Nonce is random for each encryption operation

### 4. Encryption & Integrity Verification

**On vault creation/update:**
1. Entries are serialized to binary
2. Nonce generated (12 random bytes)
3. AES-256-GCM encrypts all entries with the nonce and encryption key
4. HMAC-SHA256 computed over: Version + KDF Params + Encrypted Data
5. All components written to `vault.pqdb`

**On vault unlock/read:**
1. Vault file deserialized
2. HMAC verified against stored HMAC using verification key
3. If HMAC check fails: **operation aborted** (data tampered or wrong password)
4. AES-256-GCM decrypts using derived encryption key
5. Entries are parsed and made available

**Attack scenarios mitigated:**
- ✓ Offline brute-force: Argon2id makes each password guess expensive (64MB memory)
- ✓ Vault tampering: HMAC detects any modifications
- ✓ Wrong master password: Detected during decryption + HMAC verification
- ✓ Key derivation attacks: Domain separation prevents key reuse

### 5. Cryptographic Primitives

| Component | Algorithm | Notes |
|-----------|-----------|-------|
| KDF | Argon2id | OWASP recommended, GPU-resistant |
| Encryption | AES-256-GCM | NIST standard, authenticated encryption |
| Authentication | HMAC-SHA256 | Keyed MAC for integrity |
| Domain Sep. | SHA-256 | Counter-mode expansion |
| RNG | crypto/rand | Go stdlib, cryptographically secure |
| Hybrid KEM | Kyber768 + AES | Post-quantum resistant |

## Security Properties

### What's Protected
- ✓ All password data encrypted with AES-256-GCM
- ✓ Vault file integrity verified with HMAC
- ✓ Master password never stored
- ✓ Derived keys not stored
- ✓ Each operation uses fresh nonces
- ✓ Post-quantum resistant encryption (Kyber768)

### What's NOT Protected
- ✗ Master password (only user knows)
- ✗ Passwords in plaintext after decryption (shown in UI)
- ✗ Decrypted entries in memory during display
- ✗ File metadata (timestamps, size)

### Security Assumptions
1. **User controls master password**: Strong, random, not reused
2. **Kyber keypair is kept safe**: private.key not compromised
3. **Operating system is trusted**: No keyloggers, malware
4. **Go crypto/rand is sound**: System entropy available
5. **Argon2id parameters sufficient**: 64MB, 1 iteration reasonable for UI

## Usage Flow

### First Time Setup
```
[User launches app]
  ↓
[App checks if vault.pqdb exists]
  ↓
[No vault found]
  ↓
[Prompt user for NEW master password]
  ↓
[Generate KDF params (salt, Argon2id config)]
  ↓
[Derive encryption key + verification key]
  ↓
[Create empty encrypted vault]
  ↓
[Store KDF params and HMAC in vault.pqdb]
  ↓
[Show main password manager UI]
```

### Subsequent Launches
```
[User launches app]
  ↓
[App checks if vault.pqdb exists]
  ↓
[Vault found]
  ↓
[Prompt user to UNLOCK with master password]
  ↓
[Read vault file, extract KDF params]
  ↓
[Derive keys using provided password + stored KDF params]
  ↓
[Verify HMAC]
  ↓
[Decrypt vault contents]
  ↓
[If success: show main UI]
[If failure: show "Invalid master password" error]
```

### Adding a Password
```
[User enters password + clicks "Add"]
  ↓
[Read current vault]
  ↓
[Encrypt password with Kyber + AES]
  ↓
[Create new entry with unique ID]
  ↓
[Append entry to entries list]
  ↓
[Re-encrypt entire vault + HMAC]
  ↓
[Write vault.pqdb]
  ↓
[Show confirmation dialog]
```

## File Locations

```
PassQuantum/
├── public.key         # Kyber768 public key (can share)
├── private.key        # Kyber768 private key (MUST protect)
├── vault.pqdb         # Main encrypted password vault
└── passquantum-gui    # Executable
```

**Permissions:**
- `public.key`: 0644 (world-readable)
- `private.key`: 0600 (owner only)
- `vault.pqdb`: 0600 (owner only)

## Performance

- **Vault creation**: ~2 seconds (Argon2id with 64MB)
- **Vault unlock**: ~2 seconds (KDF verification)
- **Add password**: <100ms (encryption)
- **View passwords**: <100ms (decryption)
- **Startup**: ~1 second (keypair loading)

Argon2id memory requirement (64MB) ensures password guessing is slow even on powerful hardware.

## Testing & Verification

To manually verify security:

1. **Check no plaintext stored:**
   ```bash
   strings vault.pqdb | grep -i "password"  # Should find nothing
   hexdump -C vault.pqdb | head -20         # All binary, no ASCII passwords
   ```

2. **Test wrong password rejection:**
   - Launch app, enter wrong master password
   - Observe: "Invalid master password or vault corrupted"

3. **Test HMAC integrity:**
   - Modify vault.pqdb with any hex editor
   - Launch app with correct master password
   - Observe: "vault integrity check failed"

4. **Check KDF parameters:**
   - Each vault stores its own salt and KDF config
   - Different vaults won't use same key even with same password

## Open Source Considerations

This implementation is suitable for open-source distribution because:

1. **No secret keys in source**: All crypto keys derived from user password
2. **Auditable code**: All cryptographic operations in `core/crypto`
3. **Standard algorithms**: Argon2id, AES-256-GCM, HMAC-SHA256
4. **Well-tested libraries**: Uses Go stdlib + CIRCL for post-quantum
5. **Clear threat model**: Documented assumptions and protections
6. **Easy to review**: Modular design with clear separation of concerns

## Future Improvements

- [ ] Database backups with encryption
- [ ] Password expiration policies
- [ ] Multi-device sync (encrypted)
- [ ] Hardware security module (HSM) support
- [ ] Biometric unlock (using derived key, not password)
- [ ] 2FA for vault creation
- [ ] Secure password sharing (encrypted invite)
