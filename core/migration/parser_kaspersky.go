package migration

import (
	"bufio"
	"io"
	"strings"

	"passquantum/core/model"
)

// KasperskyTXTImporter parses the plain-text export emitted by Kaspersky
// Password Manager v24 and later. The file is organized in three sections,
// each labelled by a single word on its own line ("Websites", "Applications",
// "Notes"), with individual entries separated by a line containing "---".
// Within an entry, fields are written as "Key: value" pairs that may span
// multiple lines (Comment fields in particular).
//
// Example:
//
//	Websites
//
//	Website name: GitHub
//	Website URL: https://github.com
//	Login name: personal
//	Login: octocat
//	Password: ghpw
//	Comment: notes here
//
//	---
//
//	Website name: ...
//
// The Kaspersky .edb backup format is encrypted with a vendor-proprietary
// scheme and is intentionally out of scope.
type KasperskyTXTImporter struct{}

func init() {
	DefaultRegistry.Register(&KasperskyTXTImporter{})
}

func (KasperskyTXTImporter) ID() string           { return "kaspersky_txt" }
func (KasperskyTXTImporter) DisplayName() string  { return "Kaspersky Password Manager (TXT)" }
func (KasperskyTXTImporter) Extensions() []string { return []string{".txt"} }

func (KasperskyTXTImporter) Detect(_ string, head []byte) float64 {
	// The file always begins with "Websites" (possibly preceded by a BOM
	// and blank lines).
	first := strings.TrimSpace(firstNonEmptyLine(head))
	first = strings.TrimPrefix(first, "\ufeff")
	if strings.EqualFold(first, "Websites") {
		return 0.95
	}
	return 0
}

// section identifies which kind of record we are currently accumulating.
type kasperskySection int

const (
	ksSecNone kasperskySection = iota
	ksSecWebsites
	ksSecApplications
	ksSecNotes
)

func (KasperskyTXTImporter) Parse(r io.Reader, _ ParseOptions) (*ImportResult, error) {
	result := &ImportResult{}
	scanner := bufio.NewScanner(r)
	// Allow up to 1 MB per line for long Comment fields.
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	section := ksSecNone
	current := map[string]string{}
	lastKey := ""

	commitCurrent := func() {
		defer func() {
			current = map[string]string{}
			lastKey = ""
		}()
		if len(current) == 0 {
			return
		}
		entry := buildKasperskyEntry(section, current)
		if entry == nil {
			result.Skipped++
			return
		}
		result.Entries = append(result.Entries, *entry)
	}

	for scanner.Scan() {
		line := scanner.Text()
		// Strip a leading BOM on the very first line.
		line = strings.TrimPrefix(line, "\ufeff")
		trimmed := strings.TrimRight(line, "\r")

		// Entry separator.
		if strings.TrimSpace(trimmed) == "---" {
			commitCurrent()
			continue
		}

		// Section header.
		switch strings.TrimSpace(trimmed) {
		case "Websites":
			commitCurrent()
			section = ksSecWebsites
			continue
		case "Applications":
			commitCurrent()
			section = ksSecApplications
			continue
		case "Notes":
			commitCurrent()
			section = ksSecNotes
			continue
		}

		// Empty line: keep accumulating but record a paragraph break in the
		// last key's value (so multi-paragraph Comment fields survive).
		if strings.TrimSpace(trimmed) == "" {
			if lastKey != "" {
				current[lastKey] += "\n"
			}
			continue
		}

		// Key: value line.
		if idx := strings.Index(trimmed, ":"); idx > 0 {
			key := strings.TrimSpace(trimmed[:idx])
			value := strings.TrimSpace(trimmed[idx+1:])
			current[key] = value
			lastKey = key
			continue
		}

		// Continuation line (part of a multi-line Comment value).
		if lastKey != "" {
			current[lastKey] += "\n" + trimmed
		}
	}
	// Commit the final record if the file did not end with "---".
	commitCurrent()

	if err := scanner.Err(); err != nil {
		return result, err
	}
	return result, nil
}

func buildKasperskyEntry(section kasperskySection, fields map[string]string) *ImportedEntry {
	get := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := fields[k]; ok && strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		}
		return ""
	}

	switch section {
	case ksSecWebsites:
		name := get("Website name")
		urlStr := get("Website URL")
		// "Login" usually carries the username; "Login name" can be a
		// display alias. Prefer "Login" but fall back to "Login name".
		username := get("Login", "Login name")
		password := get("Password")
		notes := get("Comment")

		if name == "" && urlStr == "" && username == "" && password == "" {
			return nil
		}
		var urls []string
		if urlStr != "" {
			urls = []string{urlStr}
		}
		return &ImportedEntry{
			Type:     model.EntryTypePassword,
			Title:    name,
			Username: username,
			Password: []byte(password),
			URLs:     urls,
			Notes:    notes,
			Source:   "kaspersky_txt",
		}

	case ksSecApplications:
		name := get("Application name", "Name")
		username := get("Login", "Login name")
		password := get("Password")
		notes := get("Comment")

		if name == "" && username == "" && password == "" {
			return nil
		}
		return &ImportedEntry{
			Type:     model.EntryTypePassword,
			Title:    name,
			Username: username,
			Password: []byte(password),
			Notes:    notes,
			Source:   "kaspersky_txt",
		}

	case ksSecNotes:
		title := get("Name", "Title")
		// Kaspersky's note body lives under "Text" (or "Note" on some
		// localized exports).
		text := get("Text", "Note", "Content")
		if title == "" && text == "" {
			return nil
		}
		return &ImportedEntry{
			Type:   model.EntryTypeNote,
			Title:  title,
			Notes:  text,
			Source: "kaspersky_txt",
		}

	default:
		return nil
	}
}
