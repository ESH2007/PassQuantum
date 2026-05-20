package screens

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"passquantum/core/model"
	"passquantum/app"
	"passquantum/theme"
	"passquantum/ui/widgets"
)

// PasswordGeneratorSettings holds configuration for password generation
type PasswordGeneratorSettings struct {
	Length           int
	UseUppercase     bool
	UseLowercase     bool
	UseNumbers       bool
	UseSpecialChars  bool
	ExcludeAmbiguous bool
}

// DefaultPasswordGeneratorSettings returns default settings
func DefaultPasswordGeneratorSettings() PasswordGeneratorSettings {
	return PasswordGeneratorSettings{
		Length:           16,
		UseUppercase:     true,
		UseLowercase:     true,
		UseNumbers:       true,
		UseSpecialChars:  true,
		ExcludeAmbiguous: true,
	}
}

// ShowPasswordsView displays all vault items in the current vault.
func ShowPasswordsView(w fyne.Window, fyneApp fyne.App, appState *app.AppState) {
	w.SetTitle("Your Vault Items - " + appState.CurrentVault)
	w.Resize(fyne.NewSize(650, 450))

	// Run decryption in goroutine
	go func() {
		appState.Mu.Lock()
		defer appState.Mu.Unlock()

		vaultFile := app.GetVaultPath(appState.CurrentVault)
		entries, err := app.ReadVault(vaultFile, appState.MasterPassword)
		if err != nil {
			fyne.Do(func() {
				widgets.ShowAppError(fmt.Errorf("failed to read vault: %w", err), w)
			})
			return
		}
		// entries is []*model.VaultEntry
		if len(entries) == 0 {
			fyne.Do(func() {
				widgets.ShowAppInformation("No Vault Items", "No vault items stored in this vault yet.", w)
			})
			return
		}

		// Decrypt and display vault items on main thread.
		fyne.Do(func() {
			displayPasswordsList(w, fyneApp, entries, appState)
		})
	}()
}

func displayPasswordsList(w fyne.Window, fyneApp fyne.App, entries []*model.VaultEntry, appState *app.AppState) {
	var items []fyne.CanvasObject

	headerText := theme.CreateLabel(fmt.Sprintf("VAULT ITEMS: %d", len(entries)), 14, theme.ColorAccentCyn, true)
	headerSection := container.NewVBox(headerText, theme.CreateDivider())
	items = append(items, headerSection)
	items = append(items, widget.NewLabel(""))

	for i, entry := range entries {
		ss, err := app.Decapsulate(entry.KyberCiphertext, appState.PrivateKey)
		if err != nil {
			errMsg := theme.CreateLabel(fmt.Sprintf("%d. ERROR: %v", i+1, err), 10, theme.ColorDanger, false)
			items = append(items, errMsg)
			continue
		}

		plaintext, err := app.DecryptAES256GCM(entry.Nonce, entry.Ciphertext, ss)
		if err != nil {
			errMsg := theme.CreateLabel(fmt.Sprintf("%d. ERROR: %v", i+1, err), 10, theme.ColorDanger, false)
			items = append(items, errMsg)
			continue
		}

		card := createVaultItemCard(i+1, entry, plaintext, w, fyneApp, appState)
		items = append(items, card)
		items = append(items, widget.NewLabel(""))
	}

	backBtn := theme.CreateNeonButton("← BACK", func() {
		ShowMainScreen(w, fyneApp, appState)
	}, 120, 44)
	items = append(items, backBtn)

	scrollBox := container.NewVScroll(container.NewVBox(items...))
	scrollBox.SetMinSize(fyne.NewSize(900, 600))

	bgContainer := theme.CreateBackgroundContainer(container.NewPadded(scrollBox))
	w.SetContent(bgContainer)
}

func createVaultItemCard(index int, entry *model.VaultEntry, payload string, w fyne.Window, fyneApp fyne.App, appState *app.AppState) fyne.CanvasObject {
	switch entry.Type {
	case model.EntryTypeNote:
		return createNoteCard(index, entry, payload, w, fyneApp, appState)
	case model.EntryTypeCard:
		return createCardDetailsCard(index, entry, payload, w, fyneApp, appState)
	}

	if strings.HasPrefix(entry.Service, "NOTE:") {
		return createNoteCard(index, entry, payload, w, fyneApp, appState)
	}
	if strings.HasPrefix(entry.Service, "CARD:") {
		return createCardDetailsCard(index, entry, payload, w, fyneApp, appState)
	}
	return createPasswordCard(index, entry, payload, w, fyneApp, appState)
}

func createNoteCard(index int, entry *model.VaultEntry, payload string, w fyne.Window, fyneApp fyne.App, appState *app.AppState) fyne.CanvasObject {
	title := strings.TrimPrefix(entry.Service, "NOTE:")
	content := payload

	var parsed map[string]string
	if err := json.Unmarshal([]byte(payload), &parsed); err == nil {
		if v := parsed["title"]; v != "" {
			title = v
		}
		if v := parsed["content"]; v != "" {
			content = v
		}
	}

	preview := content
	if len(preview) > 120 {
		preview = preview[:120] + "..."
	}

	titleLabel := theme.CreateLabel(fmt.Sprintf("#%d - NOTE: %s", index, title), 12, theme.ColorAccentCyn, true)
	noteLabel := widget.NewLabel("📝 " + preview)
	noteLabel.Wrapping = fyne.TextWrapWord

	showingFull := false
	viewBtn := theme.CreateNeonButton("VIEW", func() {
		if showingFull {
			noteLabel.SetText("📝 " + preview)
			showingFull = false
			return
		}
		noteLabel.SetText("📝 " + content)
		showingFull = true
	}, 70, 32)

	copyBtn := theme.CreateNeonButton("COPY", func() {
		w.Clipboard().SetContent(content)
		widgets.ShowAppInformation("Copied", "Note copied to clipboard", w)
	}, 70, 32)

	deleteBtn := theme.CreateNeonButton("DELETE", func() {
		widgets.ShowAppConfirm("Delete", fmt.Sprintf("Delete note '%s'?", title), func(ok bool) {
			if ok {
				deleteEntryByID(entry.ID, "note", w, fyneApp, appState)
			}
		}, w)
	}, 80, 32)

	buttons := container.NewHBox(viewBtn, copyBtn, deleteBtn)
	return theme.CreateCard(container.NewVBox(titleLabel, noteLabel, buttons), 850, 0, true)
}

func createCardDetailsCard(index int, entry *model.VaultEntry, payload string, w fyne.Window, fyneApp fyne.App, appState *app.AppState) fyne.CanvasObject {
	title := strings.TrimPrefix(entry.Service, "CARD:")

	type cardPayload struct {
		Subtype string `json:"subtype"`
		Holder  string `json:"holder"`
		Number  string `json:"number"`
		Expiry  string `json:"expiry"`
		CVV     string `json:"cvv"`
	}

	cp := cardPayload{Subtype: entry.Username}
	if entry.CardSubtype != "" {
		cp.Subtype = entry.CardSubtype
	}
	if err := json.Unmarshal([]byte(payload), &cp); err != nil {
		cp.Number = payload
	}

	masked := "****"
	if len(cp.Number) >= 4 {
		masked = cp.Number[len(cp.Number)-4:]
	}

	titleLabel := theme.CreateLabel(fmt.Sprintf("#%d - %s CARD: %s", index, strings.ToUpper(cp.Subtype), title), 12, theme.ColorAccentCyn, true)
	holderLabel := theme.CreateLabel("👤 "+cp.Holder, 10, theme.ColorTextSec, false)
	numberLabel := widget.NewLabel("💳 **** **** **** " + masked)
	expiryLabel := theme.CreateLabel("Expiry: "+cp.Expiry, 10, theme.ColorTextSec, false)

	showingFull := false
	showBtn := theme.CreateNeonButton("SHOW", func() {
		if showingFull {
			numberLabel.SetText("💳 **** **** **** " + masked)
			showingFull = false
			return
		}
		numberLabel.SetText("💳 " + cp.Number)
		showingFull = true
	}, 70, 32)

	copyBtn := theme.CreateNeonButton("COPY", func() {
		w.Clipboard().SetContent(cp.Number)
		widgets.ShowAppInformation("Copied", "Card number copied to clipboard", w)
	}, 70, 32)

	deleteBtn := theme.CreateNeonButton("DELETE", func() {
		widgets.ShowAppConfirm("Delete", fmt.Sprintf("Delete card '%s'?", title), func(ok bool) {
			if ok {
				deleteEntryByID(entry.ID, "card", w, fyneApp, appState)
			}
		}, w)
	}, 80, 32)

	buttons := container.NewHBox(showBtn, copyBtn, deleteBtn)
	return theme.CreateCard(container.NewVBox(titleLabel, holderLabel, numberLabel, expiryLabel, buttons), 850, 0, true)
}

func createPasswordCard(index int, entry *model.VaultEntry, password string, w fyne.Window, fyneApp fyne.App, appState *app.AppState) fyne.CanvasObject {
	passwordMasked := "••••••••••"

	serviceLabel := theme.CreateLabel(fmt.Sprintf("#%d - %s", index, entry.Service), 12, theme.ColorAccentCyn, true)
	usernameLabel := theme.CreateLabel("👤 "+entry.Username, 10, theme.ColorTextSec, false)
	passwordLabel := widget.NewLabel("🔐 " + passwordMasked)

	showPasswordBtn := theme.CreateNeonButton("SHOW", func() {
		if passwordLabel.Text == "🔐 "+passwordMasked {
			passwordLabel.SetText("🔐 " + password)
		} else {
			passwordLabel.SetText("🔐 " + passwordMasked)
		}
	}, 60, 32)

	copyBtn := theme.CreateNeonButton("COPY", func() {
		w.Clipboard().SetContent(password)
		widgets.ShowAppInformation("Copied", "Password copied to clipboard!", w)
	}, 60, 32)

	editBtn := theme.CreateNeonButton("EDIT", func() {
		serviceInput := widget.NewEntry()
		serviceInput.SetText(entry.Service)

		usernameInput := widget.NewEntry()
		usernameInput.SetText(entry.Username)

		passwordInput := widget.NewPasswordEntry()
		passwordInput.SetText(password)
		passwordStrengthBar := NewStrengthBar()
		BindStrengthBar(passwordStrengthBar, passwordInput, func() []string {
			return storedVaultPasswords(appState)
		})

		// Create styled input containers
		createInputContainer := func(input fyne.CanvasObject) fyne.CanvasObject {
			bg := canvas.NewRectangle(color.NRGBA{R: 30, G: 40, B: 50, A: 255})
			bg.SetMinSize(fyne.NewSize(350, 50))
			bg.CornerRadius = theme.BorderRadius
			return container.NewMax(bg, container.NewPadded(input))
		}

		// Build form content
		formContent := container.NewVBox(
			theme.CreateLabel("EDIT PASSWORD", 14, theme.ColorAccentCyn, true),
			theme.CreateDivider(),
			widget.NewLabel(""),
			theme.CreateLabel("Service", 11, theme.ColorPurple, true),
			createInputContainer(serviceInput),
			widget.NewLabel(""),
			theme.CreateLabel("Username", 11, theme.ColorPurple, true),
			createInputContainer(usernameInput),
			widget.NewLabel(""),
			theme.CreateLabel("Password", 11, theme.ColorPurple, true),
			createInputContainer(passwordInput),
			widget.NewLabel(""),
			passwordStrengthBar,
			widget.NewLabel(""),
		)

		// Declare customDialog first so it can be referenced in button closures
		var customDialog *dialog.CustomDialog

		saveBtn := theme.CreateNeonButton("✓ SAVE", func() {
			newService := serviceInput.Text
			newUsername := usernameInput.Text
			newPassword := passwordInput.Text

			go func(id uint64) {
				appState.Mu.Lock()
				defer appState.Mu.Unlock()

				vaultFile := app.GetVaultPath(appState.CurrentVault)
				entries, err := app.ReadVault(vaultFile, appState.MasterPassword)
				if err != nil {
					fyne.Do(func() {
						widgets.ShowAppError(fmt.Errorf("failed to read vault: %w", err), w)
					})
					return
				}

				updated := false
				for _, e := range entries {
					if e.ID == id {
						e.Service = newService
						e.Username = newUsername
						if newPassword != password {
							ct, ss, cerr := app.Encapsulate(appState.PublicKey)
							if cerr != nil {
								fyne.Do(func() {
									widgets.ShowAppError(fmt.Errorf("encapsulation failed: %v", cerr), w)
								})
								return
							}

							nonce, ciphertext, cerr := app.EncryptAES256GCM(newPassword, ss)
							if cerr != nil {
								fyne.Do(func() {
									widgets.ShowAppError(fmt.Errorf("encryption failed: %v", cerr), w)
								})
								return
							}

							e.KyberCiphertext = ct
							e.Nonce = nonce
							e.Ciphertext = ciphertext
						}
						updated = true
						break
					}
				}

				if !updated {
					fyne.Do(func() {
						widgets.ShowAppError(fmt.Errorf("entry not found"), w)
					})
					return
				}

				err = app.WriteVault(entries, vaultFile, appState.MasterPassword)
				if err != nil {
					fyne.Do(func() {
						widgets.ShowAppError(fmt.Errorf("failed to save vault: %w", err), w)
					})
					return
				}

				fyne.Do(func() {
					if customDialog != nil {
						customDialog.Hide()
					}
					widgets.ShowAppInformation("Updated", "Vault item updated", w)
					ShowPasswordsView(w, fyneApp, appState)
				})
			}(entry.ID)
		}, 120, 44)

		cancelBtn := theme.CreateNeonButton("✕ CANCEL", func() {
			if customDialog != nil {
				customDialog.Hide()
			}
		}, 120, 44)

		buttonBox := container.NewHBox(cancelBtn, saveBtn)

		dialogContent := container.NewVBox(formContent, container.NewCenter(buttonBox))
		customDialog = dialog.NewCustom("Edit Password", "Close", dialogContent, w)
		customDialog.Show()
	}, 60, 32)

	deleteBtn := theme.CreateNeonButton("DELETE", func() {
		widgets.ShowAppConfirm("Delete", fmt.Sprintf("Delete password for %s?", entry.Service), func(ok bool) {
			if !ok {
				return
			}
			deleteEntryByID(entry.ID, "password", w, fyneApp, appState)
		}, w)
	}, 80, 32)

	buttonRow := container.NewHBox(showPasswordBtn, copyBtn, editBtn, deleteBtn)
	content := container.NewVBox(serviceLabel, usernameLabel, passwordLabel, buttonRow)

	return theme.CreateCard(content, 850, 0, true)
}

func deleteEntryByID(entryID uint64, entryKind string, w fyne.Window, fyneApp fyne.App, appState *app.AppState) {
	go func(id uint64) {
		appState.Mu.Lock()
		defer appState.Mu.Unlock()

		vaultFile := app.GetVaultPath(appState.CurrentVault)
		entries, err := app.ReadVault(vaultFile, appState.MasterPassword)
		if err != nil {
			fyne.Do(func() {
				widgets.ShowAppError(fmt.Errorf("failed to read vault: %w", err), w)
			})
			return
		}

		newEntries := make([]*model.VaultEntry, 0, len(entries))
		for _, e := range entries {
			if e.ID != id {
				newEntries = append(newEntries, e)
			}
		}

		err = app.WriteVault(newEntries, vaultFile, appState.MasterPassword)
		if err != nil {
			fyne.Do(func() {
				widgets.ShowAppError(fmt.Errorf("failed to delete %s: %w", entryKind, err), w)
			})
			return
		}

		fyne.Do(func() {
			widgets.ShowAppInformation("Deleted", capitalizeWord(entryKind)+" deleted successfully", w)
			ShowPasswordsView(w, fyneApp, appState)
		})
	}(entryID)
}

func capitalizeWord(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

// ShowPasswordGenerator displays the password generator screen
func ShowPasswordGenerator(w fyne.Window, fyneApp fyne.App, appState *app.AppState) {
	w.SetTitle("Password Generator - " + appState.CurrentVault)
	w.Resize(fyne.NewSize(650, 450))

	settings := DefaultPasswordGeneratorSettings()
	generatedPasswordDisplay := widget.NewLabel("")

	lengthInput := widget.NewEntry()
	lengthInput.SetText("16")
	lengthInput.OnChanged = func(s string) {
		if s != "" {
			fmt.Sscanf(s, "%d", &settings.Length)
			if settings.Length < 4 {
				settings.Length = 4
			}
			if settings.Length > 128 {
				settings.Length = 128
			}
		}
	}

	uppercaseCheck := widget.NewCheck("Uppercase Letters (A-Z)", func(b bool) {
		settings.UseUppercase = b
	})
	uppercaseCheck.SetChecked(settings.UseUppercase)

	lowercaseCheck := widget.NewCheck("Lowercase Letters (a-z)", func(b bool) {
		settings.UseLowercase = b
	})
	lowercaseCheck.SetChecked(settings.UseLowercase)

	numbersCheck := widget.NewCheck("Numbers (0-9)", func(b bool) {
		settings.UseNumbers = b
	})
	numbersCheck.SetChecked(settings.UseNumbers)

	specialCharsCheck := widget.NewCheck("Special Characters (!@#$%^&*)", func(b bool) {
		settings.UseSpecialChars = b
	})
	specialCharsCheck.SetChecked(settings.UseSpecialChars)

	ambiguousCheck := widget.NewCheck("Exclude Ambiguous (i, l, 1, L, o, 0, O)", func(b bool) {
		settings.ExcludeAmbiguous = b
	})
	ambiguousCheck.SetChecked(settings.ExcludeAmbiguous)

	generateBtn := theme.CreateNeonButton("🔄 GENERATE", func() {
		password, err := GeneratePassword(settings)
		if err != nil {
			widgets.ShowAppError(err, w)
			return
		}
		generatedPasswordDisplay.SetText(password)
	}, 160, 44)

	copyGeneratedBtn := theme.CreateNeonButton("COPY", func() {
		if generatedPasswordDisplay.Text != "" {
			w.Clipboard().SetContent(generatedPasswordDisplay.Text)
			widgets.ShowAppInformation("Copied", "Password copied to clipboard!", w)
		} else {
			widgets.ShowAppInformation("Empty", "Generate a password first!", w)
		}
	}, 100, 44)

	usePasswordBtn := theme.CreateNeonButton("USE PASSWORD", func() {
		if generatedPasswordDisplay.Text == "" {
			widgets.ShowAppInformation("Empty", "Generate a password first!", w)
			return
		}

		// Show vault selection first
		vaults := app.ListVaults()
		if len(vaults) == 0 {
			widgets.ShowAppError(fmt.Errorf("no vaults available"), w)
			return
		}

		vaultSelection := widget.NewSelect(vaults, nil)
		vaultSelection.SetSelected(appState.CurrentVault)

		serviceInput := widget.NewEntry()
		serviceInput.PlaceHolder = "Service name"

		usernameInput := widget.NewEntry()
		usernameInput.PlaceHolder = "Username or email"

		generatedPasswordInput := widget.NewEntry()
		generatedPasswordInput.SetText(generatedPasswordDisplay.Text)
		generatedPasswordInput.Disable()
		passwordStrengthBar := NewStrengthBar()
		BindStrengthBar(passwordStrengthBar, generatedPasswordInput, func() []string {
			return storedVaultPasswords(appState)
		})

		// Create styled input containers
		createInputContainer := func(input fyne.CanvasObject) fyne.CanvasObject {
			bg := canvas.NewRectangle(color.NRGBA{R: 30, G: 40, B: 50, A: 255})
			bg.SetMinSize(fyne.NewSize(350, 50))
			bg.CornerRadius = theme.BorderRadius
			return container.NewStack(bg, container.NewPadded(input))
		}

		// Create styled select container
		selectBg := canvas.NewRectangle(color.NRGBA{R: 30, G: 40, B: 50, A: 255})
		selectBg.SetMinSize(fyne.NewSize(350, 50))
		selectBg.CornerRadius = theme.BorderRadius
		selectContainer := container.NewStack(selectBg, container.NewPadded(vaultSelection))

		// Build form content
		formContent := container.NewVBox(
			theme.CreateLabel("ADD GENERATED PASSWORD", 14, theme.ColorAccentCyn, true),
			theme.CreateDivider(),
			widget.NewLabel(""),
			theme.CreateLabel("Vault", 11, theme.ColorPurple, true),
			selectContainer,
			widget.NewLabel(""),
			theme.CreateLabel("Service", 11, theme.ColorPurple, true),
			createInputContainer(serviceInput),
			widget.NewLabel(""),
			theme.CreateLabel("Username", 11, theme.ColorPurple, true),
			createInputContainer(usernameInput),
			widget.NewLabel(""),
			theme.CreateLabel("Strength Analysis", 11, theme.ColorPurple, true),
			passwordStrengthBar,
			widget.NewLabel(""),
		)

		// Declare customDialog first so it can be referenced in button closures
		var customDialog *dialog.CustomDialog

		saveBtn := theme.CreateNeonButton("✓ SAVE", func() {

			selectedVault := vaultSelection.Selected
			service := serviceInput.Text
			username := usernameInput.Text
			password := generatedPasswordDisplay.Text

			if selectedVault == "" {
				widgets.ShowAppError(fmt.Errorf("vault must be selected"), w)
				return
			}

			if service == "" {
				widgets.ShowAppError(fmt.Errorf("service name cannot be empty"), w)
				return
			}

			if password == "" {
				widgets.ShowAppError(fmt.Errorf("please generate a password first"), w)
				return
			}

			// Open the selected vault first - this updates appState with correct vault keys
			if err := app.OpenVault(appState, selectedVault, func() {
				// Now appState has the correct keys for the selected vault
				go func() {
					appState.Mu.Lock()
					defer appState.Mu.Unlock()

					// Use appState keys directly - they now match the selected vault
					vaultFile := app.GetVaultPath(selectedVault)

					entries, err := app.ReadVault(vaultFile, appState.MasterPassword)
					if err != nil {
						fyne.Do(func() {
							widgets.ShowAppError(fmt.Errorf("failed to read vault: %w", err), w)
						})
						return
					}

					ct, ss, err := app.Encapsulate(appState.PublicKey)
					if err != nil {
						fyne.Do(func() { widgets.ShowAppError(err, w) })
						return
					}

					nonce, ciphertext, err := app.EncryptAES256GCM(password, ss)
					if err != nil {
						fyne.Do(func() { widgets.ShowAppError(err, w) })
						return
					}

					entry := model.NewVaultEntry()
					entry.Type = model.EntryTypePassword
					entry.Service = service
					entry.Username = username
					entry.KyberCiphertext = ct
					entry.Nonce = nonce
					entry.Ciphertext = ciphertext

					entries = append(entries, entry)

					// CRITICAL FIX: Use appState keys directly - they match the selected vault
					// Previously this used stale snapshotted keys from before vault switch
					err = app.WriteVault(entries, vaultFile, appState.MasterPassword)

					if err != nil {
						fyne.Do(func() { widgets.ShowAppError(err, w) })
						return
					}

					fyne.Do(func() {
						if customDialog != nil {
							customDialog.Hide()
						}
						widgets.ShowAppInformation("Success", "Vault item saved to vault!", w)
					})
				}()
			}); err != nil {
				widgets.ShowAppError(err, w)
			}
		}, 120, 44)

		cancelBtn := theme.CreateNeonButton("✕ CANCEL", func() {
			if customDialog != nil {
				customDialog.Hide()
			}
		}, 120, 44)

		buttonBox := container.NewHBox(cancelBtn, saveBtn)

		dialogContent := container.NewVBox(formContent, container.NewCenter(buttonBox))
		customDialog = dialog.NewCustom("Add Generated Password", "Close", dialogContent, w)
		customDialog.Show()
	}, 160, 44)

	backBtn := theme.CreateNeonButton("← BACK", func() {
		ShowPasswordsView(w, fyneApp, appState)
	}, 100, 44)

	headerText := theme.CreateLabel("PASSWORD GENERATOR", 14, theme.ColorAccentCyn, true)
	headerSection := container.NewVBox(headerText, theme.CreateDivider())

	settingsHeader := theme.CreateLabel("⚙️ QUANTUM PARAMETERS", 11, theme.ColorPurple, true)
	lengthLabel := theme.CreateLabel("Length (4-128):", 10, theme.ColorTextSec, false)

	charTypesHeader := theme.CreateLabel("CHARACTER TYPES", 11, theme.ColorPurple, true)

	generatedLabel := theme.CreateLabel("GENERATED PASSWORD", 11, theme.ColorPurple, true)
	generatedPasswordBg := canvas.NewRectangle(color.NRGBA{R: 30, G: 40, B: 50, A: 255})
	generatedPasswordBg.SetMinSize(fyne.NewSize(750, 50))
	generatedPasswordBg.CornerRadius = theme.BorderRadius

	generatedCard := container.NewMax(generatedPasswordBg, container.NewCenter(generatedPasswordDisplay))

	actionButtons := container.NewHBox(copyGeneratedBtn, usePasswordBtn)

	content := container.NewVBox(
		headerSection,
		widget.NewLabel(""),
		container.NewCenter(generateBtn),
		widget.NewLabel(""),
		generatedLabel,
		generatedCard,
		widget.NewLabel(""),
		container.NewCenter(actionButtons),
		widget.NewLabel(""),
		settingsHeader,
		lengthLabel,
		lengthInput,
		widget.NewLabel(""),
		charTypesHeader,
		uppercaseCheck,
		lowercaseCheck,
		numbersCheck,
		specialCharsCheck,
		ambiguousCheck,
		widget.NewLabel(""),
		container.NewCenter(backBtn),
	)

	scrollBox := container.NewVScroll(content)
	scrollBox.SetMinSize(fyne.NewSize(900, 600))

	bgContainer := theme.CreateBackgroundContainer(container.NewPadded(scrollBox))
	w.SetContent(bgContainer)
}

// GeneratePassword generates a random password based on settings
func GeneratePassword(settings PasswordGeneratorSettings) (string, error) {
	const (
		uppercase    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		lowercase    = "abcdefghijklmnopqrstuvwxyz"
		numbers      = "0123456789"
		specialChars = "!@#$%^&*()_+-=[]{}|;:,.<>?"
		ambiguous    = "il1Lo0O"
	)

	var charset string

	if settings.UseUppercase {
		charset += uppercase
	}
	if settings.UseLowercase {
		charset += lowercase
	}
	if settings.UseNumbers {
		charset += numbers
	}
	if settings.UseSpecialChars {
		charset += specialChars
	}

	if settings.ExcludeAmbiguous {
		for _, char := range ambiguous {
			charset = removeChar(charset, char)
		}
	}

	if len(charset) == 0 {
		return "", fmt.Errorf("at least one character type must be selected")
	}

	password := make([]byte, settings.Length)
	for i := 0; i < settings.Length; i++ {
		password[i] = charset[rand.Intn(len(charset))]
	}

	return string(password), nil
}

// removeChar removes a character from a string
func removeChar(s string, char rune) string {
	result := ""
	for _, c := range s {
		if c != char {
			result += string(c)
		}
	}
	return result
}

// ShowPasswordGeneratorNoVault displays the password generator screen when called from vault selection
func ShowPasswordGeneratorNoVault(w fyne.Window, fyneApp fyne.App, appState *app.AppState) {
	w.SetTitle("Password Generator")
	w.Resize(fyne.NewSize(650, 450))

	settings := DefaultPasswordGeneratorSettings()
	generatedPasswordDisplay := widget.NewLabel("")

	lengthInput := widget.NewEntry()
	lengthInput.SetText("16")
	lengthInput.OnChanged = func(s string) {
		if s != "" {
			fmt.Sscanf(s, "%d", &settings.Length)
			if settings.Length < 4 {
				settings.Length = 4
			}
			if settings.Length > 128 {
				settings.Length = 128
			}
		}
	}

	uppercaseCheck := widget.NewCheck("Uppercase Letters (A-Z)", func(b bool) {
		settings.UseUppercase = b
	})
	uppercaseCheck.SetChecked(settings.UseUppercase)

	lowercaseCheck := widget.NewCheck("Lowercase Letters (a-z)", func(b bool) {
		settings.UseLowercase = b
	})
	lowercaseCheck.SetChecked(settings.UseLowercase)

	numbersCheck := widget.NewCheck("Numbers (0-9)", func(b bool) {
		settings.UseNumbers = b
	})
	numbersCheck.SetChecked(settings.UseNumbers)

	specialCharsCheck := widget.NewCheck("Special Characters (!@#$%^&*)", func(b bool) {
		settings.UseSpecialChars = b
	})
	specialCharsCheck.SetChecked(settings.UseSpecialChars)

	ambiguousCheck := widget.NewCheck("Exclude Ambiguous (i, l, 1, L, o, 0, O)", func(b bool) {
		settings.ExcludeAmbiguous = b
	})
	ambiguousCheck.SetChecked(settings.ExcludeAmbiguous)

	generateBtn := theme.CreateNeonButton("🔄 GENERATE", func() {
		password, err := GeneratePassword(settings)
		if err != nil {
			widgets.ShowAppError(err, w)
			return
		}
		generatedPasswordDisplay.SetText(password)
	}, 160, 44)

	copyGeneratedBtn := theme.CreateNeonButton("COPY", func() {
		if generatedPasswordDisplay.Text != "" {
			w.Clipboard().SetContent(generatedPasswordDisplay.Text)
			widgets.ShowAppInformation("Copied", "Password copied to clipboard!", w)
		} else {
			widgets.ShowAppInformation("Empty", "Generate a password first!", w)
		}
	}, 100, 44)

	usePasswordBtn := theme.CreateNeonButton("USE PASSWORD", func() {
		if generatedPasswordDisplay.Text == "" {
			widgets.ShowAppInformation("Empty", "Generate a password first!", w)
			return
		}

		// Show vault selection first
		vaults := app.ListVaults()
		if len(vaults) == 0 {
			widgets.ShowAppError(fmt.Errorf("no vaults available"), w)
			return
		}

		vaultSelection := widget.NewSelect(vaults, nil)
		if len(vaults) > 0 {
			vaultSelection.SetSelected(vaults[0])
		}

		serviceInput := widget.NewEntry()
		serviceInput.PlaceHolder = "Service name"

		usernameInput := widget.NewEntry()
		usernameInput.PlaceHolder = "Username or email"

		generatedPasswordInput := widget.NewEntry()
		generatedPasswordInput.SetText(generatedPasswordDisplay.Text)
		generatedPasswordInput.Disable()
		passwordStrengthBar := NewStrengthBar()
		BindStrengthBar(passwordStrengthBar, generatedPasswordInput, func() []string {
			return storedVaultPasswords(appState)
		})

		// Create styled input containers
		createInputContainer := func(input fyne.CanvasObject) fyne.CanvasObject {
			bg := canvas.NewRectangle(color.NRGBA{R: 30, G: 40, B: 50, A: 255})
			bg.SetMinSize(fyne.NewSize(350, 50))
			bg.CornerRadius = theme.BorderRadius
			return container.NewMax(bg, container.NewPadded(input))
		}

		// Create styled select container
		selectBg := canvas.NewRectangle(color.NRGBA{R: 30, G: 40, B: 50, A: 255})
		selectBg.SetMinSize(fyne.NewSize(350, 50))
		selectBg.CornerRadius = theme.BorderRadius
		selectContainer := container.NewMax(selectBg, container.NewPadded(vaultSelection))

		// Build form content
		formContent := container.NewVBox(
			theme.CreateLabel("ADD GENERATED PASSWORD", 14, theme.ColorAccentCyn, true),
			theme.CreateDivider(),
			widget.NewLabel(""),
			theme.CreateLabel("Vault", 11, theme.ColorPurple, true),
			selectContainer,
			widget.NewLabel(""),
			theme.CreateLabel("Service", 11, theme.ColorPurple, true),
			createInputContainer(serviceInput),
			widget.NewLabel(""),
			theme.CreateLabel("Username", 11, theme.ColorPurple, true),
			createInputContainer(usernameInput),
			widget.NewLabel(""),
			theme.CreateLabel("Strength Analysis", 11, theme.ColorPurple, true),
			passwordStrengthBar,
			widget.NewLabel(""),
		)

		// Declare customDialog first so it can be referenced in button closures
		var customDialog *dialog.CustomDialog

		saveBtn := theme.CreateNeonButton("✓ SAVE", func() {
			selectedVault := vaultSelection.Selected
			service := serviceInput.Text
			username := usernameInput.Text
			password := generatedPasswordDisplay.Text

			if selectedVault == "" {
				widgets.ShowAppError(fmt.Errorf("vault must be selected"), w)
				return
			}

			if service == "" {
				widgets.ShowAppError(fmt.Errorf("service name cannot be empty"), w)
				return
			}

			if password == "" {
				widgets.ShowAppError(fmt.Errorf("please generate a password first"), w)
				return
			}

			// Open the selected vault to ensure we have the right keys
			if err := app.OpenVault(appState, selectedVault, func() {
				go func() {
					appState.Mu.Lock()
					defer appState.Mu.Unlock()

					vaultFile := app.GetVaultPath(selectedVault)
					entries, err := app.ReadVault(vaultFile, appState.MasterPassword)
					if err != nil {
						fyne.Do(func() {
							widgets.ShowAppError(fmt.Errorf("failed to read vault: %w", err), w)
						})
						return
					}

					ct, ss, err := app.Encapsulate(appState.PublicKey)
					if err != nil {
						fyne.Do(func() {
							widgets.ShowAppError(fmt.Errorf("encapsulation failed: %v", err), w)
						})
						return
					}

					nonce, ciphertext, err := app.EncryptAES256GCM(password, ss)
					if err != nil {
						fyne.Do(func() {
							widgets.ShowAppError(fmt.Errorf("encryption failed: %v", err), w)
						})
						return
					}

					entry := model.NewVaultEntry()
					entry.Type = model.EntryTypePassword
					entry.Service = service
					entry.Username = username
					entry.KyberCiphertext = ct
					entry.Nonce = nonce
					entry.Ciphertext = ciphertext

					entries = append(entries, entry)

					err = app.WriteVault(entries, vaultFile, appState.MasterPassword)
					if err != nil {
						fyne.Do(func() {
							widgets.ShowAppError(fmt.Errorf("failed to save vault: %w", err), w)
						})
						return
					}

					fyne.Do(func() {
						if customDialog != nil {
							customDialog.Hide()
						}
						widgets.ShowAppInformation("Success", "Vault item saved to vault!", w)
						ShowVaultSelection(w, fyneApp, appState)
					})
				}()
			}); err != nil {
				widgets.ShowAppError(err, w)
			}
		}, 120, 44)

		cancelBtn := theme.CreateNeonButton("✕ CANCEL", func() {
			if customDialog != nil {
				customDialog.Hide()
			}
		}, 120, 44)

		buttonBox := container.NewHBox(cancelBtn, saveBtn)

		dialogContent := container.NewVBox(formContent, container.NewCenter(buttonBox))
		customDialog = dialog.NewCustom("Add Generated Password", "Close", dialogContent, w)
		customDialog.Show()
	}, 160, 44)

	backBtn := theme.CreateNeonButton("← BACK", func() {
		ShowVaultSelection(w, fyneApp, appState)
	}, 100, 44)

	headerText := theme.CreateLabel("PASSWORD GENERATOR", 14, theme.ColorAccentCyn, true)
	headerSection := container.NewVBox(headerText, theme.CreateDivider())

	settingsHeader := theme.CreateLabel("⚙️ QUANTUM PARAMETERS", 11, theme.ColorPurple, true)
	lengthLabel := theme.CreateLabel("Length (4-128):", 10, theme.ColorTextSec, false)

	charTypesHeader := theme.CreateLabel("CHARACTER TYPES", 11, theme.ColorPurple, true)

	generatedLabel := theme.CreateLabel("GENERATED PASSWORD", 11, theme.ColorPurple, true)
	generatedPasswordBg := canvas.NewRectangle(color.NRGBA{R: 30, G: 40, B: 50, A: 255})
	generatedPasswordBg.SetMinSize(fyne.NewSize(750, 50))
	generatedPasswordBg.CornerRadius = theme.BorderRadius

	generatedCard := container.NewMax(generatedPasswordBg, container.NewCenter(generatedPasswordDisplay))

	actionButtons := container.NewHBox(copyGeneratedBtn, usePasswordBtn)

	content := container.NewVBox(
		headerSection,
		widget.NewLabel(""),
		settingsHeader,
		lengthLabel,
		lengthInput,
		widget.NewLabel(""),
		charTypesHeader,
		uppercaseCheck,
		lowercaseCheck,
		numbersCheck,
		specialCharsCheck,
		ambiguousCheck,
		widget.NewLabel(""),
		container.NewCenter(generateBtn),
		widget.NewLabel(""),
		generatedLabel,
		generatedCard,
		widget.NewLabel(""),
		container.NewCenter(actionButtons),
		widget.NewLabel(""),
		container.NewCenter(backBtn),
	)

	scrollBox := container.NewVScroll(content)
	scrollBox.SetMinSize(fyne.NewSize(900, 600))

	bgContainer := theme.CreateBackgroundContainer(container.NewPadded(scrollBox))
	w.SetContent(bgContainer)
}
