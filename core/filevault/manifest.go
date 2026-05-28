package filevault

import (
	"encoding/json"
	"fmt"
	"time"
)

// FileManifest is the encrypted index of all stored files in a vault.
type FileManifest struct {
	Version int             `json:"version"`
	Files   []*FileMetadata `json:"files"`
}

// FileMetadata describes one encrypted file in the vault.
type FileMetadata struct {
	UUID            string    `json:"uuid"`
	OriginalName    string    `json:"original_name"`
	MimeType        string    `json:"mime_type"`
	Size            int64     `json:"size"`
	SHA256          string    `json:"sha256"`
	StoredAt        time.Time `json:"stored_at"`
	KyberCiphertext []byte    `json:"kyber_ct"`
}

func newManifest() *FileManifest {
	return &FileManifest{Version: 1}
}

func (m *FileManifest) add(f *FileMetadata) {
	m.Files = append(m.Files, f)
}

func (m *FileManifest) remove(uuid string) bool {
	for i, f := range m.Files {
		if f.UUID == uuid {
			m.Files = append(m.Files[:i], m.Files[i+1:]...)
			return true
		}
	}
	return false
}

func (m *FileManifest) find(uuid string) *FileMetadata {
	for _, f := range m.Files {
		if f.UUID == uuid {
			return f
		}
	}
	return nil
}

func serializeManifest(m *FileManifest) ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("filevault: marshal manifest: %w", err)
	}
	return data, nil
}

func deserializeManifest(data []byte) (*FileManifest, error) {
	var m FileManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("filevault: unmarshal manifest: %w", err)
	}
	return &m, nil
}
