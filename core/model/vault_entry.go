package model

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strings"
)

type EntryType uint8

const (
	EntryTypeUnknown EntryType = iota
	EntryTypePassword
	EntryTypeNote
	EntryTypeCard
)

// VaultEntry represents an encrypted entry stored in the vault.
// Each entry is encrypted with a unique nonce using AES-256-GCM and
// contains Kyber768 encapsulated key material for hybrid encryption.
type VaultEntry struct {
	ID              uint64 // Unique identifier (4 bytes + reserved for future use)
	Type            EntryType
	CardSubtype     string
	Service         string // Service/website name (e.g., "Gmail", "GitHub")
	Username        string // Username or email associated with the password
	KyberCiphertext []byte // Kyber768 encapsulated secret (~1088 bytes)
	Nonce           []byte // AES-GCM nonce (12 bytes)
	Ciphertext      []byte // AES-256-GCM encrypted entry payload
}

// PasswordEntry remains as an alias for backward compatibility.
type PasswordEntry = VaultEntry

// NewVaultEntry creates a new vault entry with a unique ID.
func NewVaultEntry() *VaultEntry {
	idBytes := make([]byte, 8)
	rand.Read(idBytes)
	id := binary.BigEndian.Uint64(idBytes)

	return &VaultEntry{
		ID:   id,
		Type: EntryTypePassword,
	}
}

// NewPasswordEntry creates a new vault entry with a unique ID.
// Deprecated: prefer NewVaultEntry.
func NewPasswordEntry() *PasswordEntry {
	return NewVaultEntry()
}

// SerializeV2 encodes an entry with explicit type metadata for the migrated format.
func (pe *VaultEntry) SerializeV2() []byte {
	typeByte := byte(pe.Type)
	if typeByte == byte(EntryTypeUnknown) {
		typeByte = byte(inferEntryType(pe.Service))
	}

	if len(pe.Nonce) > 255 {
		return nil
	}
	if len(pe.CardSubtype) > 255 {
		return nil
	}

	subtypeBytes := []byte(pe.CardSubtype)
	serviceBytes := []byte(pe.Service)
	usernameBytes := []byte(pe.Username)

	size := 1 + 8 + 1 + len(subtypeBytes) + 2 + len(serviceBytes) + 2 + len(usernameBytes) + 2 + len(pe.KyberCiphertext) + 1 + len(pe.Nonce) + 4 + len(pe.Ciphertext)
	data := make([]byte, size)

	idx := 0
	data[idx] = typeByte
	idx++

	binary.BigEndian.PutUint64(data[idx:idx+8], pe.ID)
	idx += 8

	data[idx] = uint8(len(subtypeBytes))
	idx++
	copy(data[idx:idx+len(subtypeBytes)], subtypeBytes)
	idx += len(subtypeBytes)

	binary.BigEndian.PutUint16(data[idx:idx+2], uint16(len(serviceBytes)))
	idx += 2
	copy(data[idx:idx+len(serviceBytes)], serviceBytes)
	idx += len(serviceBytes)

	binary.BigEndian.PutUint16(data[idx:idx+2], uint16(len(usernameBytes)))
	idx += 2
	copy(data[idx:idx+len(usernameBytes)], usernameBytes)
	idx += len(usernameBytes)

	binary.BigEndian.PutUint16(data[idx:idx+2], uint16(len(pe.KyberCiphertext)))
	idx += 2
	copy(data[idx:idx+len(pe.KyberCiphertext)], pe.KyberCiphertext)
	idx += len(pe.KyberCiphertext)

	data[idx] = uint8(len(pe.Nonce))
	idx++
	copy(data[idx:idx+len(pe.Nonce)], pe.Nonce)
	idx += len(pe.Nonce)

	binary.BigEndian.PutUint32(data[idx:idx+4], uint32(len(pe.Ciphertext)))
	idx += 4
	copy(data[idx:idx+len(pe.Ciphertext)], pe.Ciphertext)

	return data
}

func DeserializeV2(data []byte) (*VaultEntry, error) {
	if len(data) < 1+8+1+2+2+2+1+4 {
		return nil, fmt.Errorf("invalid typed entry: too short")
	}

	idx := 0
	entryType := EntryType(data[idx])
	idx++

	id := binary.BigEndian.Uint64(data[idx : idx+8])
	idx += 8

	subtypeLen := int(data[idx])
	idx++
	if len(data) < idx+subtypeLen+2 {
		return nil, fmt.Errorf("invalid typed entry: truncated subtype")
	}
	cardSubtype := string(data[idx : idx+subtypeLen])
	idx += subtypeLen

	serviceLen := int(binary.BigEndian.Uint16(data[idx : idx+2]))
	idx += 2
	if len(data) < idx+serviceLen+2 {
		return nil, fmt.Errorf("invalid typed entry: truncated service")
	}
	service := string(data[idx : idx+serviceLen])
	idx += serviceLen

	usernameLen := int(binary.BigEndian.Uint16(data[idx : idx+2]))
	idx += 2
	if len(data) < idx+usernameLen+2 {
		return nil, fmt.Errorf("invalid typed entry: truncated username")
	}
	username := string(data[idx : idx+usernameLen])
	idx += usernameLen

	kyberLen := int(binary.BigEndian.Uint16(data[idx : idx+2]))
	idx += 2
	if len(data) < idx+kyberLen+1 {
		return nil, fmt.Errorf("invalid typed entry: truncated kyber ciphertext")
	}
	kyberCiphertext := append([]byte(nil), data[idx:idx+kyberLen]...)
	idx += kyberLen

	nonceLen := int(data[idx])
	idx++
	if len(data) < idx+nonceLen+4 {
		return nil, fmt.Errorf("invalid typed entry: truncated nonce")
	}
	nonce := append([]byte(nil), data[idx:idx+nonceLen]...)
	idx += nonceLen

	ciphertextLen := int(binary.BigEndian.Uint32(data[idx : idx+4]))
	idx += 4
	if len(data) < idx+ciphertextLen {
		return nil, fmt.Errorf("invalid typed entry: truncated ciphertext")
	}
	ciphertext := append([]byte(nil), data[idx:idx+ciphertextLen]...)

	if entryType == EntryTypeUnknown {
		entryType = inferEntryType(service)
	}

	return &VaultEntry{
		ID:              id,
		Type:            entryType,
		CardSubtype:     cardSubtype,
		Service:         service,
		Username:        username,
		KyberCiphertext: kyberCiphertext,
		Nonce:           nonce,
		Ciphertext:      ciphertext,
	}, nil
}

// Serialize encodes the entry to the legacy binary format for backward compatibility.
func (pe *VaultEntry) Serialize() []byte {
	size := 8 + 2 + len(pe.Service) + 2 + len(pe.Username) + 2 + len(pe.KyberCiphertext) + 12 + 2 + len(pe.Ciphertext)
	data := make([]byte, size)

	idx := 0
	binary.BigEndian.PutUint64(data[idx:idx+8], pe.ID)
	idx += 8

	binary.BigEndian.PutUint16(data[idx:idx+2], uint16(len(pe.Service)))
	idx += 2
	copy(data[idx:], pe.Service)
	idx += len(pe.Service)

	binary.BigEndian.PutUint16(data[idx:idx+2], uint16(len(pe.Username)))
	idx += 2
	copy(data[idx:], pe.Username)
	idx += len(pe.Username)

	binary.BigEndian.PutUint16(data[idx:idx+2], uint16(len(pe.KyberCiphertext)))
	idx += 2
	copy(data[idx:], pe.KyberCiphertext)
	idx += len(pe.KyberCiphertext)

	copy(data[idx:idx+12], pe.Nonce)
	idx += 12

	binary.BigEndian.PutUint16(data[idx:idx+2], uint16(len(pe.Ciphertext)))
	idx += 2
	copy(data[idx:], pe.Ciphertext)

	return data
}

func Deserialize(data []byte) (*VaultEntry, error) {
	if len(data) < 8+2+2+2+12+2 {
		return nil, fmt.Errorf("invalid vault entry: too short")
	}

	idx := 0
	id := binary.BigEndian.Uint64(data[idx : idx+8])
	idx += 8

	serviceLen := int(binary.BigEndian.Uint16(data[idx : idx+2]))
	idx += 2
	if len(data) < idx+serviceLen {
		return nil, fmt.Errorf("invalid vault entry: truncated service")
	}
	service := string(data[idx : idx+serviceLen])
	idx += serviceLen

	usernameLen := int(binary.BigEndian.Uint16(data[idx : idx+2]))
	idx += 2
	if len(data) < idx+usernameLen {
		return nil, fmt.Errorf("invalid vault entry: truncated username")
	}
	username := string(data[idx : idx+usernameLen])
	idx += usernameLen

	kyberLen := int(binary.BigEndian.Uint16(data[idx : idx+2]))
	idx += 2
	if len(data) < idx+kyberLen+12+2 {
		return nil, fmt.Errorf("invalid vault entry: truncated kyber ciphertext")
	}
	kyberCiphertext := append([]byte(nil), data[idx:idx+kyberLen]...)
	idx += kyberLen

	nonce := append([]byte(nil), data[idx:idx+12]...)
	idx += 12

	ciphertextLen := int(binary.BigEndian.Uint16(data[idx : idx+2]))
	idx += 2
	if len(data) < idx+ciphertextLen {
		return nil, fmt.Errorf("invalid vault entry: truncated ciphertext")
	}
	ciphertext := append([]byte(nil), data[idx:idx+ciphertextLen]...)

	return &VaultEntry{
		ID:              id,
		Type:            inferEntryType(service),
		CardSubtype:     inferCardSubtype(service, username),
		Service:         service,
		Username:        username,
		KyberCiphertext: kyberCiphertext,
		Nonce:           nonce,
		Ciphertext:      ciphertext,
	}, nil
}

func inferEntryType(service string) EntryType {
	s := strings.ToUpper(service)
	if strings.HasPrefix(s, "NOTE:") {
		return EntryTypeNote
	}
	if strings.HasPrefix(s, "CARD:") {
		return EntryTypeCard
	}
	return EntryTypePassword
}

func inferCardSubtype(service string, username string) string {
	if inferEntryType(service) != EntryTypeCard {
		return ""
	}
	if strings.EqualFold(username, "credit") || strings.EqualFold(username, "debit") {
		return strings.ToLower(username)
	}
	return ""
}
