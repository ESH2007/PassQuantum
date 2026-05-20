package main

import (
	"fmt"
	"log"
	"os"

	"passquantum/core/crypto"
	"passquantum/core/model"
	"passquantum/core/storage"
)

func main() {
	vaultFile := "test_vault.pqdb"
	os.Remove(vaultFile)

	masterPassword := "testpassword123"

	fmt.Println("============================================================")
	fmt.Println("TEST 1: Create vault and add a password")
	fmt.Println("============================================================")

	// Create vault with no entries
	err := storage.WriteVault([]*model.VaultEntry{}, vaultFile, masterPassword)
	if err != nil {
		log.Fatal("Failed to create vault:", err)
	}

	// Read and verify vault is empty
	entries, err := storage.ReadVault(vaultFile, masterPassword)
	if err != nil {
		log.Fatal("Failed to read vault:", err)
	}
	fmt.Printf("Vault has %d entries (expected 0)\n\n", len(entries))

	fmt.Println("============================================================")
	fmt.Println("TEST 2: Add a password")
	fmt.Println("============================================================")

	// Generate Kyber keypair (used for per-entry encryption, separate from vault KEM)
	pubKey, privKey, err := crypto.GenerateKeypair()
	if err != nil {
		log.Fatal("Failed to generate keypair:", err)
	}

	// Create and encrypt a password entry
	ct, ss, err := crypto.Encapsulate(pubKey)
	if err != nil {
		log.Fatal("Encapsulation failed:", err)
	}

	password := "mySecretPassword123!"
	nonce, ciphertext, err := crypto.EncryptAES256GCM(password, ss)
	if err != nil {
		log.Fatal("Encryption failed:", err)
	}

	entry := model.NewVaultEntry()
	entry.KyberCiphertext = ct
	entry.Nonce = nonce
	entry.Ciphertext = ciphertext

	fmt.Printf("Created entry with encrypted password\n")
	fmt.Printf("  Kyber ciphertext size: %d bytes\n", len(ct))
	fmt.Printf("  Nonce size: %d bytes\n", len(nonce))
	fmt.Printf("  AES ciphertext size: %d bytes\n", len(ciphertext))

	// Save vault with the new entry
	err = storage.WriteVault([]*model.VaultEntry{entry}, vaultFile, masterPassword)
	if err != nil {
		log.Fatal("Failed to save vault:", err)
	}

	fmt.Println("\n============================================================")
	fmt.Println("TEST 3: Read and decrypt password")
	fmt.Println("============================================================")

	// Read vault
	entries, err = storage.ReadVault(vaultFile, masterPassword)
	if err != nil {
		log.Fatal("Failed to read vault:", err)
	}

	fmt.Printf("Vault has %d entries (expected 1)\n", len(entries))

	if len(entries) > 0 {
		// Decrypt the password
		ss, err := crypto.Decapsulate(entries[0].KyberCiphertext, privKey)
		if err != nil {
			log.Fatal("Decapsulation failed:", err)
		}

		plaintext, err := crypto.DecryptAES256GCM(entries[0].Nonce, entries[0].Ciphertext, ss)
		if err != nil {
			log.Fatal("Decryption failed:", err)
		}

		fmt.Printf("Decrypted password: %s\n", plaintext)
		fmt.Printf("Expected password:  %s\n", password)

		if plaintext == password {
			fmt.Println("\n✅ TEST PASSED: Password correctly encrypted and decrypted!")
		} else {
			fmt.Println("\n❌ TEST FAILED: Decrypted password does not match!")
		}
	} else {
		fmt.Println("\n❌ TEST FAILED: No entries found in vault!")
	}

	os.Remove(vaultFile)
}
