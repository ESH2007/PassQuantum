package migration

import (
	"errors"
	"io"

	"passquantum/core/model"
)

// KeePassImporter parses KeePassXC's default CSV export.
// Header: Group,Title,Username,Password,URL,Notes,TOTP
//
// KeePassXC supports TOTP entries by storing an otpauth URI in the TOTP
// column on the same row as the login.
type KeePassImporter struct{}

func init() {
	DefaultRegistry.Register(&KeePassImporter{})
}

func (KeePassImporter) ID() string           { return "keepass_csv" }
func (KeePassImporter) DisplayName() string  { return "KeePass / KeePassXC" }
func (KeePassImporter) Extensions() []string { return []string{".csv"} }

func (KeePassImporter) Detect(_ string, head []byte) float64 {
	if headerMatches(head, "group", "title", "username", "password", "url", "notes") {
		return 0.93
	}
	return 0
}

func (KeePassImporter) Parse(r io.Reader, _ ParseOptions) (*ImportResult, error) {
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

		title := getCSVCol(row, idx, "title")
		username := getCSVCol(row, idx, "username")
		password := getCSVCol(row, idx, "password")
		urlStr := getCSVCol(row, idx, "url")
		notes := getCSVCol(row, idx, "notes")
		totp := getCSVCol(row, idx, "totp")
		group := getCSVCol(row, idx, "group")

		if password == "" && username == "" && notes == "" && totp == "" {
			result.Skipped++
			continue
		}

		var urls []string
		if urlStr != "" {
			urls = []string{urlStr}
		}
		result.Entries = append(result.Entries, ImportedEntry{
			Type:     model.EntryTypePassword,
			Title:    title,
			Username: username,
			Password: []byte(password),
			URLs:     urls,
			TOTP:     totp,
			Notes:    notes,
			Folder:   group,
			Source:   "keepass_csv",
		})
	}
	return result, nil
}
