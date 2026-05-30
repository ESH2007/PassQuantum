package migration

import (
	"bufio"
	"encoding/csv"
	"errors"
	"io"
	"strings"
	"unicode/utf8"
)

// utf8BOM is the byte order mark some browsers prefix CSV exports with.
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// newCSVReader returns an encoding/csv reader wrapped to skip a leading
// UTF-8 BOM, accept lazy quotes (some exporters emit unescaped quotes inside
// fields) and tolerate variable-length records (so a missing trailing column
// in a row does not abort the whole import).
func newCSVReader(r io.Reader) *csv.Reader {
	br := bufio.NewReader(r)
	// Peek and skip BOM if present.
	if peek, err := br.Peek(len(utf8BOM)); err == nil {
		for i, b := range utf8BOM {
			if peek[i] != b {
				goto noBOM
			}
		}
		_, _ = br.Discard(len(utf8BOM))
	}
noBOM:
	cr := csv.NewReader(br)
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1
	cr.TrimLeadingSpace = false
	return cr
}

// readCSVHeader reads the first record as the header row, normalizing each
// column name to lowercase + trimmed whitespace, and returns a lookup map
// from normalized header name to its column index. Duplicate headers keep
// the first index.
func readCSVHeader(cr *csv.Reader) (map[string]int, []string, error) {
	rec, err := cr.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil, errors.New("empty CSV")
		}
		return nil, nil, err
	}
	cols := make([]string, len(rec))
	idx := make(map[string]int, len(rec))
	for i, raw := range rec {
		name := strings.ToLower(strings.TrimSpace(raw))
		cols[i] = name
		if _, exists := idx[name]; !exists {
			idx[name] = i
		}
	}
	return idx, cols, nil
}

// getCSVCol returns the value at the given column name from a record,
// or "" if the column was not present in the header or the row is short.
func getCSVCol(record []string, idx map[string]int, name string) string {
	i, ok := idx[name]
	if !ok || i >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[i])
}

// hasUTF8 cheaply checks that a byte slice is valid UTF-8 to avoid
// reporting binary garbage as a parse error.
func hasUTF8(b []byte) bool {
	return utf8.Valid(b)
}

// headerMatches reports whether all of the required header names are present
// (lowercased, trimmed) in the first non-empty line of head. It is meant for
// cheap Detect() implementations that only need to glance at the first line.
func headerMatches(head []byte, required ...string) bool {
	line := firstNonEmptyLine(head)
	if line == "" {
		return false
	}
	line = strings.ToLower(strings.TrimSpace(line))
	// Strip a leading UTF-8 BOM if present (U+FEFF).
	line = strings.TrimPrefix(line, "\ufeff")
	for _, want := range required {
		if !strings.Contains(line, strings.ToLower(want)) {
			return false
		}
	}
	return true
}

func firstNonEmptyLine(head []byte) string {
	s := string(head)
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) != "" {
			return line
		}
	}
	return ""
}
