package app

import (
"fmt"
"os"
"path/filepath"
"strings"

"github.com/cloudflare/circl/kem/kyber/kyber768"

"passquantum/core/crypto"
"passquantum/core/filevault"
"passquantum/core/model"
"passquantum/core/storage"
securestorage "passquantum/internal/storage"
)

// GetVaultPath returns the file path for a vault with the given name.
func GetVaultPath(vaultName string) string {
vaultName = strings.TrimSpace(vaultName)
preferredPath, preferredErr := securestorage.ResolveVaultPath(vaultName + ".enc")
if preferredErr != nil {
return vaultName + ".enc"
}

legacyPath, legacyErr := securestorage.ResolveVaultPath(vaultName + ".pqdb")
if legacyErr != nil {
return preferredPath
}

if _, err := os.Stat(preferredPath); err == nil {
return preferredPath
}

if _, err := os.Stat(legacyPath); err == nil {
return legacyPath
}

return preferredPath
}

// ListVaults returns a list of all available vault names.
func ListVaults() []string {
var vaults []string
seen := map[string]struct{}{}

vaultsDir, err := securestorage.GetVaultDir()
if err != nil {
return vaults
}

files, err := os.ReadDir(vaultsDir)
if err != nil {
return vaults
}

for _, file := range files {
if file.IsDir() {
continue
}

ext := filepath.Ext(file.Name())
if ext != ".enc" && ext != ".pqdb" {
continue
}

vaultName := strings.TrimSuffix(file.Name(), ext)
if _, exists := seen[vaultName]; exists {
continue
}

seen[vaultName] = struct{}{}
vaults = append(vaults, vaultName)
}

return vaults
}

// VaultExists reports whether any vault file exists.
func VaultExists(vaultFile string) bool {
return len(ListVaults()) > 0
}

// CreateNewVault creates a new vault using the unlocked global master password.
// Returns an error that the caller (UI layer) should display.
func CreateNewVault(appState *AppState, vaultName string) error {
return createVaultWithUnlockedSession(appState, vaultName)
}

// UnlockVault verifies the application master password and starts an unlocked session.
// Returns an error that the caller (UI layer) should display.
func UnlockVault(appState *AppState, masterPassword string) error {
return unlockAppSession(appState, masterPassword)
}

// OpenVault opens a specific vault using the already-unlocked global master password.
// onSuccess is called (on the caller's goroutine) when the vault is opened successfully.
// Returns an error that the caller (UI layer) should display.
func OpenVault(appState *AppState, vaultName string, onSuccess func()) error {
if err := openVaultWithUnlockedSession(appState, vaultName); err != nil {
return err
}
if onSuccess != nil {
onSuccess()
}
return nil
}

// ReadVault reads and decrypts a vault file and returns parsed entries.
func ReadVault(vaultFile string, password string) ([]*model.VaultEntry, error) {
return storage.ReadVault(vaultFile, password)
}

// WriteVault encrypts and writes a vault file.
func WriteVault(entries []*model.VaultEntry, vaultFile string, password string) error {
return storage.WriteVault(entries, vaultFile, password)
}

// Crypto wrappers

func GenerateKeypair() (*kyber768.PublicKey, *kyber768.PrivateKey, error) {
return crypto.GenerateKeypair()
}

func LoadKeypair(pubPath, privPath string) (*kyber768.PublicKey, *kyber768.PrivateKey, error) {
return crypto.LoadKeypair(pubPath, privPath)
}

func SaveKeypair(pubKey *kyber768.PublicKey, privKey *kyber768.PrivateKey, pubPath, privPath string) error {
return crypto.SaveKeypair(pubKey, privKey, pubPath, privPath)
}

func Encapsulate(pubKey *kyber768.PublicKey) ([]byte, []byte, error) {
return crypto.Encapsulate(pubKey)
}

func Decapsulate(ciphertext []byte, privKey *kyber768.PrivateKey) ([]byte, error) {
return crypto.Decapsulate(ciphertext, privKey)
}

func EncryptAES256GCM(plaintext string, key []byte) ([]byte, []byte, error) {
return crypto.EncryptAES256GCM(plaintext, key)
}

func DecryptAES256GCM(nonce, ciphertext, key []byte) (string, error) {
return crypto.DecryptAES256GCM(nonce, ciphertext, key)
}

// PasswordValidationResult holds the result of password validation.
type PasswordValidationResult struct {
Valid        bool
ErrorMessage string
Warnings     []string
}

// ValidatePassword performs comprehensive password validation:
// length > 8, must contain special characters, must not already exist in the vault.
func ValidatePassword(password string, vaultFile string, masterPassword string, privateKey *kyber768.PrivateKey) PasswordValidationResult {
result := PasswordValidationResult{
Valid:    true,
Warnings: []string{},
}

if len(password) <= 8 {
result.Valid = false
result.ErrorMessage = "Password must be longer than 8 characters"
return result
}

specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
hasSpecial := false
for _, char := range password {
for _, special := range specialChars {
if char == special {
hasSpecial = true
break
}
}
if hasSpecial {
break
}
}
if !hasSpecial {
result.Valid = false
result.ErrorMessage = "Password must contain at least one special character (!@#$%^&*...)"
return result
}

entries, err := ReadVault(vaultFile, masterPassword)
if err != nil {
result.Warnings = append(result.Warnings, "Could not check for duplicate passwords in vault")
return result
}

for _, entry := range entries {
ss, err := Decapsulate(entry.KyberCiphertext, privateKey)
if err != nil {
continue
}

plaintext, err := DecryptAES256GCM(entry.Nonce, entry.Ciphertext, ss)
if err != nil {
continue
}

if plaintext == password {
result.Valid = false
result.ErrorMessage = fmt.Sprintf("This password already exists in the vault (used for: %s)", entry.Service)
return result
}
}

return result
}

// InitFileStore creates (or re-opens) the file vault store for the current vault.
// It also ensures a TempTracker exists on the appState.
func InitFileStore(appState *AppState) error {
if appState.CurrentVault == "" || appState.MasterPassword == "" {
return fmt.Errorf("vault must be unlocked before initializing file store")
}
if appState.TempTracker == nil {
appState.TempTracker = filevault.NewTempTracker()
}
store, err := filevault.NewStore(
appState.CurrentVault,
appState.MasterPassword,
appState.PublicKey,
appState.PrivateKey,
appState.TempTracker,
)
if err != nil {
return fmt.Errorf("failed to init file store: %w", err)
}
if appState.FileStore != nil {
appState.FileStore.Close()
}
appState.FileStore = store
return nil
}

// EntriesByType filters vault entries to a single type.
func EntriesByType(entries []*model.VaultEntry, t model.EntryType) []*model.VaultEntry {
var filtered []*model.VaultEntry
for _, e := range entries {
if e.Type == t {
filtered = append(filtered, e)
}
}
return filtered
}
