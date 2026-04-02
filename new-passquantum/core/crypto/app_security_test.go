package crypto

import "testing"

func TestCreateAndVerifyAppSecurityProfile(t *testing.T) {
	_, privateKey, err := GenerateKeypair()
	if err != nil {
		t.Fatalf("GenerateKeypair() error = %v", err)
	}

	profile, encryptionKey, verificationKey, err := CreateAppSecurityProfile("correct horse battery staple", privateKey)
	if err != nil {
		t.Fatalf("CreateAppSecurityProfile() error = %v", err)
	}

	if profile.FormatVersion != AppSecurityFormatVersion {
		t.Fatalf("CreateAppSecurityProfile() format version = %d, want %d", profile.FormatVersion, AppSecurityFormatVersion)
	}

	if len(profile.PrivateKeyFingerprint) == 0 {
		t.Fatal("CreateAppSecurityProfile() fingerprint was empty")
	}

	verifiedEncryptionKey, verifiedVerificationKey, verified, fingerprintMatches, err := VerifyAppSecurityProfile(profile, "correct horse battery staple", privateKey)
	if err != nil {
		t.Fatalf("VerifyAppSecurityProfile() error = %v", err)
	}

	if !verified {
		t.Fatal("VerifyAppSecurityProfile() verified = false, want true")
	}

	if !fingerprintMatches {
		t.Fatal("VerifyAppSecurityProfile() fingerprintMatches = false, want true")
	}

	if string(encryptionKey) != string(verifiedEncryptionKey) {
		t.Fatal("VerifyAppSecurityProfile() encryption key mismatch")
	}

	if string(verificationKey) != string(verifiedVerificationKey) {
		t.Fatal("VerifyAppSecurityProfile() verification key mismatch")
	}

	_, _, verified, fingerprintMatches, err = VerifyAppSecurityProfile(profile, "wrong password", privateKey)
	if err != nil {
		t.Fatalf("VerifyAppSecurityProfile() wrong password error = %v", err)
	}

	if verified {
		t.Fatal("VerifyAppSecurityProfile() with wrong password verified = true, want false")
	}

	if !fingerprintMatches {
		t.Fatal("VerifyAppSecurityProfile() with wrong password fingerprintMatches = false, want true")
	}
}

func TestVerifyAppSecurityProfileFingerprintMismatch(t *testing.T) {
	_, firstPrivateKey, err := GenerateKeypair()
	if err != nil {
		t.Fatalf("GenerateKeypair() first key error = %v", err)
	}

	_, secondPrivateKey, err := GenerateKeypair()
	if err != nil {
		t.Fatalf("GenerateKeypair() second key error = %v", err)
	}

	profile, _, _, err := CreateAppSecurityProfile("top secret", firstPrivateKey)
	if err != nil {
		t.Fatalf("CreateAppSecurityProfile() error = %v", err)
	}

	_, _, verified, fingerprintMatches, err := VerifyAppSecurityProfile(profile, "top secret", secondPrivateKey)
	if err != nil {
		t.Fatalf("VerifyAppSecurityProfile() mismatch error = %v", err)
	}

	if verified {
		t.Fatal("VerifyAppSecurityProfile() verified = true for mismatched key")
	}

	if fingerprintMatches {
		t.Fatal("VerifyAppSecurityProfile() fingerprintMatches = true for mismatched key")
	}
}
