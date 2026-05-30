package migration

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"
)

// zipMagic is the 4-byte signature of every ZIP local file header. ZIP-based
// exports (.1pux, Dashlane .zip, Proton Pass .zip) all start with this.
var zipMagic = []byte{0x50, 0x4B, 0x03, 0x04}

// looksLikeZIP reports whether the head buffer starts with the ZIP magic.
// Used by archive-based importers in their Detect() implementations.
func looksLikeZIP(head []byte) bool {
	return len(head) >= 4 && bytes.Equal(head[:4], zipMagic)
}

// readWholeReader pulls the entire reader into memory, bounded by
// MaxFileSize+1 so the caller can detect a too-large file by comparing the
// returned length against MaxFileSize.
func readWholeReader(r io.Reader) ([]byte, error) {
	buf, err := io.ReadAll(io.LimitReader(r, MaxFileSize+1))
	if err != nil {
		return nil, err
	}
	if len(buf) > MaxFileSize {
		return nil, ErrFileTooLarge
	}
	return buf, nil
}

// readWholeZIP reads r entirely into memory and returns a zip.Reader over it.
// The returned cleanup function does nothing (kept for symmetry).
//
// Reading into memory rather than tempfiling is safe because MaxFileSize is
// already enforced upstream; for a 100 MB import the heap pressure is
// negligible compared to the post-quantum encryption work that follows.
func readWholeZIP(r io.Reader) (*zip.Reader, []byte, error) {
	raw, err := readWholeReader(r)
	if err != nil {
		return nil, nil, err
	}
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return nil, nil, fmt.Errorf("open zip: %w", err)
	}
	return zr, raw, nil
}

// findZIPFile searches a ZIP archive for the first entry whose name matches
// any of the candidate suffixes (case-insensitive). It returns the entry's
// decompressed bytes or an empty slice if no match is found. Directory
// entries are skipped.
func findZIPFile(zr *zip.Reader, candidates ...string) ([]byte, string, error) {
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		lower := strings.ToLower(f.Name)
		for _, want := range candidates {
			if strings.HasSuffix(lower, strings.ToLower(want)) {
				rc, err := f.Open()
				if err != nil {
					return nil, f.Name, fmt.Errorf("open zip entry %s: %w", f.Name, err)
				}
				data, err := io.ReadAll(io.LimitReader(rc, MaxFileSize+1))
				rc.Close()
				if err != nil {
					return nil, f.Name, fmt.Errorf("read zip entry %s: %w", f.Name, err)
				}
				if len(data) > MaxFileSize {
					return nil, f.Name, ErrFileTooLarge
				}
				return data, f.Name, nil
			}
		}
	}
	return nil, "", nil
}

// listZIPFilesBySuffix returns the decompressed bytes of every ZIP entry
// whose name ends with one of the suffixes (case-insensitive). Used by
// the Dashlane importer to enumerate the multiple CSVs in its archive.
func listZIPFilesBySuffix(zr *zip.Reader, suffix string) (map[string][]byte, error) {
	suffix = strings.ToLower(suffix)
	out := make(map[string][]byte)
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(f.Name), suffix) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("open zip entry %s: %w", f.Name, err)
		}
		data, err := io.ReadAll(io.LimitReader(rc, MaxFileSize+1))
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("read zip entry %s: %w", f.Name, err)
		}
		if len(data) > MaxFileSize {
			return nil, ErrFileTooLarge
		}
		out[f.Name] = data
	}
	return out, nil
}
