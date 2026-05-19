package storage

import (
	"path/filepath"
	"testing"

	"passquantum/core/model"
	securestorage "passquantum/internal/storage"
)

func TestWriteReadVaultTypedFormatRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, filepath.Base(tempDir)+"-typed-roundtrip.pqdb")
	password := "typed-pass"

	passwordEntry := model.NewVaultEntry()
	passwordEntry.Type = model.EntryTypePassword
	passwordEntry.Service = "GitHub"
	passwordEntry.Username = "alice@example.com"
	passwordEntry.KyberCiphertext = []byte{1, 2, 3}
	passwordEntry.Nonce = []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	passwordEntry.Ciphertext = []byte("enc-password")

	noteEntry := model.NewVaultEntry()
	noteEntry.Type = model.EntryTypeNote
	noteEntry.Service = "NOTE:Server"
	noteEntry.Username = "note"
	noteEntry.KyberCiphertext = []byte{4, 5, 6}
	noteEntry.Nonce = []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	noteEntry.Ciphertext = []byte("enc-note")

	cardEntry := model.NewVaultEntry()
	cardEntry.Type = model.EntryTypeCard
	cardEntry.CardSubtype = "credit"
	cardEntry.Service = "CARD:Main"
	cardEntry.Username = "Credit"
	cardEntry.KyberCiphertext = []byte{7, 8, 9}
	cardEntry.Nonce = []byte{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3}
	cardEntry.Ciphertext = []byte("enc-card")

	entries := []*model.VaultEntry{passwordEntry, noteEntry, cardEntry}
	if err := WriteVault(entries, vaultPath, password); err != nil {
		t.Fatalf("WriteVault() error = %v", err)
	}

	loaded, err := ReadVault(vaultPath, password)
	if err != nil {
		t.Fatalf("ReadVault() error = %v", err)
	}
	if len(loaded) != 3 {
		t.Fatalf("ReadVault() entries length = %d, want 3", len(loaded))
	}

	if loaded[1].Type != model.EntryTypeNote {
		t.Fatalf("note entry type = %d, want %d", loaded[1].Type, model.EntryTypeNote)
	}
	if loaded[2].Type != model.EntryTypeCard {
		t.Fatalf("card entry type = %d, want %d", loaded[2].Type, model.EntryTypeCard)
	}
	if loaded[2].CardSubtype != "credit" {
		t.Fatalf("card subtype = %q, want %q", loaded[2].CardSubtype, "credit")
	}
}

// TestReadVaultLegacyFormatReturnsError verifies that files without the PQVT magic
// header are rejected with an error requiring re-encryption.
func TestReadVaultLegacyFormatReturnsError(t *testing.T) {
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, filepath.Base(tempDir)+"-legacy-format.pqdb")
	if err := securestorage.WriteVaultFile(vaultPath, []byte("OLDFORMAT_NOT_PQVT")); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := ReadVault(vaultPath, "any-password")
	if err == nil {
		t.Fatal("ReadVault() with legacy format should return an error, got nil")
	}
}
