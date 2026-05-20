package storage

import (
	"fmt"
	"os"

	"passquantum/core/crypto"
	"passquantum/core/model"
	securestorage "passquantum/internal/storage"
)

const DefaultVaultFile = "vault.pqdb"

// WriteVault encrypts and writes vault entries using the PQ vault format.
// The password is the user's master password; all KDF and key material is
// derived internally on each call.
func WriteVault(entries []*model.VaultEntry, vaultPath string, password string) error {
	plaintext := serializeEntries(entries)

	vaultData, err := crypto.PQVaultEncrypt(plaintext, password)
	if err != nil {
		return fmt.Errorf("failed to encrypt vault: %w", err)
	}

	if err := securestorage.WriteVaultFile(vaultPath, vaultData); err != nil {
		return fmt.Errorf("failed to write vault file: %w", err)
	}

	return nil
}

// ReadVault reads and decrypts a vault file.
// It auto-detects the vault format:
//   - "PQVT" magic → PQ vault (Argon2id + Kyber-768 KEM + Dilithium3 signature)
//   - legacy format → auto-migrated on first open (re-encrypted to PQ format in-place)
func ReadVault(vaultPath string, password string) ([]*model.VaultEntry, error) {
	vaultData, err := securestorage.ReadVaultFile(vaultPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*model.VaultEntry{}, nil
		}
		return nil, fmt.Errorf("failed to read vault file: %w", err)
	}

	var plaintext []byte

	if crypto.IsPQVaultFormat(vaultData) {
		plaintext, err = crypto.PQVaultDecrypt(vaultData, password)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt vault: %w", err)
		}
	} else {
		// Legacy format (Argon2id + HMAC-SHA256 + AES-256-GCM).
		// Auto-migrate: decrypt with old scheme, re-encrypt with PQ format, write back.
		plaintext, err = decryptLegacyVault(vaultData, password)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt legacy vault (wrong password?): %w", err)
		}

		// Re-encrypt with the new PQ format and write back immediately.
		newData, migrateErr := crypto.PQVaultEncrypt(plaintext, password)
		if migrateErr == nil {
			_ = securestorage.WriteVaultFile(vaultPath, newData) // best-effort; don't fail the read
		}
	}

	entries, err := deserializeEntries(plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to parse vault entries: %w", err)
	}

	return entries, nil
}

// decryptLegacyVault decrypts a vault file written in the pre-PQ format:
//
//	Version(1) | KDFParamsLen(1) | KDFParams(26) | HMAC(32) | EncDataLen(4) | EncData
func decryptLegacyVault(data []byte, password string) ([]byte, error) {
	vault, err := crypto.VaultFileDeserialize(data)
	if err != nil {
		return nil, fmt.Errorf("invalid legacy vault header: %w", err)
	}

	encKey, verKey, err := crypto.DeriveKeys(password, vault.KDFParams)
	if err != nil {
		return nil, fmt.Errorf("key derivation failed: %w", err)
	}

	plaintext, err := crypto.DecryptVault(vault, encKey, verKey)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

// VaultExists checks if the vault file exists.
func VaultExists(vaultPath string) bool {
	resolvedPath, err := securestorage.ResolveVaultPath(vaultPath)
	if err != nil {
		return false
	}

	_, err = os.Stat(resolvedPath)
	return err == nil
}

// DeleteVault removes the vault file (destructive — data loss!).
func DeleteVault(vaultPath string) error {
	resolvedPath, err := securestorage.ResolveVaultPath(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to resolve vault path: %w", err)
	}

	if err := os.Remove(resolvedPath); err != nil {
		return fmt.Errorf("failed to remove vault file %q: %w", resolvedPath, err)
	}

	return nil
}
