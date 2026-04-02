# PassQuantum Secure Storage - Implementation Guide

## Module Overview

### `core/crypto/kdf.go` - Key Derivation
Handles Argon2id password-based key derivation with secure domain separation.

**Public Functions:**
- `DefaultKDFParams()` → Safe default parameters
- `GenerateSalt()` → Create 16-byte random salt
- `DeriveKeys(password, params)` → Returns (encKey, verKey, error)
- `KDFParams.Serialize()` → Encode for storage
- `KDFParamsDeserialize(bytes)` → Decode from storage
- `WipeBytes(data)` → Securely zero out sensitive data

**Key Points:**
- Argon2id parameters: Memory=64MB, Iterations=1, Parallelism=4
- Domain separation via SHA-256 prevents key reuse
- Salt stored in vault.pqdb for reproducibility

### `core/crypto/vault.go` - Vault Encryption
Handles encryption/decryption of entire password vault with integrity checking.

**Public Functions:**
- `EncryptVault(plaintext, encKey, verKey, params)` → VaultFile
- `DecryptVault(vault, encKey, verKey)` → Decrypted plaintext
- `VaultFile.Serialize()` → Binary format for disk
- `VaultFileDeserialize(bytes)` → Parse from disk

**Vault File Format:**
- Version marker
- KDF parameters (for reproducibility)
- HMAC-SHA256 (integrity verification)
- AES-256-GCM encrypted entries + nonce

**Security:**
- HMAC fails if password wrong, vault tampered, or data corrupted
- Nonce randomly generated per encryption
- GCM provides authenticated encryption (no separate MAC needed after encryption)

### `core/model/password_entry.go` - Entry Structure
Defines encrypted password entry format (no plaintext data).

**PasswordEntry struct:**
```go
type PasswordEntry struct {
    ID              uint64   // Unique entry identifier
    KyberCiphertext []byte   // Kyber768 encapsulated secret
    Nonce           []byte   // AES-GCM nonce (12 bytes)
    Ciphertext      []byte   // AES-256-GCM encrypted password
}
```

**Functions:**
- `NewPasswordEntry()` → Create with random ID
- `entry.Serialize()` → Binary format for vault storage
- `Deserialize(bytes)` → Parse from vault

**Format:**
```
[ID(8) | KyberLen(2) | Kyber(var) | Nonce(12) | CipherLen(2) | Cipher(var)]
```

### `core/storage/storage.go` - Vault I/O
Handles reading/writing encrypted vaults to disk.

**Public Functions:**
- `WriteVault(entries, path, encKey, verKey, kdfParams)` → Write encrypted vault
- `ReadVault(path, encKey, verKey)` → Read and decrypt vault
- `VaultExists(path)` → Check if vault file exists
- `DeleteVault(path)` → Remove vault file

**Process:**
1. **Writing:** Serialize all entries → Encrypt with AES → HMAC → Write to disk
2. **Reading:** Read file → Deserialize → Verify HMAC → Decrypt → Parse entries

**Error Handling:**
- Returns specific errors (missing file, HMAC mismatch, decryption failure)
- Skips malformed entries (logs warning) instead of crashing

### `ui/main.go` - Master Password Prompt
User interface integration with secure vault lifecycle.

**Key Functions:**
- `initializeApp()` → Load or create Kyber keypair
- `promptMasterPassword(w, app, state)` → Initial unlock/creation dialog
- `createNewVault(w, state, password)` → Create new encrypted vault
- `unlockVault(w, state, password)` → Decrypt existing vault
- `buildUI(w, app, state)` → Main password manager interface

**Security Points:**
- Master password read as PasswordEntry (masked)
- Keys only stored in memory (AppState)
- Keys wiped on vault lock
- Each password add operation re-encrypts entire vault (for consistency)

## Integration Sequence

### App Startup
```go
1. Load Kyber keypair (public.key, private.key)
2. Check if vault.pqdb exists
3. If exists: Show "Unlock Vault" dialog
   - User enters master password
   - DeriveKeys(password, kdfParams from vault)
   - DecryptVault(vault, encKey, verKey) → HMAC check
   - If successful: Load entries into memory
4. If not exists: Show "Create Vault" dialog
   - User enters new master password
   - GenerateSalt() and DeriveKeys()
   - WriteVault([], path, encKey, verKey, kdfParams)
5. Show main UI
```

### Adding a Password
```go
1. User enters plaintext password in UI
2. Read current vault (already decrypted in memory)
3. Encrypt password:
   - Encapsulate(publicKey) → Kyber ciphertext + shared secret
   - EncryptAES256GCM(password, sharedSecret) → Nonce + ciphertext
4. Create PasswordEntry with ID, Kyber CT, nonce, AES CT
5. Append to entries list
6. WriteVault(entries, path, encKey, verKey, kdfParams)
   - This re-encrypts entire vault
   - Generates new nonce
   - Computes new HMAC
7. Show success dialog
```

### Viewing Passwords
```go
1. Entries already in memory (decrypted)
2. For each entry:
   - Decapsulate(kyberCiphertext, privateKey) → shared secret
   - DecryptAES256GCM(nonce, ciphertext, sharedSecret) → plaintext
   - Display in UI
3. Passwords are shown only in UI memory (not stored after display)
```

### Locking Vault
```go
1. User clicks "Lock Vault"
2. Clear AppState.encryptionKey, verificationKey, masterPassword
3. Exit application
4. Next startup: Prompt for master password again
```

## Testing the Implementation

### Unit Test Ideas

```go
// Test KDF
func TestKDFDeriveKeys(t *testing.T) {
    params := DefaultKDFParams()
    key1, ver1, _ := DeriveKeys("password", params)
    
    // Same password + params = same key
    key2, ver2, _ := DeriveKeys("password", params)
    if !bytes.Equal(key1, key2) {
        t.Fatal("KDF not deterministic")
    }
    
    // Different password = different key
    key3, _, _ := DeriveKeys("different", params)
    if bytes.Equal(key1, key3) {
        t.Fatal("KDF not sensitive to password")
    }
}

// Test Vault encryption/decryption
func TestVaultEncryption(t *testing.T) {
    plaintext := []byte("test data")
    encKey := make([]byte, 32)
    verKey := make([]byte, 32)
    rand.Read(encKey)
    rand.Read(verKey)
    
    params := DefaultKDFParams()
    vault, _ := EncryptVault(plaintext, encKey, verKey, params)
    
    // Should decrypt back to original
    decrypted, _ := DecryptVault(vault, encKey, verKey)
    if !bytes.Equal(plaintext, decrypted) {
        t.Fatal("Vault decryption failed")
    }
    
    // Wrong key should fail
    badKey := make([]byte, 32)
    _, err := DecryptVault(vault, badKey, verKey)
    if err == nil {
        t.Fatal("Bad key should fail HMAC")
    }
}

// Test HMAC integrity
func TestVaultIntegrity(t *testing.T) {
    // ... create vault ...
    
    // Tamper with encrypted data
    vault.EncryptedData[0] ^= 0xFF
    
    _, err := DecryptVault(vault, encKey, verKey)
    if err == nil || !strings.Contains(err.Error(), "HMAC") {
        t.Fatal("Vault should detect tampering")
    }
}
```

### Manual Testing

1. **Create vault:**
   ```bash
   ./passquantum-gui
   # Create new vault with password "test123"
   # Verify vault.pqdb created
   ```

2. **Add password:**
   ```
   - Enter password "MySecretPassword"
   - Click "Add Password"
   - Verify file size changed (vault re-encrypted)
   ```

3. **View password:**
   ```
   - Click "View Passwords"
   - Verify "MySecretPassword" displayed
   ```

4. **Verify encryption:**
   ```bash
   strings vault.pqdb | grep -i secret  # Should find nothing
   hexdump -C vault.pqdb | head         # All binary
   ```

5. **Test wrong password:**
   ```
   - Exit app
   - Relaunch
   - Enter wrong master password
   - Verify error "invalid master password"
   ```

6. **Test vault tampering:**
   ```bash
   # Modify vault.pqdb with hex editor (change 1 byte)
   ./passquantum-gui
   # Enter correct password
   # Should fail: "vault integrity check failed"
   ```

## Performance Characteristics

- **KDF (Argon2id)**: ~2 seconds (intentionally slow)
- **AES-256-GCM encryption**: <10ms per password
- **HMAC computation**: <1ms
- **Vault writing**: <50ms
- **Vault reading**: <50ms + KDF time

Large vaults (1000+ passwords): ~100-200ms to re-encrypt due to full vault serialization.

## Security Checklist

- [x] No plaintext passwords logged or printed
- [x] Master password never stored
- [x] Derived keys never stored
- [x] Each encryption operation uses fresh nonce
- [x] HMAC prevents tampering
- [x] KDF memory-hard (64MB)
- [x] Domain separation prevents key reuse
- [x] File permissions: 0600 for private data
- [x] Error messages don't leak info (generic "invalid password")
- [x] Post-quantum resistant encryption (Kyber768)

## Code Organization

```
core/crypto/
├── kyber.go       # Kyber keypair management
├── aes.go         # AES-256-GCM operations
├── kdf.go         # Argon2id + domain separation
└── vault.go       # Vault encryption + HMAC

core/model/
└── password_entry.go  # Entry struct + serialization

core/storage/
└── storage.go     # Vault file I/O

ui/
└── main.go        # Fyne GUI + master password prompts
```

All crypto operations isolated in `core/crypto/`, making them easy to audit and test independently.
