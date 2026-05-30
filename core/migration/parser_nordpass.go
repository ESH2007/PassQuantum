package migration

import (
	"errors"
	"io"
	"strings"

	"passquantum/core/model"
)

// NordPassImporter parses NordPass's CSV export. NordPass uses a single wide
// table that mixes logins, cards and identities, with a "type" column for
// disambiguation:
//
//	name,url,username,password,note,cardholdername,cardnumber,cvc,
//	  expirydate,zipcode,folder,full_name,phone_number,email,address1,
//	  address2,city,country,state,type
type NordPassImporter struct{}

func init() {
	DefaultRegistry.Register(&NordPassImporter{})
}

func (NordPassImporter) ID() string           { return "nordpass_csv" }
func (NordPassImporter) DisplayName() string  { return "NordPass" }
func (NordPassImporter) Extensions() []string { return []string{".csv"} }

func (NordPassImporter) Detect(_ string, head []byte) float64 {
	if headerMatches(head, "name", "url", "username", "password", "cardholdername", "cardnumber") {
		return 0.95
	}
	return 0
}

func (NordPassImporter) Parse(r io.Reader, _ ParseOptions) (*ImportResult, error) {
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

		typeField := strings.ToLower(getCSVCol(row, idx, "type"))
		name := getCSVCol(row, idx, "name")
		folder := getCSVCol(row, idx, "folder")
		note := getCSVCol(row, idx, "note")

		switch typeField {
		case "credit_card", "card", "debit_card":
			entry := ImportedEntry{
				Type:   model.EntryTypeCard,
				Title:  name,
				Notes:  note,
				Folder: folder,
				Source: "nordpass_csv",
				Card: &CardData{
					Subtype:  "credit",
					Holder:   getCSVCol(row, idx, "cardholdername"),
					Number:   []byte(getCSVCol(row, idx, "cardnumber")),
					CVV:      []byte(getCSVCol(row, idx, "cvc")),
					ExpMonth: extractExpMonth(getCSVCol(row, idx, "expirydate")),
					ExpYear:  extractExpYear(getCSVCol(row, idx, "expirydate")),
				},
			}
			if typeField == "debit_card" {
				entry.Card.Subtype = "debit"
			}
			result.Entries = append(result.Entries, entry)

		case "secure_note", "note":
			if note == "" {
				result.Skipped++
				continue
			}
			result.Entries = append(result.Entries, ImportedEntry{
				Type:   model.EntryTypeNote,
				Title:  name,
				Notes:  note,
				Folder: folder,
				Source: "nordpass_csv",
			})

		default: // login / password (empty type field too)
			username := getCSVCol(row, idx, "username")
			password := getCSVCol(row, idx, "password")
			urlStr := getCSVCol(row, idx, "url")

			if password == "" && username == "" && note == "" {
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
				Notes:    note,
				Folder:   folder,
				Source:   "nordpass_csv",
			})
		}
	}
	return result, nil
}

// extractExpMonth pulls the month out of "MM/YY", "MM/YYYY" or "MM-YYYY"
// style values. Returns "" if the input is empty.
func extractExpMonth(exp string) string {
	exp = strings.TrimSpace(exp)
	if exp == "" {
		return ""
	}
	for _, sep := range []string{"/", "-", " "} {
		if i := strings.Index(exp, sep); i > 0 {
			return strings.TrimSpace(exp[:i])
		}
	}
	return exp
}

// extractExpYear is the symmetric counterpart of extractExpMonth.
func extractExpYear(exp string) string {
	exp = strings.TrimSpace(exp)
	if exp == "" {
		return ""
	}
	for _, sep := range []string{"/", "-", " "} {
		if i := strings.Index(exp, sep); i > 0 {
			return strings.TrimSpace(exp[i+1:])
		}
	}
	return ""
}
