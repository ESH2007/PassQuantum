package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

// NewAES256GCM builds an AES-256-GCM AEAD from key. key must be a valid AES key
// length (16/24/32 bytes); callers using a longer shared secret should pass
// key[:32]. Centralizing this avoids repeating the aes.NewCipher + cipher.NewGCM
// dance across the vault and file-vault pipelines.
func NewAES256GCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// EncryptAES256GCM encrypts plaintext using AES-256-GCM with a given shared secret key
// Returns the nonce and ciphertext
func EncryptAES256GCM(plaintext string, sharedSecret []byte) ([]byte, []byte, error) {
	gcm, err := NewAES256GCM(sharedSecret[:32])
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	return nonce, ciphertext, nil
}

// DecryptAES256GCM decrypts ciphertext using AES-256-GCM with a given shared secret key
// Returns the plaintext
func DecryptAES256GCM(nonce []byte, ciphertext []byte, sharedSecret []byte) (string, error) {
	gcm, err := NewAES256GCM(sharedSecret[:32])
	if err != nil {
		return "", err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
