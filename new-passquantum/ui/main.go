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
	securestorage "passquantum/internal/storage"
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
	securityProfile        *crypto.AppSecurityProfile
	mu                     sync.Mutex
	isUnlocked             bool
	isTraining             bool
	currentVault           string
	startupWarning         string
	faceGuard              *FaceGuard
	// lockApp is called from any goroutine to lock the app immediately;
	// it clears sensitive state and returns the user to the login screen.
	lockApp func()
}

func main() {
	normalizeLocaleForFyne()
	_ = os.Setenv("OPENCV_LOG_LEVEL", "ERROR")

	if _, err := securestorage.GetVaultDir(); err != nil {
		log.Printf("WARNING: failed to initialize secure vault directory: %v", err)
	} else if err := securestorage.ValidateVaultPermissions(); err != nil {
		log.Printf("WARNING: failed to validate vault permissions: %v", err)
	}

	myApp := app.NewWithID("com.passquantum.app")
	setApplicationIcon(myApp)
	w := myApp.NewWindow("PassQuantum - Post-Quantum Safe Password Manager")
	w.SetTitle("PassQuantum - Post-Quantum Safe Password Manager")
	// w.Resize(fyne.NewSize(500, 350))
	RestoreThemeOnLaunch(myApp, w)

	// Initialize crypto
	appState := initializeApp()

	// Initialize face recognition guard (warn-only on failure — app proceeds without it)
	if guard, err := NewFaceGuard(); err != nil {
		log.Printf("[FaceGuard] WARNING: could not create face guard: %v", err)
	} else {
		appState.faceGuard = guard
		if err := guard.Launch(); err != nil {
			log.Printf("[FaceGuard] WARNING: could not launch face_guard.py: %v", err)
		}
		go guard.Listen()

		// Wire the global lock-on-face-loss callback.
		// OnLost fires whenever Python sends FACE_LOST (face absent for 5 s).
		// We lock only when the app is actually unlocked and not in training
		// (during training the user deliberately moves their face).
		guard.OnLost = func() {
			appState.mu.Lock()
			unlocked := appState.isUnlocked
			training := appState.isTraining
			appState.mu.Unlock()

			if !unlocked || training {
				return
			}

			log.Println("[FaceGuard] FACE_LOST: locking app and killing monitored processes")

			// Kill any user-selected companion apps before locking.
			go killProcessesByName(loadKillApps(myApp.Preferences()))

			// Lock the UI on the Fyne goroutine.
			fyne.Do(func() {
				if appState.lockApp != nil {
					appState.lockApp()
				}
			})
		}

		guard.OnOK = func() {
			log.Println("[FaceGuard] FACE_OK: face recognised")
		}
	}

	// lockApp clears sensitive state and returns the user to the login screen.
	// Assigned here (after guard init) so the closure captures w and myApp.
	appState.lockApp = func() {
		appState.clearSensitiveState()
		PromptMasterPassword(w, myApp, appState)
	}

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
