package screens

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strings"

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

// ShowPasswordsView redirects to the main screen items view.
func ShowPasswordsView(w fyne.Window, fyneApp fyne.App, appState *app.AppState) {
	ShowMainScreen(w, fyneApp, appState)
}

func createVaultItemCard(index int, entry *model.VaultEntry, payload string, w fyne.Window, fyneApp fyne.App, appState *app.AppState) fyne.CanvasObject {
	switch entry.Type {
	case model.EntryTypeNote:
		return createNoteCard(index, entry, payload, w, fyneApp, appState)
	case model.EntryTypeCard:
		return createCardDetailsCard(index, entry, payload, w, fyneApp, appState)
	case model.EntryTypeTOTP:
		return createTOTPItemCard(index, entry, payload, w, fyneApp, appState)
	}

	if strings.HasPrefix(entry.Service, "NOTE:") {
		return createNoteCard(index, entry, payload, w, fyneApp, appState)
	}
	if strings.HasPrefix(entry.Service, "CARD:") {
		return createCardDetailsCard(index, entry, payload, w, fyneApp, appState)
	}
	if strings.HasPrefix(entry.Service, "TOTP:") {
		return createTOTPItemCard(index, entry, payload, w, fyneApp, appState)
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

	icon := theme.TypeIcon(theme.IconNote, theme.ColorAccentCyan)
	titleTxt := canvas.NewText(title, theme.ColorTextPrimary)
	titleTxt.TextSize = 13
	titleTxt.TextStyle = fyne.TextStyle{Bold: true}

	badge := theme.KindBadge("Note")
	titleRow := container.NewHBox(titleTxt, badge)

	noteLabel := canvas.NewText(preview, theme.ColorTextSecondary)
	noteLabel.TextSize = 11
	noteLabel.TextStyle = fyne.TextStyle{Monospace: true}

	showingFull := false
	viewBtn := theme.CreateSmallIconButton(theme.IconEye, func() {
		if showingFull {
			noteLabel.Text = preview
			showingFull = false
		} else {
			noteLabel.Text = content
			showingFull = true
		}
		noteLabel.Refresh()
	})

	copyBtn := theme.CreateSmallIconButton(theme.IconCopy, func() {
		w.Clipboard().SetContent(content)
		widgets.ShowAppInformation("Copied", "Note copied to clipboard", w)
	})

	deleteBtn := theme.CreateSmallIconButton(theme.IconTrash, func() {
		widgets.ShowAppConfirm("Delete", fmt.Sprintf("Delete note '%s'?", title), func(ok bool) {
			if ok {
				deleteEntryByID(entry.ID, "note", w, fyneApp, appState)
			}
		}, w)
	})

	left := container.NewHBox(icon, container.NewVBox(titleRow, noteLabel))
	buttons := container.NewHBox(viewBtn, copyBtn, deleteBtn)
	row := container.NewBorder(nil, nil, left, buttons)

	return theme.CardWithHeader("", "", nil, row)
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
		log.Printf("[Vault] WARNING: card entry %q has invalid JSON payload: %v", entry.Service, err)
		cp.Number = payload
	}

	masked := "****"
	if len(cp.Number) >= 4 {
		masked = cp.Number[len(cp.Number)-4:]
	}

	icon := theme.TypeIcon(theme.IconCard, theme.ColorAccentCyan)
	titleTxt := canvas.NewText(title, theme.ColorTextPrimary)
	titleTxt.TextSize = 13
	titleTxt.TextStyle = fyne.TextStyle{Bold: true}

	badge := theme.KindBadge(strings.ToUpper(cp.Subtype) + " Card")
	titleRow := container.NewHBox(titleTxt, badge)

	numberTxt := canvas.NewText("**** **** **** "+masked, theme.ColorTextSecondary)
	numberTxt.TextSize = 11
	numberTxt.TextStyle = fyne.TextStyle{Monospace: true}

	holderTxt := canvas.NewText(cp.Holder+" | Exp: "+cp.Expiry, theme.ColorFg2)
	holderTxt.TextSize = 11

	showingFull := false
	showBtn := theme.CreateSmallIconButton(theme.IconEye, func() {
		if showingFull {
			numberTxt.Text = "**** **** **** " + masked
			showingFull = false
		} else {
			numberTxt.Text = cp.Number
			showingFull = true
		}
		numberTxt.Refresh()
	})

	copyBtn := theme.CreateSmallIconButton(theme.IconCopy, func() {
		w.Clipboard().SetContent(cp.Number)
		widgets.ShowAppInformation("Copied", "Card number copied to clipboard", w)
	})

	deleteBtn := theme.CreateSmallIconButton(theme.IconTrash, func() {
		widgets.ShowAppConfirm("Delete", fmt.Sprintf("Delete card '%s'?", title), func(ok bool) {
			if ok {
				deleteEntryByID(entry.ID, "card", w, fyneApp, appState)
			}
		}, w)
	})

	left := container.NewHBox(icon, container.NewVBox(titleRow, container.NewVBox(numberTxt, holderTxt)))
	buttons := container.NewHBox(showBtn, copyBtn, deleteBtn)
	row := container.NewBorder(nil, nil, left, buttons)

	return theme.CardWithHeader("", "", nil, row)
}

func createPasswordCard(index int, entry *model.VaultEntry, password string, w fyne.Window, fyneApp fyne.App, appState *app.AppState) fyne.CanvasObject {
	icon := theme.TypeIcon(theme.IconKey, theme.ColorAccentCyan)

	titleTxt := canvas.NewText(entry.Service, theme.ColorTextPrimary)
	titleTxt.TextSize = 13
	titleTxt.TextStyle = fyne.TextStyle{Bold: true}

	badge := theme.KindBadge("Password")
	titleRow := container.NewHBox(titleTxt, badge)

	maskedPassword := canvas.NewText(entry.Username+" : ............", theme.ColorTextSecondary)
	maskedPassword.TextSize = 11
	maskedPassword.TextStyle = fyne.TextStyle{Monospace: true}

	showing := false
	showBtn := theme.CreateSmallIconButton(theme.IconEye, func() {
		if showing {
			maskedPassword.Text = entry.Username + " : ............"
			showing = false
		} else {
			maskedPassword.Text = entry.Username + " : " + password
			showing = true
		}
		maskedPassword.Refresh()
	})

	copyBtn := theme.CreateSmallIconButton(theme.IconCopy, func() {
		w.Clipboard().SetContent(password)
		widgets.ShowAppInformation("Copied", "Password copied to clipboard!", w)
	})

	editBtn := theme.CreateSmallIconButton(theme.IconEdit, func() {
		showEditPasswordDialog(entry, password, w, fyneApp, appState)
	})

	deleteBtn := theme.CreateSmallIconButton(theme.IconTrash, func() {
		widgets.ShowAppConfirm("Delete", fmt.Sprintf("Delete password for %s?", entry.Service), func(ok bool) {
			if ok {
				deleteEntryByID(entry.ID, "password", w, fyneApp, appState)
			}
		}, w)
	})

	left := container.NewHBox(icon, container.NewVBox(titleRow, maskedPassword))
	buttons := container.NewHBox(showBtn, copyBtn, editBtn, deleteBtn)
	row := container.NewBorder(nil, nil, left, buttons)

	return theme.CardWithHeader("", "", nil, row)
}

func showEditPasswordDialog(entry *model.VaultEntry, password string, w fyne.Window, fyneApp fyne.App, appState *app.AppState) {
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

	formContent := container.NewVBox(
		theme.SectionEyebrow("EDIT PASSWORD"),
		theme.FieldLabel("SERVICE", nil),
		serviceInput,
		theme.FieldLabel("USERNAME", nil),
		usernameInput,
		theme.FieldLabel("PASSWORD", nil),
		passwordInput,
		passwordStrengthBar,
	)

	var customDialog *dialog.CustomDialog

	saveBtn := theme.CreatePrimaryButton("Save changes", func() {
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
				ShowMainScreen(w, fyneApp, appState)
			})
		}(entry.ID)
	})

	cancelBtn := theme.CreateGhostButton("Cancel", func() {
		if customDialog != nil {
			customDialog.Hide()
		}
	})

	buttonBox := container.NewHBox(cancelBtn, saveBtn)
	dialogContent := container.NewVBox(formContent, container.NewCenter(buttonBox))
	customDialog = dialog.NewCustom("Edit Password", "Close", dialogContent, w)
	customDialog.Show()
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
			ShowMainScreen(w, fyneApp, appState)
		})
	}(entryID)
}

func capitalizeWord(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
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
	charsetLen := big.NewInt(int64(len(charset)))
	for i := 0; i < settings.Length; i++ {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("crypto/rand failure: %w", err)
		}
		password[i] = charset[n.Int64()]
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

// ShowPasswordGeneratorNoVault redirects to the vault selection screen.
func ShowPasswordGeneratorNoVault(w fyne.Window, fyneApp fyne.App, appState *app.AppState) {
	ShowVaultSelection(w, fyneApp, appState)
}
