package filevault

import (
	"bytes"
	"crypto/rand"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptDecryptRoundTrip_Small(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}

	original := []byte("Hello, PassQuantum File Vault!")

	var encrypted bytes.Buffer
	if err := EncryptFile(bytes.NewReader(original), &encrypted, key); err != nil {
		t.Fatalf("EncryptFile: %v", err)
	}

	var decrypted bytes.Buffer
	if err := DecryptFile(bytes.NewReader(encrypted.Bytes()), &decrypted, key); err != nil {
		t.Fatalf("DecryptFile: %v", err)
	}

	if !bytes.Equal(original, decrypted.Bytes()) {
		t.Errorf("round-trip mismatch:\n  got  %q\n  want %q", decrypted.Bytes(), original)
	}
}

func TestEncryptDecryptRoundTrip_ExactChunkSize(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	original := make([]byte, ChunkSize)
	rand.Read(original)

	var encrypted bytes.Buffer
	if err := EncryptFile(bytes.NewReader(original), &encrypted, key); err != nil {
		t.Fatalf("EncryptFile: %v", err)
	}

	var decrypted bytes.Buffer
	if err := DecryptFile(bytes.NewReader(encrypted.Bytes()), &decrypted, key); err != nil {
		t.Fatalf("DecryptFile: %v", err)
	}

	if !bytes.Equal(original, decrypted.Bytes()) {
		t.Error("round-trip mismatch for exact chunk-size file")
	}
}

func TestEncryptDecryptRoundTrip_MultiChunk(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	// 2.5 chunks worth of data
	original := make([]byte, ChunkSize*2+ChunkSize/2)
	rand.Read(original)

	var encrypted bytes.Buffer
	if err := EncryptFile(bytes.NewReader(original), &encrypted, key); err != nil {
		t.Fatalf("EncryptFile: %v", err)
	}

	var decrypted bytes.Buffer
	if err := DecryptFile(bytes.NewReader(encrypted.Bytes()), &decrypted, key); err != nil {
		t.Fatalf("DecryptFile: %v", err)
	}

	if !bytes.Equal(original, decrypted.Bytes()) {
		t.Error("round-trip mismatch for multi-chunk file")
	}
}

func TestEncryptDecryptRoundTrip_Empty(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	var encrypted bytes.Buffer
	if err := EncryptFile(bytes.NewReader(nil), &encrypted, key); err != nil {
		t.Fatalf("EncryptFile empty: %v", err)
	}

	var decrypted bytes.Buffer
	if err := DecryptFile(bytes.NewReader(encrypted.Bytes()), &decrypted, key); err != nil {
		t.Fatalf("DecryptFile empty: %v", err)
	}

	if decrypted.Len() != 0 {
		t.Errorf("expected empty output, got %d bytes", decrypted.Len())
	}
}

func TestDecryptFile_WrongKey(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	rand.Read(key1)
	rand.Read(key2)

	original := []byte("sensitive data")
	var encrypted bytes.Buffer
	if err := EncryptFile(bytes.NewReader(original), &encrypted, key1); err != nil {
		t.Fatalf("EncryptFile: %v", err)
	}

	var decrypted bytes.Buffer
	err := DecryptFile(bytes.NewReader(encrypted.Bytes()), &decrypted, key2)
	if err == nil {
		t.Fatal("expected error decrypting with wrong key")
	}
}

func TestEncryptFileWithProgress(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	size := int64(ChunkSize*3 + 1000)
	original := make([]byte, size)
	rand.Read(original)

	var progressCalls []int64
	onProgress := func(processed int64) {
		progressCalls = append(progressCalls, processed)
	}

	var encrypted bytes.Buffer
	if err := EncryptFileWithProgress(bytes.NewReader(original), &encrypted, key, size, onProgress); err != nil {
		t.Fatalf("EncryptFileWithProgress: %v", err)
	}

	if len(progressCalls) != 4 { // 3 full chunks + 1 partial
		t.Errorf("expected 4 progress calls, got %d", len(progressCalls))
	}
	if progressCalls[len(progressCalls)-1] != size {
		t.Errorf("final progress = %d, want %d", progressCalls[len(progressCalls)-1], size)
	}
}

func TestDecryptFileWithProgress(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	size := int64(ChunkSize + 500)
	original := make([]byte, size)
	rand.Read(original)

	var encrypted bytes.Buffer
	EncryptFile(bytes.NewReader(original), &encrypted, key)

	var progressCalls []int64
	var decrypted bytes.Buffer
	if err := DecryptFileWithProgress(bytes.NewReader(encrypted.Bytes()), &decrypted, key, size, func(p int64) {
		progressCalls = append(progressCalls, p)
	}); err != nil {
		t.Fatalf("DecryptFileWithProgress: %v", err)
	}

	if len(progressCalls) != 2 {
		t.Errorf("expected 2 progress calls, got %d", len(progressCalls))
	}
	if !bytes.Equal(original, decrypted.Bytes()) {
		t.Error("round-trip mismatch with progress")
	}
}

func TestDeriveChunkNonce_Uniqueness(t *testing.T) {
	base := make([]byte, 12)
	rand.Read(base)

	seen := make(map[string]bool)
	for i := uint32(0); i < 1000; i++ {
		n := deriveChunkNonce(base, i)
		key := string(n)
		if seen[key] {
			t.Fatalf("duplicate nonce at chunk %d", i)
		}
		seen[key] = true
	}
}

func TestSecureDelete(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secret.txt")

	if err := os.WriteFile(path, []byte("top secret content"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := SecureDelete(path); err != nil {
		t.Fatalf("SecureDelete: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("file still exists after SecureDelete")
	}
}

func TestSecureDelete_Nonexistent(t *testing.T) {
	if err := SecureDelete("/nonexistent/path/file.txt"); err != nil {
		t.Errorf("SecureDelete on nonexistent file should not error, got: %v", err)
	}
}

func TestManifestSerializeDeserialize(t *testing.T) {
	m := newManifest()
	m.add(&FileMetadata{
		UUID:         "test-uuid-1",
		OriginalName: "document.pdf",
		MimeType:     "application/pdf",
		Size:         12345,
		SHA256:       "abcdef1234567890",
	})
	m.add(&FileMetadata{
		UUID:         "test-uuid-2",
		OriginalName: "photo.jpg",
		MimeType:     "image/jpeg",
		Size:         54321,
		SHA256:       "0987654321fedcba",
	})

	data, err := serializeManifest(m)
	if err != nil {
		t.Fatalf("serializeManifest: %v", err)
	}

	restored, err := deserializeManifest(data)
	if err != nil {
		t.Fatalf("deserializeManifest: %v", err)
	}

	if len(restored.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(restored.Files))
	}
	if restored.Files[0].UUID != "test-uuid-1" {
		t.Errorf("file 0 UUID = %q", restored.Files[0].UUID)
	}
	if restored.Files[1].OriginalName != "photo.jpg" {
		t.Errorf("file 1 name = %q", restored.Files[1].OriginalName)
	}
}

func TestManifestAddRemoveFind(t *testing.T) {
	m := newManifest()

	f1 := &FileMetadata{UUID: "aaa", OriginalName: "a.txt"}
	f2 := &FileMetadata{UUID: "bbb", OriginalName: "b.txt"}
	m.add(f1)
	m.add(f2)

	if got := m.find("aaa"); got == nil || got.OriginalName != "a.txt" {
		t.Error("find(aaa) failed")
	}
	if got := m.find("ccc"); got != nil {
		t.Error("find(ccc) should return nil")
	}

	if !m.remove("aaa") {
		t.Error("remove(aaa) returned false")
	}
	if len(m.Files) != 1 {
		t.Errorf("expected 1 file after remove, got %d", len(m.Files))
	}
	if m.remove("aaa") {
		t.Error("second remove(aaa) should return false")
	}
}

func TestTempTracker(t *testing.T) {
	dir := t.TempDir()
	tracker := NewTempTracker()

	// Create two temp files
	for i := 0; i < 2; i++ {
		path := filepath.Join(dir, TempFilePath("test.txt"))
		os.MkdirAll(filepath.Dir(path), 0700)
		os.WriteFile(path, []byte("temp data"), 0600)
		tracker.Add(path)
	}

	tracker.CleanupAll()

	// After cleanup, tracker should be empty
	tracker.mu.Lock()
	if len(tracker.paths) != 0 {
		t.Error("tracker not empty after CleanupAll")
	}
	tracker.mu.Unlock()
}

func TestShortKey(t *testing.T) {
	shortKey := make([]byte, 16)
	err := EncryptFile(bytes.NewReader([]byte("test")), io.Discard, shortKey)
	if err == nil {
		t.Error("expected error for short key")
	}
}

func TestMimeFromExt(t *testing.T) {
	tests := []struct {
		ext, want string
	}{
		{".pdf", "application/pdf"},
		{".png", "image/png"},
		{".jpg", "image/jpeg"},
		{".txt", "text/plain"},
		{".unknown", "application/octet-stream"},
	}
	for _, tc := range tests {
		if got := mimeFromExt(tc.ext); got != tc.want {
			t.Errorf("mimeFromExt(%q) = %q, want %q", tc.ext, got, tc.want)
		}
	}
}
