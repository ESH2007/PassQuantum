package migration

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"passquantum/core/model"
)

// DashlaneImporter handles Dashlane's ZIP export. The archive bundles up to
// five CSVs:
//
//   - credentials.csv  — logins (always present)
//   - payments.csv     — credit cards
//   - securenotes.csv  — secure notes
//   - ids.csv          — government IDs
//   - personalinfo.csv — identity / contact data
//
// Phase 2 implements credentials + securenotes + payments. ids.csv and
// personalinfo.csv are folded into secure notes so the user does not lose
// the data.
type DashlaneImporter struct{}

func init() {
	DefaultRegistry.Register(&DashlaneImporter{})
}

func (DashlaneImporter) ID() string           { return "dashlane_zip" }
func (DashlaneImporter) DisplayName() string  { return "Dashlane" }
func (DashlaneImporter) Extensions() []string { return []string{".zip"} }

func (DashlaneImporter) Detect(filename string, head []byte) float64 {
	if !looksLikeZIP(head) {
		return 0
	}
	if strings.Contains(strings.ToLower(filename), "dashlane") {
		return 0.9
	}
	// We cannot peek inside the ZIP cheaply here; let other ZIP-based
	// importers (1Password, Proton Pass) win when their hints are stronger.
	return 0.15
}

func (DashlaneImporter) Parse(r io.Reader, _ ParseOptions) (*ImportResult, error) {
	raw, err := readWholeReader(r)
	if err != nil {
		return nil, err
	}
	zr, _, err := readWholeZIP(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	csvs, err := listZIPFilesBySuffix(zr, ".csv")
	if err != nil {
		return nil, err
	}
	if len(csvs) == 0 {
		return nil, errors.New("dashlane: no CSV files found inside archive")
	}

	result := &ImportResult{}
	for name, data := range csvs {
		lower := strings.ToLower(name)
		switch {
		case strings.Contains(lower, "credential"):
			parseDashlaneCredentials(bytes.NewReader(data), result)
		case strings.Contains(lower, "payment"):
			parseDashlanePayments(bytes.NewReader(data), result)
		case strings.Contains(lower, "securenote"):
			parseDashlaneSecureNotes(bytes.NewReader(data), result)
		case strings.Contains(lower, "id"):
			parseDashlaneAsNotes(bytes.NewReader(data), name, result)
		case strings.Contains(lower, "personal"):
			parseDashlaneAsNotes(bytes.NewReader(data), name, result)
		default:
			parseDashlaneAsNotes(bytes.NewReader(data), name, result)
		}
	}
	return result, nil
}

func parseDashlaneCredentials(r io.Reader, result *ImportResult) {
	cr := newCSVReader(r)
	idx, _, err := readCSVHeader(cr)
	if err != nil {
		return
	}
	for {
		row, err := cr.Read()
		if err != nil {
			if err == io.EOF {
				return
			}
			result.Skipped++
			continue
		}
		title := getCSVCol(row, idx, "title")
		username := getCSVCol(row, idx, "username")
		password := getCSVCol(row, idx, "password")
		urlStr := getCSVCol(row, idx, "url")
		note := getCSVCol(row, idx, "note")
		category := getCSVCol(row, idx, "category")
		otp := getCSVCol(row, idx, "otpsecret")

		if password == "" && username == "" && note == "" {
			result.Skipped++
			continue
		}

		fields := map[string]string{}
		for _, alt := range []string{"username2", "username3", "email"} {
			if v := getCSVCol(row, idx, alt); v != "" && v != username {
				fields[alt] = v
			}
		}

		var urls []string
		if urlStr != "" {
			urls = []string{urlStr}
		}
		var totp string
		if otp != "" {
			totp = NormalizeTOTP(otp, title, username)
		}
		entry := ImportedEntry{
			Type:     model.EntryTypePassword,
			Title:    title,
			Username: username,
			Password: []byte(password),
			URLs:     urls,
			Notes:    note,
			Folder:   category,
			TOTP:     totp,
			Source:   "dashlane_zip",
		}
		if len(fields) > 0 {
			entry.Fields = fields
		}
		result.Entries = append(result.Entries, entry)
	}
}

func parseDashlanePayments(r io.Reader, result *ImportResult) {
	cr := newCSVReader(r)
	idx, _, err := readCSVHeader(cr)
	if err != nil {
		return
	}
	for {
		row, err := cr.Read()
		if err != nil {
			if err == io.EOF {
				return
			}
			result.Skipped++
			continue
		}
		// Dashlane payment columns: type,account_name,account_holder,
		// cc_number,code,expiration_month,expiration_year,...
		typ := strings.ToLower(getCSVCol(row, idx, "type"))
		name := getCSVCol(row, idx, "account_name")
		holder := getCSVCol(row, idx, "account_holder")
		number := getCSVCol(row, idx, "cc_number")
		code := getCSVCol(row, idx, "code")
		month := getCSVCol(row, idx, "expiration_month")
		year := getCSVCol(row, idx, "expiration_year")

		if number == "" {
			result.Skipped++
			continue
		}

		subtype := "credit"
		if strings.Contains(typ, "debit") {
			subtype = "debit"
		}

		result.Entries = append(result.Entries, ImportedEntry{
			Type:   model.EntryTypeCard,
			Title:  name,
			Source: "dashlane_zip",
			Card: &CardData{
				Subtype:  subtype,
				Holder:   holder,
				Number:   []byte(number),
				CVV:      []byte(code),
				ExpMonth: month,
				ExpYear:  year,
			},
		})
	}
}

func parseDashlaneSecureNotes(r io.Reader, result *ImportResult) {
	cr := newCSVReader(r)
	idx, _, err := readCSVHeader(cr)
	if err != nil {
		return
	}
	for {
		row, err := cr.Read()
		if err != nil {
			if err == io.EOF {
				return
			}
			result.Skipped++
			continue
		}
		title := getCSVCol(row, idx, "title")
		note := getCSVCol(row, idx, "note")
		category := getCSVCol(row, idx, "category")
		if title == "" && note == "" {
			result.Skipped++
			continue
		}
		result.Entries = append(result.Entries, ImportedEntry{
			Type:   model.EntryTypeNote,
			Title:  title,
			Notes:  note,
			Folder: category,
			Source: "dashlane_zip",
		})
	}
}

// parseDashlaneAsNotes ingests an unstructured Dashlane CSV (ids.csv,
// personalinfo.csv or any unrecognized table) as one secure note per row.
// Every column becomes a "key: value" line inside the note body so the data
// is preserved verbatim without polluting the vault with custom types.
func parseDashlaneAsNotes(r io.Reader, sourceName string, result *ImportResult) {
	cr := newCSVReader(r)
	_, cols, err := readCSVHeader(cr)
	if err != nil {
		return
	}
	for {
		row, err := cr.Read()
		if err != nil {
			if err == io.EOF {
				return
			}
			result.Skipped++
			continue
		}
		var b strings.Builder
		titleCandidate := ""
		hasContent := false
		for i, value := range row {
			if i >= len(cols) {
				break
			}
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			hasContent = true
			b.WriteString(cols[i])
			b.WriteString(": ")
			b.WriteString(value)
			b.WriteByte('\n')
			if titleCandidate == "" && (cols[i] == "title" || cols[i] == "name") {
				titleCandidate = value
			}
		}
		if !hasContent {
			result.Skipped++
			continue
		}
		title := titleCandidate
		if title == "" {
			title = strings.TrimSuffix(sourceName, ".csv")
		}
		result.Entries = append(result.Entries, ImportedEntry{
			Type:   model.EntryTypeNote,
			Title:  title,
			Notes:  strings.TrimRight(b.String(), "\n"),
			Source: "dashlane_zip",
		})
	}
}
