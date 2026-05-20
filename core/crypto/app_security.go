package crypto

import (
	"crypto/sha256"
	"fmt"

	"github.com/cloudflare/circl/kem/kyber/kyber768"
)

// AppSecurityFormatVersion is the current on-disk profile version.
const AppSecurityFormatVersion uint8 = 1

// AppSecurityProfile stores the app-level master password verifier.
type AppSecurityProfile struct {
	FormatVersion         uint8             `json:"format_version"`
	PrivateKeyFingerprint []byte            `json:"private_key_fingerprint"`
	KDFParams             KDFParams         `json:"kdf_params"`
	Verifier              []byte            `json:"verifier"`
}

// PrivateKeyFingerprint returns a stable fingerprint for the current private key.
func PrivateKeyFingerprint(privateKey *kyber768.PrivateKey) ([]byte, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("private key is required")
	}

	privateKeyBytes, err := privateKey.MarshalBinary()
	if err != nil {
		return nil, err
	}

	fingerprint := sha256.Sum256(privateKeyBytes)
	return fingerprint[:], nil
}

// CreateAppSecurityProfile builds a verifier profile and the derived session keys.
func CreateAppSecurityProfile(masterPassword string, privateKey *kyber768.PrivateKey) (*AppSecurityProfile, []byte, []byte, error) {
	if masterPassword == "" {
		return nil, nil, nil, fmt.Errorf("master password cannot be empty")
	}

	fingerprint, err := PrivateKeyFingerprint(privateKey)
	if err != nil {
		return nil, nil, nil, err
	}

	params := DefaultKDFParams()
	params.Salt, err = GenerateSalt()
	if err != nil {
		return nil, nil, nil, err
	}

	encryptionKey, verificationKey, err := DeriveKeys(masterPassword, params)
	if err != nil {
		return nil, nil, nil, err
	}

	profile := &AppSecurityProfile{
		FormatVersion:         AppSecurityFormatVersion,
		PrivateKeyFingerprint: append([]byte(nil), fingerprint...),
		KDFParams:             params,
		Verifier:              computeAppVerifier(verificationKey, fingerprint),
	}

	return profile, encryptionKey, verificationKey, nil
}

// VerifyAppSecurityProfile checks whether the provided password matches the stored verifier.
func VerifyAppSecurityProfile(profile *AppSecurityProfile, masterPassword string, privateKey *kyber768.PrivateKey) ([]byte, []byte, bool, bool, error) {
	if profile == nil {
		return nil, nil, false, false, fmt.Errorf("security profile is required")
	}

	if profile.FormatVersion > AppSecurityFormatVersion {
		return nil, nil, false, false, fmt.Errorf("unsupported security profile version: %d (this build supports up to version %d)", profile.FormatVersion, AppSecurityFormatVersion)
	}

	fingerprint, err := PrivateKeyFingerprint(privateKey)
	if err != nil {
		return nil, nil, false, false, err
	}

	if !sha256Equal(profile.PrivateKeyFingerprint, fingerprint) {
		return nil, nil, false, false, nil
	}

	encryptionKey, verificationKey, err := DeriveKeys(masterPassword, profile.KDFParams)
	if err != nil {
		return nil, nil, false, true, err
	}

	verified := sha256Equal(profile.Verifier, computeAppVerifier(verificationKey, fingerprint))
	return encryptionKey, verificationKey, verified, true, nil
}

func computeAppVerifier(verificationKey, fingerprint []byte) []byte {
	h := sha256.New()
	h.Write([]byte("passquantum-app-verifier-v1"))
	h.Write(fingerprint)
	h.Write(verificationKey)
	return h.Sum(nil)
}

func sha256Equal(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}

	var diff byte
	for index := range left {
		diff |= left[index] ^ right[index]
	}

	return diff == 0
}
