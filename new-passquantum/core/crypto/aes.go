package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

// EncryptAES256GCM encrypts plaintext using AES-256-GCM with a given shared secret key
// Returns the nonce and ciphertext
func EncryptAES256GCM(plaintext string, sharedSecret []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(sharedSecret[:32])
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
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
	block, err := aes.NewCipher(sharedSecret[:32])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
