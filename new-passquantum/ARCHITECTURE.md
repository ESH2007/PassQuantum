# PassQuantum - Technical Architecture

> Comprehensive technical documentation covering cryptographic design, system architecture, module APIs, and implementation details.

## 📋 Table of Contents

- [System Overview](#system-overview)
- [Architecture Patterns](#architecture-patterns)
- [Cryptographic Design](#cryptographic-design)
- [Module Reference](#module-reference)
- [Data Models](#data-models)
- [Storage Format](#storage-format)
- [Security Properties](#security-properties)
- [API Reference](#api-reference)
- [Performance Characteristics](#performance-characteristics)
- [Testing Guidelines](#testing-guidelines)

---

## System Overview

PassQuantum is built with a clean, modular architecture that separates concerns across four main packages:

```
┌─────────────────────────────────────────────────────────┐
│                      UI Layer (ui/)                     │
│  • Fyne desktop GUI                                     │
│  • User interaction handling                            │
│  • Screen navigation                                    │
└──────────────────┬──────────────────────────────────────┘
                   │
        ┌──────────┴──────────────────────┐
        │                                 │
        ▼                                 ▼
┌──────────────────┐            ┌──────────────────┐
│  Storage Layer   │            │  Crypto Layer    │
│   (storage/)     │            │   (crypto/)      │
│  • Vault I/O     │            │  • Kyber768      │
│  • File mgmt     │            │  • AES-256-GCM   │
│  • Serialization │            │  • Argon2id KDF  │
└──────────────────┘            └──────────────────┘
        │                                 │
        └──────────┬────────────────────┬─┘
                   │                    │
                   ▼                    ▼
            ┌──────────────────┐
            │   Model Layer    │
            │    (model/)      │
            │  • Data structs  │
            │  • Entry format  │
            └──────────────────┘
```

### Design Principles

1. **Separation of Concerns**: Each package has a single, well-defined responsibility
2. **No Global State**: All state passed as function parameters or in structs
3. **Testability**: Core packages have no UI dependencies
4. **Security by Default**: Sensible defaults, fail-secure error handling
5. **Modularity**: Easy to swap implementations or add features

---

## Architecture Patterns

### Package Organization

```
new-passquantum/
├── core/
│   ├── crypto/        # Cryptographic primitives (pure functions)
│   ├── model/         # Data structures and serialization
│   └── storage/       # File I/O operations
└── ui/               # User interface (Fyne-based)
```

### Dependency Flow

```
UI → Storage → Crypto
     ↓         ↓
     Model ←───┘
```

**Rules**:
- UI can import all other packages
- Storage can import crypto and model
- Crypto can import model (for entry encryption)
- Model is self-contained (no imports from other packages)

### Data Flow

#### Adding a Password
```
1. User Input (UI)
   ↓
2. Validation
   ↓
3. Kyber Encapsulation (crypto/kyber.go)
   ↓
4. AES Encryption (crypto/aes.go)
   ↓
5. Create Entry (model/vault_entry.go)
   ↓
6. Serialize & Encrypt Vault (crypto/vault.go)
   ↓
7. Write to Disk (storage/storage.go)
   ↓
8. UI Feedback
```

#### Viewing Passwords
```
1. Read Vault (storage/storage.go)
   ↓
2. Decrypt Vault (crypto/vault.go)
   ↓
3. Parse Entries (model/vault_entry.go)
   ↓
4. For each entry:
   - Kyber Decapsulation (crypto/kyber.go)
   - AES Decryption (crypto/aes.go)
   ↓
5. Display in UI
```

---

## Cryptographic Design

### Key Hierarchy

```
User Master Password (user input)
    ↓
┌─────────────────────────────────────────────────┐
│ Argon2id Key Derivation Function                │
│ • Memory: 64 MB (GPU-resistant)                 │
│ • Iterations: 1 (interactive use)               │
│ • Parallelism: 4 threads                        │
│ • Salt: 16 bytes (cryptographically random)     │
└─────────────────────────────────────────────────┘
    ↓
Master Key (64 bytes)
    ↓
┌─────────────────────────────────────────────────┐
│ Domain Separation (SHA-256 with label)          │
│ • Prevents key reuse across purposes            │
│ • Counter-mode expansion                        │
└─────────────────────────────────────────────────┘
    ↓
    ├─ Label "encryption" → Encryption Key (32 bytes)
    │  └─ AES-256-GCM for vault encryption
    │
    └─ Label "verification" → Verification Key (32 bytes)
       └─ HMAC-SHA256 for integrity
```

### Encryption Pipeline

#### Vault-Level Encryption
```
1. Serialize all vault entries → Plaintext
2. Generate random 12-byte nonce
3. AES-256-GCM(plaintext, encryption_key, nonce) → Ciphertext
4. HMAC-SHA256(version + kdf_params + ciphertext, verification_key) → MAC
5. Write: Version | KDF Params | MAC | Nonce | Ciphertext
```

#### Entry-Level Encryption (Individual Password)
```
1. Kyber768.Encapsulate(public_key) → (kyber_ct, shared_secret)
2. Generate random 12-byte nonce
3. AES-256-GCM(password, shared_secret, nonce) → Ciphertext
4. Create Entry: ID | Service | Username | Kyber_CT | Nonce | Ciphertext
```

### Cryptographic Primitives

| Component | Algorithm | Key Size | Notes |
|-----------|-----------|----------|-------|
| KDF | Argon2id | 64 bytes output | Memory-hard, GPU-resistant |
| Vault Encryption | AES-256-GCM | 256 bits | Authenticated encryption |
| Integrity | HMAC-SHA256 | 256 bits | Prevents tampering |
| Key Expansion | SHA-256 | 256 bits | Domain separation |
| Post-Quantum KEM | Kyber768 | ~1184 bytes CT | NIST PQC standard |
| Entry Encryption | AES-256-GCM | 256 bits | Per-password encryption |

---

## Module Reference

### core/crypto - Cryptographic Operations

#### kdf.go - Key Derivation
- **DefaultKDFParams()**: Returns secure default parameters
- **GenerateSalt()**: Creates cryptographically random 16-byte salt
- **DeriveKeys()**: Derives encryption and verification keys from master password
- **KDFParams.Serialize()**: Encodes parameters for storage
- **WipeBytes()**: Securely zeros out sensitive data

#### vault.go - Vault Encryption
- **EncryptVault()**: Encrypts vault contents with AES-256-GCM and computes HMAC
- **DecryptVault()**: Decrypts vault and verifies HMAC
- **VaultFile.Serialize()**: Converts vault to binary format
- **VaultFileDeserialize()**: Parses vault from binary format

#### kyber.go - Post-Quantum KEM
- **GenerateKeypair()**: Creates new Kyber768 keypair
- **LoadKeypair()**: Loads keypair from disk
- **SaveKeypair()**: Saves keypair with proper permissions
- **Encapsulate()**: Generates shared secret and ciphertext
- **Decapsulate()**: Recovers shared secret from ciphertext

#### aes.go - Symmetric Encryption
- **EncryptAES256GCM()**: Encrypts plaintext with AES-256-GCM
- **DecryptAES256GCM()**: Decrypts ciphertext with AES-256-GCM

### core/model - Data Structures

#### vault_entry.go
```go
type VaultEntry struct {
    ID              uint64   // Unique entry identifier
   Type            EntryType
   CardSubtype     string
    Service         string   // Service/website name
    Username        string   // Associated username or email
    KyberCiphertext []byte   // Kyber768 encapsulated secret
    Nonce           []byte   // AES-GCM nonce (12 bytes)
   Ciphertext      []byte   // AES-256-GCM encrypted entry payload
}
```

- **NewVaultEntry()**: Creates entry with random ID
- **Serialize()**: Converts entry to binary format
- **Deserialize()**: Parses entry from binary format

### core/storage - File I/O

#### storage.go
- **WriteVault()**: Encrypts and writes entire vault to disk
- **ReadVault()**: Reads and decrypts vault from disk
- **VaultExists()**: Checks if vault file exists
- **DeleteVault()**: Removes vault file from disk

### ui - User Interface

#### main.go - Application Entry
- **initializeApp()**: Loads or creates Kyber keypair
- **main()**: Application entry point

#### login_screen.go - Authentication
- **PromptMasterPassword()**: Displays login/creation screen

#### vault_selection.go - Vault Management 
- **ShowVaultSelection()**: Displays vault list and management
- **createVaultCard()**: Creates vault card UI component

#### main_screen.go - Vault Manager
- **ShowMainScreen()**: Main vault item entry interface

#### passwords_view.go - Vault Item Display
- **ShowPasswordsView()**: Displays all vault items in a vault
- **createPasswordCard()**: Creates individual password card

#### settings_screen.go - Settings
- **ShowSettingsScreen()**: Tabbed settings interface
- **buildSecuritySettings()**: Security tab
- **buildVaultSettings()**: Vault management tab
- **buildDisplaySettings()**: Display customization tab
- **buildBackupSettings()**: Backup configuration tab
- **buildAboutSettings()**: About/info tab

#### helpers.go - Utilities
- **ListVaults()**: Returns all available vault names
- **GetVaultPath()**: Returns full path for vault file
- **CreateNewVault()**: Creates encrypted vault with master password
- **UnlockVault()**: Decrypts existing vault
- **AddPasswordToVault()**: Encrypts and adds password
- **DeletePasswordFromVault()**: Removes password from vault

---

## Data Models

### Vault File Structure

```
┌─────────────────────────────────────────────────────────┐
│ HEADER                                                  │
├─────────────────────────────────────────────────────────┤
│ Version (1 byte): 0x01                                  │
│ KDF Params Length (1 byte): 26                          │
│ KDF Parameters (26 bytes):                              │
│   • Salt (16 bytes)                                     │
│   • Memory (4 bytes, uint32 big-endian)                │
│   • Iterations (4 bytes, uint32 big-endian)            │
│   • Parallelism (1 byte)                               │
│   • Version (1 byte)                                   │
├─────────────────────────────────────────────────────────┤
│ INTEGRITY                                               │
├─────────────────────────────────────────────────────────┤
│ HMAC-SHA256 (32 bytes)                                  │
├─────────────────────────────────────────────────────────┤
│ ENCRYPTED DATA                                          │
├─────────────────────────────────────────────────────────┤
│ Encrypted Data Length (4 bytes, uint32)                │
│ AES-GCM Nonce (12 bytes)                               │
│ AES-GCM Ciphertext (variable + 16-byte tag)           │
└─────────────────────────────────────────────────────────┘
```

### Entry Serialization Format

```
EntryID (8 bytes, uint64 big-endian)
ServiceLen (2 bytes, uint16 big-endian)
Service (variable UTF-8)
UsernameLen (2 bytes, uint16 big-endian)
Username (variable UTF-8)
KyberLen (2 bytes, uint16 big-endian)
KyberCiphertext (~1088 bytes)
Nonce (12 bytes)
CiphertextLen (2 bytes, uint16 big-endian)
Ciphertext (variable + 16-byte tag)
```

---

## Storage Format

### File Locations
```
new-passquantum/
├── vaults/              # Encrypted vault files
│   ├── Personal.pqdb    # Example vault
│   ├── Work.pqdb        # Example vault
│   └── Finance.pqdb     # Example vault
├── public.key           # Kyber768 public key (1184 bytes)
├── private.key          # Kyber768 private key (2400 bytes)
└── passquantum          # Executable
```

### File Permissions
- **vaults/*.pqdb**: 0600 (owner read/write only)
- **private.key**: 0600 (owner read/write only)
- **public.key**: 0644 (world readable)

---

## Security Properties

### Threat Model

| Threat | Mitigation | Status |
|--------|-----------|--------|
| Offline brute-force | Argon2id (64MB) | ✅ Mitigated |
| Vault tampering | HMAC-SHA256 | ✅ Detected |
| Wrong password | HMAC + decrypt fail | ✅ Detected |
| Nonce reuse | Random generation | ✅ Prevented |
| Key derivation attacks | Domain separation | ✅ Prevented |
| Post-quantum attacks | Kyber768 | ✅ Resistant |
| Memory scraping | Wiping on lock | ⚠️ Partial |
| Malware/keylogger | OS security | ❌ Out of scope |

### Security Assumptions
1. User chooses strong master password
2. Kyber private key is protected
3. Operating system is trusted
4. Crypto libraries are sound
5. System entropy is available

---

## API Reference

### Complete Function Signatures

```go
// KDF
func DefaultKDFParams() KDFParams
func GenerateSalt() ([]byte, error)
func DeriveKeys(password string, params KDFParams) (encKey, verKey []byte, err error)
func WipeBytes(data []byte)

// Vault
func EncryptVault(plaintext []byte, encKey, verKey []byte, params KDFParams) (*VaultFile, error)
func DecryptVault(vault *VaultFile, encKey, verKey []byte) ([]byte, error)

// Kyber
func GenerateKeypair() (*kyber768.PublicKey, *kyber768.PrivateKey, error)
func Encapsulate(publicKey *kyber768.PublicKeyfunc) (ciphertext, sharedSecret []byte, err error)
func Decapsulate(ciphertext []byte, privateKey *kyber768.PrivateKey) (sharedSecret []byte, err error)

// AES
func EncryptAES256GCM(plaintext string, sharedSecret []byte) (nonce, ciphertext []byte, err error)
func DecryptAES256GCM(nonce, ciphertext, sharedSecret []byte) (plaintext string, err error)

// Model
func NewVaultEntry() *VaultEntry
func (e *VaultEntry) Serialize() []byte
func Deserialize(data []byte) (*VaultEntry, error)

// Storage
func WriteVault(entries []*model.VaultEntry, vaultPath string, 
                encKey, verKey []byte, params crypto.KDFParams) error
func ReadVault(vaultPath string, encKey, verKey []byte) ([]*model.VaultEntry, error)
func VaultExists(vaultPath string) bool
func DeleteVault(vaultPath string) error

// UI Helpers
func ListVaults() []string
func GetVaultPath(vaultName string) string
func CreateNewVault(w interface{}, appState *AppState, masterPassword, vaultName string) bool
func UnlockVault(w interface{}, appState *AppState, vaultPath, masterPassword string) bool
func AddPasswordToVault(appState *AppState, service, username, password string) error
func DeletePasswordFromVault(appState *AppState, entryID uint64) error
```

---

## Performance Characteristics

### Computational Costs

| Operation | Time | Notes |
|-----------|------|-------|
| Argon2id KDF | ~2s | Intentionally expensive |
| Kyber Encapsulate | ~0.5ms | Fast KEM |
| Kyber Decapsulate | ~0.7ms | Slightly slower |
| AES-256-GCM Encrypt | <0.1ms | Hardware-accelerated |
| AES-256-GCM Decrypt | <0.1ms | Hardware-accelerated |
| HMAC-SHA256 | <0.1ms | Fast hash |

### Scalability

| Vault Size | Entries | Add Time | View Time |
|------------|---------|----------|-----------|
| Small | 1-10 | <100ms | <100ms |
| Medium | 11-100 | <200ms | <200ms |
| Large | 101-1000 | <500ms | <500ms |
| Very Large | 1001+ | ~1s+ | ~1s+ |

---

## Testing Guidelines

### Unit Testing

```go
// Test KDF determinism
func TestKDFDeterminism(t *testing.T) {
    params := crypto.DefaultKDFParams()
    key1, ver1, _ := crypto.DeriveKeys("password", params)
    key2, ver2, _ := crypto.DeriveKeys("password", params)
    assert.Equal(t, key1, key2)
}

// Test vault round-trip
func TestVaultRoundTrip(t *testing.T) {
    plaintext := []byte("test data")
    // ... encrypt and decrypt ...
    assert.Equal(t, plaintext, decrypted)
}

// Test HMAC integrity
func TestVaultTampering(t *testing.T) {
    // ... tamper with vault ...
    _, err := crypto.DecryptVault(vault, encKey, verKey)
    assert.Error(t, err)
}
```

### Integration Testing

```bash
# Verify no plaintext
strings vaults/test.pqdb | grep -i "password"

# Test wrong password
./passquantum  # Enter wrong password

# Test tampering
xxd vaults/test.pqdb  # Modify bytes
./passquantum  # Should detect tampering
```

### Performance Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./core/crypto
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=. ./core/crypto
go tool pprof mem.prof
```

---

## Extension Points

### Adding New Cryptographic Algorithm

```go
// 1. Create core/crypto/new_algo.go
func EncryptNewAlgo(plaintext string, key []byte) ([]byte, error) {
    // Implementation
}

// 2. Update UI to use new algorithm
```

### Adding New Storage Backend

```go
// 1. Create core/storage/cloud.go
func WriteVaultToCloud(vault *crypto.VaultFile, cloudPath string) error {
    // Cloud upload implementation
}

// 2. Update UI to offer cloud sync
```

### Adding New UI Screen

```go
// 1. Create ui/new_screen.go
func ShowNewScreen(w fyne.Window, fyneApp fyne.App, appState *AppState) {
    // New screen implementation
}

// 2. Add navigation from existing screens
```

---

**PassQuantum Architecture** - Secure by design, modular by nature. 🔐
