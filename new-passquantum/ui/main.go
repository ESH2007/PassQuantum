package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/cloudflare/circl/kem/kyber/kyber768"

	"passquantum/core/crypto"
	"passquantum/core/model"
	"passquantum/core/storage"
)

const (
	pubKeyPath  = "public.key"
	privKeyPath = "private.key"
	vaultFile   = "vault.pqdb"
)

type AppState struct {
	publicKey       *kyber768.PublicKey
	privateKey      *kyber768.PrivateKey
	masterPassword  string
	encryptionKey   []byte
	verificationKey []byte
	kdfParams       crypto.KDFParams
	mu              sync.Mutex
	isUnlocked      bool
}

func main() {
	myApp := app.New()
	w := myApp.NewWindow("PassQuantum - Post-Quantum Safe Password Manager")
	w.SetTitle("PassQuantum - Post-Quantum Safe Password Manager")
	w.Resize(fyne.NewSize(500, 400))

	// Initialize crypto
	appState := initializeApp()

	// Show master password prompt on startup
	promptMasterPassword(w, myApp, appState)

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

func promptMasterPassword(w fyne.Window, fyneApp fyne.App, appState *AppState) {
	// Check if vault exists
	vaultExists := storage.VaultExists(vaultFile)

	// Create dialog for master password
	passwordInput := widget.NewPasswordEntry()
	passwordInput.PlaceHolder = "Enter master password"

	if vaultExists {
		// Vault exists - need to unlock it
		unlockBtn := widget.NewButton("Unlock Vault", func() {
			password := passwordInput.Text
			if password == "" {
				dialog.ShowError(fmt.Errorf("master password cannot be empty"), w)
				return
			}

			// Try to unlock
			if unlockVault(w, appState, password) {
				// Unlock successful - show main UI
				mainContent := buildUI(w, fyneApp, appState)
				w.SetContent(mainContent)
			}
		})

		// Create dialog
		content := container.NewVBox(
			widget.NewLabel("Vault exists. Enter your master password to unlock:"),
			passwordInput,
			unlockBtn,
		)

		w.SetContent(content)
	} else {
		// No vault exists - create new one
		createBtn := widget.NewButton("Create New Vault", func() {
			password := passwordInput.Text
			if password == "" {
				dialog.ShowError(fmt.Errorf("master password cannot be empty"), w)
				return
			}

			// Create new vault
			if createNewVault(w, appState, password) {
				// Creation successful - show main UI
				mainContent := buildUI(w, fyneApp, appState)
				w.SetContent(mainContent)
			}
		})

		// Create dialog
		content := container.NewVBox(
			widget.NewLabel("No vault found. Create a new master password:"),
			passwordInput,
			createBtn,
		)

		w.SetContent(content)
	}
}

func createNewVault(w fyne.Window, appState *AppState, masterPassword string) bool {
	// Generate KDF parameters
	kdfParams := crypto.DefaultKDFParams()
	salt, err := crypto.GenerateSalt()
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to generate salt: %w", err), w)
		return false
	}
	kdfParams.Salt = salt

	// Derive keys from master password
	encKey, verKey, err := crypto.DeriveKeys(masterPassword, kdfParams)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to derive keys: %w", err), w)
		return false
	}

	// Save empty vault
	err = storage.WriteVault([]*model.PasswordEntry{}, vaultFile, encKey, verKey, kdfParams)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to create vault: %w", err), w)
		return false
	}

	// Store in app state
	appState.masterPassword = masterPassword
	appState.encryptionKey = encKey
	appState.verificationKey = verKey
	appState.kdfParams = kdfParams
	appState.isUnlocked = true

	dialog.ShowInformation("Success", "Vault created successfully!", w)
	return true
}

func unlockVault(w fyne.Window, appState *AppState, masterPassword string) bool {
	// Read vault file to get KDF parameters
	vaultData, err := os.ReadFile(vaultFile)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to read vault: %w", err), w)
		return false
	}

	vault, err := crypto.VaultFileDeserialize(vaultData)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to parse vault: %w", err), w)
		return false
	}

	// Derive keys using the provided master password and stored KDF params
	encKey, verKey, err := crypto.DeriveKeys(masterPassword, vault.KDFParams)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to derive keys: %w", err), w)
		return false
	}

	// Try to decrypt vault - this verifies the master password
	_, err = crypto.DecryptVault(vault, encKey, verKey)
	if err != nil {
		dialog.ShowError(fmt.Errorf("invalid master password or vault corrupted: %w", err), w)
		return false
	}

	// Store in app state
	appState.masterPassword = masterPassword
	appState.encryptionKey = encKey
	appState.verificationKey = verKey
	appState.kdfParams = vault.KDFParams
	appState.isUnlocked = true

	return true
}

func buildUI(w fyne.Window, fyneApp fyne.App, appState *AppState) *fyne.Container {
	// Password input field (masked)
	passwordInput := widget.NewPasswordEntry()
	passwordInput.PlaceHolder = "Enter password"

	// Add Password button
	addBtn := widget.NewButton("Add Password", func() {
		pass := passwordInput.Text
		if pass == "" {
			dialog.ShowError(fmt.Errorf("password cannot be empty"), w)
			return
		}

		// Run encryption in goroutine to avoid blocking UI
		go func() {
			appState.mu.Lock()
			defer appState.mu.Unlock()

			// Load current vault
			entries, err := storage.ReadVault(vaultFile, appState.encryptionKey, appState.verificationKey)
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("failed to read vault: %w", err), w)
				})
				return
			}

			// Encrypt password using Kyber + AES
			ct, ss, err := crypto.Encapsulate(appState.publicKey)
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("encapsulation failed: %v", err), w)
				})
				return
			}

			nonce, ciphertext, err := crypto.EncryptAES256GCM(pass, ss)
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("encryption failed: %v", err), w)
				})
				return
			}

			// Create new entry
			entry := model.NewPasswordEntry()
			entry.KyberCiphertext = ct
			entry.Nonce = nonce
			entry.Ciphertext = ciphertext

			// Add to vault
			entries = append(entries, entry)

			// Save updated vault
			err = storage.WriteVault(entries, vaultFile, appState.encryptionKey, appState.verificationKey, appState.kdfParams)
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("failed to save password: %v", err), w)
				})
				return
			}

			// Clear input and show success on main thread
			fyne.Do(func() {
				passwordInput.SetText("")
				dialog.ShowInformation("Success", "Password saved successfully!", w)
			})
		}()
	})

	// View Passwords button
	viewBtn := widget.NewButton("View Passwords", func() {
		// Run decryption in goroutine
		go func() {
			appState.mu.Lock()
			defer appState.mu.Unlock()

			entries, err := storage.ReadVault(vaultFile, appState.encryptionKey, appState.verificationKey)
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("failed to read vault: %w", err), w)
				})
				return
			}

			if len(entries) == 0 {
				fyne.Do(func() {
					dialog.ShowInformation("No Passwords", "No passwords stored yet.", w)
				})
				return
			}

			// Decrypt and display passwords on main thread
			fyne.Do(func() {
				showPasswordsWindow(w, fyneApp, entries, appState)
			})
		}()
	})

	// Lock vault button
	lockBtn := widget.NewButton("Lock Vault", func() {
		appState.mu.Lock()
		appState.isUnlocked = false
		appState.masterPassword = ""
		appState.encryptionKey = make([]byte, 0)
		appState.verificationKey = make([]byte, 0)
		appState.mu.Unlock()

		fyneApp.Quit()
	})

	// Layout
	buttonBox := container.NewVBox(
		widget.NewLabel("PassQuantum - Post-Quantum Safe Password Manager"),
		widget.NewLabel("(Vault is encrypted and secured)"),
		widget.NewLabel(""),
		widget.NewLabel("Enter a new password:"),
		passwordInput,
		addBtn,
		viewBtn,
		lockBtn,
	)

	return container.NewVBox(buttonBox)
}

func showPasswordsWindow(parentWindow fyne.Window, fyneApp fyne.App, entries []*model.PasswordEntry, appState *AppState) {
	// Create new window for displaying passwords
	decryptWindow := fyneApp.NewWindow("Your Passwords")
	decryptWindow.SetTitle("Your Passwords")
	decryptWindow.Resize(fyne.NewSize(500, 450))

	// Decrypt all passwords
	var items []fyne.CanvasObject

	// Add header
	items = append(items, widget.NewLabel(fmt.Sprintf("Total passwords: %d", len(entries))))
	items = append(items, widget.NewLabel(""))

	for i, entry := range entries {
		// Decapsulate to get shared secret
		ss, err := crypto.Decapsulate(entry.KyberCiphertext, appState.privateKey)
		if err != nil {
			items = append(items, widget.NewLabel(fmt.Sprintf("%d. ERROR Decapsulation: %v", i+1, err)))
			continue
		}

		// Decrypt using the shared secret
		plaintext, err := crypto.DecryptAES256GCM(entry.Nonce, entry.Ciphertext, ss)
		if err != nil {
			items = append(items, widget.NewLabel(fmt.Sprintf("%d. ERROR Decryption: %v", i+1, err)))
			continue
		}

		// Create a label with the decrypted password
		items = append(items, widget.NewLabel(fmt.Sprintf("%d. %s", i+1, plaintext)))
	}

	// If no items were added besides the header, show a message
	if len(items) == 2 {
		items = append(items, widget.NewLabel("No passwords could be displayed"))
	}

	// Add close button
	closeBtn := widget.NewButton("Close", func() {
		decryptWindow.Close()
	})
	items = append(items, widget.NewLabel(""))
	items = append(items, closeBtn)

	// Create scrollable list of passwords
	scrollBox := container.NewVScroll(container.NewVBox(items...))
	scrollBox.SetMinSize(fyne.NewSize(500, 400))

	decryptWindow.SetContent(scrollBox)
	decryptWindow.Show()
}
