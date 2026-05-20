package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"io"

	"golang.org/x/crypto/argon2"
)

// KDFParams contains the Argon2id parameters
type KDFParams struct {
	Salt        []byte // 16 bytes
	Memory      uint32 // 64 MB for security
	Iterations  uint32 // 1 iteration is fast
	Parallelism uint8  // 4 cores
}

// DefaultKDFParams returns secure defaults for password derivation
func DefaultKDFParams() KDFParams {
	return KDFParams{
		Salt:        nil,       // Will be generated
		Memory:      64 * 1024, // 64 MB
		Iterations:  1,
		Parallelism: 4,
	}
}

// GenerateSalt creates a random 16-byte salt
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// DeriveKeys takes a master password and KDF params, returns encryption key and verification key
// Uses domain separation to ensure keys are independent
func DeriveKeys(masterPassword string, params KDFParams) (encryptionKey []byte, verificationKey []byte, err error) {
	if len(params.Salt) == 0 {
		params.Salt, err = GenerateSalt()
		if err != nil {
			return nil, nil, err
		}
	}

	// Derive a long master key using Argon2id
	// Output: 64 bytes (32 for encryption key + 32 for verification key)
	masterKey := argon2.IDKey(
		[]byte(masterPassword),
		params.Salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		64,
	)

	// Domain separation: encrypt different key material for different purposes
	// Encryption key: first 32 bytes derived with domain separator
	encryptionKey = deriveKeyWithDomain(masterKey, "encryption", 32)

	// Verification key: second half with different domain separator
	verificationKey = deriveKeyWithDomain(masterKey, "verification", 32)

	return encryptionKey, verificationKey, nil
}

// deriveKeyWithDomain uses HKDF-like domain separation to derive a key for a specific purpose
func deriveKeyWithDomain(masterKey []byte, domain string, keyLen int) []byte {
	// Simple HKDF-inspired domain separation using SHA-256
	h := sha256.New()
	h.Write([]byte(domain))
	h.Write(masterKey)
	hash := h.Sum(nil)

	// If we need exactly 32 bytes, return the hash
	if keyLen == 32 {
		return hash
	}

	// For other sizes, expand using SHA-256 in counter mode
	key := make([]byte, 0, keyLen)
	counter := uint32(0)

	for len(key) < keyLen {
		h := sha256.New()
		binary.Write(h, binary.BigEndian, counter)
		h.Write(hash)
		key = append(key, h.Sum(nil)...)
		counter++
	}

	return key[:keyLen]
}

// WipeBytes securely overwrites sensitive data
func WipeBytes(data []byte) {
	if len(data) == 0 {
		return
	}
	// Fill with zeros
	copy(data, make([]byte, len(data)))
}

// KDFParamsSerialize encodes KDF parameters to bytes for storage
func (p KDFParams) Serialize() []byte {
	data := make([]byte, 1+16+4+4+1)
	data[0] = 1 // version
	copy(data[1:17], p.Salt)
	binary.BigEndian.PutUint32(data[17:21], p.Memory)
	binary.BigEndian.PutUint32(data[21:25], p.Iterations)
	data[25] = p.Parallelism
	return data
}

// KDFParamsDeserialize decodes KDF parameters from bytes
func KDFParamsDeserialize(data []byte) (KDFParams, error) {
	if len(data) < 26 {
		return KDFParams{}, io.ErrUnexpectedEOF
	}

	return KDFParams{
		Salt:        append([]byte(nil), data[1:17]...),
		Memory:      binary.BigEndian.Uint32(data[17:21]),
		Iterations:  binary.BigEndian.Uint32(data[21:25]),
		Parallelism: data[25],
	}, nil
}
