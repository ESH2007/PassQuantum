package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"

	pqapp "passquantum/app"
	"passquantum/bridge"
	"passquantum/core/crypto"
	"passquantum/core/filevault"
	securestorage "passquantum/internal/storage"
	"passquantum/theme"
	"passquantum/ui/screens"
)

func main() {
	normalizeLocaleForFyne()
	_ = os.Setenv("OPENCV_LOG_LEVEL", "ERROR")

	if _, err := securestorage.GetVaultDir(); err != nil {
		log.Printf("WARNING: failed to initialize secure vault directory: %v", err)
	} else if err := securestorage.ValidateVaultPermissions(); err != nil {
		log.Printf("WARNING: failed to validate vault permissions: %v", err)
	}

	if err := filevault.CleanupOrphans(); err != nil {
		log.Printf("WARNING: failed to cleanup orphan temp files: %v", err)
	}

	// Route face_data.npy into the vault directory so it shares the same
	// location and security policy as the .enc files.
	if vaultDir, err := securestorage.GetVaultDir(); err == nil {
		_ = os.Setenv("PASSQUANTUM_WORK_DIR", vaultDir)
	}

	myApp := fyneapp.NewWithID("com.passquantum.app")
	myApp.Settings().SetTheme(&theme.QuantumTheme{})
	screens.SetApplicationIcon(myApp)
	myApp.Lifecycle().SetOnStarted(func() {
		go maximizeWindow()
	})
	w := myApp.NewWindow("PassQuantum - Post-Quantum Safe Password Manager")
	screens.RestoreThemeOnLaunch(myApp, w)

	// Initialize crypto keypair
	appState := initializeApp()

	// Initialize face recognition guard (warn-only on failure — app proceeds without it)
	if guard, err := bridge.NewFaceGuard(); err != nil {
		log.Printf("[FaceGuard] WARNING: could not create face guard: %v", err)
	} else {
		appState.FaceGuard = guard
		if err := guard.Launch(); err != nil {
			log.Printf("[FaceGuard] WARNING: could not launch face_guard.py: %v", err)
		}
		go guard.Listen()

		// OnLost fires whenever Python sends FACE_LOST (face absent for 5 s).
		// Lock only when the app is actually unlocked and not in training.
		guard.OnLost = func() {
			appState.Mu.Lock()
			unlocked := appState.IsUnlocked
			training := appState.IsTraining
			appState.Mu.Unlock()

			if !unlocked || training {
				return
			}

			log.Println("[FaceGuard] FACE_LOST: locking app and killing monitored processes")

			go bridge.KillProcessesByName(bridge.LoadKillApps(myApp.Preferences()))

			fyne.Do(func() {
				if appState.LockApp != nil {
					appState.LockApp()
				}
			})
		}

		guard.OnOK = func() {
			log.Println("[FaceGuard] FACE_OK: face recognised")
		}
	}

	// LockApp clears sensitive state and returns the user to the login screen.
	// Assigned after guard init so the closure captures w and myApp.
	appState.LockApp = func() {
		appState.ClearSensitiveState()
		screens.PromptMasterPassword(w, myApp, appState)
	}

	screens.PromptMasterPassword(w, myApp, appState)

	w.SetOnClosed(func() {
		if appState.FaceGuard != nil {
			appState.FaceGuard.Shutdown()
		}
	})

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

func initializeApp() *pqapp.AppState {
	appState := &pqapp.AppState{}

	pubKeyPath, err := securestorage.GetSecureFilePath(pqapp.PubKeyPath)
	if err != nil {
		log.Fatal("Failed to resolve public key path:", err)
	}
	privKeyPath, err := securestorage.GetSecureFilePath(pqapp.PrivKeyPath)
	if err != nil {
		log.Fatal("Failed to resolve private key path:", err)
	}

	pubKey, privKey, err := crypto.LoadKeypair(pubKeyPath, privKeyPath)
	if err != nil {
		pubKey, privKey, err = crypto.GenerateKeypair()
		if err != nil {
			log.Fatal("Failed to generate keypair:", err)
		}
		if err = crypto.SaveKeypair(pubKey, privKey, pubKeyPath, privKeyPath); err != nil {
			log.Fatal("Failed to save keypair:", err)
		}
	}

	appState.PublicKey = pubKey
	appState.PrivateKey = privKey

	return appState
}

// maximizeWindow asks the window manager to maximize the PassQuantum window.
// Fyne v2 has no native maximize API, so we use wmctrl (X11/XWayland) with
// xdotool as a fallback.  A short sleep lets the window become visible first.
func maximizeWindow() {
	time.Sleep(300 * time.Millisecond)

	// wmctrl: available on most Linux desktops, works on X11 and XWayland
	if err := exec.Command("wmctrl", "-r", "PassQuantum", "-b", "add,maximized_vert,maximized_horz").Run(); err == nil {
		return
	}

	// xdotool fallback
	out, err := exec.Command("xdotool", "search", "--name", "PassQuantum").Output()
	if err != nil {
		log.Printf("[maximize] no wmctrl or xdotool found; install one to auto-maximize: %v", err)
		return
	}
	for _, wid := range strings.Fields(string(out)) {
		exec.Command("xdotool", "windowmaximize", wid).Run()
	}
}
