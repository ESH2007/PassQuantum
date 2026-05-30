package migration

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"

	"passquantum/core/model"
)

// ---------------- Proton Pass ----------------

func TestProtonPass_SkipsRecycleBinAndTrashedItems(t *testing.T) {
	imp := ProtonPassImporter{}
	res, err := imp.Parse(openFixture(t, "protonpass.json"), ParseOptions{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	// item-3 is trashed (state == 2); the Recycle Bin vault is skipped entirely.
	if len(res.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(res.Entries))
	}
}

func TestProtonPass_UsesItemEmailWhenUsernameEmpty(t *testing.T) {
	imp := ProtonPassImporter{}
	res, _ := imp.Parse(openFixture(t, "protonpass.json"), ParseOptions{})
	var github *ImportedEntry
	for i := range res.Entries {
		if res.Entries[i].Title == "GitHub" {
			github = &res.Entries[i]
		}
	}
	if github == nil {
		t.Fatal("GitHub entry missing")
	}
	if github.Username != "octocat@example.com" {
		t.Errorf("expected itemEmail to fill username, got %q", github.Username)
	}
}

func TestProtonPass_DetectsJSONExport(t *testing.T) {
	head := readHead(t, "protonpass.json")
	results := DefaultRegistry.Detect("data.json", head)
	if len(results) == 0 || results[0].Importer.ID() != "protonpass" {
		var top string
		if len(results) > 0 {
			top = results[0].Importer.ID()
		}
		t.Fatalf("expected protonpass to win, got %q", top)
	}
}

// ---------------- Kaspersky ----------------

func TestKaspersky_ParsesAllSections(t *testing.T) {
	imp := KasperskyTXTImporter{}
	res, err := imp.Parse(openFixture(t, "kaspersky.txt"), ParseOptions{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var pw, note int
	for _, e := range res.Entries {
		switch e.Type {
		case model.EntryTypePassword:
			pw++
		case model.EntryTypeNote:
			note++
		}
	}
	// 2 websites + 1 application = 3 passwords; 1 secure note.
	if pw != 3 {
		t.Errorf("expected 3 password entries, got %d", pw)
	}
	if note != 1 {
		t.Errorf("expected 1 note entry, got %d", note)
	}
}

func TestKaspersky_MultiLineComment(t *testing.T) {
	imp := KasperskyTXTImporter{}
	res, _ := imp.Parse(openFixture(t, "kaspersky.txt"), ParseOptions{})
	var gmail *ImportedEntry
	for i := range res.Entries {
		if res.Entries[i].Title == "Gmail" {
			gmail = &res.Entries[i]
		}
	}
	if gmail == nil {
		t.Fatal("Gmail missing")
	}
	if !strings.Contains(gmail.Notes, "multiple lines") {
		t.Errorf("multi-line comment lost: %q", gmail.Notes)
	}
}

func TestKaspersky_DetectsByLeadingSection(t *testing.T) {
	head := readHead(t, "kaspersky.txt")
	results := DefaultRegistry.Detect("export.txt", head)
	if len(results) == 0 || results[0].Importer.ID() != "kaspersky_txt" {
		t.Fatalf("expected kaspersky_txt to win, got results = %+v", results)
	}
}

// ---------------- 1Password 1PUX ----------------

const sample1PUX = `{
  "accounts": [
    {
      "attrs": { "name": "Personal" },
      "vaults": [
        {
          "attrs": { "name": "Private" },
          "items": [
            {
              "uuid": "i1",
              "categoryUuid": "001",
              "overview": { "title": "GitHub", "url": "https://github.com", "urls": [] },
              "details": {
                "loginFields": [
                  { "designation": "username", "value": "octocat" },
                  { "designation": "password", "value": "ghpw" }
                ],
                "notesPlain": "Personal",
                "sections": [
                  {
                    "title": "Two-Factor",
                    "fields": [
                      { "title": "one-time password", "id": "TOTP_x", "value": { "totp": "otpauth://totp/GitHub:octocat?secret=ABCDEFGH" } }
                    ]
                  }
                ]
              }
            },
            {
              "uuid": "i2",
              "categoryUuid": "002",
              "overview": { "title": "Visa", "url": "" },
              "details": {
                "sections": [
                  {
                    "title": "card",
                    "fields": [
                      { "title": "cardholder name", "id": "cardholder", "value": { "string": "Alice Doe" } },
                      { "title": "number", "id": "ccnum", "value": { "creditCardNumber": "4111111111111111" } },
                      { "title": "verification number", "id": "cvv", "value": { "concealed": "123" } },
                      { "title": "expiry date", "id": "expiry", "value": { "monthYear": "12/2027" } }
                    ]
                  }
                ]
              }
            },
            {
              "uuid": "i3",
              "categoryUuid": "003",
              "overview": { "title": "SSH passphrase" },
              "details": {
                "notesPlain": "the passphrase is BLAH",
                "sections": []
              }
            },
            {
              "uuid": "i4",
              "categoryUuid": "001",
              "trashed": 1700000000,
              "overview": { "title": "Trashed login" },
              "details": {
                "loginFields": [
                  { "designation": "username", "value": "x" },
                  { "designation": "password", "value": "y" }
                ]
              }
            }
          ]
        }
      ]
    }
  ]
}`

func build1PUX(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, err := w.Create("export.data")
	if err != nil {
		t.Fatalf("zip create: %v", err)
	}
	if _, err := f.Write([]byte(sample1PUX)); err != nil {
		t.Fatalf("zip write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}

func TestOnePassword_ParsesCategories(t *testing.T) {
	imp := OnePasswordImporter{}
	res, err := imp.Parse(bytes.NewReader(build1PUX(t)), ParseOptions{})
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
	if res.Skipped != 1 {
		t.Errorf("expected 1 skipped (trashed item), got %d", res.Skipped)
	}
}

func TestOnePassword_ExtractsTOTPFromSection(t *testing.T) {
	imp := OnePasswordImporter{}
	res, _ := imp.Parse(bytes.NewReader(build1PUX(t)), ParseOptions{})
	var github *ImportedEntry
	for i := range res.Entries {
		if res.Entries[i].Title == "GitHub" {
			github = &res.Entries[i]
		}
	}
	if github == nil {
		t.Fatal("GitHub entry missing")
	}
	if !strings.HasPrefix(github.TOTP, "otpauth://") {
		t.Errorf("expected otpauth URI in TOTP, got %q", github.TOTP)
	}
}

func TestOnePassword_CardFields(t *testing.T) {
	imp := OnePasswordImporter{}
	res, _ := imp.Parse(bytes.NewReader(build1PUX(t)), ParseOptions{})
	var card *ImportedEntry
	for i := range res.Entries {
		if res.Entries[i].Type == model.EntryTypeCard {
			card = &res.Entries[i]
		}
	}
	if card == nil || card.Card == nil {
		t.Fatal("card entry missing")
	}
	if string(card.Card.Number) != "4111111111111111" {
		t.Errorf("card number = %q", string(card.Card.Number))
	}
	if card.Card.Holder != "Alice Doe" {
		t.Errorf("holder = %q", card.Card.Holder)
	}
	if card.Card.ExpMonth != "12" || card.Card.ExpYear != "2027" {
		t.Errorf("expiry = %q/%q", card.Card.ExpMonth, card.Card.ExpYear)
	}
}

// ---------------- Dashlane ----------------

func buildDashlaneZIP(t *testing.T) []byte {
	t.Helper()
	files := map[string]string{
		"credentials.csv": "" +
			"username,username2,username3,title,password,note,url,category,otpSecret\n" +
			"octocat,,,GitHub,ghpw,Personal,https://github.com,Dev,\n" +
			"alice@example.com,,,Gmail,gmailpw,,https://mail.google.com,Email,JBSWY3DPEHPK3PXP\n",
		"securenotes.csv": "" +
			"title,note,category\n" +
			"SSH Key,Passphrase is BLAH,Notes\n",
		"payments.csv": "" +
			"type,account_name,account_holder,cc_number,code,expiration_month,expiration_year\n" +
			"credit_card,My Visa,Alice Doe,4111111111111111,123,12,2027\n",
	}
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatalf("zip create: %v", err)
		}
		if _, err := f.Write([]byte(content)); err != nil {
			t.Fatalf("zip write: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}

func TestDashlane_ParsesAllSheets(t *testing.T) {
	imp := DashlaneImporter{}
	res, err := imp.Parse(bytes.NewReader(buildDashlaneZIP(t)), ParseOptions{})
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
	if pw != 2 {
		t.Errorf("expected 2 passwords, got %d", pw)
	}
	if card != 1 {
		t.Errorf("expected 1 card, got %d", card)
	}
	if note != 1 {
		t.Errorf("expected 1 note, got %d", note)
	}
}

func TestDashlane_NormalizesBareTOTPSecret(t *testing.T) {
	imp := DashlaneImporter{}
	res, _ := imp.Parse(bytes.NewReader(buildDashlaneZIP(t)), ParseOptions{})
	var gmail *ImportedEntry
	for i := range res.Entries {
		if res.Entries[i].Title == "Gmail" {
			gmail = &res.Entries[i]
		}
	}
	if gmail == nil {
		t.Fatal("Gmail missing")
	}
	if !strings.HasPrefix(gmail.TOTP, "otpauth://") {
		t.Errorf("expected normalized otpauth URI, got %q", gmail.TOTP)
	}
	if !strings.Contains(gmail.TOTP, "JBSWY3DPEHPK3PXP") {
		t.Errorf("secret missing: %q", gmail.TOTP)
	}
}
