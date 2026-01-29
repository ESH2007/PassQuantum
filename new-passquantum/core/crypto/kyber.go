package crypto

import (
	"os"

	"github.com/cloudflare/circl/kem/kyber/kyber768"
)

// GenerateKeypair generates a new Kyber768 keypair
func GenerateKeypair() (*kyber768.PublicKey, *kyber768.PrivateKey, error) {
	return kyber768.GenerateKeyPair(nil)
}

// SaveKeypair saves the Kyber768 keypair to disk
func SaveKeypair(publicKey *kyber768.PublicKey, privateKey *kyber768.PrivateKey, pubPath, privPath string) error {
	pubBytes, err := publicKey.MarshalBinary()
	if err != nil {
		return err
	}

	err = os.WriteFile(pubPath, pubBytes, 0644)
	if err != nil {
		return err
	}

	privBytes, err := privateKey.MarshalBinary()
	if err != nil {
		return err
	}

	err = os.WriteFile(privPath, privBytes, 0600)
	if err != nil {
		return err
	}

	return nil
}

// LoadKeypair loads the Kyber768 keypair from disk
func LoadKeypair(pubPath, privPath string) (*kyber768.PublicKey, *kyber768.PrivateKey, error) {
	pubBytes, err := os.ReadFile(pubPath)
	if err != nil {
		return nil, nil, err
	}

	privBytes, err := os.ReadFile(privPath)
	if err != nil {
		return nil, nil, err
	}

	scheme := kyber768.Scheme()

	publicKey, err := scheme.UnmarshalBinaryPublicKey(pubBytes)
	if err != nil {
		return nil, nil, err
	}

	privateKey, err := scheme.UnmarshalBinaryPrivateKey(privBytes)
	if err != nil {
		return nil, nil, err
	}

	pk := publicKey.(*kyber768.PublicKey)
	sk := privateKey.(*kyber768.PrivateKey)

	return pk, sk, nil
}

// Encapsulate performs Kyber768 encapsulation with a public key
// Returns the ciphertext and shared secret
func Encapsulate(publicKey *kyber768.PublicKey) ([]byte, []byte, error) {
	ct := make([]byte, kyber768.CiphertextSize)
	ss := make([]byte, kyber768.SharedKeySize)

	publicKey.EncapsulateTo(ct, ss, nil)

	return ct, ss, nil
}

// Decapsulate performs Kyber768 decapsulation with a private key
// Returns the shared secret
func Decapsulate(encapsulatedSecret []byte, privateKey *kyber768.PrivateKey) ([]byte, error) {
	ss := make([]byte, kyber768.SharedKeySize)
	privateKey.DecapsulateTo(ss, encapsulatedSecret)

	return ss, nil
}
