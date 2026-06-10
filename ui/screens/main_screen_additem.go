package screens

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"passquantum/app"
	"passquantum/core/model"
	"passquantum/core/totp"
	"passquantum/strength"
	"passquantum/theme"
	"passquantum/ui/widgets"
)

func (ns *NavigationState) createPasswordsView() fyne.CanvasObject {
	type cardPayload struct {
		Subtype string `json:"subtype"`
		Holder  string `json:"holder"`
		Number  string `json:"number"`
		Expiry  string `json:"expiry"`
		CVV     string `json:"cvv"`
	}

	backBtn := theme.CreateGhostButton("All items", func() {
		ns.switchView(NavViewItems)
	})

	header := theme.PageHeader(
		"PASSQUANTUM / "+ns.appState.CurrentVault,
		"Add vault item",
		"Create a new encrypted entry in this vault.",
		backBtn,
	)

	passwordInput := widget.NewPasswordEntry()
	passwordInput.PlaceHolder = "Enter password"

	// Segmented strength meter
	strengthMeter := theme.SegmentedStrengthMeterFlex(0)
	strengthLabel := canvas.NewText("", theme.ColorTextSecondary)
	strengthLabel.TextSize = 11
	strengthLabel.TextStyle = fyne.TextStyle{Monospace: true}
	strengthCrack := canvas.NewText("", theme.ColorFg2)
	strengthCrack.TextSize = 11
	strengthCrack.TextStyle = fyne.TextStyle{Monospace: true}

	strengthIssues := container.NewVBox(
		newStrengthText("Start typing to analyze password strength.", theme.ColorFg2, 11, false),
	)

	strengthContainer := container.NewVBox(
		container.NewBorder(nil, nil, nil, strengthLabel, strengthMeter),
		strengthCrack,
		strengthIssues,
	)

	updateAddStrength := func(value string) {
		result := strength.Analyze(value, storedVaultPasswords(ns.appState))

		if result.EasterEggMode {
			level := int(result.Score) + 1
			if level > 5 {
				level = 5
			}
			strengthMeter = theme.SegmentedStrengthMeterFlex(level)
			strengthLabel.Text = "Password Game Mode"
			strengthLabel.Refresh()
			strengthCrack.Text = "neal.fun trigger detected"
			strengthCrack.Refresh()
			strengthIssues.Objects = []fyne.CanvasObject{NewEasterEggPanel(result.EasterEggRules)}
			strengthIssues.Refresh()
			return
		}

		level := 0
		if value != "" {
			level = int(result.Score) + 1
			if level > 5 {
				level = 5
			}
		}

		newMeter := theme.SegmentedStrengthMeterFlex(level)
		strengthContainer.Objects[0] = container.NewBorder(nil, nil, nil, strengthLabel, newMeter)
		strengthContainer.Refresh()

		if value == "" {
			strengthLabel.Text = ""
			strengthLabel.Refresh()
			strengthCrack.Text = ""
			strengthCrack.Refresh()
			strengthIssues.Objects = []fyne.CanvasObject{
				newStrengthText("Start typing to analyze password strength.", theme.ColorFg2, 11, false),
			}
		} else {
			strengthLabel.Text = result.ScoreLabel
			strengthLabel.Refresh()
			strengthCrack.Text = "Crack time: " + result.CrackTime
			strengthCrack.Refresh()
			strengthIssues.Objects = []fyne.CanvasObject{NewIssuesList(result.Issues)}
		}
		strengthIssues.Refresh()
	}

	passwordInput.OnChanged = updateAddStrength
	updateAddStrength(passwordInput.Text)

	serviceInput := widget.NewEntry()
	serviceInput.PlaceHolder = "e.g. github.com"

	usernameInput := widget.NewEntry()
	usernameInput.PlaceHolder = "Username or email"

	noteTitleInput := widget.NewEntry()
	noteTitleInput.PlaceHolder = "Note title"
	noteContentInput := widget.NewMultiLineEntry()
	noteContentInput.SetMinRowsVisible(5)
	noteContentInput.PlaceHolder = "Plaintext is encrypted at rest. Markdown is preserved."

	cardTypeSelect := widget.NewSelect([]string{"Credit", "Debit", "Prepaid"}, nil)
	cardTypeSelect.SetSelected("Credit")
	cardNameInput := widget.NewEntry()
	cardNameInput.PlaceHolder = "Card nickname"
	cardHolderInput := widget.NewEntry()
	cardHolderInput.PlaceHolder = "Card holder"
	cardNumberInput := widget.NewEntry()
	cardNumberInput.PlaceHolder = "Card number"
	cardExpiryInput := widget.NewEntry()
	cardExpiryInput.PlaceHolder = "MM/YY"
	cardCVVInput := widget.NewPasswordEntry()
	cardCVVInput.PlaceHolder = "CVV"

	// TOTP fields
	totpIssuerInput := widget.NewEntry()
	totpIssuerInput.PlaceHolder = "e.g. GitHub"
	totpAccountInput := widget.NewEntry()
	totpAccountInput.PlaceHolder = "user@example.com"
	totpSecretInput := widget.NewEntry()
	totpSecretInput.PlaceHolder = "Base32 secret (e.g. JBSWY3DPEHPK3PXP)"
	totpAlgorithmSelect := widget.NewSelect([]string{"SHA1", "SHA256", "SHA512"}, nil)
	totpAlgorithmSelect.SetSelected("SHA1")
	totpDigitsSelect := widget.NewSelect([]string{"6", "7", "8"}, nil)
	totpDigitsSelect.SetSelected("6")
	totpPeriodSelect := widget.NewSelect([]string{"30", "60", "90"}, nil)
	totpPeriodSelect.SetSelected("30")

	itemTypes := []string{"Password", "Cyphered Note", "Card", "TOTP"}
	activeItemType := 0 // 0=Password, 1=Note, 2=Card, 3=TOTP

	passwordSection := container.NewVBox(
		theme.FieldLabel("SERVICE", nil),
		serviceInput,
		theme.FieldLabel("USERNAME / EMAIL", nil),
		usernameInput,
		theme.FieldLabel("PASSWORD", nil),
		passwordInput,
		strengthContainer,
	)

	noteSection := container.NewVBox(
		theme.FieldLabel("NOTE TITLE", nil),
		noteTitleInput,
		theme.FieldLabel("CYPHERED NOTE", nil),
		noteContentInput,
	)

	cardSection := container.NewVBox(
		theme.FieldLabel("CARD TYPE", nil),
		cardTypeSelect,
		theme.FieldLabel("CARD NICKNAME", nil),
		cardNameInput,
		theme.FieldLabel("CARD HOLDER", nil),
		cardHolderInput,
		theme.FieldLabel("CARD NUMBER", nil),
		cardNumberInput,
		container.NewGridWithColumns(2,
			container.NewVBox(theme.FieldLabel("EXPIRY", nil), cardExpiryInput),
			container.NewVBox(theme.FieldLabel("CVV", nil), cardCVVInput),
		),
	)

	totpSection := container.NewVBox(
		theme.FieldLabel("ISSUER", nil),
		totpIssuerInput,
		theme.FieldLabel("ACCOUNT", nil),
		totpAccountInput,
		theme.FieldLabel("SECRET (BASE32)", nil),
		totpSecretInput,
		container.NewGridWithColumns(3,
			container.NewVBox(theme.FieldLabel("ALGORITHM", nil), totpAlgorithmSelect),
			container.NewVBox(theme.FieldLabel("DIGITS", nil), totpDigitsSelect),
			container.NewVBox(theme.FieldLabel("PERIOD (s)", nil), totpPeriodSelect),
		),
	)

	var formContent *fyne.Container
	var typeTabs fyne.CanvasObject
	var buildTypeTabs func()

	refreshSections := func(idx int) {
		activeItemType = idx
		passwordSection.Hide()
		noteSection.Hide()
		cardSection.Hide()
		totpSection.Hide()

		switch idx {
		case 1:
			noteSection.Show()
		case 2:
			cardSection.Show()
		case 3:
			totpSection.Show()
		default:
			passwordSection.Show()
		}

		if formContent != nil {
			formContent.Refresh()
		}

		// Rebuild tabs to update active indicator
		buildTypeTabs()
		if formContent != nil {
			formContent.Objects[0] = typeTabs
			formContent.Refresh()
		}
	}

	buildTypeTabs = func() {
		typeTabs = theme.UnderlineTabs(itemTypes, activeItemType, refreshSections)
	}
	buildTypeTabs()
	refreshSections(0)

	saveBtn := theme.CreatePrimaryButton("Save item", func() {
		itemType := itemTypes[activeItemType]
		service := serviceInput.Text
		username := usernameInput.Text
		secret := passwordInput.Text
		entryType := model.EntryTypePassword
		cardSubtype := ""

		switch itemType {
		case "Cyphered Note":
			entryType = model.EntryTypeNote
			title := noteTitleInput.Text
			content := noteContentInput.Text
			if title == "" || content == "" {
				widgets.ShowAppError(fmt.Errorf("note title and content cannot be empty"), ns.window)
				return
			}
			notePayload, _ := json.Marshal(map[string]string{
				"type":    "note",
				"title":   title,
				"content": content,
			})
			service = "NOTE:" + title
			username = "note"
			secret = string(notePayload)
		case "Card":
			entryType = model.EntryTypeCard
			cardSubtype = cardTypeSelect.Selected
			if cardNameInput.Text == "" || cardHolderInput.Text == "" || cardNumberInput.Text == "" {
				widgets.ShowAppError(fmt.Errorf("card name, holder and number cannot be empty"), ns.window)
				return
			}
			cp := cardPayload{
				Subtype: cardTypeSelect.Selected,
				Holder:  cardHolderInput.Text,
				Number:  cardNumberInput.Text,
				Expiry:  cardExpiryInput.Text,
				CVV:     cardCVVInput.Text,
			}
			cardJSON, _ := json.Marshal(cp)
			service = "CARD:" + cardNameInput.Text
			username = cardTypeSelect.Selected
			secret = string(cardJSON)
		case "TOTP":
			entryType = model.EntryTypeTOTP
			if totpSecretInput.Text == "" || totpIssuerInput.Text == "" {
				widgets.ShowAppError(fmt.Errorf("TOTP issuer and secret are required"), ns.window)
				return
			}
			digits, _ := strconv.Atoi(totpDigitsSelect.Selected)
			period, _ := strconv.Atoi(totpPeriodSelect.Selected)
			params := &totp.TOTPParams{
				Secret:    totpSecretInput.Text,
				Algorithm: totp.Algorithm(totpAlgorithmSelect.Selected),
				Digits:    digits,
				Period:    period,
				Issuer:    totpIssuerInput.Text,
				Account:   totpAccountInput.Text,
			}
			if err := totp.Validate(params); err != nil {
				widgets.ShowAppError(fmt.Errorf("invalid TOTP params: %w", err), ns.window)
				return
			}
			totpJSON, _ := totp.Serialize(params)
			service = "TOTP:" + totpIssuerInput.Text
			username = totpAccountInput.Text
			secret = string(totpJSON)
		default:
			if secret == "" {
				widgets.ShowAppError(fmt.Errorf("password cannot be empty"), ns.window)
				return
			}
			if service == "" {
				widgets.ShowAppError(fmt.Errorf("service name cannot be empty"), ns.window)
				return
			}
		}

		clearInputs := func() {
			serviceInput.SetText("")
			usernameInput.SetText("")
			passwordInput.SetText("")
			noteTitleInput.SetText("")
			noteContentInput.SetText("")
			cardNameInput.SetText("")
			cardHolderInput.SetText("")
			cardNumberInput.SetText("")
			cardExpiryInput.SetText("")
			cardCVVInput.SetText("")
			totpIssuerInput.SetText("")
			totpAccountInput.SetText("")
			totpSecretInput.SetText("")
		}

		// writeEntry encrypts the secret and either appends a new entry or
		// replaces the crypto fields of an existing one (keeping its ID).
		writeEntry := func(replaceID uint64, replaceExisting bool, successMsg string) {
			ns.appState.Mu.Lock()
			defer ns.appState.Mu.Unlock()

			vaultFile := app.GetVaultPath(ns.appState.CurrentVault)
			entries, err := app.ReadVault(vaultFile, ns.appState.MasterPassword)
			if err != nil {
				fyne.Do(func() {
					widgets.ShowAppError(fmt.Errorf("failed to read vault: %w", err), ns.window)
				})
				return
			}

			ct, ss, err := app.Encapsulate(ns.appState.PublicKey)
			if err != nil {
				fyne.Do(func() {
					widgets.ShowAppError(fmt.Errorf("encapsulation failed: %v", err), ns.window)
				})
				return
			}

			nonce, ciphertext, err := app.EncryptAES256GCM(secret, ss)
			if err != nil {
				fyne.Do(func() {
					widgets.ShowAppError(fmt.Errorf("encryption failed: %v", err), ns.window)
				})
				return
			}

			if replaceExisting {
				var target *model.VaultEntry
				for _, e := range entries {
					if e != nil && e.ID == replaceID {
						target = e
						break
					}
				}
				if target == nil {
					fyne.Do(func() {
						widgets.ShowAppError(fmt.Errorf("entry no longer exists; please retry"), ns.window)
					})
					return
				}
				target.KyberCiphertext = ct
				target.Nonce = nonce
				target.Ciphertext = ciphertext
			} else {
				entry := model.NewVaultEntry()
				entry.KyberCiphertext = ct
				entry.Nonce = nonce
				entry.Ciphertext = ciphertext
				entry.Type = entryType
				entry.CardSubtype = cardSubtype
				entry.Service = service
				entry.Username = username
				entries = append(entries, entry)
			}

			if err := app.WriteVault(entries, vaultFile, ns.appState.MasterPassword); err != nil {
				fyne.Do(func() {
					widgets.ShowAppError(fmt.Errorf("failed to save vault item: %v", err), ns.window)
				})
				return
			}

			fyne.Do(func() {
				clearInputs()
				widgets.ShowAppInformation("Success", successMsg, ns.window)
			})
		}

		go func() {
			ns.appState.Mu.Lock()
			vaultFile := app.GetVaultPath(ns.appState.CurrentVault)
			entries, err := app.ReadVault(vaultFile, ns.appState.MasterPassword)
			if err != nil {
				ns.appState.Mu.Unlock()
				fyne.Do(func() {
					widgets.ShowAppError(fmt.Errorf("failed to read vault: %w", err), ns.window)
				})
				return
			}

			dup := app.FindDuplicateEntry(entries, entryType, service, username)
			ns.appState.Mu.Unlock()

			if dup != nil {
				dupID := dup.ID
				displayService := strings.TrimPrefix(service, "TOTP:")
				fyne.Do(func() {
					widgets.ShowAppConfirm(
						"Replace existing entry?",
						fmt.Sprintf("An entry for '%s' / '%s' already exists. Replace it?", displayService, username),
						func(confirmed bool) {
							if !confirmed {
								return
							}
							go writeEntry(dupID, true, "Entry replaced successfully!")
						},
						ns.window,
					)
				})
				return
			}

			writeEntry(0, false, "Item saved successfully!")
		}()
	})

	cancelBtn := theme.CreateGhostButton("Cancel", func() {
		ns.switchView(NavViewItems)
	})

	formContent = container.NewVBox(
		typeTabs,
		passwordSection,
		noteSection,
		cardSection,
		totpSection,
	)

	footer := theme.FormFooter("Encrypted on save: AES-256-GCM", cancelBtn, saveBtn)

	formCard := theme.CardWithHeader("", "", nil, formContent)

	return container.NewVBox(header, formCard, footer)
}
