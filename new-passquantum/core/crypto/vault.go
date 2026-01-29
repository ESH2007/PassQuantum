package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
)

// VaultFile represents an encrypted vault file structure
type VaultFile struct {
	Version       uint8
	KDFParams     KDFParams
	HMAC          [32]byte // SHA256 HMAC for integrity
	EncryptedData []byte
}

// EncryptVault encrypts password entries into a vault file
// The vault file contains:
// - Version (1 byte)
// - KDF params (26 bytes)
// - HMAC (32 bytes)
// - Encrypted data (variable)
func EncryptVault(plaintext []byte, encryptionKey []byte, verificationKey []byte, params KDFParams) (*VaultFile, error) {
	// Generate a random nonce for AES-GCM
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Create AES-GCM cipher
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Encrypt the plaintext
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Prepend nonce to ciphertext
	encryptedData := append(nonce, ciphertext...)

	vault := &VaultFile{
		Version:       1,
		KDFParams:     params,
		EncryptedData: encryptedData,
	}

	// Compute HMAC for integrity verification
	// HMAC over: version + KDF params + encrypted data
	h := hmac.New(sha256.New, verificationKey)
	h.Write([]byte{vault.Version})
	h.Write(vault.KDFParams.Serialize())
	h.Write(vault.EncryptedData)
	copy(vault.HMAC[:], h.Sum(nil))

	return vault, nil
}

// DecryptVault decrypts a vault file and verifies integrity
func DecryptVault(vault *VaultFile, encryptionKey []byte, verificationKey []byte) ([]byte, error) {
	// Verify HMAC
	h := hmac.New(sha256.New, verificationKey)
	h.Write([]byte{vault.Version})
	h.Write(vault.KDFParams.Serialize())
	h.Write(vault.EncryptedData)
	expectedHMAC := h.Sum(nil)

	if !hmac.Equal(expectedHMAC, vault.HMAC[:]) {
		return nil, fmt.Errorf("vault integrity check failed: HMAC mismatch")
	}

	// Extract nonce and ciphertext
	if len(vault.EncryptedData) < 12 {
		return nil, fmt.Errorf("invalid encrypted data: too short")
	}

	nonce := vault.EncryptedData[:12]
	ciphertext := vault.EncryptedData[12:]

	// Create AES-GCM cipher
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// VaultFileSerialize encodes a vault file to bytes for storage
func (v *VaultFile) Serialize() []byte {
	kdfData := v.KDFParams.Serialize()

	// Layout:
	// - Version (1 byte)
	// - KDF params length (1 byte, should be 26)
	// - KDF params (26 bytes)
	// - HMAC (32 bytes)
	// - Encrypted data length (4 bytes, big-endian)
	// - Encrypted data (variable)

	totalLen := 1 + 1 + len(kdfData) + 32 + 4 + len(v.EncryptedData)
	data := make([]byte, totalLen)

	idx := 0

	// Write version
	data[idx] = v.Version
	idx++

	// Write KDF params length
	data[idx] = byte(len(kdfData))
	idx++

	// Write KDF params
	copy(data[idx:], kdfData)
	idx += len(kdfData)

	// Write HMAC
	copy(data[idx:], v.HMAC[:])
	idx += 32

	// Write encrypted data length
	binary.BigEndian.PutUint32(data[idx:idx+4], uint32(len(v.EncryptedData)))
	idx += 4

	// Write encrypted data
	copy(data[idx:], v.EncryptedData)

	return data
}

// VaultFileDeserialize decodes a vault file from bytes
func VaultFileDeserialize(data []byte) (*VaultFile, error) {
	if len(data) < 1+1+26+32+4 {
		return nil, io.ErrUnexpectedEOF
	}

	idx := 0

	version := data[idx]
	idx++

	if version != 1 {
		return nil, fmt.Errorf("unsupported vault version: %d", version)
	}

	kdfLen := int(data[idx])
	idx++

	if len(data) < idx+kdfLen+32+4 {
		return nil, io.ErrUnexpectedEOF
	}

	kdfData := data[idx : idx+kdfLen]
	idx += kdfLen

	kdfParams, err := KDFParamsDeserialize(kdfData)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize KDF params: %w", err)
	}

	var hmac [32]byte
	copy(hmac[:], data[idx:idx+32])
	idx += 32

	encryptedLen := binary.BigEndian.Uint32(data[idx : idx+4])
	idx += 4

	if len(data) < idx+int(encryptedLen) {
		return nil, io.ErrUnexpectedEOF
	}

	encryptedData := append([]byte(nil), data[idx:idx+int(encryptedLen)]...)

	return &VaultFile{
		Version:       version,
		KDFParams:     kdfParams,
		HMAC:          hmac,
		EncryptedData: encryptedData,
	}, nil
}
