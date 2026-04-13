package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"passquantum/core/crypto"
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

// ReencryptVaultFile decrypts a vault with the current password and returns its re-encrypted bytes.
func ReencryptVaultFile(vaultPath string, currentPassword string, newPassword string) ([]byte, crypto.KDFParams, error) {
	vaultData, err := os.ReadFile(vaultPath)
	if err != nil {
		return nil, crypto.KDFParams{}, fmt.Errorf("failed to read vault file: %w", err)
	}

	vault, err := crypto.VaultFileDeserialize(vaultData)
	if err != nil {
		return nil, crypto.KDFParams{}, fmt.Errorf("failed to deserialize vault: %w", err)
	}

	currentEncryptionKey, currentVerificationKey, err := crypto.DeriveKeys(currentPassword, vault.KDFParams)
	if err != nil {
		return nil, crypto.KDFParams{}, fmt.Errorf("failed to derive current vault keys: %w", err)
	}

	entries, err := ReadVault(vaultPath, currentEncryptionKey, currentVerificationKey)
	if err != nil {
		return nil, crypto.KDFParams{}, fmt.Errorf("failed to decrypt vault with current password: %w", err)
	}

	newParams := crypto.DefaultKDFParams()
	newParams.Salt, err = crypto.GenerateSalt()
	if err != nil {
		return nil, crypto.KDFParams{}, fmt.Errorf("failed to generate new vault salt: %w", err)
	}

	newEncryptionKey, newVerificationKey, err := crypto.DeriveKeys(newPassword, newParams)
	if err != nil {
		return nil, crypto.KDFParams{}, fmt.Errorf("failed to derive new vault keys: %w", err)
	}

	plaintext := serializeEntries(entries)
	reencryptedVault, err := crypto.EncryptVault(plaintext, newEncryptionKey, newVerificationKey, newParams)
	if err != nil {
		return nil, crypto.KDFParams{}, fmt.Errorf("failed to re-encrypt vault: %w", err)
	}

	return reencryptedVault.Serialize(), newParams, nil
}

func resolveSecurityMetadataPath(path string) string {
	if path == "" {
		return DefaultAppSecurityMetadataPath
	}

	return path
}
