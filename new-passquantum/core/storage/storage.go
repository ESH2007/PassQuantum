package storage

import (
	"fmt"
	"os"

	"passquantum/core/crypto"
	"passquantum/core/model"
)

const DefaultVaultFile = "vault.pqdb"

// WriteVault encrypts and writes vault entries to a vault file.
// All data is encrypted - no plaintext stored
func WriteVault(entries []*model.VaultEntry, vaultPath string, encryptionKey []byte, verificationKey []byte, kdfParams crypto.KDFParams) error {
	plaintext := serializeEntries(entries)

	// Encrypt the vault
	vault, err := crypto.EncryptVault(plaintext, encryptionKey, verificationKey, kdfParams)
	if err != nil {
		return fmt.Errorf("failed to encrypt vault: %w", err)
	}

	// Write to disk
	vaultData := vault.Serialize()
	err = os.WriteFile(vaultPath, vaultData, 0600)
	if err != nil {
		return fmt.Errorf("failed to write vault file: %w", err)
	}

	return nil
}

// ReadVault reads and decrypts a vault file
// Returns the decrypted vault entries.
func ReadVault(vaultPath string, encryptionKey []byte, verificationKey []byte) ([]*model.VaultEntry, error) {
	// Read vault file from disk
	vaultData, err := os.ReadFile(vaultPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*model.VaultEntry{}, nil
		}
		return nil, fmt.Errorf("failed to read vault file: %w", err)
	}

	// Deserialize vault file
	vault, err := crypto.VaultFileDeserialize(vaultData)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize vault: %w", err)
	}

	// Decrypt vault
	plaintext, err := crypto.DecryptVault(vault, encryptionKey, verificationKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt vault: %w", err)
	}

	entries, err := deserializeEntries(plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to parse vault entries: %w", err)
	}

	return entries, nil
}

// VaultExists checks if the vault file exists
func VaultExists(vaultPath string) bool {
	_, err := os.Stat(vaultPath)
	return err == nil
}

// DeleteVault removes the vault file (careful - data loss!)
func DeleteVault(vaultPath string) error {
	return os.Remove(vaultPath)
}
