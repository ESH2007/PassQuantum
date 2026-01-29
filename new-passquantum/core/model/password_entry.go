package model

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

// PasswordEntry represents an encrypted password entry stored in the vault
// Each entry is encrypted with a unique nonce using AES-256-GCM
// The entry also contains the Kyber768 encapsulated secret for hybrid encryption
type PasswordEntry struct {
	ID              uint64 // Unique identifier (4 bytes + reserved for future use)
	KyberCiphertext []byte // Kyber768 encapsulated secret (~1088 bytes)
	Nonce           []byte // AES-GCM nonce (12 bytes)
	Ciphertext      []byte // AES-256-GCM encrypted password (variable)
}

// NewPasswordEntry creates a new password entry with a unique ID
func NewPasswordEntry() *PasswordEntry {
	idBytes := make([]byte, 8)
	rand.Read(idBytes)
	id := binary.BigEndian.Uint64(idBytes)

	return &PasswordEntry{
		ID: id,
	}
}

// Serialize encodes the password entry to binary format for storage in vault
// Format:
// - ID (8 bytes, big-endian)
// - KyberCiphertext length (2 bytes, big-endian)
// - KyberCiphertext (variable)
// - Nonce (12 bytes, fixed)
// - Ciphertext length (2 bytes, big-endian)
// - Ciphertext (variable)
func (pe *PasswordEntry) Serialize() []byte {
	// Calculate total size
	size := 8 + 2 + len(pe.KyberCiphertext) + 12 + 2 + len(pe.Ciphertext)
	data := make([]byte, size)

	idx := 0

	// Write ID
	binary.BigEndian.PutUint64(data[idx:idx+8], pe.ID)
	idx += 8

	// Write KyberCiphertext length and data
	binary.BigEndian.PutUint16(data[idx:idx+2], uint16(len(pe.KyberCiphertext)))
	idx += 2
	copy(data[idx:], pe.KyberCiphertext)
	idx += len(pe.KyberCiphertext)

	// Write Nonce (always 12 bytes)
	copy(data[idx:idx+12], pe.Nonce)
	idx += 12

	// Write Ciphertext length and data
	binary.BigEndian.PutUint16(data[idx:idx+2], uint16(len(pe.Ciphertext)))
	idx += 2
	copy(data[idx:], pe.Ciphertext)

	return data
}

// Deserialize decodes a binary-encoded password entry
func Deserialize(data []byte) (*PasswordEntry, error) {
	if len(data) < 8+2+12+2 {
		return nil, fmt.Errorf("invalid password entry: too short")
	}

	idx := 0

	// Read ID
	id := binary.BigEndian.Uint64(data[idx : idx+8])
	idx += 8

	// Read KyberCiphertext length and data
	kyberLen := int(binary.BigEndian.Uint16(data[idx : idx+2]))
	idx += 2
	if len(data) < idx+kyberLen+12+2 {
		return nil, fmt.Errorf("invalid password entry: truncated kyber ciphertext")
	}

	kyberCiphertext := append([]byte(nil), data[idx:idx+kyberLen]...)
	idx += kyberLen

	// Read Nonce (12 bytes)
	nonce := append([]byte(nil), data[idx:idx+12]...)
	idx += 12

	// Read Ciphertext length and data
	ciphertextLen := int(binary.BigEndian.Uint16(data[idx : idx+2]))
	idx += 2
	if len(data) < idx+ciphertextLen {
		return nil, fmt.Errorf("invalid password entry: truncated ciphertext")
	}

	ciphertext := append([]byte(nil), data[idx:idx+ciphertextLen]...)

	return &PasswordEntry{
		ID:              id,
		KyberCiphertext: kyberCiphertext,
		Nonce:           nonce,
		Ciphertext:      ciphertext,
	}, nil
}
