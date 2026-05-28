package filevault

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/cloudflare/circl/kem/kyber/kyber768"
	"github.com/google/uuid"

	"passquantum/core/crypto"
)

const (
	manifestFile   = "manifest.enc"
	maxFileSize    = 1 << 30 // 1 GB default limit
	filesDirPerms  = 0700
	storedFilePerms = 0600
)

// Store manages encrypted file storage for a single vault.
type Store struct {
	vaultDir    string
	password    string
	pubKey      *kyber768.PublicKey
	privKey     *kyber768.PrivateKey
	manifest    *FileManifest
	tempTracker *TempTracker
}

// NewStore opens (or creates) a file store for the given vault.
func NewStore(vaultName, password string, pubKey *kyber768.PublicKey, privKey *kyber768.PrivateKey, tracker *TempTracker) (*Store, error) {
	baseDir, err := getFilesDir(vaultName)
	if err != nil {
		return nil, err
	}
	s := &Store{
		vaultDir:    baseDir,
		password:    password,
		pubKey:      pubKey,
		privKey:     privKey,
		tempTracker: tracker,
	}
	if err := s.LoadManifest(); err != nil {
		return nil, err
	}
	return s, nil
}

// LoadManifest reads and decrypts the manifest, or creates a new one.
func (s *Store) LoadManifest() error {
	path := filepath.Join(s.vaultDir, manifestFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			s.manifest = newManifest()
			return nil
		}
		return fmt.Errorf("filevault: read manifest: %w", err)
	}

	plaintext, err := crypto.PQVaultDecrypt(data, s.password)
	if err != nil {
		return fmt.Errorf("filevault: decrypt manifest: %w", err)
	}

	m, err := deserializeManifest(plaintext)
	if err != nil {
		return err
	}
	s.manifest = m
	return nil
}

// SaveManifest encrypts and writes the manifest to disk.
func (s *Store) SaveManifest() error {
	data, err := serializeManifest(s.manifest)
	if err != nil {
		return err
	}

	encrypted, err := crypto.PQVaultEncrypt(data, s.password)
	if err != nil {
		return fmt.Errorf("filevault: encrypt manifest: %w", err)
	}

	path := filepath.Join(s.vaultDir, manifestFile)
	if err := os.WriteFile(path, encrypted, storedFilePerms); err != nil {
		return fmt.Errorf("filevault: write manifest: %w", err)
	}
	return nil
}

// StoreFile encrypts a file from srcPath and adds it to the vault.
func (s *Store) StoreFile(srcPath string, onProgress func(int64)) (*FileMetadata, error) {
	info, err := os.Stat(srcPath)
	if err != nil {
		return nil, fmt.Errorf("filevault: stat source: %w", err)
	}
	if info.Size() > maxFileSize {
		return nil, fmt.Errorf("filevault: file too large (%d bytes, max %d)", info.Size(), maxFileSize)
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return nil, fmt.Errorf("filevault: open source: %w", err)
	}
	defer src.Close()

	// Compute SHA-256 while reading
	hasher := sha256.New()
	tee := io.TeeReader(src, hasher)

	// Per-file Kyber KEM
	ct, ss, err := crypto.Encapsulate(s.pubKey)
	if err != nil {
		return nil, fmt.Errorf("filevault: encapsulate: %w", err)
	}
	defer crypto.WipeBytes(ss)

	fileUUID := uuid.New().String()
	dstPath := filepath.Join(s.vaultDir, fileUUID+".bin")

	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, storedFilePerms)
	if err != nil {
		return nil, fmt.Errorf("filevault: create dest: %w", err)
	}
	defer dst.Close()

	if err := EncryptFileWithProgress(tee, dst, ss, info.Size(), onProgress); err != nil {
		os.Remove(dstPath)
		return nil, err
	}

	// Detect MIME type from file extension (no need to re-read)
	mimeType := http.DetectContentType([]byte{})
	ext := filepath.Ext(srcPath)
	if m := mimeFromExt(ext); m != "" {
		mimeType = m
	}

	meta := &FileMetadata{
		UUID:            fileUUID,
		OriginalName:    filepath.Base(srcPath),
		MimeType:        mimeType,
		Size:            info.Size(),
		SHA256:          fmt.Sprintf("%x", hasher.Sum(nil)),
		StoredAt:        time.Now(),
		KyberCiphertext: ct,
	}

	s.manifest.add(meta)
	if err := s.SaveManifest(); err != nil {
		os.Remove(dstPath)
		s.manifest.remove(fileUUID)
		return nil, err
	}

	return meta, nil
}

// RetrieveFile decrypts a stored file to dstPath.
func (s *Store) RetrieveFile(fileUUID, dstPath string, onProgress func(int64)) error {
	meta := s.manifest.find(fileUUID)
	if meta == nil {
		return fmt.Errorf("filevault: file %q not found in manifest", fileUUID)
	}

	ss, err := crypto.Decapsulate(meta.KyberCiphertext, s.privKey)
	if err != nil {
		return fmt.Errorf("filevault: decapsulate: %w", err)
	}
	defer crypto.WipeBytes(ss)

	srcPath := filepath.Join(s.vaultDir, fileUUID+".bin")
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("filevault: open encrypted: %w", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("filevault: create dest: %w", err)
	}
	defer dst.Close()

	return DecryptFileWithProgress(src, dst, ss, meta.Size, onProgress)
}

// OpenFile decrypts to a temp file and opens it with the system default app.
// Returns the temp file path (tracked for cleanup).
func (s *Store) OpenFile(fileUUID string) (string, error) {
	meta := s.manifest.find(fileUUID)
	if meta == nil {
		return "", fmt.Errorf("filevault: file %q not found", fileUUID)
	}

	tmpPath := TempFilePath(meta.OriginalName)
	if err := s.RetrieveFile(fileUUID, tmpPath, nil); err != nil {
		return "", err
	}

	if s.tempTracker != nil {
		s.tempTracker.Add(tmpPath)
	}

	if err := openWithSystem(tmpPath); err != nil {
		return tmpPath, fmt.Errorf("filevault: system open: %w", err)
	}

	return tmpPath, nil
}

// DecryptToMemory decrypts a stored file entirely into memory and returns
// the plaintext bytes. Use only for small files (e.g. image thumbnails).
func (s *Store) DecryptToMemory(fileUUID string) ([]byte, error) {
	meta := s.manifest.find(fileUUID)
	if meta == nil {
		return nil, fmt.Errorf("filevault: file %q not found in manifest", fileUUID)
	}

	ss, err := crypto.Decapsulate(meta.KyberCiphertext, s.privKey)
	if err != nil {
		return nil, fmt.Errorf("filevault: decapsulate: %w", err)
	}
	defer crypto.WipeBytes(ss)

	srcPath := filepath.Join(s.vaultDir, fileUUID+".bin")
	src, err := os.Open(srcPath)
	if err != nil {
		return nil, fmt.Errorf("filevault: open encrypted: %w", err)
	}
	defer src.Close()

	var buf bytes.Buffer
	if err := DecryptFile(src, &buf, ss); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DeleteFile securely removes an encrypted file and its manifest entry.
func (s *Store) DeleteFile(fileUUID string) error {
	meta := s.manifest.find(fileUUID)
	if meta == nil {
		return fmt.Errorf("filevault: file %q not found in manifest", fileUUID)
	}

	s.manifest.remove(fileUUID)

	binPath := filepath.Join(s.vaultDir, fileUUID+".bin")
	if err := SecureDelete(binPath); err != nil {
		s.manifest.add(meta)
		return err
	}

	return s.SaveManifest()
}

// ListFiles returns all file metadata in the manifest.
func (s *Store) ListFiles() []*FileMetadata {
	if s.manifest == nil {
		return nil
	}
	return s.manifest.Files
}

// Close flushes the manifest and cleans up temp files.
func (s *Store) Close() {
	_ = s.SaveManifest()
	if s.tempTracker != nil {
		s.tempTracker.CleanupAll()
	}
	crypto.WipeBytes([]byte(s.password))
}

func getFilesDir(vaultName string) (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("filevault: config dir: %w", err)
	}

	dir := filepath.Join(configDir, "passquantum", "files", vaultName)
	if err := os.MkdirAll(dir, filesDirPerms); err != nil {
		return "", fmt.Errorf("filevault: create files dir: %w", err)
	}
	return dir, nil
}

func openWithSystem(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", path)
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	return cmd.Start()
}

func mimeFromExt(ext string) string {
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".doc", ".docx":
		return "application/msword"
	case ".xls", ".xlsx":
		return "application/vnd.ms-excel"
	case ".ppt", ".pptx":
		return "application/vnd.ms-powerpoint"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".txt":
		return "text/plain"
	case ".zip":
		return "application/zip"
	case ".tar", ".gz", ".tgz":
		return "application/gzip"
	default:
		return "application/octet-stream"
	}
}
