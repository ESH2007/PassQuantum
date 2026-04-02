package storage

import (
	"fmt"
	"os"

	"passquantum/core/crypto"
	"passquantum/core/model"
)

const DefaultVaultFile = "vault.pqdb"

// WriteVault encrypts and writes password entries to a vault file
// All data is encrypted - no plaintext stored
func WriteVault(entries []*model.PasswordEntry, vaultPath string, encryptionKey []byte, verificationKey []byte, kdfParams crypto.KDFParams) error {
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
// Returns the decrypted password entries
func ReadVault(vaultPath string, encryptionKey []byte, verificationKey []byte) ([]*model.PasswordEntry, error) {
	// Read vault file from disk
	vaultData, err := os.ReadFile(vaultPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*model.PasswordEntry{}, nil
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

	// Parse entries from plaintext
	entries := make([]*model.PasswordEntry, 0)
	idx := 0

	for idx < len(plaintext) {
		// Entry format: ID(8) + ServiceLen(2) + Service + UsernameLen(2) + Username + KyberLen(2) + Kyber + Nonce(12) + CiphertextLen(2) + Ciphertext
		// Minimum entry size: 8 + 2 + 0 + 2 + 0 + 2 + 0 + 12 + 2 + 0 = 30 bytes
		if idx+30 > len(plaintext) {
			break
		}

		// Read Service length at offset 8
		serviceLenPos := idx + 8
		serviceLen := int(plaintext[serviceLenPos])<<8 | int(plaintext[serviceLenPos+1])

		// Read Service
		servicePosStart := idx + 10
		servicePos := servicePosStart + serviceLen

		// Check bounds for Service
		if servicePos > len(plaintext) {
			break
		}

		// Read Username length
		usernameLenPos := servicePos
		if usernameLenPos+2 > len(plaintext) {
			break
		}
		usernameLen := int(plaintext[usernameLenPos])<<8 | int(plaintext[usernameLenPos+1])

		// Read Username
		usernamePosStart := usernameLenPos + 2
		usernamePos := usernamePosStart + usernameLen

		// Check bounds for Username
		if usernamePos+2 > len(plaintext) {
			break
		}

		// Read Kyber length
		kyberLenPos := usernamePos
		kyberLen := int(plaintext[kyberLenPos])<<8 | int(plaintext[kyberLenPos+1])

		// Read Kyber ciphertext
		kyberPosStart := kyberLenPos + 2
		kyberPos := kyberPosStart + kyberLen

		// Check bounds for Kyber
		if kyberPos+12+2 > len(plaintext) {
			break
		}

		// Read Nonce (12 bytes)
		noncePos := kyberPos
		nonceEnd := noncePos + 12

		// Read Ciphertext length
		ciphertextLenPos := nonceEnd
		ciphertextLen := int(plaintext[ciphertextLenPos])<<8 | int(plaintext[ciphertextLenPos+1])

		// Read Ciphertext
		ciphertextPosStart := ciphertextLenPos + 2
		ciphertextPos := ciphertextPosStart + ciphertextLen

		// Check bounds for Ciphertext
		if ciphertextPos > len(plaintext) {
			break
		}

		// Total entry size
		totalEntrySize := ciphertextPos - idx

		entry, err := model.Deserialize(plaintext[idx:ciphertextPos])
		if err != nil {
			break
		}

		entries = append(entries, entry)
		idx += totalEntrySize
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
