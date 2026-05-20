package storage

import (
	"path/filepath"
	"testing"

	"passquantum/core/crypto"
	"passquantum/core/model"
	securestorage "passquantum/internal/storage"
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
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, filepath.Base(tempDir)+"-vault.pqdb")

	entry := model.NewVaultEntry()
	entry.Service = "GitHub"
	entry.Username = "alice@example.com"
	entry.KyberCiphertext = []byte{1, 2, 3, 4}
	entry.Nonce = []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	entry.Ciphertext = []byte{9, 8, 7, 6, 5}

	if err := WriteVault([]*model.VaultEntry{entry}, vaultPath, "old-password"); err != nil {
		t.Fatalf("WriteVault() error = %v", err)
	}

	rotatedData, err := ReencryptVaultFile(vaultPath, "old-password", "new-password")
	if err != nil {
		t.Fatalf("ReencryptVaultFile() error = %v", err)
	}

	if err := securestorage.WriteVaultFile(vaultPath, rotatedData); err != nil {
		t.Fatalf("WriteFile() rotated vault error = %v", err)
	}

	if _, err := ReadVault(vaultPath, "old-password"); err == nil {
		t.Fatal("ReadVault() with original password succeeded after rotation, want failure")
	}

	entries, err := ReadVault(vaultPath, "new-password")
	if err != nil {
		t.Fatalf("ReadVault() with new password error = %v", err)
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
