package migration

import (
	"errors"
	"io"
	"strings"

	"passquantum/core/model"
)

// ChromiumImporter parses the CSV exported from any Chromium-based browser
// (Chrome, Brave, Edge, Opera, Vivaldi). The header is canonical and
// identical across these browsers: "name,url,username,password,note".
//
// Older Chrome versions emit only the first four columns without the note
// column; the parser handles both shapes via column name lookup.
type ChromiumImporter struct{}

func init() {
	DefaultRegistry.Register(&ChromiumImporter{})
}

func (ChromiumImporter) ID() string             { return "chromium_csv" }
func (ChromiumImporter) DisplayName() string    { return "Chrome / Brave / Edge / Opera / Vivaldi" }
func (ChromiumImporter) Extensions() []string   { return []string{".csv"} }

func (ChromiumImporter) Detect(filename string, head []byte) float64 {
	// Strict Chromium signature: name,url,username,password [,note]
	if headerMatches(head, "name", "url", "username", "password") {
		// Disambiguate from Bitwarden CSV (which also has name,url) by
		// checking for absence of Bitwarden-specific columns.
		line := strings.ToLower(firstNonEmptyLine(head))
		if strings.Contains(line, "login_uri") || strings.Contains(line, "login_username") {
			return 0
		}
		return 0.95
	}
	return 0
}

func (ChromiumImporter) Parse(r io.Reader, _ ParseOptions) (*ImportResult, error) {
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

		title := getCSVCol(row, idx, "name")
		urlStr := getCSVCol(row, idx, "url")
		username := getCSVCol(row, idx, "username")
		password := getCSVCol(row, idx, "password")
		note := getCSVCol(row, idx, "note")

		if password == "" && username == "" && note == "" {
			result.Skipped++
			continue
		}

		var urls []string
		if urlStr != "" {
			urls = []string{urlStr}
		}
		entry := ImportedEntry{
			Type:     model.EntryTypePassword,
			Title:    title,
			Username: username,
			Password: []byte(password),
			URLs:     urls,
			Notes:    note,
			Source:   "chromium_csv",
		}
		result.Entries = append(result.Entries, entry)
	}
	return result, nil
}
