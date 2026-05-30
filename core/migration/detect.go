package migration

import (
	"io"
	"os"
)

// headBytes is how many bytes of a file the detectors look at. Big enough to
// see a CSV header line and the opening of a JSON object/array.
const headBytes = 8 * 1024

// DetectFile sniffs the file at path and returns the available importers
// ordered by descending confidence. The returned slice may be empty.
//
// The function reads at most headBytes from disk; it does not load the
// full file. Callers should treat the returned slice as advisory: the UI
// always lets the user pick a different importer manually.
func DetectFile(path string) ([]DetectionResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	head := make([]byte, headBytes)
	n, err := io.ReadFull(f, head)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, err
	}
	return DefaultRegistry.Detect(path, head[:n]), nil
}

// ValidateSize checks that the file at path is within MaxFileSize.
func ValidateSize(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Size() > MaxFileSize {
		return ErrFileTooLarge
	}
	return nil
}

// OpenLimited opens the file at path for reading, wrapping it in an
// io.LimitedReader capped at MaxFileSize. The caller is responsible for
// closing the returned os.File via the cleanup function.
func OpenLimited(path string) (io.Reader, func() error, error) {
	if err := ValidateSize(path); err != nil {
		return nil, nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	limited := io.LimitReader(f, MaxFileSize+1)
	return limited, f.Close, nil
}
