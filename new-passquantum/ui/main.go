package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/cloudflare/circl/kem/kyber/kyber768"

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
}

func main() {
	normalizeLocaleForFyne()
	_ = os.Setenv("OPENCV_LOG_LEVEL", "ERROR")

	myApp := app.NewWithID("com.passquantum.app")
	setApplicationIcon(myApp)
	w := myApp.NewWindow("PassQuantum - Post-Quantum Safe Password Manager")
	w.SetTitle("PassQuantum - Post-Quantum Safe Password Manager")
	// w.Resize(fyne.NewSize(500, 350))
	RestoreThemeOnLaunch(myApp, w)

	// Initialize crypto
	appState := initializeApp()

	// Show master password prompt on startup
	PromptMasterPassword(w, myApp, appState)

	w.ShowAndRun()
}

// normalizeLocaleForFyne avoids startup locale parse failures when environments use
// bare C/POSIX locale identifiers.
func normalizeLocaleForFyne() {
	fix := func(key string) {
		v := strings.TrimSpace(os.Getenv(key))
		if v == "" || v == "C" || v == "POSIX" {
			_ = os.Setenv(key, "en_US.UTF-8")
		}
	}

	fix("LC_ALL")
	fix("LANG")
	fix("LANGUAGE")
	fix("LC_MESSAGES")
	fix("LC_CTYPE")
}

func setApplicationIcon(myApp fyne.App) {
	if customPath := myApp.Preferences().StringWithFallback("custom_icon_path", ""); customPath != "" {
		data, err := os.ReadFile(customPath)
		if err == nil && len(data) > 0 {
			myApp.SetIcon(fyne.NewStaticResource(filepath.Base(customPath), data))
			return
		}
		// Stale path (file moved/deleted) — clear it and fall through to bundled default.
		myApp.Preferences().SetString("custom_icon_path", "")
	}

	iconCandidates := []string{
		"Icon.png",
		filepath.Join("..", "Icon.png"),
		filepath.Join("new-passquantum", "Icon.png"),
		filepath.Join("build", "windows", "Icon.png"),
	}

	for _, iconPath := range iconCandidates {
		data, err := os.ReadFile(iconPath)
		if err != nil || len(data) == 0 {
			continue
		}
		myApp.SetIcon(fyne.NewStaticResource("Icon.png", data))
		return
	}
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
