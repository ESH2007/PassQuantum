package main

import (
	"log"
	"sync"

	"fyne.io/fyne/v2/app"
	"github.com/cloudflare/circl/kem/kyber/kyber768"

	"passquantum/core/biometric"
	"passquantum/core/crypto"
)

const (
	pubKeyPath  = "public.key"
	privKeyPath = "private.key"
)

type AppState struct {
	publicKey              *kyber768.PublicKey
	privateKey             *kyber768.PrivateKey
	masterPassword         string
	sessionEncryptionKey   []byte
	sessionVerificationKey []byte
	encryptionKey          []byte
	verificationKey        []byte
	kdfParams              crypto.KDFParams
	securityProfile        *crypto.AppSecurityProfile
	mu                     sync.Mutex
	isUnlocked             bool
	currentVault           string
	startupWarning         string

	// Biometric runtime state — populated from the security profile on each unlock.
	biometricEnabled   bool
	biometricTemplate  []float32 // cleared on lock; reloaded from profile on next unlock
	biometricThreshold float32
	biometricStopCheck func() // cancel function for the continuous check goroutine

	// biometricRuntime is loaded once when biometric is first used and kept alive
	// for the session to avoid repeated model-load overhead.
	biometricRuntime biometric.RuntimeHandle
}

func main() {
	myApp := app.New()
	w := myApp.NewWindow("PassQuantum - Post-Quantum Safe Password Manager")
	w.SetTitle("PassQuantum - Post-Quantum Safe Password Manager")
	// w.Resize(fyne.NewSize(500, 350))

	// Initialize crypto
	appState := initializeApp()

	// Show master password prompt on startup
	PromptMasterPassword(w, myApp, appState)

	w.ShowAndRun()
}

func initializeApp() *AppState {
	appState := &AppState{}

	// Try to load existing keypair
	pubKey, privKey, err := crypto.LoadKeypair(pubKeyPath, privKeyPath)
	if err != nil {
		// Generate new keypair if not found
		pubKey, privKey, err = crypto.GenerateKeypair()
		if err != nil {
			log.Fatal("Failed to generate keypair:", err)
		}

		// Save the keypair
		err = crypto.SaveKeypair(pubKey, privKey, pubKeyPath, privKeyPath)
		if err != nil {
			log.Fatal("Failed to save keypair:", err)
		}
	}

	appState.publicKey = pubKey
	appState.privateKey = privKey

	return appState
}
