# PassQuantum UI Component API Reference

## Quick Navigation
- [Main Components](#main-components)
- [Data Models](#data-models)
- [Helper Functions](#helper-functions)
- [Types & Constants](#types--constants)

---

## Main Components

### 1. LoginScreen (`login_screen.go`)

#### PromptMasterPassword()
```go
func PromptMasterPassword(w fyne.Window, fyneApp fyne.App, appState *AppState)
```
**Purpose:** Display the master password authentication screen

**Parameters:**
- `w` - Fyne window for rendering
- `fyneApp` - Application instance
- `appState` - Mutable application state

**Behavior:**
- Detects existing vaults
- Shows login form with neon styling
- Routes to vault selection or vault creation
- Loads logo image if available

**Example:**
```go
PromptMasterPassword(window, app, appState)
```

---

### 2. VaultSelection (`vault_selection.go`)

#### ShowVaultSelection()
```go
func ShowVaultSelection(w fyne.Window, fyneApp fyne.App, appState *AppState)
```
**Purpose:** Display vault management interface

**Parameters:**
- `w` - Window reference
- `fyneApp` - Application instance
- `appState` - Current app state

**Displays:**
- List of all available vaults
- Create new vault button
- Delete vault button (per vault)
- Settings button
- Lock & Exit button

#### createVaultCard()
```go
func createVaultCard(w fyne.Window, fyneApp fyne.App, appState *AppState, vaultName string) fyne.CanvasObject
```
**Returns:** Styled vault card UI component

---

### 3. MainScreen (`main_screen.go`)

#### ShowMainScreen()
```go
func ShowMainScreen(w fyne.Window, fyneApp fyne.App, appState *AppState)
```
**Purpose:** Display main password manager interface

**Features:**
- Service name input field
- Username/email input field
- Password input field (masked)
- Add Password button
- View All Passwords button
- Back to Vaults button
- Lock & Exit button

**Example Flow:**
```go
// User enters service details
serviceInput.SetText("Gmail")
usernameInput.SetText("user@example.com")
passwordInput.SetText("SecurePassword123!")
// Click Add Password
// Password is encrypted with Kyber + AES-256-GCM and saved
```

---

### 4. PasswordsView (`passwords_view.go`)

#### ShowPasswordsView()
```go
func ShowPasswordsView(w fyne.Window, fyneApp fyne.App, appState *AppState)
```
**Purpose:** Display all passwords in current vault

**Features:**
- Decrypts passwords on-demand
- Shows total password count
- Displays in card format
- Non-blocking UI with goroutines

#### displayPasswordsList()
```go
func displayPasswordsList(w fyne.Window, fyneApp fyne.App, entries []*model.PasswordEntry, appState *AppState)
```
**Parameters:**
- `entries` - Encrypted password entries to display

**Card Features:**
- Service name with index
- Username display
- Masked password with Show/Hide toggle
- Copy to clipboard button
- Delete button

#### createPasswordCard()
```go
func createPasswordCard(index int, service, username, password string) fyne.CanvasObject
```
**Parameters:**
- `index` - Entry number (1-based)
- `service` - Service/website name
- `username` - Associated username
- `password` - Decrypted password

**Returns:** Styled password entry card

---

### 5. SettingsScreen (`settings_screen.go`)

#### ShowSettingsScreen()
```go
func ShowSettingsScreen(w fyne.Window, fyneApp fyne.App, appState *AppState)
```
**Purpose:** Display settings interface with tabs

**Tabs:**
1. Security - Password & session management
2. Vault - Vault maintenance & backup/restore
3. Display - UI customization
4. Backup - Automated backup settings
5. About - Application info

#### buildSecuritySettings()
```go
func buildSecuritySettings(w fyne.Window, fyneApp fyne.App, appState *AppState) *fyne.Container
```
**Returns:** Security settings tab content

**Components:**
- Password strength selector
- Master password change button
- Two-factor authentication toggle
- Session timeout configuration
- Clipboard timeout configuration

#### buildVaultSettings()
```go
func buildVaultSettings(w fyne.Window, fyneApp fyne.App, appState *AppState) *fyne.Container
```
**Returns:** Vault settings tab content

**Components:**
- Current vault info
- Vault statistics
- Compact vault button
- Export vault button
- Import vault button

#### buildDisplaySettings()
```go
func buildDisplaySettings(w fyne.Window, fyneApp fyne.App, appState *AppState) *fyne.Container
```
**Returns:** Display settings tab content

**Components:**
- Theme selector (Dark/Light/System)
- Font size selector
- Password visibility options
- Confirmation settings

#### buildBackupSettings()
```go
func buildBackupSettings(w fyne.Window, fyneApp fyne.App, appState *AppState) *fyne.Container
```
**Returns:** Backup settings tab content

**Components:**
- Auto-backup toggle
- Backup frequency selector
- Manual backup button
- Restore from backup button

#### buildAboutSettings()
```go
func buildAboutSettings(w fyne.Window, fyneApp fyne.App, appState *AppState) *fyne.Container
```
**Returns:** About settings tab content

**Components:**
- Application branding
- Feature list
- Version info
- Documentation link
- Update checker

#### showChangeMasterPasswordDialog()
```go
func showChangeMasterPasswordDialog(w fyne.Window, appState *AppState)
```
**Purpose:** Display master password change dialog

**Inputs:**
- Current password verification
- New password
- New password confirmation

---

## Data Models

### AppState
```go
type AppState struct {
    publicKey       *kyber768.PublicKey      // Kyber public key
    privateKey      *kyber768.PrivateKey     // Kyber private key
    masterPassword  string                   // Current session master password
    encryptionKey   []byte                   // Derived encryption key
    verificationKey []byte                   // Verification/HMAC key
    kdfParams       crypto.KDFParams         // KDF parameters (salt, etc)
    mu              sync.Mutex               // Thread safety lock
    isUnlocked      bool                     // Session state flag
    currentVault    string                   // Currently open vault name
}
```

### PasswordEntry (from core/model)
```go
type PasswordEntry struct {
    ID              uint64  // Unique entry identifier
    Service         string  // Service/website name (NEW)
    Username        string  // Username or email (NEW)
    KyberCiphertext []byte  // Kyber768 encapsulated secret
    Nonce           []byte  // AES-GCM nonce (12 bytes)
    Ciphertext      []byte  // AES-256-GCM encrypted password
}
```

---

## Helper Functions

### Vault Management

#### ListVaults()
```go
func ListVaults() []string
```
**Returns:** List of available vault names from `vaults/` directory

#### GetVaultPath()
```go
func GetVaultPath(vaultName string) string
```
**Returns:** Full file path for vault (e.g., `vaults/Gmail.pqdb`)

#### VaultExists()
```go
func VaultExists(vaultFile string) bool
```
**Returns:** True if any vault exists

#### CreateNewVault()
```go
func CreateNewVault(w interface{}, appState *AppState, masterPassword, vaultName string) bool
```
**Parameters:**
- `w` - Window (for error dialogs)
- `appState` - State to update
- `masterPassword` - Master password for vault
- `vaultName` - Name for new vault

**Returns:** Success status

**Operations:**
- Generates KDF parameters
- Derives encryption keys
- Creates empty vault file
- Updates app state

#### UnlockVault()
```go
func UnlockVault(w interface{}, appState *AppState, masterPassword string) bool
```
**Validates** master password correctness

#### OpenVault()
```go
func OpenVault(w, fyneApp interface{}, appState *AppState, vaultName string)
```
**Purpose:** Open specific vault and show main screen

---

### Storage Operations

#### ReadVault()
```go
func ReadVault(vaultFile string, encKey, verKey []byte) (interface{}, error)
```
**Parameters:**
- `vaultFile` - Path to vault file
- `encKey` - Encryption key
- `verKey` - Verification key

**Returns:** `[]*model.PasswordEntry` or error

**Type Assert:**
```go
entries := vaultData.([]*model.PasswordEntry)
```

#### WriteVault()
```go
func WriteVault(entries []*model.PasswordEntry, vaultFile string, 
                encKey, verKey []byte, kdfParams crypto.KDFParams) error
```
**Parameters:**
- `entries` - Password entries to save
- `vaultFile` - Destination file path
- `encKey` - Encryption key
- `verKey` - Verification key
- `kdfParams` - KDF parameters for vault

**Returns:** Error if operation fails

---

### Cryptographic Operations

#### SaveKeypair()
```go
func SaveKeypair(pubKey *kyber768.PublicKey, privKey *kyber768.PrivateKey, 
                 pubPath, privPath string) error
```

#### LoadKeypair()
```go
func LoadKeypair(pubPath, privPath string) (*kyber768.PublicKey, *kyber768.PrivateKey, error)
```

#### Encapsulate()
```go
func Encapsulate(pubKey *kyber768.PublicKey) ([]byte, []byte, error)
```
**Returns:** 
- Kyber ciphertext
- Shared secret
- Error

#### Decapsulate()
```go
func Decapsulate(ciphertext []byte, privKey *kyber768.PrivateKey) ([]byte, error)
```
**Returns:** Shared secret or error

#### EncryptAES256GCM()
```go
func EncryptAES256GCM(plaintext string, key []byte) ([]byte, []byte, error)
```
**Returns:**
- Nonce
- Ciphertext
- Error

#### DecryptAES256GCM()
```go
func DecryptAES256GCM(nonce, ciphertext, key []byte) (string, error)
```
**Returns:** Decrypted plaintext or error

---

## Types & Constants

### Constants (in main.go)
```go
const (
    pubKeyPath  = "public.key"   // Public key file
    privKeyPath = "private.key"  // Private key file
)
```

### Vault Directory Structure
```
project/
├── vaults/
│   ├── Gmail.pqdb
│   ├── GitHub.pqdb
│   └── Work.pqdb
├── public.key
└── private.key
```

---

## Usage Examples

### Creating a New Vault
```go
CreateNewVault(window, appState, "MyMasterPassword", "MyVault")
ShowVaultSelection(window, app, appState)
```

### Adding a Password
```go
// In main_screen.go add button callback:
entry := model.NewPasswordEntry()
entry.Service = "Gmail"
entry.Username = "user@gmail.com"
entry.KyberCiphertext = ct
entry.Nonce = nonce
entry.Ciphertext = ciphertext

entries = append(entries, entry)
WriteVault(entries, vaultFile, encKey, verKey, kdfParams)
```

### Viewing Passwords
```go
ShowPasswordsView(window, app, appState)
// Internally:
entries, _ := ReadVault(vaultFile, encKey, verKey)
for _, entry := range entries {
    ss, _ := Decapsulate(entry.KyberCiphertext, appState.privateKey)
    password, _ := DecryptAES256GCM(entry.Nonce, entry.Ciphertext, ss)
    // Display password
}
```

---

## Error Handling

All functions that can fail return `error` as last return value.

**Common Error Cases:**
- Master password mismatch
- Vault file not found
- Cryptographic operation failure
- File I/O errors
- Invalid serialized data

**Error Display:**
```go
if err != nil {
    dialog.ShowError(fmt.Errorf("Operation failed: %w", err), window)
}
```

---

## Thread Safety

- `AppState.mu` protects concurrent access
- All vault operations locked during execution
- UI updates on main thread via `fyne.Do()`

```go
appState.mu.Lock()
defer appState.mu.Unlock()
// Safe operations here
```
