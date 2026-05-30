package migration

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"passquantum/core/model"
)

// ---------------- 1PUX schema ----------------
//
// 1PUX is a ZIP archive containing export.data (a JSON document) plus the
// export's attachments (under files/). The JSON schema is:
//
//   accounts: [
//     {
//       attrs: { name, ... },
//       vaults: [
//         {
//           attrs: { name, ... },
//           items: [
//             {
//               uuid, categoryUuid (kind),
//               overview: { title, url, urls: [{label, url}] },
//               details:  {
//                 loginFields: [{ designation: "username"|"password", value }],
//                 notesPlain: "...",
//                 sections: [{ title, fields: [{ title, value: {string/totp/...}}] }]
//               }
//             }
//           ]
//         }
//       ]
//     }
//   ]
//
// Category UUIDs (the bits we care about):
//   001 = Login
//   002 = Credit Card
//   003 = Secure Note
//   004 = Identity
//   005 = Password (standalone)

type onePuxExport struct {
	Accounts []onePuxAccount `json:"accounts"`
}

type onePuxAccount struct {
	Attrs  onePuxAttrs    `json:"attrs"`
	Vaults []onePuxVault  `json:"vaults"`
}

type onePuxVault struct {
	Attrs onePuxAttrs   `json:"attrs"`
	Items []onePuxItem  `json:"items"`
}

type onePuxAttrs struct {
	Name string `json:"name"`
}

type onePuxItem struct {
	UUID         string          `json:"uuid"`
	CategoryUUID string          `json:"categoryUuid"`
	Overview     onePuxOverview  `json:"overview"`
	Details      onePuxDetails   `json:"details"`
	TrashedTime  int64           `json:"trashed"`
}

type onePuxOverview struct {
	Title string         `json:"title"`
	URL   string         `json:"url"`
	URLs  []onePuxURLRef `json:"urls"`
	Tags  []string       `json:"tags"`
}

type onePuxURLRef struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type onePuxDetails struct {
	LoginFields []onePuxLoginField `json:"loginFields"`
	NotesPlain  string             `json:"notesPlain"`
	Sections    []onePuxSection    `json:"sections"`
}

type onePuxLoginField struct {
	Designation string `json:"designation"`
	Name        string `json:"name"`
	Value       string `json:"value"`
}

type onePuxSection struct {
	Title  string         `json:"title"`
	Name   string         `json:"name"`
	Fields []onePuxField  `json:"fields"`
}

type onePuxField struct {
	Title string                 `json:"title"`
	ID    string                 `json:"id"`
	Value map[string]json.RawMessage `json:"value"`
}

// ---------------- Importer ----------------

// OnePasswordImporter parses 1Password's .1pux export archive.
type OnePasswordImporter struct{}

func init() {
	DefaultRegistry.Register(&OnePasswordImporter{})
}

func (OnePasswordImporter) ID() string           { return "1password_1pux" }
func (OnePasswordImporter) DisplayName() string  { return "1Password (1PUX)" }
func (OnePasswordImporter) Extensions() []string { return []string{".1pux", ".zip"} }

func (OnePasswordImporter) Detect(filename string, head []byte) float64 {
	lower := strings.ToLower(filename)
	if !looksLikeZIP(head) {
		return 0
	}
	if strings.HasSuffix(lower, ".1pux") {
		return 0.98
	}
	// Generic .zip — only claim if the filename hints at 1Password.
	if strings.Contains(lower, "1password") {
		return 0.7
	}
	return 0
}

func (OnePasswordImporter) Parse(r io.Reader, _ ParseOptions) (*ImportResult, error) {
	raw, err := readWholeReader(r)
	if err != nil {
		return nil, err
	}
	zr, _, err := readWholeZIP(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	data, _, err := findZIPFile(zr, "export.data")
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("1pux: export.data not found inside archive")
	}

	var export onePuxExport
	if err := json.Unmarshal(data, &export); err != nil {
		return nil, fmt.Errorf("1pux json: %w", err)
	}

	result := &ImportResult{}
	for _, acc := range export.Accounts {
		for _, vault := range acc.Vaults {
			folder := strings.TrimSpace(vault.Attrs.Name)
			for i := range vault.Items {
				item := &vault.Items[i]
				if item.TrashedTime > 0 {
					result.Skipped++
					continue
				}
				entry, ok := onePuxConvertItem(item, folder)
				if !ok {
					result.Skipped++
					continue
				}
				result.Entries = append(result.Entries, entry)
			}
		}
	}
	return result, nil
}

func onePuxConvertItem(item *onePuxItem, folder string) (ImportedEntry, bool) {
	base := ImportedEntry{
		Title:  strings.TrimSpace(item.Overview.Title),
		Notes:  strings.TrimSpace(item.Details.NotesPlain),
		Folder: folder,
		Tags:   item.Overview.Tags,
		Source: "1password_1pux",
	}
	if u := strings.TrimSpace(item.Overview.URL); u != "" {
		base.URLs = []string{u}
	}
	for _, ref := range item.Overview.URLs {
		if v := strings.TrimSpace(ref.URL); v != "" {
			base.URLs = append(base.URLs, v)
		}
	}

	// Collect TOTP secrets from the sections, regardless of category.
	totp := firstTOTPInSections(item.Details.Sections)

	switch item.CategoryUUID {
	case "001", "005": // Login / Password
		username, password := loginCredsFromFields(item.Details.LoginFields)
		base.Type = model.EntryTypePassword
		base.Username = username
		base.Password = []byte(password)
		base.TOTP = totp
		if extras := flattenOnePuxSections(item.Details.Sections); len(extras) > 0 {
			base.Fields = extras
		}
		return base, true

	case "003": // Secure Note
		base.Type = model.EntryTypeNote
		if extras := flattenOnePuxSections(item.Details.Sections); len(extras) > 0 {
			// Append custom section data to the note body so nothing is lost.
			var b strings.Builder
			b.WriteString(base.Notes)
			b.WriteString("\n\n--- Sections ---\n")
			for k, v := range extras {
				b.WriteString(k)
				b.WriteString(": ")
				b.WriteString(v)
				b.WriteByte('\n')
			}
			base.Notes = strings.TrimSpace(b.String())
		}
		return base, true

	case "002": // Credit Card
		card := cardFromSections(item.Details.Sections)
		if card == nil {
			return ImportedEntry{}, false
		}
		base.Type = model.EntryTypeCard
		base.Card = card
		return base, true

	case "004": // Identity
		// VaultEntry has no native identity type; flatten the data into a
		// secure note so it is not silently dropped.
		extras := flattenOnePuxSections(item.Details.Sections)
		var b strings.Builder
		b.WriteString(base.Notes)
		if len(extras) > 0 {
			b.WriteString("\n\n--- Identity ---\n")
			for k, v := range extras {
				b.WriteString(k)
				b.WriteString(": ")
				b.WriteString(v)
				b.WriteByte('\n')
			}
		}
		base.Type = model.EntryTypeNote
		base.Notes = strings.TrimSpace(b.String())
		return base, true

	default:
		return ImportedEntry{}, false
	}
}

// loginCredsFromFields extracts the canonical username/password pair from a
// 1Password loginFields array. Designations are the stable identifier.
func loginCredsFromFields(fields []onePuxLoginField) (username, password string) {
	for _, f := range fields {
		switch strings.ToLower(f.Designation) {
		case "username":
			username = f.Value
		case "password":
			password = f.Value
		}
	}
	return
}

// firstTOTPInSections walks every section's fields and returns the first
// value rendered as a TOTP (value object has a "totp" key).
func firstTOTPInSections(sections []onePuxSection) string {
	for _, s := range sections {
		for _, f := range s.Fields {
			if raw, ok := f.Value["totp"]; ok {
				var v string
				if err := json.Unmarshal(raw, &v); err == nil && v != "" {
					return v
				}
			}
		}
	}
	return ""
}

// flattenOnePuxSections folds the contents of every section into a flat
// map[string]string the mapper can attach to the encrypted notes payload.
// Concealed (password / totp) values are replaced with placeholders so we
// never write a secret into a non-secret field.
func flattenOnePuxSections(sections []onePuxSection) map[string]string {
	out := map[string]string{}
	for _, s := range sections {
		for _, f := range s.Fields {
			key := strings.TrimSpace(f.Title)
			if key == "" {
				key = strings.TrimSpace(f.ID)
			}
			if key == "" {
				continue
			}
			if _, hidden := f.Value["concealed"]; hidden {
				out[key] = "(hidden value)"
				continue
			}
			if _, hidden := f.Value["totp"]; hidden {
				out[key] = "(totp value)"
				continue
			}
			for _, kind := range []string{"string", "email", "url", "phone", "date", "monthYear"} {
				if raw, ok := f.Value[kind]; ok {
					var v string
					if err := json.Unmarshal(raw, &v); err == nil && v != "" {
						out[key] = v
					}
					break
				}
			}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// cardFromSections recognizes a credit-card item by scanning the standard
// 1Password section fields (cardholder / number / expiry / cvv). Any field
// it cannot map is dropped on the floor; the section flattener separately
// captures the rest into the entry's Fields map.
func cardFromSections(sections []onePuxSection) *CardData {
	card := &CardData{Subtype: "credit"}
	have := false

	for _, s := range sections {
		for _, f := range s.Fields {
			key := strings.ToLower(strings.TrimSpace(f.Title) + " " + strings.TrimSpace(f.ID))
			value := onePuxStringValue(f.Value)
			switch {
			case strings.Contains(key, "cardholder") || strings.Contains(key, "ccname"):
				if value != "" {
					card.Holder = value
					have = true
				}
			case strings.Contains(key, "ccnum") || strings.Contains(key, "number"):
				if value != "" {
					card.Number = []byte(value)
					have = true
				}
			case strings.Contains(key, "cvv") || strings.Contains(key, "verification"):
				if value != "" {
					card.CVV = []byte(value)
					have = true
				}
			case strings.Contains(key, "expiry") || strings.Contains(key, "expdate") || strings.Contains(key, "monthyear"):
				month, year := splitProtonExpiry(value)
				if month != "" {
					card.ExpMonth, card.ExpYear = month, year
					have = true
				}
			case strings.Contains(key, "type") && (strings.Contains(value, "debit") || strings.Contains(value, "credit")):
				if strings.Contains(strings.ToLower(value), "debit") {
					card.Subtype = "debit"
				}
			}
		}
	}
	if !have {
		return nil
	}
	return card
}

func onePuxStringValue(v map[string]json.RawMessage) string {
	for _, kind := range []string{"string", "email", "url", "phone", "date", "monthYear", "creditCardNumber", "creditCardType"} {
		if raw, ok := v[kind]; ok {
			var s string
			if err := json.Unmarshal(raw, &s); err == nil {
				return s
			}
		}
	}
	return ""
}
