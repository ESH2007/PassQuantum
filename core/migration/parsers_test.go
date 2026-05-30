package migration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"passquantum/core/model"
)

// openFixture opens a testdata file and registers cleanup. The returned
// reader is passed straight to Importer.Parse.
func openFixture(t *testing.T, name string) *os.File {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("open fixture %s: %v", name, err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

// readHead returns the first headBytes bytes for detection tests.
func readHead(t *testing.T, name string) []byte {
	t.Helper()
	f := openFixture(t, name)
	buf := make([]byte, headBytes)
	n, _ := f.Read(buf)
	return buf[:n]
}

// ---------------- Detection ----------------

func TestDetection_ChromiumWinsOnChromiumFile(t *testing.T) {
	head := readHead(t, "chromium.csv")
	results := DefaultRegistry.Detect("chromium.csv", head)
	if len(results) == 0 {
		t.Fatal("no importer matched")
	}
	if results[0].Importer.ID() != "chromium_csv" {
		t.Fatalf("expected chromium_csv to win, got %s with %.2f",
			results[0].Importer.ID(), results[0].Score)
	}
	if results[0].Score < 0.9 {
		t.Errorf("expected confident score, got %.2f", results[0].Score)
	}
}

func TestDetection_FirefoxWinsOnFirefoxFile(t *testing.T) {
	head := readHead(t, "firefox.csv")
	results := DefaultRegistry.Detect("firefox.csv", head)
	if len(results) == 0 {
		t.Fatal("no importer matched")
	}
	if results[0].Importer.ID() != "firefox_csv" {
		t.Fatalf("expected firefox_csv to win, got %s", results[0].Importer.ID())
	}
}

func TestDetection_LastPassWinsOnLastPassFile(t *testing.T) {
	head := readHead(t, "lastpass.csv")
	results := DefaultRegistry.Detect("lastpass.csv", head)
	if len(results) == 0 {
		t.Fatal("no importer matched")
	}
	if results[0].Importer.ID() != "lastpass_csv" {
		t.Fatalf("expected lastpass_csv to win, got %s", results[0].Importer.ID())
	}
}

func TestDetection_BitwardenJSONWinsOnJSONFile(t *testing.T) {
	head := readHead(t, "bitwarden.json")
	results := DefaultRegistry.Detect("bitwarden.json", head)
	if len(results) == 0 {
		t.Fatal("no importer matched")
	}
	if results[0].Importer.ID() != "bitwarden_json" {
		t.Fatalf("expected bitwarden_json to win, got %s", results[0].Importer.ID())
	}
}

func TestDetection_GenericIsLowConfidenceFallback(t *testing.T) {
	// A CSV with unknown header should still be detected by the generic
	// importer, but with a low score.
	head := []byte("foo,bar,baz\n1,2,3\n")
	results := DefaultRegistry.Detect("anything.csv", head)
	var foundGeneric bool
	for _, r := range results {
		if r.Importer.ID() == "generic_csv" {
			foundGeneric = true
			if r.Score >= 0.5 {
				t.Errorf("generic CSV score too high (%.2f); should remain a fallback", r.Score)
			}
		}
	}
	if !foundGeneric {
		t.Error("generic CSV should always surface as a fallback for .csv files")
	}
}

// ---------------- Per-parser correctness ----------------

func TestChromiumParser(t *testing.T) {
	imp := ChromiumImporter{}
	res, err := imp.Parse(openFixture(t, "chromium.csv"), ParseOptions{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(res.Entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(res.Entries))
	}
	if res.Skipped != 1 {
		t.Errorf("expected 1 skipped (empty row), got %d", res.Skipped)
	}
	// First entry: GitHub.
	got := res.Entries[0]
	if got.Title != "GitHub" {
		t.Errorf("title = %q", got.Title)
	}
	if string(got.Password) != "hunter2pass!" {
		t.Errorf("password mismatch: got %q", string(got.Password))
	}
	if len(got.URLs) != 1 || got.URLs[0] != "https://github.com" {
		t.Errorf("urls = %v", got.URLs)
	}
}

func TestFirefoxParser(t *testing.T) {
	imp := FirefoxImporter{}
	res, err := imp.Parse(openFixture(t, "firefox.csv"), ParseOptions{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	// The third row has empty username and password, but the parser
	// still emits it because Firefox's empty-password rows often carry
	// httpRealm metadata. In our fixture the empty row should be skipped.
	if len(res.Entries) < 2 {
		t.Fatalf("expected at least 2 entries, got %d", len(res.Entries))
	}
	if string(res.Entries[0].Password) != "bobpw!" {
		t.Errorf("first password mismatch: got %q", string(res.Entries[0].Password))
	}
}

func TestLastPassParser_SplitsSecureNotes(t *testing.T) {
	imp := LastPassImporter{}
	res, err := imp.Parse(openFixture(t, "lastpass.csv"), ParseOptions{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var noteCount, pwCount int
	for _, e := range res.Entries {
		switch e.Type {
		case model.EntryTypeNote:
			noteCount++
		case model.EntryTypePassword:
			pwCount++
		}
	}
	if noteCount != 1 {
		t.Errorf("expected 1 secure note, got %d", noteCount)
	}
	if pwCount != 3 {
		t.Errorf("expected 3 passwords, got %d", pwCount)
	}
}

func TestKeePassParser(t *testing.T) {
	imp := KeePassImporter{}
	res, err := imp.Parse(openFixture(t, "keepass.csv"), ParseOptions{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(res.Entries) < 2 {
		t.Fatalf("expected at least 2 entries, got %d", len(res.Entries))
	}
	// Second entry has an otpauth URI in its TOTP column.
	gmail := res.Entries[1]
	if !strings.HasPrefix(gmail.TOTP, "otpauth://") {
		t.Errorf("expected otpauth URI, got %q", gmail.TOTP)
	}
}

func TestNordPassParser_Mixed(t *testing.T) {
	imp := NordPassImporter{}
	res, err := imp.Parse(openFixture(t, "nordpass.csv"), ParseOptions{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var pw, card, note int
	for _, e := range res.Entries {
		switch e.Type {
		case model.EntryTypePassword:
			pw++
		case model.EntryTypeCard:
			card++
		case model.EntryTypeNote:
			note++
		}
	}
	if pw != 1 || card != 1 || note != 1 {
		t.Errorf("expected 1/1/1 pw/card/note, got %d/%d/%d", pw, card, note)
	}
}

func TestBitwardenCSVParser(t *testing.T) {
	imp := BitwardenCSVImporter{}
	res, err := imp.Parse(openFixture(t, "bitwarden.csv"), ParseOptions{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(res.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(res.Entries))
	}
	// Gmail entry should carry a TOTP value.
	gmail := res.Entries[1]
	if gmail.TOTP == "" {
		t.Errorf("expected non-empty TOTP on gmail entry")
	}
}

func TestBitwardenJSONParser(t *testing.T) {
	imp := BitwardenJSONImporter{}
	res, err := imp.Parse(openFixture(t, "bitwarden.json"), ParseOptions{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(res.Entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(res.Entries))
	}
	var card *ImportedEntry
	for i := range res.Entries {
		if res.Entries[i].Type == model.EntryTypeCard {
			card = &res.Entries[i]
		}
	}
	if card == nil {
		t.Fatal("missing card entry")
	}
	if card.Card == nil || string(card.Card.Number) != "4111111111111111" {
		t.Errorf("card number mismatch")
	}
	if card.Card.Holder != "Alice Doe" {
		t.Errorf("card holder = %q", card.Card.Holder)
	}
}

// ---------------- Generic CSV ----------------

func TestGenericCSV_HeuristicMapping(t *testing.T) {
	csv := "name,login,secret,website\n" +
		"GitHub,octocat,ghpw,https://github.com\n" +
		"Empty,,,\n"
	imp := GenericCSVImporter{}
	res, err := imp.Parse(strings.NewReader(csv), ParseOptions{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(res.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(res.Entries))
	}
	got := res.Entries[0]
	if got.Title != "GitHub" || got.Username != "octocat" || string(got.Password) != "ghpw" {
		t.Errorf("mapped fields wrong: %+v", got)
	}
	if len(got.URLs) != 1 || got.URLs[0] != "https://github.com" {
		t.Errorf("urls = %v", got.URLs)
	}
}

func TestGenericCSV_RequiresPasswordColumn(t *testing.T) {
	csv := "foo,bar,baz\n1,2,3\n"
	imp := GenericCSVImporter{}
	_, err := imp.Parse(strings.NewReader(csv), ParseOptions{})
	if err == nil {
		t.Fatal("expected error when no password-like column is present")
	}
}
