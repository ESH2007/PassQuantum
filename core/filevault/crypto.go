package filevault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
)

const ChunkSize = 64 * 1024 // 64 KB

// File format:
//   [12B base_nonce][4B chunk_size LE]
//   repeated: [4B encrypted_chunk_len LE][chunk_ciphertext + 16B GCM tag]

// EncryptFile encrypts src into dst using AES-256-GCM in 64 KB chunks.
// sharedSecret must be exactly 32 bytes.
func EncryptFile(src io.Reader, dst io.Writer, sharedSecret []byte) error {
	return EncryptFileWithProgress(src, dst, sharedSecret, 0, nil)
}

// DecryptFile decrypts src into dst, reversing EncryptFile.
func DecryptFile(src io.Reader, dst io.Writer, sharedSecret []byte) error {
	return DecryptFileWithProgress(src, dst, sharedSecret, 0, nil)
}

// EncryptFileWithProgress encrypts with an optional progress callback.
// onProgress receives the cumulative number of plaintext bytes processed.
func EncryptFileWithProgress(src io.Reader, dst io.Writer, sharedSecret []byte, totalSize int64, onProgress func(int64)) error {
	if len(sharedSecret) < 32 {
		return fmt.Errorf("filevault: shared secret must be at least 32 bytes")
	}

	block, err := aes.NewCipher(sharedSecret[:32])
	if err != nil {
		return fmt.Errorf("filevault: aes init: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("filevault: gcm init: %w", err)
	}

	baseNonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(baseNonce); err != nil {
		return fmt.Errorf("filevault: nonce generation: %w", err)
	}

	// Write header: base_nonce + chunk_size
	if _, err := dst.Write(baseNonce); err != nil {
		return fmt.Errorf("filevault: write nonce: %w", err)
	}
	var csLE [4]byte
	binary.LittleEndian.PutUint32(csLE[:], ChunkSize)
	if _, err := dst.Write(csLE[:]); err != nil {
		return fmt.Errorf("filevault: write chunk size: %w", err)
	}

	buf := make([]byte, ChunkSize)
	var processed int64
	var chunkIdx uint32

	for {
		n, readErr := io.ReadFull(src, buf)
		if n > 0 {
			nonce := deriveChunkNonce(baseNonce, chunkIdx)
			sealed := gcm.Seal(nil, nonce, buf[:n], nil)

			var lenBuf [4]byte
			binary.LittleEndian.PutUint32(lenBuf[:], uint32(len(sealed)))
			if _, err := dst.Write(lenBuf[:]); err != nil {
				return fmt.Errorf("filevault: write chunk len: %w", err)
			}
			if _, err := dst.Write(sealed); err != nil {
				return fmt.Errorf("filevault: write chunk: %w", err)
			}

			processed += int64(n)
			chunkIdx++
			if onProgress != nil {
				onProgress(processed)
			}
		}
		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("filevault: read source: %w", readErr)
		}
	}

	return nil
}

// DecryptFileWithProgress decrypts with an optional progress callback.
// onProgress receives the cumulative number of plaintext bytes written.
func DecryptFileWithProgress(src io.Reader, dst io.Writer, sharedSecret []byte, totalSize int64, onProgress func(int64)) error {
	if len(sharedSecret) < 32 {
		return fmt.Errorf("filevault: shared secret must be at least 32 bytes")
	}

	block, err := aes.NewCipher(sharedSecret[:32])
	if err != nil {
		return fmt.Errorf("filevault: aes init: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("filevault: gcm init: %w", err)
	}

	// Read header
	baseNonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(src, baseNonce); err != nil {
		return fmt.Errorf("filevault: read nonce: %w", err)
	}
	var csLE [4]byte
	if _, err := io.ReadFull(src, csLE[:]); err != nil {
		return fmt.Errorf("filevault: read chunk size: %w", err)
	}

	var processed int64
	var chunkIdx uint32

	for {
		var lenBuf [4]byte
		_, readErr := io.ReadFull(src, lenBuf[:])
		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("filevault: read chunk len: %w", readErr)
		}

		chunkLen := binary.LittleEndian.Uint32(lenBuf[:])
		if chunkLen > ChunkSize+uint32(gcm.Overhead())+1024 {
			return fmt.Errorf("filevault: chunk too large (%d bytes)", chunkLen)
		}

		sealed := make([]byte, chunkLen)
		if _, err := io.ReadFull(src, sealed); err != nil {
			return fmt.Errorf("filevault: read chunk data: %w", err)
		}

		nonce := deriveChunkNonce(baseNonce, chunkIdx)
		plaintext, err := gcm.Open(nil, nonce, sealed, nil)
		if err != nil {
			return fmt.Errorf("filevault: decrypt chunk %d: %w", chunkIdx, err)
		}

		if _, err := dst.Write(plaintext); err != nil {
			return fmt.Errorf("filevault: write plaintext: %w", err)
		}

		processed += int64(len(plaintext))
		chunkIdx++
		if onProgress != nil {
			onProgress(processed)
		}
	}

	return nil
}

// deriveChunkNonce XORs a 32-bit chunk counter into the last 4 bytes of baseNonce.
func deriveChunkNonce(baseNonce []byte, chunkIdx uint32) []byte {
	nonce := make([]byte, len(baseNonce))
	copy(nonce, baseNonce)
	offset := len(nonce) - 4
	existing := binary.LittleEndian.Uint32(nonce[offset:])
	binary.LittleEndian.PutUint32(nonce[offset:], existing^chunkIdx)
	return nonce
}
