package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"passquantum/core/crypto"
	securestorage "passquantum/internal/storage"
)

const DefaultAppSecurityMetadataPath = "app-security.pqmeta"

func AppSecurityProfileExists(path string) bool {
	_, err := os.Stat(resolveSecurityMetadataPath(path))
	return err == nil
}

func LoadAppSecurityProfile(path string) (*crypto.AppSecurityProfile, error) {
	profilePath := resolveSecurityMetadataPath(path)
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return nil, err
	}

	var profile crypto.AppSecurityProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to decode app security profile: %w", err)
	}

	return &profile, nil
}

func SaveAppSecurityProfile(path string, profile *crypto.AppSecurityProfile) error {
	if profile == nil {
		return fmt.Errorf("security profile is required")
	}

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode app security profile: %w", err)
	}

	if err := os.WriteFile(resolveSecurityMetadataPath(path), data, 0600); err != nil {
		return fmt.Errorf("failed to write app security profile: %w", err)
	}

	return nil
}

// ReencryptVaultFile decrypts a vault with currentPassword and returns the
// bytes of the same vault re-encrypted with newPassword using the PQ format.
// The caller is responsible for atomically replacing the vault file.
func ReencryptVaultFile(vaultPath string, currentPassword string, newPassword string) ([]byte, error) {
	entries, err := ReadVault(vaultPath, currentPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt vault with current password: %w", err)
	}

	plaintext := serializeEntries(entries)

	newData, err := crypto.PQVaultEncrypt(plaintext, newPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to re-encrypt vault: %w", err)
	}

	return newData, nil
}

func resolveSecurityMetadataPath(path string) string {
	name := path
	if name == "" {
		name = DefaultAppSecurityMetadataPath
	}

	if filepath.IsAbs(name) {
		return name
	}

	if resolved, err := securestorage.GetSecureFilePath(name); err == nil {
		return resolved
	}

	return name
}
