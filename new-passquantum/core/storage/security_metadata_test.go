package storage

import (
	"os"
	"path/filepath"
	"testing"

	"passquantum/core/crypto"
	"passquantum/core/model"
)

func TestSaveAndLoadAppSecurityProfile(t *testing.T) {
	_, privateKey, err := crypto.GenerateKeypair()
	if err != nil {
		t.Fatalf("GenerateKeypair() error = %v", err)
	}

	profile, _, _, err := crypto.CreateAppSecurityProfile("metadata-password", privateKey)
	if err != nil {
		t.Fatalf("CreateAppSecurityProfile() error = %v", err)
	}

	profilePath := filepath.Join(t.TempDir(), "app-security-test.json")
	if err := SaveAppSecurityProfile(profilePath, profile); err != nil {
		t.Fatalf("SaveAppSecurityProfile() error = %v", err)
	}

	loadedProfile, err := LoadAppSecurityProfile(profilePath)
	if err != nil {
		t.Fatalf("LoadAppSecurityProfile() error = %v", err)
	}

	if loadedProfile.FormatVersion != profile.FormatVersion {
		t.Fatalf("LoadAppSecurityProfile() format version = %d, want %d", loadedProfile.FormatVersion, profile.FormatVersion)
	}

	if string(loadedProfile.PrivateKeyFingerprint) != string(profile.PrivateKeyFingerprint) {
		t.Fatal("LoadAppSecurityProfile() fingerprint mismatch")
	}

	if string(loadedProfile.Verifier) != string(profile.Verifier) {
		t.Fatal("LoadAppSecurityProfile() verifier mismatch")
	}
}

func TestReencryptVaultFile(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "vault.pqdb")

	originalParams := crypto.DefaultKDFParams()
	var err error
	originalParams.Salt, err = crypto.GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() error = %v", err)
	}

	originalEncryptionKey, originalVerificationKey, err := crypto.DeriveKeys("old-password", originalParams)
	if err != nil {
		t.Fatalf("DeriveKeys() original error = %v", err)
	}

	entry := model.NewPasswordEntry()
	entry.Service = "GitHub"
	entry.Username = "alice@example.com"
	entry.KyberCiphertext = []byte{1, 2, 3, 4}
	entry.Nonce = []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	entry.Ciphertext = []byte{9, 8, 7, 6, 5}

	if err := WriteVault([]*model.PasswordEntry{entry}, vaultPath, originalEncryptionKey, originalVerificationKey, originalParams); err != nil {
		t.Fatalf("WriteVault() error = %v", err)
	}

	rotatedData, rotatedParams, err := ReencryptVaultFile(vaultPath, "old-password", "new-password")
	if err != nil {
		t.Fatalf("ReencryptVaultFile() error = %v", err)
	}

	if err := os.WriteFile(vaultPath, rotatedData, 0600); err != nil {
		t.Fatalf("WriteFile() rotated vault error = %v", err)
	}

	if _, err := ReadVault(vaultPath, originalEncryptionKey, originalVerificationKey); err == nil {
		t.Fatal("ReadVault() with original password succeeded after rotation, want failure")
	}

	rotatedEncryptionKey, rotatedVerificationKey, err := crypto.DeriveKeys("new-password", rotatedParams)
	if err != nil {
		t.Fatalf("DeriveKeys() rotated error = %v", err)
	}

	entries, err := ReadVault(vaultPath, rotatedEncryptionKey, rotatedVerificationKey)
	if err != nil {
		t.Fatalf("ReadVault() with rotated password error = %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("ReadVault() entries length = %d, want 1", len(entries))
	}

	if entries[0].Service != entry.Service || entries[0].Username != entry.Username {
		t.Fatal("ReadVault() entry metadata changed during rotation")
	}

	if string(entries[0].Ciphertext) != string(entry.Ciphertext) {
		t.Fatal("ReadVault() ciphertext changed during rotation")
	}
}

// TestSaveAndLoadProfileWithBiometricFields verifies that biometric settings and
// template bytes survive a save/load round-trip through JSON.
func TestSaveAndLoadProfileWithBiometricFields(t *testing.T) {
	_, privateKey, err := crypto.GenerateKeypair()
	if err != nil {
		t.Fatalf("GenerateKeypair() error = %v", err)
	}

	profile, _, _, err := crypto.CreateAppSecurityProfile("biometric-test-password", privateKey)
	if err != nil {
		t.Fatalf("CreateAppSecurityProfile() error = %v", err)
	}

	// Attach biometric data.
	profile.Biometric = crypto.BiometricSettings{Enabled: true, Threshold: 0.97}
	profile.BiometricTemplate = []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	profilePath := filepath.Join(t.TempDir(), "biometric-profile.json")
	if err := SaveAppSecurityProfile(profilePath, profile); err != nil {
		t.Fatalf("SaveAppSecurityProfile() error = %v", err)
	}

	loaded, err := LoadAppSecurityProfile(profilePath)
	if err != nil {
		t.Fatalf("LoadAppSecurityProfile() error = %v", err)
	}

	if !loaded.Biometric.Enabled {
		t.Fatal("LoadAppSecurityProfile() Biometric.Enabled = false, want true")
	}

	if loaded.Biometric.Threshold != profile.Biometric.Threshold {
		t.Fatalf("LoadAppSecurityProfile() Biometric.Threshold = %v, want %v",
			loaded.Biometric.Threshold, profile.Biometric.Threshold)
	}

	if string(loaded.BiometricTemplate) != string(profile.BiometricTemplate) {
		t.Fatal("LoadAppSecurityProfile() BiometricTemplate mismatch")
	}
}

// TestLoadLegacyV1ProfileNoBiometricFields ensures that a v1 profile (without biometric
// fields) loads correctly with biometric disabled as the default.
func TestLoadLegacyV1ProfileNoBiometricFields(t *testing.T) {
	// Construct a minimal v1-style JSON without any biometric keys.
	v1JSON := `{
		"format_version": 1,
		"private_key_fingerprint": "AAEC",
		"kdf_params": {
			"salt": "AAEC",
			"memory": 65536,
			"iterations": 1,
			"parallelism": 4
		},
		"verifier": "AAEC"
	}`

	profilePath := filepath.Join(t.TempDir(), "v1-profile.json")
	if err := os.WriteFile(profilePath, []byte(v1JSON), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	loaded, err := LoadAppSecurityProfile(profilePath)
	if err != nil {
		t.Fatalf("LoadAppSecurityProfile(v1) error = %v", err)
	}

	if loaded.Biometric.Enabled {
		t.Fatal("LoadAppSecurityProfile(v1) Biometric.Enabled = true, want false (default)")
	}

	if loaded.BiometricTemplate != nil {
		t.Fatal("LoadAppSecurityProfile(v1) BiometricTemplate != nil, want nil (default)")
	}
}
