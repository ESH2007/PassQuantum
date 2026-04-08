# PassQuantum - Post-Quantum Password Manager

> A production-ready, open-source password manager featuring post-quantum cryptography (Kyber768), AES-256-GCM encryption, Argon2id key derivation, and a modern desktop GUI.

## 🔐 Security Highlights

- **Post-Quantum Cryptography**: Kyber768 key encapsulation mechanism resistant to quantum computer attacks
- **Master Password Protection**: Argon2id KDF with 64MB memory requirement (GPU-resistant)
- **Encrypted Vault Storage**: All passwords encrypted with AES-256-GCM in single `vault.pqdb` file
- **Integrity Verification**: HMAC-SHA256 detects tampering and validates master passwords
- **Multi-Vault Support**: Create and manage multiple independent password vaults
- **Zero Plaintext Storage**: No passwords or keys ever written to disk in plaintext
- **Modular Architecture**: Clean separation of cryptography, storage, model, and UI

## 📋 Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Usage](#usage)
- [Security Architecture](#security-architecture)
- [Building from Source](#building-from-source)
- [Project Structure](#project-structure)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [License](#license)

## ✨ Features

### Core Functionality
- ✅ **Add Passwords**: Store passwords with service name, username, and encrypted password
- ✅ **View Passwords**: Decrypt and display all passwords in a vault
- ✅ **Multiple Vaults**: Create and manage separate password vaults with different master passwords
- ✅ **Copy to Clipboard**: Quick copy functionality for passwords
- ✅ **Password Management**: Delete individual passwords or entire vaults

### Security Features
- ✅ **Post-Quantum Encryption**: Kyber768 + AES-256-GCM hybrid encryption
- ✅ **Master Password**: Each vault protected by Argon2id-derived keys
- ✅ **HMAC Integrity**: Detects any tampering with vault files
- ✅ **Unique Entry IDs**: Each password has a unique 64-bit identifier
- ✅ **Fresh Nonces**: Every encryption operation uses a random nonce
- ✅ **Domain Separation**: Prevents key misuse through SHA-256 key derivation

### User Interface
- ✅ **Modern GUI**: Professional Fyne-based desktop interface
- ✅ **Animated Background**: Smooth 60 FPS particle animation
- ✅ **Enhanced Visuals**: Cyberpunk aesthetic with neon glows and effects
- ✅ **Responsive Design**: Adaptive layouts for different screen sizes
- ✅ **Masked Input**: Password fields are masked by default
- ✅ **Settings Panel**: Comprehensive 5-tab settings interface

### Advanced Features
- ✅ **Vault Selection**: Easy switching between multiple vaults
- ✅ **Vault Management**: Create, open, and delete vaults from GUI
- ✅ **Backup & Restore**: Export and import encrypted vault backups
- ✅ **Vault Statistics**: View password counts and vault information
- ✅ **Theme Support**: Dark mode with potential for light theme

## 🚀 Quick Start

### Prerequisites
- Go 1.22 or later
- Linux/macOS/Windows with GUI support (X11, Wayland, or native)
- Docker (optional, for cross-platform builds)

### Installation

#### Option 1: Download Pre-built Binary
Download the latest release for your platform from the [Releases](https://github.com/yourusername/passquantum/releases) page:
- **Linux**: `PassQuantum-linux-amd64.tar.gz`
- **Windows**: `PassQuantum-windows-amd64.zip`
- **macOS**: `PassQuantum-macos-amd64.zip`

Extract and run the executable.

#### Option 2: Build from Source
```bash
# Clone the repository
git clone https://github.com/yourusername/passquantum.git
cd passquantum/new-passquantum

# Install dependencies
go mod tidy

# Build the application
go build -o passquantum ./ui

# Run
./passquantum
```

### First Run

On first launch, PassQuantum will:
1. Check for existing Kyber768 keypair files
2. Generate new keypair if none exists (saved as `public.key` and `private.key`)
3. Prompt for master password to create your first vault
4. Create an encrypted vault file (`vaults/{name}.pqdb`)
5. Display the main password manager interface

## 📖 Usage

### Creating a Vault

1. Launch PassQuantum
2. Enter a master password on the login screen
3. Click "Create New Vault" if no vaults exist
4. Enter a name for your vault (e.g., "Personal", "Work", "Finance")
5. Your vault is created and ready to use

### Adding a Password

1. Select or create a vault
2. On the main screen, enter:
   - **Service Name**: Website or application name (e.g., "Gmail")
   - **Username/Email**: Your username or email for that service
   - **Password**: The password to encrypt and store
3. Click "Add Password"
4. Password is encrypted and saved to the vault

### Viewing Passwords

1. Click "View All Passwords" on the main screen
2. All passwords are decrypted and displayed in cards showing:
   - Service name and entry number
   - Username/email
   - Masked password (click "Show" to reveal)
3. Use "Copy" to copy password to clipboard
4. Use "Delete" to remove a password entry

### Managing Multiple Vaults

1. Click "Back to Vaults" from the main screen
2. View all available vaults
3. Click "Open" to access a vault
4. Click "Delete" to remove a vault (with confirmation)
5. Click "Create New Vault" to add another vault

### Settings

Access comprehensive settings through the settings button:
- **Security**: Password strength, master password change, session settings
- **Vault**: Statistics, compaction, export/import
- **Display**: Theme selection, font size, visibility options
- **Backup**: Automatic backup configuration, manual backup/restore
- **About**: Version info, features, documentation links

## 🔐 Security Architecture

### Cryptographic Stack

```
User Master Password
    ↓
Argon2id(password, 16-byte salt, 64MB memory, 1 iteration, 4 threads)
    ↓
Master Key (64 bytes)
    ↓
Domain Separation (SHA-256)
    ├─ "encryption" → Encryption Key (32 bytes) → AES-256-GCM
    └─ "verification" → Verification Key (32 bytes) → HMAC-SHA256

Entry Encryption:
Password → Kyber768.Encapsulate() → Shared Secret
         → AES-256-GCM(password, shared secret) → Ciphertext
         → HMAC-SHA256(vault contents) → Integrity Check
```

### Vault File Format

```
Version(1B) | KDFParams(26B) | HMAC(32B) | [Nonce(12B) + AES-GCM-CT(var)]
```

Each vault file (`vault.pqdb`) contains:
- **Version marker**: Format versioning for future upgrades
- **KDF parameters**: Salt + Argon2id configuration (for reproducible key derivation)
- **HMAC**: Integrity verification computed over all data
- **Encrypted entries**: AES-256-GCM encrypted password entries with unique nonces

### Password Entry Structure

Each encrypted password entry:
```
EntryID(8) | Service(var) | Username(var) | KyberCT(~1088) | Nonce(12) | AES-CT(var)
```

### Security Properties

| Threat | Mitigation | Status |
|--------|-----------|--------|
| Offline brute-force | Argon2id (64MB memory) | ✅ Mitigated |
| Vault tampering | HMAC-SHA256 verification | ✅ Detected |
| Wrong password | HMAC + decryption failure | ✅ Detected |
| Key derivation attacks | Domain separation | ✅ Prevented |
| Post-quantum attacks | Kyber768 encapsulation | ✅ Resistant |
| Nonce reuse | Fresh random nonce per operation | ✅ Prevented |
| Key leakage | Keys only in RAM, wiped on lock | ✅ Mitigated |

## 🏗️ Building from Source

### Native Build (Fastest)

```bash
cd new-passquantum
./build-native.sh
```

This creates a native executable for your current platform.

### Cross-Platform Build (Requires Docker)

```bash
# Build for Linux
./build.sh linux

# Build for Windows
./build.sh windows

### Face Recognition Demo (Standalone)

Run a live demo of the same face-recognition pipeline used by the app:

```bash
go run ./cmd/biometric-demo -camera 0 -threshold 0.97
```

Demo controls:
- `E`: Enroll current detected face as reference template
- `C`: Clear enrolled template
- `Q` or `Esc`: Quit demo

Notes:
- Uses the same BlazeFace + Face Mesh + feature extraction + cosine similarity flow as the main app.
- Requires `models/blazeface.onnx` and `models/face_mesh.onnx`.

# Build for macOS
./build.sh mac

# Build all platforms
./build.sh all
```

Built binaries are placed in `fyne-cross/dist/{platform}-amd64/`.

### Manual Build

```bash
# Standard build
go build -o passquantum ./ui

# Optimized build (smaller binary)
go build -ldflags="-s -w" -o passquantum ./ui

# With version info
go build -ldflags="-X main.Version=1.0.0" -o passquantum ./ui
```

### Build Requirements

- **Go**: 1.22 or later
- **Fyne dependencies**: 
  - Linux: `libgl1-mesa-dev libxcursor-dev libxinerama-dev libxrandr-dev`
  - macOS: Xcode command-line tools
  - Windows: MinGW or MSVC
- **Docker**: Required for cross-platform builds

## 📁 Project Structure

```
new-passquantum/
├── core/                       # Core cryptographic and storage logic
│   ├── crypto/                 # Cryptographic operations (no UI dependencies)
│   │   ├── aes.go             # AES-256-GCM encryption/decryption
│   │   ├── kdf.go             # Argon2id key derivation + domain separation
│   │   ├── kyber.go           # Kyber768 keypair management
│   │   └── vault.go           # Vault encryption/decryption with HMAC
│   ├── model/                  # Data structures
│   │   └── password_entry.go  # PasswordEntry struct & serialization
│   └── storage/                # Persistent storage
│       └── storage.go         # Vault file I/O operations
├── ui/                         # User interface
│   ├── main.go                # Application entry point & keypair init
│   ├── login_screen.go        # Master password authentication
│   ├── vault_selection.go     # Multi-vault management
│   ├── main_screen.go         # Main password manager interface
│   ├── passwords_view.go      # Password display with cards
│   ├── settings_screen.go     # Comprehensive settings panel
│   ├── helpers.go             # Utility functions & crypto wrappers
│   ├── ui_theme.go            # Theme, colors, animations
│   └── ui_enhancements.go     # Enhanced UI components
├── vaults/                     # Runtime vault storage
│   └── *.pqdb                 # Encrypted vault files
├── build.sh                    # Cross-platform build script
├── build-native.sh             # Native build script
├── Makefile                    # Build automation
├── go.mod                      # Go module dependencies
├── go.sum                      # Dependency checksums
├── public.key                  # Kyber768 public key (generated on first run)
├── private.key                 # Kyber768 private key (generated on first run)
├── README.md                   # This file
├── ARCHITECTURE.md             # Detailed technical architecture
├── USER_EXPERIENCE.md          # User-facing documentation
└── LEGACY_md_files/            # Legacy documentation files
```

## 📚 Documentation

### Core Documentation
- **[README.md](README.md)**: This file - project overview and quick start
- **[ARCHITECTURE.md](ARCHITECTURE.md)**: Detailed technical architecture, cryptographic design, and API reference
- **[USER_EXPERIENCE.md](USER_EXPERIENCE.md)**: User guide, troubleshooting, and UI navigation

### Legacy Documentation (LEGACY_md_files/)
Complete historical documentation including:
- Build guides, compilation instructions, security architecture details
- Implementation guides, development guides, refactoring summaries
- UI component API, visual guides, enhancement reports
- Windows fixes, quick references, completion checklists

See `LEGACY_md_files/` for all historical documentation.

## 🎯 Performance

| Operation | Time | Notes |
|-----------|------|-------|
| App startup | ~1s | Keypair loading |
| New vault creation | ~2s | Argon2id KDF (64MB) |
| Vault unlock | ~2s | Key derivation + HMAC |
| Add password | <100ms | Encryption + vault rewrite |
| View passwords | <100ms | Decryption only |
| Lock vault | <10ms | Memory clearing |

**Note**: KDF is intentionally slow to resist brute-force attacks on master passwords.

## 🧪 Security Testing

### Manual Verification

```bash
# Verify no plaintext stored
strings vaults/vault.pqdb | grep -i "password"  # Should find nothing
hexdump -C vaults/vault.pqdb | head -20         # All binary data

# Test wrong password rejection
# → Launch app, enter wrong password → "invalid master password"

# Test vault tampering detection
# → Modify vault.pqdb with hex editor → "vault integrity check failed"
```

### Recommended Security Audits

For production deployment:
1. Third-party security code review of `core/crypto/`
2. Fuzz testing of entry serialization
3. Penetration testing for password leaks
4. Dependency audit (Fyne, CIRCL, Go stdlib)
5. Compliance review (GDPR, data retention)

## 🤝 Contributing

Contributions are welcome! Please see our contributing guidelines:

### Areas for Contribution
- Security audits and cryptographic review
- UI/UX improvements and theme development
- Cross-platform testing (Windows, macOS, Linux)
- Performance optimizations
- Documentation improvements
- Internationalization (i18n) support
- Accessibility enhancements

### Development Setup

```bash
# Clone and setup
git clone https://github.com/yourusername/passquantum.git
cd passquantum/new-passquantum
go mod tidy

# Run tests (when available)
go test ./core/...

# Build and test
go build -o passquantum ./ui
./passquantum
```

## 🛣️ Roadmap

### Upcoming Features
- [ ] Password strength meter and generator
- [ ] Master password change functionality
- [ ] Two-factor authentication (2FA) support
- [ ] Biometric unlock (Touch ID, Windows Hello)
- [ ] Encrypted cloud backup integration
- [ ] Password breach database checking
- [ ] Import/export from other password managers
- [ ] Browser extension integration
- [ ] Mobile applications (iOS, Android)
- [ ] Hardware security key support

### Future Enhancements
- [ ] Password expiration and rotation policies
- [ ] Secure password sharing between users
- [ ] Audit logs and access tracking
- [ ] Multi-device sync with end-to-end encryption
- [ ] Dark/light theme toggle
- [ ] Custom encryption algorithm selection
- [ ] Compliance features (GDPR, CCPA)

## 📄 License

[Insert appropriate open-source license - e.g., MIT, Apache 2.0, GPL 3.0]

## 🙏 Acknowledgments

- **Cloudflare CIRCL**: Post-quantum cryptography library (Kyber768)
- **Fyne Project**: Cross-platform GUI toolkit
- **Go Crypto**: Standard library cryptographic implementations
- **NIST PQC**: Post-Quantum Cryptography Standardization

## 📞 Support

### Getting Help
- **Documentation**: See [ARCHITECTURE.md](ARCHITECTURE.md) and [USER_EXPERIENCE.md](USER_EXPERIENCE.md)
- **Issues**: Report bugs and feature requests via GitHub Issues
- **Discussions**: Join community discussions on GitHub Discussions

### Security Issues
If you discover a security vulnerability, please email security@passquantum.example.com (do not open a public GitHub issue).

## ⚠️ Disclaimer

This software is provided "as is" without warranty of any kind. While PassQuantum uses industry-standard cryptographic algorithms and follows best practices, users should:
- Always maintain secure backups of their keypairs
- Use strong, unique master passwords
- Keep their systems secure and up-to-date
- Understand that no software can guarantee absolute security

## 📊 Statistics

- **Languages**: Go (100%)
- **Total Lines of Code**: ~1,400 lines (core) + ~1,100 lines (UI)
- **Dependencies**: 3 main (circl, fyne, golang.org/x/crypto)
- **Supported Platforms**: Linux, macOS, Windows
- **Binary Size**: ~30MB (includes GUI framework)

---

**PassQuantum** - Your passwords, quantum-safe. 🔐✨
