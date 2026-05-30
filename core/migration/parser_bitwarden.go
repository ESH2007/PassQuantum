package migration

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"passquantum/core/model"
)

// ---------------- CSV ----------------

// BitwardenCSVImporter parses Bitwarden's unencrypted CSV export.
// Header:
//
//	folder,favorite,type,name,notes,fields,reprompt,
//	  login_uri,login_username,login_password,login_totp
//
// Only login and secure_note records are present in CSV; cards and
// identities require the JSON export.
type BitwardenCSVImporter struct{}

func init() {
	DefaultRegistry.Register(&BitwardenCSVImporter{})
}

func (BitwardenCSVImporter) ID() string           { return "bitwarden_csv" }
func (BitwardenCSVImporter) DisplayName() string  { return "Bitwarden (CSV)" }
func (BitwardenCSVImporter) Extensions() []string { return []string{".csv"} }

func (BitwardenCSVImporter) Detect(_ string, head []byte) float64 {
	if headerMatches(head, "login_uri", "login_username", "login_password") {
		return 0.97
	}
	return 0
}

func (BitwardenCSVImporter) Parse(r io.Reader, _ ParseOptions) (*ImportResult, error) {
	cr := newCSVReader(r)
	idx, _, err := readCSVHeader(cr)
	if err != nil {
		return nil, err
	}

	result := &ImportResult{}
	for {
		row, err := cr.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			result.Skipped++
			continue
		}

		t := strings.ToLower(getCSVCol(row, idx, "type"))
		name := getCSVCol(row, idx, "name")
		notes := getCSVCol(row, idx, "notes")
		folder := getCSVCol(row, idx, "folder")

		switch t {
		case "note", "secure_note":
			if notes == "" && name == "" {
				result.Skipped++
				continue
			}
			result.Entries = append(result.Entries, ImportedEntry{
				Type:   model.EntryTypeNote,
				Title:  name,
				Notes:  notes,
				Folder: folder,
				Source: "bitwarden_csv",
			})
		default:
			username := getCSVCol(row, idx, "login_username")
			password := getCSVCol(row, idx, "login_password")
			urlStr := getCSVCol(row, idx, "login_uri")
			totp := getCSVCol(row, idx, "login_totp")

			if password == "" && username == "" && totp == "" {
				result.Skipped++
				continue
			}
			var urls []string
			if urlStr != "" {
				urls = []string{urlStr}
			}
			result.Entries = append(result.Entries, ImportedEntry{
				Type:     model.EntryTypePassword,
				Title:    name,
				Username: username,
				Password: []byte(password),
				URLs:     urls,
				TOTP:     totp,
				Notes:    notes,
				Folder:   folder,
				Source:   "bitwarden_csv",
			})
		}
	}
	return result, nil
}

// ---------------- JSON ----------------

type bitwardenJSON struct {
	Encrypted bool                  `json:"encrypted"`
	Folders   []bitwardenJSONFolder `json:"folders"`
	Items     []bitwardenJSONItem   `json:"items"`
}

type bitwardenJSONFolder struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type bitwardenJSONItem struct {
	Type     int                    `json:"type"` // 1=login, 2=note, 3=card, 4=identity
	Name     string                 `json:"name"`
	Notes    string                 `json:"notes"`
	FolderID string                 `json:"folderId"`
	Favorite bool                   `json:"favorite"`
	Fields   []bitwardenJSONField   `json:"fields"`
	Login    *bitwardenJSONLogin    `json:"login,omitempty"`
	Card     *bitwardenJSONCard     `json:"card,omitempty"`
	Identity *bitwardenJSONIdentity `json:"identity,omitempty"`
}

type bitwardenJSONField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  int    `json:"type"`
}

type bitwardenJSONLogin struct {
	Username string             `json:"username"`
	Password string             `json:"password"`
	TOTP     string             `json:"totp"`
	URIs     []bitwardenJSONURI `json:"uris"`
}

type bitwardenJSONURI struct {
	URI string `json:"uri"`
}

type bitwardenJSONCard struct {
	CardholderName string `json:"cardholderName"`
	Brand          string `json:"brand"`
	Number         string `json:"number"`
	ExpMonth       string `json:"expMonth"`
	ExpYear        string `json:"expYear"`
	Code           string `json:"code"`
}

type bitwardenJSONIdentity struct {
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Address1   string `json:"address1"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
	Country    string `json:"country"`
}

// BitwardenJSONImporter parses Bitwarden's unencrypted JSON export. The
// encrypted variant (encrypted:true) is not supported in Phase 1.
type BitwardenJSONImporter struct{}

func init() {
	DefaultRegistry.Register(&BitwardenJSONImporter{})
}

func (BitwardenJSONImporter) ID() string           { return "bitwarden_json" }
func (BitwardenJSONImporter) DisplayName() string  { return "Bitwarden (JSON)" }
func (BitwardenJSONImporter) Extensions() []string { return []string{".json"} }

func (BitwardenJSONImporter) Detect(_ string, head []byte) float64 {
	h := string(head)
	// Look for the unambiguous Bitwarden top-level keys.
	if strings.Contains(h, "\"encrypted\"") && strings.Contains(h, "\"items\"") &&
		(strings.Contains(h, "\"folders\"") || strings.Contains(h, "\"login\"")) {
		return 0.95
	}
	return 0
}

func (BitwardenJSONImporter) Parse(r io.Reader, _ ParseOptions) (*ImportResult, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var data bitwardenJSON
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("bitwarden json parse: %w", err)
	}
	if data.Encrypted {
		return nil, errors.New("encrypted Bitwarden exports are not supported in this version")
	}

	folderByID := make(map[string]string, len(data.Folders))
	for _, f := range data.Folders {
		folderByID[f.ID] = f.Name
	}

	result := &ImportResult{}
	for _, item := range data.Items {
		folder := folderByID[item.FolderID]
		fields := flattenBWFields(item.Fields)

		switch item.Type {
		case 1: // login
			if item.Login == nil {
				result.Skipped++
				continue
			}
			var urls []string
			for _, u := range item.Login.URIs {
				if u.URI != "" {
					urls = append(urls, u.URI)
				}
			}
			result.Entries = append(result.Entries, ImportedEntry{
				Type:     model.EntryTypePassword,
				Title:    item.Name,
				Username: item.Login.Username,
				Password: []byte(item.Login.Password),
				URLs:     urls,
				TOTP:     item.Login.TOTP,
				Notes:    item.Notes,
				Folder:   folder,
				Fields:   fields,
				Source:   "bitwarden_json",
			})

		case 2: // secure note
			if item.Notes == "" && item.Name == "" {
				result.Skipped++
				continue
			}
			result.Entries = append(result.Entries, ImportedEntry{
				Type:   model.EntryTypeNote,
				Title:  item.Name,
				Notes:  item.Notes,
				Folder: folder,
				Fields: fields,
				Source: "bitwarden_json",
			})

		case 3: // card
			if item.Card == nil {
				result.Skipped++
				continue
			}
			result.Entries = append(result.Entries, ImportedEntry{
				Type:   model.EntryTypeCard,
				Title:  item.Name,
				Notes:  item.Notes,
				Folder: folder,
				Fields: fields,
				Source: "bitwarden_json",
				Card: &CardData{
					Subtype:  "credit",
					Brand:    item.Card.Brand,
					Holder:   item.Card.CardholderName,
					Number:   []byte(item.Card.Number),
					ExpMonth: item.Card.ExpMonth,
					ExpYear:  item.Card.ExpYear,
					CVV:      []byte(item.Card.Code),
				},
			})

		case 4: // identity — Phase 2 will route this to a dedicated entry type;
			// for now we capture it as a secure note so the data is not lost.
			if item.Identity == nil {
				result.Skipped++
				continue
			}
			id := item.Identity
			content := strings.Join([]string{
				"Name: " + strings.TrimSpace(id.FirstName+" "+id.LastName),
				"Email: " + id.Email,
				"Phone: " + id.Phone,
				"Address: " + id.Address1,
				"City: " + id.City,
				"State: " + id.State,
				"Postal code: " + id.PostalCode,
				"Country: " + id.Country,
			}, "\n")
			result.Entries = append(result.Entries, ImportedEntry{
				Type:   model.EntryTypeNote,
				Title:  item.Name,
				Notes:  content + "\n\n" + item.Notes,
				Folder: folder,
				Source: "bitwarden_json",
			})

		default:
			result.Skipped++
		}
	}
	return result, nil
}

func flattenBWFields(fields []bitwardenJSONField) map[string]string {
	if len(fields) == 0 {
		return nil
	}
	out := make(map[string]string, len(fields))
	for _, f := range fields {
		if f.Name == "" {
			continue
		}
		// type 1 in Bitwarden's schema is a hidden / password-style field.
		// We do not include those values verbatim to avoid creating a second
		// plaintext copy of a secret inside a separate note: we just record
		// that a hidden field existed.
		if f.Type == 1 {
			out[f.Name] = "(hidden value)"
			continue
		}
		out[f.Name] = f.Value
	}
	return out
}
