# PassQuantum Development Guide

## Project Structure Recap

```
core/crypto/      - Cryptographic operations (Kyber768, AES-256)
core/model/       - Data structures (PasswordEntry)
core/storage/     - File I/O operations
ui/               - Fyne GUI application
```

## Adding New Features

### Adding a New Encryption Algorithm

1. **Create new file** in `core/crypto/`:
   ```bash
   # e.g., for RSA support
   core/crypto/rsa.go
   ```

2. **Implement functions** following the pattern:
   ```go
   package crypto
   
   // Functions should be pure or have clear side effects
   // Should NOT import any UI packages
   
   func EncryptRSA(plaintext string, publicKey *rsa.PublicKey) ([]byte, error) {
       // Implementation
   }
   
   func DecryptRSA(ciphertext []byte, privateKey *rsa.PrivateKey) (string, error) {
       // Implementation
   }
   ```

3. **Update UI** to call new functions in `ui/main.go`:
   ```go
   // In the encrypt goroutine:
   ciphertext, err := crypto.EncryptRSA(pass, appState.publicKey)
   ```

### Adding a New Storage Format

1. **Create new file** in `core/storage/`:
   ```bash
   core/storage/json_storage.go
   ```

2. **Implement storage functions**:
   ```go
   package storage
   
   func WritePasswordJSON(entry *model.PasswordEntry, filePath string) error {
       // JSON serialization logic
   }
   
   func ReadPasswordsJSON(filePath string) ([]*model.PasswordEntry, error) {
       // JSON deserialization logic
   }
   ```

3. **The UI remains unchanged** - it just calls different functions

### Adding UI Controls

1. **Update** `ui/main.go` in the `buildUI()` function:
   ```go
   // Add new widget
   newButton := widget.NewButton("My Feature", func() {
       // Call core functions here
       // Update UI in response
   })
   
   // Add to layout
   buttonBox := container.NewVBox(
       // ... existing widgets ...
       newButton,
   )
   ```

2. **For async operations**, use goroutines:
   ```go
   go func() {
       appState.mu.Lock()
       defer appState.mu.Unlock()
       
       // Long-running operation
       result := someExpensiveOperation()
       
       // Update UI on main thread
       dialog.ShowInformation("Result", result, w)
   }()
   ```

## Testing

### Unit Testing Core Packages

Create test files alongside implementation:

```bash
core/crypto/kyber_test.go
core/crypto/aes_test.go
core/storage/storage_test.go
core/model/password_entry_test.go
```

**Example test**:
```go
// core/crypto/aes_test.go
package crypto

import (
    "testing"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
    plaintext := "test_password"
    key := make([]byte, 32)
    
    nonce, ciphertext, err := EncryptAES256GCM(plaintext, key)
    if err != nil {
        t.Fatalf("Encryption failed: %v", err)
    }
    
    decrypted, err := DecryptAES256GCM(nonce, ciphertext, key)
    if err != nil {
        t.Fatalf("Decryption failed: %v", err)
    }
    
    if decrypted != plaintext {
        t.Errorf("Round-trip failed: got %q, want %q", decrypted, plaintext)
    }
}
```

**Run tests**:
```bash
go test ./core/...
go test ./core/crypto
go test ./core/storage
```

## Build & Deployment

### Building with Custom Configuration

```bash
# Standard build
go build -o passquantum-gui ./ui

# With optimizations
go build -ldflags="-s -w" -o passquantum-gui ./ui

# With version info
go build -ldflags="-X main.Version=1.0.0" -o passquantum-gui ./ui

# For different OS/arch
GOOS=windows GOARCH=amd64 go build -o passquantum.exe ./ui
GOOS=darwin GOARCH=amd64 go build -o passquantum-mac ./ui
```

### Cross-Platform Building

Fyne supports Linux, macOS, and Windows, but requires platform-specific setup:

**Linux**:
```bash
# Install dependencies
sudo apt-get install libgl1-mesa-dev libxcursor-dev libxinerama-dev libxrandr-dev

# Build
go build -o passquantum-gui ./ui
```

**macOS**:
```bash
# Install Xcode command line tools first
xcode-select --install

# Build
go build -o passquantum-gui ./ui
```

**Windows**:
```powershell
# Using MinGW or MSVC
go build -o passquantum.exe ./ui
```

## Code Style Guidelines

### Naming Conventions

- **Functions**: PascalCase for exported, camelCase for private
  ```go
  func EncryptPassword()  // ✓ Exported
  func encryptPassword()  // ✓ Private
  ```

- **Packages**: lowercase, single word preferred
  ```go
  package crypto    // ✓ Good
  package encrypt   // ✓ Also acceptable
  package cryptography  // ✗ Too long
  ```

- **Variables**: camelCase for readability
  ```go
  encapsulatedSecret := ...  // ✓ Clear
  ct := ...                   // ✗ Unclear abbreviation
  ```

### Error Handling

Always propagate errors explicitly:

```go
// ✓ Good - caller decides what to do
func EncryptAES256GCM(...) ([]byte, []byte, error) {
    block, err := aes.NewCipher(sharedSecret[:32])
    if err != nil {
        return nil, nil, err  // ✓ Propagate
    }
    // ...
}

// ✗ Bad - silently fails
func EncryptAES256GCM(...) []byte {
    block, _ := aes.NewCipher(sharedSecret[:32])  // ✗ Ignore error
    // ...
}
```

### Documentation

Export functions should have doc comments:

```go
// EncryptAES256GCM encrypts plaintext using AES-256-GCM
// with the given shared secret.
//
// The nonce is randomly generated (96 bits for GCM).
// The ciphertext is authenticated.
//
// Returns (nonce, ciphertext, error).
func EncryptAES256GCM(plaintext string, sharedSecret []byte) ([]byte, []byte, error) {
```

## Performance Optimization

### Profiling

Run with pprof to identify bottlenecks:

```bash
go build -o passquantum-gui ./ui
./passquantum-gui -cpuprofile=cpu.prof -memprofile=mem.prof

# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

### Common Bottlenecks

1. **Key Generation** - Don't regenerate unless necessary
2. **File I/O** - Batch reads/writes when possible
3. **Goroutine Overhead** - Only use for truly async operations
4. **Crypto Operations** - They're slow by design (security)

### Optimization Strategies

1. **Cache keypairs** in memory (done: AppState struct)
2. **Lazy-load** password list (not done: could be improved)
3. **Use sync.Pool** for temporary buffers (not needed currently)
4. **Avoid repeated serialization** (done: base64 only in storage)

## Dependency Management

### Adding Dependencies

```bash
# Add a new dependency
go get github.com/user/package

# Update to specific version
go get github.com/user/package@v1.2.3

# Update all dependencies
go get -u ./...

# Clean up unused dependencies
go mod tidy
```

### Current Dependencies

```
github.com/cloudflare/circl (v1.3.7)
  ├── Kyber768 implementation
  └── Post-quantum cryptography

fyne.io/fyne/v2 (v2.6.0)
  ├── GUI framework
  └── Cross-platform desktop UI
```

### Avoiding Dependency Hell

1. Pin versions in go.mod when stability is critical
2. Use `go get -u ./...` carefully (may break things)
3. Always run tests after updating dependencies
4. Check release notes for breaking changes

## Debugging

### Enable Verbose Output

```bash
# Build with debug info
go build -v -x ./ui

# Run with logging
FYNE_THEME=dark ./passquantum-gui  # Try dark theme
FYNE_LOG=1 ./passquantum-gui      # Enable logging
```

### Common Issues

**Issue**: Application doesn't start
```bash
# Check dependencies
go mod verify

# Rebuild from scratch
go clean
go build -v ./ui
```

**Issue**: Encryption/decryption fails
- Check that private key file exists and is readable
- Verify password entry format in `passwords.txt`
- Look for disk permission issues

**Issue**: GUI doesn't display correctly
- Ensure X11/Wayland is running
- Try `FYNE_BACKEND=gl ./passquantum-gui`
- Update Fyne: `go get -u fyne.io/fyne/v2`

## Security Considerations for Development

### Code Review Checklist

- ✓ No hardcoded keys or secrets
- ✓ All cryptographic operations in `core/crypto`
- ✓ Error messages don't leak sensitive data
- ✓ Random number generation uses `crypto/rand`
- ✓ Key material cleared from memory when done (C crypto libraries do this)
- ✓ File permissions set correctly (0600 for private keys)

### Common Security Mistakes

1. **Using math/rand for keys** - Always use crypto/rand
2. **Storing plaintext passwords** - Only store encrypted
3. **Logging sensitive data** - Remove debug output before commit
4. **Hardcoding test keys** - Never commit real keys
5. **Ignoring errors** - Errors often indicate security issues

## Contributing

1. **Fork and clone** the repository
2. **Create feature branch**: `git checkout -b feature/my-feature`
3. **Make changes** following code style guidelines
4. **Test thoroughly**: `go test ./...`
5. **Commit with clear messages**: `git commit -m "Add feature X"`
6. **Push and create PR**: `git push origin feature/my-feature`

### Commit Message Guidelines

```
feat: Add password strength indicator
^--- Type

fix: Resolve crash when viewing empty password list
refactor: Split UI main.go into smaller files
docs: Update USER_GUIDE with new features
test: Add comprehensive crypto tests

Types: feat, fix, refactor, docs, test, perf, security
```
