package migration

import (
	"errors"
	"io"

	"passquantum/core/model"
)

// FirefoxImporter parses the CSV exported by Firefox's about:logins page.
// Canonical header:
//
//	url,username,password,httpRealm,formActionOrigin,guid,
//	  timeCreated,timeLastUsed,timePasswordChanged
//
// Firefox does not include a separate "name" column, so the title is
// derived from the URL host by the mapper.
type FirefoxImporter struct{}

func init() {
	DefaultRegistry.Register(&FirefoxImporter{})
}

func (FirefoxImporter) ID() string           { return "firefox_csv" }
func (FirefoxImporter) DisplayName() string  { return "Mozilla Firefox" }
func (FirefoxImporter) Extensions() []string { return []string{".csv"} }

func (FirefoxImporter) Detect(_ string, head []byte) float64 {
	// httpRealm + formActionOrigin are unambiguously Firefox.
	if headerMatches(head, "httprealm", "formactionorigin") {
		return 0.98
	}
	return 0
}

func (FirefoxImporter) Parse(r io.Reader, _ ParseOptions) (*ImportResult, error) {
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

		if password == "" && username == "" {
			result.Skipped++
			continue
		}

		var urls []string
		if urlStr != "" {
			urls = []string{urlStr}
		}
		entry := ImportedEntry{
			Type:     model.EntryTypePassword,
			Title:    "", // mapper derives from URL host
			Username: username,
			Password: []byte(password),
			URLs:     urls,
			Source:   "firefox_csv",
		}
		result.Entries = append(result.Entries, entry)
	}
	return result, nil
}
