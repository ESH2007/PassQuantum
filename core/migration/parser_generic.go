package migration

import (
	"errors"
	"io"
	"strings"

	"passquantum/core/model"
)

// GenericCSVImporter is the universal fallback. It tries to recognize common
// column names heuristically; if the UI passes ParseOptions.ColumnMapping it
// uses that mapping verbatim instead.
type GenericCSVImporter struct{}

func init() {
	DefaultRegistry.Register(&GenericCSVImporter{})
}

func (GenericCSVImporter) ID() string           { return "generic_csv" }
func (GenericCSVImporter) DisplayName() string  { return "Generic CSV (auto-map columns)" }
func (GenericCSVImporter) Extensions() []string { return []string{".csv"} }

// Detect always returns a low non-zero score so this importer surfaces as a
// fallback when no other importer claims the file, but never wins over a
// specific match.
func (GenericCSVImporter) Detect(_ string, head []byte) float64 {
	if hasUTF8(head) && strings.Contains(string(head), ",") {
		return 0.2
	}
	return 0
}

// columnAliases maps a destination field to the lowercase header names that
// frequently identify it across exporters. The order within each list is
// significant: the first matching alias wins.
var columnAliases = map[string][]string{
	"title":    {"title", "name", "service", "site"},
	"username": {"username", "user", "login", "email", "account"},
	"password": {"password", "pass", "passwd", "secret"},
	"url":      {"url", "uri", "website", "link", "address"},
	"notes":    {"notes", "note", "comment", "description"},
	"totp":     {"totp", "otp", "2fa", "totpsecret"},
	"folder":   {"folder", "group", "category", "grouping"},
}

func (GenericCSVImporter) Parse(r io.Reader, opts ParseOptions) (*ImportResult, error) {
	cr := newCSVReader(r)
	idx, _, err := readCSVHeader(cr)
	if err != nil {
		return nil, err
	}

	// Build the destination -> column index map.
	colByField := make(map[string]int)
	if len(opts.ColumnMapping) > 0 {
		for field, headerName := range opts.ColumnMapping {
			if i, ok := idx[strings.ToLower(strings.TrimSpace(headerName))]; ok {
				colByField[field] = i
			}
		}
	} else {
		for field, aliases := range columnAliases {
			for _, alias := range aliases {
				if i, ok := idx[alias]; ok {
					colByField[field] = i
					break
				}
			}
		}
	}

	if _, ok := colByField["password"]; !ok {
		return nil, errors.New("generic CSV: could not find a password column; please provide an explicit column mapping")
	}

	get := func(row []string, field string) string {
		i, ok := colByField[field]
		if !ok || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
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

		title := get(row, "title")
		username := get(row, "username")
		password := get(row, "password")
		urlStr := get(row, "url")
		notes := get(row, "notes")
		totp := get(row, "totp")
		folder := get(row, "folder")

		if password == "" && username == "" && notes == "" {
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
			Notes:    notes,
			TOTP:     totp,
			Folder:   folder,
			Source:   "generic_csv",
		})
	}
	return result, nil
}
