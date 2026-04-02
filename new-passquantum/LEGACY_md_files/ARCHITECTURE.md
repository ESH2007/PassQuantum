# PassQuantum - Refactored Architecture

A post-quantum safe password manager with a clean, modular architecture and desktop GUI using Fyne.

## Architecture Overview

The application has been refactored from a monolithic CLI design into a modular, layered architecture:

```
PassQuantum/
├── core/
│   ├── crypto/        # Cryptographic operations (no UI)
│   │   ├── kyber.go   # Kyber768 keypair management
│   │   └── aes.go     # AES-256-GCM encryption/decryption
│   ├── model/         # Data models
│   │   └── password_entry.go  # PasswordEntry struct & serialization
│   └── storage/       # Persistent storage
│       └── storage.go # File I/O for password entries
├── ui/
│   └── main.go        # Fyne GUI application
├── go.mod
├── go.sum
└── main.go.backup     # Original CLI version
```

## Package Responsibilities

### `core/crypto`
**Responsibility**: Cryptographic operations only. No UI dependencies, no global state.

**Files**:
- **kyber.go**: Kyber768 post-quantum key encapsulation
  - `GenerateKeypair()` - Generate new keypair
  - `LoadKeypair(pubPath, privPath)` - Load keypair from disk
  - `SaveKeypair(pub, priv, pubPath, privPath)` - Persist keypair
  - `Encapsulate(publicKey)` - Create encapsulated secret
  - `Decapsulate(secret, privateKey)` - Recover shared secret

- **aes.go**: AES-256-GCM symmetric encryption
  - `EncryptAES256GCM(plaintext, sharedSecret)` - Encrypt with shared key
  - `DecryptAES256GCM(nonce, ciphertext, sharedSecret)` - Decrypt with shared key

### `core/model`
**Responsibility**: Define data structures and serialization logic.

**Files**:
- **password_entry.go**: Password entry data model
  - `PasswordEntry` struct:
    - `EncapsulatedSecret []byte` - Kyber768 ciphertext
    - `Nonce []byte` - GCM nonce
    - `Ciphertext []byte` - Encrypted password
  - `Serialize()` - Encode entry to storage format (base64)
  - `Deserialize(string)` - Parse entry from file

### `core/storage`
**Responsibility**: File I/O only. Handles persisting encrypted passwords.

**Files**:
- **storage.go**: Password file management
  - `WritePassword(entry, path)` - Append entry to file
  - `ReadPasswords(path)` - Parse all entries from file
  - `DeletePasswordFile(path)` - Remove password database
  - `FileExists(path)` - Check if database exists
  - Error handling for missing files, malformed entries

### `ui`
**Responsibility**: GUI only. Calls core/* functions, no business logic.

**Files**:
- **main.go**: Fyne desktop application
  - `main()` - Application entry point, window setup
  - `initializeApp()` - Initialize keypair (load or generate)
  - `buildUI()` - Create GUI widgets and callbacks
  - `showPasswordsWindow()` - Display decrypted passwords in new window

**Features**:
- Masked password input field
- "Add Password" button (runs encryption in goroutine)
- "View Passwords" button (decrypts & displays in scrollable list)
- "Exit" button
- Error dialogs for user feedback
- Non-blocking async operations with goroutines

## Cryptographic Design

**Encryption Pipeline**:
1. Kyber768 encapsulation with public key → encapsulated secret + shared secret
2. AES-256-GCM encryption of password with shared secret → nonce + ciphertext
3. All three components base64-encoded and persisted to disk

**Storage Format**:
```
[base64(encapsulated_secret), base64(nonce), base64(ciphertext)], \n
```

**Decryption Pipeline**:
1. Read entry from file, base64-decode all components
2. Kyber768 decapsulation with private key → recover shared secret
3. AES-256-GCM decryption with shared secret → plaintext password

## Building & Running

### Prerequisites
- Go 1.22+
- Fyne v2.6.0+ (automatically downloaded via `go mod`)
- Linux with GUI support (X11/Wayland)

### Build
```bash
cd new-passquantum
go mod tidy
go build -o passquantum-app ./ui
```

### Run
```bash
./passquantum-app
```

The application will:
1. Check for `public.key` and `private.key` in the current directory
2. If not found, generate a new Kyber768 keypair and save it
3. Load the GUI window
4. Store encrypted passwords in `passwords.txt`

## Design Principles

### Separation of Concerns
- **Crypto** handles cryptography
- **Storage** handles I/O
- **Model** handles data structures  
- **UI** handles only presentation and user events

### No Global State
- All state is passed as function parameters
- Mutex-protected shared state in UI (AppState struct)
- Concurrent operations via goroutines don't block UI

### Testability
- Core packages have no UI dependencies
- Functions are pure or have clear side effects
- Easy to unit test crypto, storage, model packages in isolation

### Performance
- Goroutines used for encryption/decryption to avoid UI blocking
- No unnecessary async overhead for simple operations
- Direct function calls for I/O within core packages
- Fast startup (no WebView, no JVM)

## Comparison with Original

| Aspect | Original | Refactored |
|--------|----------|-----------|
| Architecture | Monolithic CLI | Modular packages |
| UI | Terminal-only | Desktop GUI (Fyne) |
| Crypto | Mixed with I/O | Pure, isolated |
| Storage | Mixed with crypto | Dedicated package |
| Testing | Difficult | Easy (core packages) |
| Async | Blocking | Non-blocking UI |
| Performance | Terminal input delay | Instant response |

## Files Modified/Created

- ✅ `core/crypto/kyber.go` - New
- ✅ `core/crypto/aes.go` - New
- ✅ `core/model/password_entry.go` - New
- ✅ `core/storage/storage.go` - New
- ✅ `ui/main.go` - New (replaces old main.go)
- ✅ `main.go.backup` - Backup of original
- ✅ `go.mod` - Updated with Fyne dependency
- ✅ `go.sum` - Generated

## Future Enhancements

- Add search/filter functionality in password list
- Add password strength indicator
- Add copy-to-clipboard for decrypted passwords
- Add dark mode theme
- Add master password protection
- Add password generation utility
- Add multi-user support
- Add cloud backup option

## Security Notes

- Private keys stored unencrypted in `private.key` with 0600 permissions
- No master password protection (consider adding)
- Decrypted passwords stored in memory only during display
- File permissions on `passwords.txt` are 0644 (readable by all users)
