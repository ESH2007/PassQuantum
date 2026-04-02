# PassQuantum UI Refactoring & Multi-Vault Implementation

## Overview
The PassQuantum UI has been completely refactored into modular, reusable components. The application now supports multiple named vaults, enhanced password management, and a comprehensive settings system.

## Project Structure

### New File Organization
```
ui/
├── main.go                 # Application entry point & AppState
├── login_screen.go         # Master password authentication UI
├── vault_selection.go      # Vault management & selection
├── main_screen.go          # Main password manager interface
├── passwords_view.go       # Password display & management
├── settings_screen.go      # Application settings
└── helpers.go              # Utility functions & wrappers
```

### Key Components

#### 1. **main.go** - Core Application
- `AppState` struct: Holds encryption keys, vault info, and session state
- `initializeApp()`: Initializes Kyber keypair (generates or loads)
- Application entry point with window setup

**Changes:**
- Added `currentVault` field to track active vault
- Simplified keypair initialization

#### 2. **login_screen.go** - Authentication
- Beautiful neon-styled login interface
- Master password validation
- Auto-detects existing vaults vs new setup

**Features:**
- Logo loading with fallback
- Password strength visual feedback
- Seamless transition to vault selection

#### 3. **vault_selection.go** - Multi-Vault Management
- Display all available vaults
- Create new vault with custom name
- Delete vault with confirmation
- Switch between vaults

**Key Functions:**
```go
ListVaults() []string                    // Get all vault names
GetVaultPath(name string) string         // Resolve vault file path
ShowVaultSelection(w, fyneApp, appState) // Display vault UI
createVaultCard(...)                     // Individual vault UI component
showCreateVaultDialog(...)               // New vault form
showDeleteVaultDialog(...)               // Safe deletion dialog
```

#### 4. **main_screen.go** - Password Manager
- Add new password entries with:
  - Service name (e.g., "Gmail", "GitHub")
  - Username/Email
  - Password
- View all passwords in current vault
- Lock vault and exit

**Enhanced Features:**
- Service & username metadata storage
- Kyber + AES-256-GCM hybrid encryption
- Goroutine-based non-blocking encryption

#### 5. **passwords_view.go** - Password Display
- Encrypted password listing in card format
- Individual password cards with:
  - Service name and metadata
  - Show/Hide password toggle
  - Copy to clipboard button
  - Delete password option
- Beautiful card-based layout

#### 6. **settings_screen.go** - Comprehensive Settings
Rich settings interface organized into tabs:

**Security Tab:**
- Password strength requirement selector
- Change master password
- Two-factor authentication (future)
- Session auto-lock timeout
- Clipboard clear timeout

**Vault Tab:**
- Current vault info
- Vault statistics
- Vault compaction/optimization
- Encrypted vault export/import

**Display Tab:**
- Theme selection (Dark/Light/System)
- Font size adjustment
- Password visibility options
- Action confirmation settings

**Backup Tab:**
- Automatic backup toggle
- Backup frequency selector
- Manual backup button
- Restore from backup
- Cloud backup (future)

**About Tab:**
- Application branding
- Version info
- Feature list
- Developer credit
- Documentation link
- Update checker

#### 7. **helpers.go** - Utility Layer
Vault and crypto helper functions:

```go
// Vault Management
ListVaults()
GetVaultPath(name)
VaultExists()
CreateNewVault(password, name)
UnlockVault(password)
OpenVault(name)

// Storage Operations
ReadVault(file, encKey, verKey)
WriteVault(entries, file, encKey, verKey, kdfParams)

// Crypto Wrappers
SaveKeypair / LoadKeypair
Encapsulate / Decapsulate
EncryptAES256GCM / DecryptAES256GCM
```

## Enhanced Data Model

### Updated PasswordEntry Structure
```go
type PasswordEntry struct {
    ID                uint64  // Unique identifier
    Service          string  // Service name (NEW)
    Username         string  // Username/email (NEW)
    KyberCiphertext  []byte  // Kyber encapsulated secret
    Nonce            []byte  // AES-GCM nonce
    Ciphertext       []byte  // Encrypted password
}
```

**Serialization Format:**
- ID (8 bytes)
- Service length + data (variable)
- Username length + data (variable)
- Kyber ciphertext length + data (variable)
- Nonce (12 bytes)
- Ciphertext length + data (variable)

## Multi-Vault Architecture

### Vault Storage
- Vaults stored in: `vaults/` directory
- File naming: `{vaultName}.pqdb`
- Each vault is independently encrypted with its own master password

### Vault Workflow
```
1. User launches app → Login Screen
2. Master password entry → Vault Selection
3. Vault list displayed (auto-detected from vaults/)
4. Select vault → Password Manager Main Screen
5. Add/View/Delete passwords
6. Settings accessible from vault selection
7. Lock vault → Back to login
```

### Master Password Handling
- Master password derived to encryption key via Argon2id KDF
- Separate KDF params for each vault
- Password validation on vault unlock

## UI Flow

```
Start
  ↓
[Login Screen]
  ↓ Master Password
[Vault Selection]
  ├→ Create New Vault
  ├→ Delete Vault
  ├→ Settings ← Comprehensive options
  └→ Select Vault
      ↓
  [Main Screen]
  ├→ Add Password
  ├→ View Passwords ← [Passwords View]
  │                    ├→ Show/Hide
  │                    ├→ Copy
  │                    └→ Delete
  ├→ Back to Vaults
  └→ Lock & Exit
```

## Security Features

1. **Post-Quantum Cryptography**
   - Kyber-768 for hybrid encryption
   - No quantum threat

2. **Strong Encryption**
   - AES-256-GCM for password encryption
   - Unique nonce per entry

3. **Key Derivation**
   - Argon2id for master password → key
   - Configurable KDF parameters
   - Salt generation per vault

4. **Session Management**
   - Optional auto-lock timeout
   - Clipboard auto-clear
   - Encrypted in-memory storage

## Future Enhancement Ideas

1. **Advanced Features**
   - Password generation with strength meter
   - Breach detection integration
   - Two-factor authentication support
   - Cloud backup synchronization

2. **UI Improvements**
   - Search/filter passwords
   - Password categorization/tags
   - Dark mode with custom themes
   - Password strength meter

3. **Security**
   - Biometric unlock support
   - Master password recovery
   - Vault sharing features
   - Activity logging

4. **Export/Import**
   - Multiple format support
   - Scheduled backups
   - Vault migration

## Compilation

The refactored code compiles successfully:
```bash
cd /home/lenovo/dev/PassQuantum/new-passquantum/ui
go build -o passquantum
```

## Dependencies

- `fyne.io/fyne/v2` - GUI framework
- `github.com/cloudflare/circl` - Kyber implementation
- `passquantum/core/crypto` - Encryption logic
- `passquantum/core/model` - Data models
- `passquantum/core/storage` - Vault I/O

## Testing

Run the application:
```bash
./passquantum
```

### Test Scenarios
1. Create new vault
2. Add multiple passwords
3. Switch between vaults
4. View and decrypt passwords
5. Delete passwords
6. Change master password (via settings)
7. Lock and re-unlock vault

## Notes

- Each vault operates independently
- Vaults are automatically discovered from filesystem
- Settings are per-session (local application state)
- Master passwords are never stored
- All passwords encrypted before storage
- Non-blocking UI with goroutines for crypto operations
