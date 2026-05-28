package filevault

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const tempPrefix = "pq_filevault_"

// TempTracker keeps track of decrypted temporary files so they can be
// securely wiped on vault lock or application exit.
type TempTracker struct {
	mu    sync.Mutex
	paths []string
}

func NewTempTracker() *TempTracker {
	return &TempTracker{}
}

// Add registers a temporary file path for later cleanup.
func (t *TempTracker) Add(path string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.paths = append(t.paths, path)
}

// CleanupAll securely deletes all tracked temp files.
func (t *TempTracker) CleanupAll() {
	t.mu.Lock()
	paths := make([]string, len(t.paths))
	copy(paths, t.paths)
	t.paths = nil
	t.mu.Unlock()

	for _, p := range paths {
		_ = SecureDelete(p)
	}
}

// SecureDelete overwrites a file with random bytes, syncs, then removes it.
func SecureDelete(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("filevault: stat for secure delete: %w", err)
	}

	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("filevault: open for overwrite: %w", err)
	}

	size := info.Size()
	buf := make([]byte, 4096)
	var written int64
	for written < size {
		n := int64(len(buf))
		if remaining := size - written; remaining < n {
			n = remaining
		}
		if _, err := rand.Read(buf[:n]); err != nil {
			f.Close()
			return fmt.Errorf("filevault: random fill: %w", err)
		}
		if _, err := f.Write(buf[:n]); err != nil {
			f.Close()
			return fmt.Errorf("filevault: overwrite: %w", err)
		}
		written += n
	}

	_ = f.Sync()
	f.Close()

	return os.Remove(path)
}

// CleanupOrphans removes any leftover temp files from a previous session.
// Call this on application startup.
func CleanupOrphans() error {
	tmpDir := os.TempDir()
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return fmt.Errorf("filevault: read temp dir: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), tempPrefix) {
			_ = SecureDelete(filepath.Join(tmpDir, e.Name()))
		}
	}
	return nil
}

// TempFilePath returns a path in the system temp directory with the pq prefix
// and the original file extension preserved (so OS can open it with the right app).
func TempFilePath(originalName string) string {
	ext := filepath.Ext(originalName)
	return filepath.Join(os.TempDir(), tempPrefix+randomHex(8)+ext)
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}
