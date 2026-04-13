package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudflare/circl/kem/kyber/kyber768"

	"fyne.io/fyne/v2"

	"passquantum/core/crypto"
	"passquantum/core/model"
	"passquantum/core/storage"
)

// VaultHelper functions for managing multiple vaults

// GetVaultPath returns the file path for a vault with the given name
func GetVaultPath(vaultName string) string {
	vaultsDir := "vaults"
	// Create vaults directory if it doesn't exist
	os.MkdirAll(vaultsDir, 0755)
	return filepath.Join(vaultsDir, vaultName+".pqdb")
}

// ListVaults returns a list of all available vault names
func ListVaults() []string {
	vaultsDir := "vaults"
	var vaults []string

	// Create directory if it doesn't exist
	if _, err := os.Stat(vaultsDir); os.IsNotExist(err) {
		os.MkdirAll(vaultsDir, 0755)
		return vaults
	}

	files, err := ioutil.ReadDir(vaultsDir)
	if err != nil {
		return vaults
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".pqdb" {
			// Remove .pqdb extension to get vault name
			vaultName := file.Name()[:len(file.Name())-5]
			vaults = append(vaults, vaultName)
		}
	}

	return vaults
}

// VaultExists checks if any vault exists
func VaultExists(vaultFile string) bool {
	vaults := ListVaults()
	return len(vaults) > 0
}

// CreateNewVault creates a new vault using the unlocked global master password.
func CreateNewVault(w interface{}, appState *AppState, vaultName string) bool {
	if err := createVaultWithUnlockedSession(appState, vaultName); err != nil {
		showWindowError(w, err)
		return false
	}

	return true
}

// UnlockVault verifies the application master password and starts an unlocked session.
func UnlockVault(w interface{}, appState *AppState, masterPassword string) bool {
	if err := unlockAppSession(appState, masterPassword); err != nil {
		showWindowError(w, err)
		return false
	}

	return true
}

// OpenVault opens a specific vault using the already-unlocked global master password.
func OpenVault(w interface{}, fyneApp interface{}, appState *AppState, vaultName string, onSuccess func()) {
	win, _ := w.(fyne.Window)
	app, _ := fyneApp.(fyne.App)

	if err := openVaultWithUnlockedSession(appState, vaultName); err != nil {
		showWindowError(w, err)
		return
	}

	if onSuccess != nil {
		onSuccess()
		return
	}

	if win != nil && app != nil {
		ShowMainScreen(win, app, appState)
	}
}

// Storage wrappers that use the core storage package

// ReadVault reads and decrypts a vault file and returns parsed entries
func ReadVault(vaultFile string, encryptionKey, verificationKey []byte) ([]*model.VaultEntry, error) {
	return storage.ReadVault(vaultFile, encryptionKey, verificationKey)
}

// WriteVault encrypts and writes a vault file
func WriteVault(entries []*model.VaultEntry, vaultFile string, encryptionKey, verificationKey []byte, kdfParams crypto.KDFParams) error {
	return storage.WriteVault(entries, vaultFile, encryptionKey, verificationKey, kdfParams)
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

// PasswordValidationResult holds the result of password validation
type PasswordValidationResult struct {
	Valid        bool
	ErrorMessage string
	Warnings     []string
}

// ValidatePassword performs comprehensive password validation
// Rules:
// 1. Length must be > 8 characters
// 2. Must contain special characters
// 3. Must NOT already exist in the vault
func ValidatePassword(password string, vaultFile string, encryptionKey, verificationKey []byte, privateKey *kyber768.PrivateKey) PasswordValidationResult {
	result := PasswordValidationResult{
		Valid:    true,
		Warnings: []string{},
	}

	// Rule 1: Check length > 8
	if len(password) <= 8 {
		result.Valid = false
		result.ErrorMessage = "Password must be longer than 8 characters"
		return result
	}

	// Rule 2: Must contain special characters
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

	// Rule 3: Must NOT already exist in the vault
	entries, err := ReadVault(vaultFile, encryptionKey, verificationKey)
	if err != nil {
		// If we can't read the vault, we can't check for duplicates
		// But we don't want to block the user, so we just warn
		result.Warnings = append(result.Warnings, "Could not check for duplicate passwords in vault")
		return result
	}

	// Decrypt all passwords and check for duplicates
	for _, entry := range entries {
		ss, err := Decapsulate(entry.KyberCiphertext, privateKey)
		if err != nil {
			continue // Skip entries we can't decrypt
		}

		plaintext, err := DecryptAES256GCM(entry.Nonce, entry.Ciphertext, ss)
		if err != nil {
			continue // Skip entries we can't decrypt
		}

		if plaintext == password {
			result.Valid = false
			result.ErrorMessage = fmt.Sprintf("This password already exists in the vault (used for: %s)", entry.Service)
			return result
		}
	}

	return result
}
