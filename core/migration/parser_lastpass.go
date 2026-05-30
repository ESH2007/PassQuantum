package migration

import (
	"errors"
	"io"

	"passquantum/core/model"
)

// LastPassImporter parses LastPass CSV exports.
// Header: url,username,password,totp,extra,name,grouping,fav
//
// Quirks:
//   - "url" == "http://sn" identifies a secure note (no URL).
//   - "extra" is the notes field.
//   - "grouping" is the folder.
//   - "totp" may carry an otpauth URI or bare base32 secret.
type LastPassImporter struct{}

func init() {
	DefaultRegistry.Register(&LastPassImporter{})
}

func (LastPassImporter) ID() string           { return "lastpass_csv" }
func (LastPassImporter) DisplayName() string  { return "LastPass" }
func (LastPassImporter) Extensions() []string { return []string{".csv"} }

func (LastPassImporter) Detect(_ string, head []byte) float64 {
	if headerMatches(head, "url", "username", "password", "totp", "extra", "name", "grouping") {
		return 0.96
	}
	return 0
}

func (LastPassImporter) Parse(r io.Reader, _ ParseOptions) (*ImportResult, error) {
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

		urlStr := getCSVCol(row, idx, "url")
		username := getCSVCol(row, idx, "username")
		password := getCSVCol(row, idx, "password")
		totp := getCSVCol(row, idx, "totp")
		extra := getCSVCol(row, idx, "extra")
		name := getCSVCol(row, idx, "name")
		grouping := getCSVCol(row, idx, "grouping")

		// Secure notes use the sentinel URL "http://sn".
		if urlStr == "http://sn" {
			if extra == "" {
				result.Skipped++
				continue
			}
			result.Entries = append(result.Entries, ImportedEntry{
				Type:   model.EntryTypeNote,
				Title:  name,
				Notes:  extra,
				Folder: grouping,
				Source: "lastpass_csv",
			})
			continue
		}

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
			Notes:    extra,
			Folder:   grouping,
			Source:   "lastpass_csv",
		})
	}
	return result, nil
}
