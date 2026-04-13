package storage

import (
	"os"
	"path/filepath"
	"testing"

	"passquantum/core/crypto"
	"passquantum/core/model"
)

func TestWriteReadVaultTypedFormatRoundTrip(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "typed-roundtrip.pqdb")
	params, encKey, verKey := mustDeriveVaultKeys(t, "typed-pass")

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
	if err := WriteVault(entries, vaultPath, encKey, verKey, params); err != nil {
		t.Fatalf("WriteVault() error = %v", err)
	}

	loaded, err := ReadVault(vaultPath, encKey, verKey)
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

func TestReadVaultLegacyFormatAutoMigration(t *testing.T) {
	vaultPath := filepath.Join(t.TempDir(), "legacy-format.pqdb")
	params, encKey, verKey := mustDeriveVaultKeys(t, "legacy-pass")

	legacyNote := model.NewVaultEntry()
	legacyNote.Service = "NOTE:Legacy"
	legacyNote.Username = "note"
	legacyNote.KyberCiphertext = []byte{11, 12}
	legacyNote.Nonce = []byte{4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4}
	legacyNote.Ciphertext = []byte("legacy-note")

	legacyCard := model.NewVaultEntry()
	legacyCard.Service = "CARD:Legacy"
	legacyCard.Username = "Credit"
	legacyCard.KyberCiphertext = []byte{13, 14}
	legacyCard.Nonce = []byte{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5}
	legacyCard.Ciphertext = []byte("legacy-card")

	legacyPlaintext := append(legacyNote.Serialize(), legacyCard.Serialize()...)
	vault, err := crypto.EncryptVault(legacyPlaintext, encKey, verKey, params)
	if err != nil {
		t.Fatalf("EncryptVault() error = %v", err)
	}
	if err := os.WriteFile(vaultPath, vault.Serialize(), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	loaded, err := ReadVault(vaultPath, encKey, verKey)
	if err != nil {
		t.Fatalf("ReadVault() legacy error = %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("ReadVault() legacy entries length = %d, want 2", len(loaded))
	}

	if loaded[0].Type != model.EntryTypeNote {
		t.Fatalf("legacy note type = %d, want %d", loaded[0].Type, model.EntryTypeNote)
	}
	if loaded[1].Type != model.EntryTypeCard {
		t.Fatalf("legacy card type = %d, want %d", loaded[1].Type, model.EntryTypeCard)
	}
	if loaded[1].CardSubtype != "credit" {
		t.Fatalf("legacy card subtype = %q, want %q", loaded[1].CardSubtype, "credit")
	}

	// First save should rewrite to typed format (auto-migration on save).
	if err := WriteVault(loaded, vaultPath, encKey, verKey, params); err != nil {
		t.Fatalf("WriteVault() auto-migration save error = %v", err)
	}

	loadedAgain, err := ReadVault(vaultPath, encKey, verKey)
	if err != nil {
		t.Fatalf("ReadVault() after migration save error = %v", err)
	}
	if len(loadedAgain) != 2 {
		t.Fatalf("ReadVault() after migration length = %d, want 2", len(loadedAgain))
	}
	if loadedAgain[1].Type != model.EntryTypeCard {
		t.Fatalf("migrated card type = %d, want %d", loadedAgain[1].Type, model.EntryTypeCard)
	}
}

func mustDeriveVaultKeys(t *testing.T, password string) (crypto.KDFParams, []byte, []byte) {
	t.Helper()
	params := crypto.DefaultKDFParams()
	salt, err := crypto.GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() error = %v", err)
	}
	params.Salt = salt

	encryptionKey, verificationKey, err := crypto.DeriveKeys(password, params)
	if err != nil {
		t.Fatalf("DeriveKeys() error = %v", err)
	}

	return params, encryptionKey, verificationKey
}
