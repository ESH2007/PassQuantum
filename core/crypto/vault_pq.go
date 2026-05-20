package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	cryptoRand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/cloudflare/circl/kem/kyber/kyber768"
	dilithiumMode3 "github.com/cloudflare/circl/sign/dilithium/mode3"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/hkdf"
)

// pqVaultMagic is the 4-byte header magic for the PQ vault format.
var pqVaultMagic = [4]byte{'P', 'Q', 'V', 'T'}

const pqVaultVersion = 0x01

// Argon2id parameters for PQ vault key derivation.
const (
	pqArgonSaltSize  = 32         // 32-byte salt (longer than legacy 16-byte)
	pqArgonIter      = 2          // time cost
	pqArgonMemKB     = 64 * 1024  // 64 MB memory cost
	pqArgonThreads   = 4          // parallelism
	pqMasterKeySize  = 32         // master_key output length in bytes
)

// Sentinel errors for PQ vault operations.
var (
	ErrPQInvalidHeader    = errors.New("pq_vault: invalid or unrecognized header")
	ErrPQSignatureInvalid = errors.New("pq_vault: Dilithium3 signature verification failed")
	ErrPQAuthFailed       = errors.New("pq_vault: AES-256-GCM authentication failed (wrong password or corrupted data)")
)

// IsPQVaultFormat returns true when data begins with the "PQVT" magic marker.
func IsPQVaultFormat(data []byte) bool {
	return len(data) >= 4 &&
		data[0] == 'P' && data[1] == 'Q' && data[2] == 'V' && data[3] == 'T'
}

// PQVaultEncrypt encrypts plaintext using the full post-quantum pipeline:
//
//	Argon2id(password, fresh 32-byte salt) → 32-byte master_key
//	HKDF(master_key, "passquantum_kyber_seed_v1")      → Kyber-768 keypair
//	HKDF(master_key, "passquantum_dilithium_seed_v1")  → Dilithium3 keypair
//	Kyber-768 KEM encapsulation (ephemeral)            → shared_secret + kyber_ct
//	HKDF(shared_secret, "passquantum_aes_key_v1")      → 32-byte AES key
//	HKDF(shared_secret, "passquantum_nonce_v1")        → 12-byte GCM nonce
//	AES-256-GCM                                        → encrypted payload
//	Dilithium3.Sign(header ‖ payload)                  → signature
//
// The returned bytes are the complete vault file ready to write to disk.
func PQVaultEncrypt(plaintext []byte, password string) ([]byte, error) {
	// ── Step 1: Fresh 32-byte Argon2id salt per save. ─────────────────────────
	salt := make([]byte, pqArgonSaltSize)
	if _, err := cryptoRand.Read(salt); err != nil {
		return nil, fmt.Errorf("pq_vault: salt generation failed: %w", err)
	}

	// ── Step 2: Argon2id → 32-byte master_key. ────────────────────────────────
	masterKey := argon2.IDKey(
		[]byte(password), salt,
		pqArgonIter, pqArgonMemKB, pqArgonThreads,
		pqMasterKeySize,
	)
	defer WipeBytes(masterKey)

	// ── Step 3: Derive Kyber-768 seed (64 bytes) from master_key. ─────────────
	// Domain string "passquantum_kyber_seed_v1" ensures key separation from Dilithium.
	kyberSeed := make([]byte, kyber768.KeySeedSize)
	if err := hkdfExpand(masterKey, []byte("passquantum_kyber_seed_v1"), kyberSeed); err != nil {
		return nil, fmt.Errorf("pq_vault: Kyber seed derivation failed: %w", err)
	}
	defer WipeBytes(kyberSeed)

	// ── Step 4: Derive Dilithium3 seed (32 bytes) from master_key. ────────────
	dilSeed := make([]byte, dilithiumMode3.SeedSize)
	if err := hkdfExpand(masterKey, []byte("passquantum_dilithium_seed_v1"), dilSeed); err != nil {
		return nil, fmt.Errorf("pq_vault: Dilithium3 seed derivation failed: %w", err)
	}
	defer WipeBytes(dilSeed)

	// ── Step 5: Derive keypairs deterministically from seeds. ─────────────────
	kyberPK, _ := kyber768.NewKeyFromSeed(kyberSeed)

	var dilSeedArr [dilithiumMode3.SeedSize]byte
	copy(dilSeedArr[:], dilSeed)
	dilPK, dilSK := dilithiumMode3.NewKeyFromSeed(&dilSeedArr)
	_ = dilPK // used only on the verify path; included here for symmetry

	// ── Step 6: Ephemeral Kyber KEM encapsulation → shared_secret + kyber_ct. ─
	// A fresh random encapsulation seed is used on every save; this keeps the
	// shared_secret (and therefore the AES key) unique per ciphertext even if
	// the password never changes.
	kyberCT := make([]byte, kyber768.CiphertextSize)
	sharedSecret := make([]byte, kyber768.SharedKeySize)
	kyberPK.EncapsulateTo(kyberCT, sharedSecret, nil) // nil → random seed
	defer WipeBytes(sharedSecret)

	// ── Step 7: HKDF(shared_secret) → AES-256-GCM key (32 bytes). ────────────
	aesKey := make([]byte, 32)
	if err := hkdfExpand(sharedSecret, []byte("passquantum_aes_key_v1"), aesKey); err != nil {
		return nil, fmt.Errorf("pq_vault: AES key derivation failed: %w", err)
	}
	defer WipeBytes(aesKey)

	// ── Step 8: HKDF(shared_secret) → 12-byte nonce (separate context). ───────
	nonce := make([]byte, 12)
	if err := hkdfExpand(sharedSecret, []byte("passquantum_nonce_v1"), nonce); err != nil {
		return nil, fmt.Errorf("pq_vault: nonce derivation failed: %w", err)
	}

	// ── Step 9: AES-256-GCM encrypt the plaintext. ────────────────────────────
	// Seal appends the 16-byte GCM authentication tag to the ciphertext.
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("pq_vault: AES init failed: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("pq_vault: GCM init failed: %w", err)
	}
	payload := gcm.Seal(nil, nonce, plaintext, nil)

	// ── Step 10: Build the plaintext header. ──────────────────────────────────
	//   4 bytes  – magic "PQVT"
	//   1 byte   – format version (0x01)
	//   32 bytes – Argon2id salt
	//   4 bytes  – Argon2id iterations  (uint32 LE)
	//   4 bytes  – Argon2id memory KB   (uint32 LE)
	//   4 bytes  – Kyber ciphertext length (uint32 LE)
	//   N bytes  – Kyber ciphertext (1088 bytes for Kyber-768)
	//   12 bytes – AES-GCM nonce
	header := make([]byte, 0, 4+1+32+4+4+4+len(kyberCT)+12)
	header = append(header, pqVaultMagic[:]...)
	header = append(header, pqVaultVersion)
	header = append(header, salt...)
	header = binary.LittleEndian.AppendUint32(header, pqArgonIter)
	header = binary.LittleEndian.AppendUint32(header, pqArgonMemKB)
	header = binary.LittleEndian.AppendUint32(header, uint32(len(kyberCT)))
	header = append(header, kyberCT...)
	header = append(header, nonce...)

	// ── Step 11: Dilithium3 signs (header ‖ payload). ─────────────────────────
	// The signature covers both the plaintext header AND the encrypted payload,
	// preventing any tampering with header fields or ciphertext bytes.
	toSign := make([]byte, 0, len(header)+len(payload))
	toSign = append(toSign, header...)
	toSign = append(toSign, payload...)

	var sigBuf [dilithiumMode3.SignatureSize]byte
	dilithiumMode3.SignTo(dilSK, toSign, sigBuf[:])

	// ── Step 12: Assemble final vault file: header ‖ payload ‖ sig_len ‖ sig. ─
	sigLen := uint32(dilithiumMode3.SignatureSize)
	out := make([]byte, 0, len(header)+len(payload)+4+dilithiumMode3.SignatureSize)
	out = append(out, header...)
	out = append(out, payload...)
	out = binary.LittleEndian.AppendUint32(out, sigLen)
	out = append(out, sigBuf[:]...)

	return out, nil
}

// PQVaultDecrypt decrypts a PQ vault, enforcing strict verify-before-decrypt order:
//
//  1. Parse and validate header magic + version.
//  2. Verify Dilithium3 signature over (header ‖ payload) — fails fast on any
//     tampering or wrong-password scenario at the signature layer.
//  3. Reconstruct master_key via Argon2id (password + stored salt).
//  4. Derive Kyber-768 private key from master_key via HKDF.
//  5. Kyber decapsulation → shared_secret.
//  6. Derive AES-256-GCM key + nonce from shared_secret via HKDF.
//  7. Decrypt and authenticate payload; return ErrPQAuthFailed on tag mismatch.
func PQVaultDecrypt(data []byte, password string) ([]byte, error) {
	// ── Step 1: Validate magic + version + minimum size. ──────────────────────
	const minHeaderSize = 4 + 1 + 32 + 4 + 4 + 4 // magic+ver+salt+iter+mem+ctLen
	if len(data) < minHeaderSize {
		return nil, ErrPQInvalidHeader
	}
	if data[0] != 'P' || data[1] != 'Q' || data[2] != 'V' || data[3] != 'T' {
		return nil, ErrPQInvalidHeader
	}
	if data[4] != pqVaultVersion {
		return nil, fmt.Errorf("pq_vault: unsupported version 0x%02x", data[4])
	}

	// Parse header fields.
	idx := 5
	salt := data[idx : idx+32]
	idx += 32
	argonIter := binary.LittleEndian.Uint32(data[idx : idx+4])
	idx += 4
	argonMemKB := binary.LittleEndian.Uint32(data[idx : idx+4])
	idx += 4
	kyberCTLen := binary.LittleEndian.Uint32(data[idx : idx+4])
	idx += 4

	if idx+int(kyberCTLen)+12 > len(data) {
		return nil, ErrPQInvalidHeader
	}
	kyberCT := data[idx : idx+int(kyberCTLen)]
	idx += int(kyberCTLen)
	nonce := data[idx : idx+12]
	idx += 12
	headerEnd := idx // header boundary

	// Signature layout from the tail of the file:
	//   … payload … | sig_len (4 LE) | sig (SignatureSize bytes)
	// Because Dilithium3.SignatureSize is a fixed constant, we can find the
	// signature deterministically without parsing the payload length.
	const sigSize = dilithiumMode3.SignatureSize
	sigFieldStart := len(data) - sigSize - 4
	if sigFieldStart < headerEnd {
		return nil, ErrPQInvalidHeader
	}
	storedSigLen := binary.LittleEndian.Uint32(data[sigFieldStart : sigFieldStart+4])
	if int(storedSigLen) != sigSize {
		return nil, ErrPQInvalidHeader
	}
	sigBytes := data[len(data)-sigSize:]
	payload := data[headerEnd:sigFieldStart]
	header := data[:headerEnd]

	// ── Step 2: Derive Dilithium3 public key and verify signature
	//           BEFORE any decryption attempt. ────────────────────────────────
	masterKey := argon2.IDKey(
		[]byte(password), salt,
		argonIter, argonMemKB, pqArgonThreads,
		pqMasterKeySize,
	)
	defer WipeBytes(masterKey)

	dilSeed := make([]byte, dilithiumMode3.SeedSize)
	if err := hkdfExpand(masterKey, []byte("passquantum_dilithium_seed_v1"), dilSeed); err != nil {
		return nil, fmt.Errorf("pq_vault: Dilithium3 seed derivation failed: %w", err)
	}
	defer WipeBytes(dilSeed)

	var dilSeedArr [dilithiumMode3.SeedSize]byte
	copy(dilSeedArr[:], dilSeed)
	dilPK, _ := dilithiumMode3.NewKeyFromSeed(&dilSeedArr)

	toVerify := make([]byte, 0, len(header)+len(payload))
	toVerify = append(toVerify, header...)
	toVerify = append(toVerify, payload...)

	if !dilithiumMode3.Verify(dilPK, toVerify, sigBytes) {
		// Signature mismatch: either wrong password or tampered file.
		return nil, ErrPQSignatureInvalid
	}

	// ── Steps 3–4: Derive Kyber-768 private key from master_key. ──────────────
	// master_key is already computed above; we now use it for the KEM path.
	kyberSeed := make([]byte, kyber768.KeySeedSize)
	if err := hkdfExpand(masterKey, []byte("passquantum_kyber_seed_v1"), kyberSeed); err != nil {
		return nil, fmt.Errorf("pq_vault: Kyber seed derivation failed: %w", err)
	}
	defer WipeBytes(kyberSeed)

	// ── Step 5: Kyber decapsulation → shared_secret. ─────────────────────────
	_, kyberSK := kyber768.NewKeyFromSeed(kyberSeed)
	sharedSecret := make([]byte, kyber768.SharedKeySize)
	kyberSK.DecapsulateTo(sharedSecret, kyberCT)
	defer WipeBytes(sharedSecret)

	// ── Step 6: Derive AES-256-GCM key + nonce from shared_secret. ────────────
	aesKey := make([]byte, 32)
	if err := hkdfExpand(sharedSecret, []byte("passquantum_aes_key_v1"), aesKey); err != nil {
		return nil, fmt.Errorf("pq_vault: AES key derivation failed: %w", err)
	}
	defer WipeBytes(aesKey)

	// The nonce in the header is exactly the HKDF-derived nonce stored during
	// encryption; use the stored header nonce for AES-GCM to remain byte-exact.
	_ = nonce // referenced below

	// ── Step 7: AES-256-GCM decrypt + authenticate. ───────────────────────────
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("pq_vault: AES init failed: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("pq_vault: GCM init failed: %w", err)
	}
	plaintext, err := gcm.Open(nil, nonce, payload, nil)
	if err != nil {
		return nil, ErrPQAuthFailed
	}

	return plaintext, nil
}

// hkdfExpand derives len(out) bytes from secret using HKDF-SHA256 with the
// given info string. salt is nil (HKDF will use a zero-filled salt of hash
// size, which is appropriate when secret is already high-entropy key material).
func hkdfExpand(secret, info, out []byte) error {
	r := hkdf.New(sha256.New, secret, nil, info)
	if _, err := io.ReadFull(r, out); err != nil {
		return err
	}
	return nil
}
