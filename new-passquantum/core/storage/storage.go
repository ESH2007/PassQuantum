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
	// Serialize all entries into binary format
	plaintext := make([]byte, 0)
	for _, entry := range entries {
		plaintext = append(plaintext, entry.Serialize()...)
	}

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
		// Minimum entry size: 8 + 2 + 0 + 12 + 2 = 24 bytes
		if idx+24 > len(plaintext) {
			break
		}

		// Try to read an entry
		// Entry format: ID(8) + KyberLen(2) + Kyber(variable) + Nonce(12) + CiphertextLen(2) + Ciphertext(variable)
		// All multi-byte values are in big-endian format

		// Read KyberLen in big-endian format at offset 8
		kyberLen := int(plaintext[idx+8])<<8 | int(plaintext[idx+9])

		// Position after ID and KyberLen and Kyber data
		posAfterKyber := idx + 8 + 2 + kyberLen

		// Position of Nonce is right after Kyber data
		posNonce := posAfterKyber
		posAfterNonce := posNonce + 12

		// Position of CiphertextLen is after Nonce
		posCiphertextLen := posAfterNonce

		// Check if we have room for CiphertextLen (2 bytes)
		if posCiphertextLen+2 > len(plaintext) {
			break
		}

		// Read CiphertextLen in big-endian format
		ciphertextLen := int(plaintext[posCiphertextLen])<<8 | int(plaintext[posCiphertextLen+1])

		// Total entry size: ID(8) + KyberLen(2) + Kyber(kyberLen) + Nonce(12) + CiphertextLen(2) + Ciphertext(ciphertextLen)
		totalEntrySize := 8 + 2 + kyberLen + 12 + 2 + ciphertextLen

		// Check if we have the full entry
		if idx+totalEntrySize > len(plaintext) {
			break
		}

		// Extract and parse this entry
		entryData := plaintext[idx : idx+totalEntrySize]
		entry, err := model.Deserialize(entryData)
		if err != nil {
			// Skip malformed entries
			fmt.Fprintf(os.Stderr, "warning: skipped malformed entry: %v\n", err)
			idx += totalEntrySize
			continue
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
