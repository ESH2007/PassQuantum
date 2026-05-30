package migration

import (
	"encoding/json"
	"strings"
	"testing"

	"passquantum/core/crypto"
	"passquantum/core/model"
)

func TestMapper_PasswordOnly(t *testing.T) {
	pub, priv, err := crypto.GenerateKeypair()
	if err != nil {
		t.Fatalf("keys: %v", err)
	}

	entries := []ImportedEntry{
		{
			Type:     model.EntryTypePassword,
			Title:    "GitHub",
			Username: "octocat",
			Password: []byte("hunter2!"),
			URLs:     []string{"https://github.com"},
			Source:   "test",
		},
	}

	result, err := MapAndEncrypt(entries, pub, nil, DupSkip)
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	if len(result.NewEntries) != 1 {
		t.Fatalf("expected 1 new entry, got %d", len(result.NewEntries))
	}
	ve := result.NewEntries[0]
	if ve.Type != model.EntryTypePassword {
		t.Errorf("type = %v", ve.Type)
	}
	if !strings.Contains(strings.ToLower(ve.Service), "github") {
		t.Errorf("service = %q", ve.Service)
	}
	if ve.Username != "octocat" {
		t.Errorf("username = %q", ve.Username)
	}

	// Round-trip: decrypt the payload and verify it matches the original
	// password verbatim (no JSON wrapping for password-only entries).
	ss, err := crypto.Decapsulate(ve.KyberCiphertext, priv)
	if err != nil {
		t.Fatalf("decap: %v", err)
	}
	plain, err := crypto.DecryptAES256GCM(ve.Nonce, ve.Ciphertext, ss)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if plain != "hunter2!" {
		t.Errorf("plaintext = %q", plain)
	}
}

func TestMapper_PasswordSecretIsWipedAfterMap(t *testing.T) {
	pub, _, err := crypto.GenerateKeypair()
	if err != nil {
		t.Fatalf("keys: %v", err)
	}
	pwBytes := []byte("topsecret123")
	entries := []ImportedEntry{{
		Type:     model.EntryTypePassword,
		Title:    "X",
		Password: pwBytes,
	}}
	_, err = MapAndEncrypt(entries, pub, nil, DupSkip)
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	// After mapping, the original byte slice must be zeroed.
	for _, b := range pwBytes {
		if b != 0 {
			t.Fatalf("password bytes not wiped: %q", string(pwBytes))
		}
	}
}

func TestMapper_PasswordWithNotesEmitsCompanion(t *testing.T) {
	pub, priv, err := crypto.GenerateKeypair()
	if err != nil {
		t.Fatalf("keys: %v", err)
	}
	entries := []ImportedEntry{
		{
			Type:     model.EntryTypePassword,
			Title:    "GitHub",
			Username: "octocat",
			Password: []byte("ghpw"),
			URLs:     []string{"https://github.com", "https://github.com/settings"},
			Notes:    "Recovery codes: abc-def",
			Folder:   "Dev",
			Source:   "test",
		},
	}
	result, err := MapAndEncrypt(entries, pub, nil, DupSkip)
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	if len(result.NewEntries) != 2 {
		t.Fatalf("expected 2 entries (password + companion note), got %d", len(result.NewEntries))
	}

	var note *model.VaultEntry
	for _, e := range result.NewEntries {
		if e.Type == model.EntryTypeNote {
			note = e
		}
	}
	if note == nil {
		t.Fatal("missing companion note")
	}
	ss, _ := crypto.Decapsulate(note.KyberCiphertext, priv)
	plain, _ := crypto.DecryptAES256GCM(note.Nonce, note.Ciphertext, ss)
	var parsed notePayload
	if err := json.Unmarshal([]byte(plain), &parsed); err != nil {
		t.Fatalf("note payload not JSON: %v", err)
	}
	if !strings.Contains(parsed.Content, "Recovery codes") {
		t.Errorf("note content missing user notes: %q", parsed.Content)
	}
	if !strings.Contains(parsed.Content, "Folder: Dev") {
		t.Errorf("note content missing folder: %q", parsed.Content)
	}
}

func TestMapper_PasswordWithEmbeddedTOTPSplits(t *testing.T) {
	pub, priv, err := crypto.GenerateKeypair()
	if err != nil {
		t.Fatalf("keys: %v", err)
	}
	entries := []ImportedEntry{
		{
			Type:     model.EntryTypePassword,
			Title:    "Gmail",
			Username: "alice@example.com",
			Password: []byte("gmailpw"),
			URLs:     []string{"https://mail.google.com"},
			TOTP:     "JBSWY3DPEHPK3PXP",
			Source:   "test",
		},
	}
	result, err := MapAndEncrypt(entries, pub, nil, DupSkip)
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	if len(result.NewEntries) != 2 {
		t.Fatalf("expected 2 entries (password + totp), got %d", len(result.NewEntries))
	}
	var totp *model.VaultEntry
	for _, e := range result.NewEntries {
		if e.Type == model.EntryTypeTOTP {
			totp = e
		}
	}
	if totp == nil {
		t.Fatal("missing TOTP entry")
	}
	if !strings.HasPrefix(totp.Service, "TOTP:") {
		t.Errorf("totp service prefix wrong: %q", totp.Service)
	}
	// Decrypted TOTP payload should be a JSON document with the secret.
	ss, _ := crypto.Decapsulate(totp.KyberCiphertext, priv)
	plain, _ := crypto.DecryptAES256GCM(totp.Nonce, totp.Ciphertext, ss)
	if !strings.Contains(plain, "JBSWY3DPEHPK3PXP") {
		t.Errorf("totp payload missing secret")
	}
}

func TestMapper_Card(t *testing.T) {
	pub, priv, err := crypto.GenerateKeypair()
	if err != nil {
		t.Fatalf("keys: %v", err)
	}
	entries := []ImportedEntry{{
		Type:  model.EntryTypeCard,
		Title: "My Visa",
		Card: &CardData{
			Subtype:  "credit",
			Holder:   "Alice Doe",
			Number:   []byte("4111111111111111"),
			ExpMonth: "12",
			ExpYear:  "2027",
			CVV:      []byte("123"),
		},
		Source: "test",
	}}
	result, err := MapAndEncrypt(entries, pub, nil, DupSkip)
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	if len(result.NewEntries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result.NewEntries))
	}
	ve := result.NewEntries[0]
	if ve.Type != model.EntryTypeCard {
		t.Fatalf("type = %v", ve.Type)
	}
	if !strings.HasPrefix(ve.Service, "CARD:") {
		t.Errorf("service prefix wrong: %q", ve.Service)
	}
	if ve.CardSubtype != "credit" {
		t.Errorf("subtype = %q", ve.CardSubtype)
	}
	ss, _ := crypto.Decapsulate(ve.KyberCiphertext, priv)
	plain, _ := crypto.DecryptAES256GCM(ve.Nonce, ve.Ciphertext, ss)
	if !strings.Contains(plain, "4111111111111111") {
		t.Errorf("card payload missing number")
	}
}

func TestMapper_DupSkip(t *testing.T) {
	pub, _, err := crypto.GenerateKeypair()
	if err != nil {
		t.Fatalf("keys: %v", err)
	}

	existing := []*model.VaultEntry{{
		Type:     model.EntryTypePassword,
		Service:  "github.com",
		Username: "octocat",
	}}
	entries := []ImportedEntry{{
		Type:     model.EntryTypePassword,
		Title:    "GitHub",
		Username: "octocat",
		Password: []byte("newpw"),
		URLs:     []string{"https://github.com"},
	}}

	result, err := MapAndEncrypt(entries, pub, existing, DupSkip)
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	if result.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", result.Skipped)
	}
	if len(result.NewEntries) != 0 {
		t.Errorf("expected 0 new entries, got %d", len(result.NewEntries))
	}
}

func TestMapper_DupReplaceRewritesCrypto(t *testing.T) {
	pub, priv, err := crypto.GenerateKeypair()
	if err != nil {
		t.Fatalf("keys: %v", err)
	}

	existing := []*model.VaultEntry{{
		Type:            model.EntryTypePassword,
		Service:         "github.com",
		Username:        "octocat",
		KyberCiphertext: []byte("old-ct"),
		Nonce:           []byte("old-nonce"),
		Ciphertext:      []byte("old-payload"),
	}}
	entries := []ImportedEntry{{
		Type:     model.EntryTypePassword,
		Title:    "GitHub",
		Username: "octocat",
		Password: []byte("newpw"),
		URLs:     []string{"https://github.com"},
	}}

	result, err := MapAndEncrypt(entries, pub, existing, DupReplace)
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	if result.Replaced != 1 {
		t.Errorf("expected 1 replaced, got %d", result.Replaced)
	}
	if len(result.NewEntries) != 0 {
		t.Errorf("expected 0 new entries, got %d", len(result.NewEntries))
	}
	// Existing entry's crypto fields must have been overwritten.
	if string(existing[0].Ciphertext) == "old-payload" {
		t.Error("existing ciphertext was not replaced")
	}
	// Decryption of the rewritten entry yields the new password.
	ss, _ := crypto.Decapsulate(existing[0].KyberCiphertext, priv)
	plain, _ := crypto.DecryptAES256GCM(existing[0].Nonce, existing[0].Ciphertext, ss)
	if plain != "newpw" {
		t.Errorf("decrypted plaintext = %q, want newpw", plain)
	}
}
