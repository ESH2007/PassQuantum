package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"passquantum/core/model"
)

// Current active view in the navigation
type NavView int

const (
	NavViewVaults NavView = iota
	NavViewPasswords
	NavViewGenerator
	NavViewChecker
	NavViewSettings
)

// NavigationState tracks the current view
type NavigationState struct {
	currentView      NavView
	window           fyne.Window
	app              fyne.App
	appState         *AppState
	contentContainer *fyne.Container
	sidebarContainer *fyne.Container
}

func ShowMainScreen(w fyne.Window, fyneApp fyne.App, appState *AppState) {
	w.SetTitle("PassQuantum - " + appState.currentVault)
	w.Resize(fyne.NewSize(1100, 700))

	// Create navigation state
	navState := &NavigationState{
		currentView: NavViewPasswords,
		window:      w,
		app:         fyneApp,
		appState:    appState,
	}

	// Create content container that will be dynamically updated
	navState.contentContainer = container.NewMax()
	navState.sidebarContainer = container.NewMax()

	// Build initial UI
	navState.rebuildUI()

	bgContainer := CreateBackgroundContainer(navState.sidebarContainer)
	w.SetContent(bgContainer)
}

// rebuildUI rebuilds the entire UI with current state
func (ns *NavigationState) rebuildUI() {
	// Create navigation items with current active state
	navItems := []NavigationItem{
		{
			Icon:     "📦",
			Label:    "Vaults",
			OnClick:  func() { ns.switchView(NavViewVaults) },
			IsActive: ns.currentView == NavViewVaults,
		},
		{
			Icon:     "🔑",
			Label:    "Passwords",
			OnClick:  func() { ns.switchView(NavViewPasswords) },
			IsActive: ns.currentView == NavViewPasswords,
		},
		{
			Icon:     "🔐",
			Label:    "Generate",
			OnClick:  func() { ns.switchView(NavViewGenerator) },
			IsActive: ns.currentView == NavViewGenerator,
		},
		{
			Icon:     "🔍",
			Label:    "Check Password",
			OnClick:  func() { ns.switchView(NavViewChecker) },
			IsActive: ns.currentView == NavViewChecker,
		},
		{
			Icon:     "⚙️",
			Label:    "Settings",
			OnClick:  func() { ns.switchView(NavViewSettings) },
			IsActive: ns.currentView == NavViewSettings,
		},
	}

	lockItem := NavigationItem{
		Icon:  "🔒",
		Label: "Lock & Exit",
		OnClick: func() {
			ns.appState.clearSensitiveState()
			ns.app.Quit()
		},
		IsActive: false,
	}

	// Create sidebar
	sidebar := CreateNavigationSidebar(navItems, lockItem, 220)

	// Update content based on current view
	ns.updateContent()

	// Create split layout: sidebar on left, content on right
	mainLayout := container.NewBorder(nil, nil, sidebar, nil, ns.contentContainer)

	// Update the main container
	ns.sidebarContainer.Objects = []fyne.CanvasObject{mainLayout}
	ns.sidebarContainer.Refresh()
}

// switchView changes the active view and updates UI
func (ns *NavigationState) switchView(view NavView) {
	ns.currentView = view
	ns.rebuildUI()
}

// updateContent updates the content area based on current view
func (ns *NavigationState) updateContent() {
	var content fyne.CanvasObject

	switch ns.currentView {
	case NavViewVaults:
		content = ns.createVaultsView()
	case NavViewPasswords:
		content = ns.createPasswordsView()
	case NavViewGenerator:
		content = ns.createGeneratorView()
	case NavViewChecker:
		content = ns.createCheckerView()
	case NavViewSettings:
		content = ns.createSettingsView()
	default:
		content = ns.createPasswordsView()
	}

	ns.contentContainer.Objects = []fyne.CanvasObject{content}
	ns.contentContainer.Refresh()
}

// createPasswordsView creates the add password form view
func (ns *NavigationState) createPasswordsView() fyne.CanvasObject {
	passwordInput := widget.NewPasswordEntry()
	passwordInput.PlaceHolder = "Enter password"

	serviceInput := widget.NewEntry()
	serviceInput.PlaceHolder = "Service name (e.g., Gmail, GitHub)"

	usernameInput := widget.NewEntry()
	usernameInput.PlaceHolder = "Username or email"

	addBtn := CreateNeonButton("➕ SAVE PASSWORD", func() {
		service := serviceInput.Text
		username := usernameInput.Text
		password := passwordInput.Text

		if password == "" {
			ShowAppError(fmt.Errorf("password cannot be empty"), ns.window)
			return
		}

		if service == "" {
			ShowAppError(fmt.Errorf("service name cannot be empty"), ns.window)
			return
		}

		go func() {
			ns.appState.mu.Lock()
			defer ns.appState.mu.Unlock()

			vaultFile := GetVaultPath(ns.appState.currentVault)
			entries, err := ReadVault(vaultFile, ns.appState.encryptionKey, ns.appState.verificationKey)
			if err != nil {
				fyne.Do(func() {
					ShowAppError(fmt.Errorf("failed to read vault: %w", err), ns.window)
				})
				return
			}

			ct, ss, err := Encapsulate(ns.appState.publicKey)
			if err != nil {
				fyne.Do(func() {
					ShowAppError(fmt.Errorf("encapsulation failed: %v", err), ns.window)
				})
				return
			}

			nonce, ciphertext, err := EncryptAES256GCM(password, ss)
			if err != nil {
				fyne.Do(func() {
					ShowAppError(fmt.Errorf("encryption failed: %v", err), ns.window)
				})
				return
			}

			entry := model.NewPasswordEntry()
			entry.KyberCiphertext = ct
			entry.Nonce = nonce
			entry.Ciphertext = ciphertext
			entry.Service = service
			entry.Username = username

			entries = append(entries, entry)

			err = WriteVault(entries, vaultFile, ns.appState.encryptionKey, ns.appState.verificationKey, ns.appState.kdfParams)
			if err != nil {
				fyne.Do(func() {
					ShowAppError(fmt.Errorf("failed to save password: %v", err), ns.window)
				})
				return
			}

			fyne.Do(func() {
				serviceInput.SetText("")
				usernameInput.SetText("")
				passwordInput.SetText("")
				ShowAppInformation("Success", "✓ Password saved successfully!", ns.window)
			})
		}()
	}, 220, 48)

	viewBtn := CreateNeonButton("📋 VIEW ALL", func() {
		ShowPasswordsView(ns.window, ns.app, ns.appState)
	}, 150, 48)

	// Enhanced header
	headerText := CreateHeaderText("ADD NEW PASSWORD", 18)
	headerSection := container.NewVBox(headerText, CreateGlowingDivider())

	// Enhanced styled inputs
	serviceLabel := CreateLabel("SERVICE NAME", 11, ColorPurple, true)
	usernameLabel := CreateLabel("USERNAME / EMAIL", 11, ColorPurple, true)
	passwordLabel := CreateLabel("PASSWORD", 11, ColorPurple, true)

	styledServiceInput := CreateStyledInput(serviceInput, 650, 42)
	styledUsernameInput := CreateStyledInput(usernameInput, 650, 42)
	styledPasswordInput := CreateStyledInput(passwordInput, 650, 42)

	formContent := container.NewVBox(
		headerSection,
		widget.NewLabel(""),
		serviceLabel,
		container.NewCenter(styledServiceInput),
		widget.NewLabel(""),
		usernameLabel,
		container.NewCenter(styledUsernameInput),
		widget.NewLabel(""),
		passwordLabel,
		container.NewCenter(styledPasswordInput),
		widget.NewLabel(""),
		widget.NewLabel(""),
		container.NewCenter(addBtn),
		widget.NewLabel(""),
		container.NewCenter(viewBtn),
	)

	formCard := CreateEnhancedCard(formContent, 750, 520)

	vaultHeaderText := CreateLabel("📦 VAULT: "+ns.appState.currentVault, 13, ColorAccentCyan, true)

	mainContent := container.NewVBox(
		container.NewCenter(vaultHeaderText),
		widget.NewLabel(""),
		container.NewCenter(formCard),
	)

	return container.NewPadded(container.NewVScroll(mainContent))
}

// createVaultsView creates the vault selection view
func (ns *NavigationState) createVaultsView() fyne.CanvasObject {
	vaults := ListVaults()

	headerText := CreateLabel("YOUR VAULTS", 18, ColorAccentCyan, true)
	headerSection := container.NewVBox(headerText, CreateDivider())

	var vaultItems []fyne.CanvasObject
	if len(vaults) == 0 {
		emptyMsg := CreateLabel("No vaults found. Create one to get started.", 12, ColorTextSecondary, false)
		vaultItems = append(vaultItems, container.NewCenter(emptyMsg))
	} else {
		for _, vaultName := range vaults {
			vaultCard := createVaultCard(ns.window, ns.app, ns.appState, vaultName)
			vaultItems = append(vaultItems, vaultCard)
			vaultItems = append(vaultItems, widget.NewLabel(""))
		}
	}

	newVaultBtn := CreateNeonButton("+ CREATE VAULT", func() {
		showCreateVaultDialog(ns.window, ns.app, ns.appState)
	}, 200, 44)

	/*
		generatePasswordBtn := CreateNeonButton("🔐 GENERATE PASSWORD", func() {
			ShowPasswordGeneratorNoVault(ns.window, ns.app, ns.appState)
		}, 220, 44)
	*/

	scrollContent := container.NewVBox(vaultItems...)
	scrollBox := container.NewVScroll(scrollContent)
	scrollBox.SetMinSize(fyne.NewSize(750, 420))

	buttonContainer := container.NewHBox(newVaultBtn) //generatePasswordBtn

	mainContent := container.NewVBox(
		headerSection,
		widget.NewLabel(""),
		scrollBox,
		widget.NewLabel(""),
		container.NewCenter(buttonContainer),
	)

	return container.NewPadded(mainContent)
}

// createGeneratorView creates the password generator view
func (ns *NavigationState) createGeneratorView() fyne.CanvasObject {
	settings := DefaultPasswordGeneratorSettings()
	generatedPasswordDisplay := widget.NewEntry()
	generatedPasswordDisplay.PlaceHolder = "Generated password will appear here"
	generatedPasswordDisplay.MultiLine = false

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

	generateBtn := CreateNeonButton("🔄 GENERATE", func() {
		password, err := GeneratePassword(settings)
		if err != nil {
			ShowAppError(err, ns.window)
			return
		}
		generatedPasswordDisplay.SetText(password)
	}, 160, 44)

	copyGeneratedBtn := CreateNeonButton("COPY", func() {
		if generatedPasswordDisplay.Text != "" {
			ns.window.Clipboard().SetContent(generatedPasswordDisplay.Text)
			ShowAppInformation("Copied", "Password copied to clipboard!", ns.window)
		} else {
			ShowAppInformation("Empty", "Generate a password first!", ns.window)
		}
	}, 100, 44)

	saveToVaultBtn := CreateNeonButton("💾 SAVE TO VAULT", func() {
		password := generatedPasswordDisplay.Text
		if password == "" {
			ShowAppInformation("Empty", "Generate a password first!", ns.window)
			return
		}
		showSaveGeneratedPasswordDialog(ns.window, ns.app, ns.appState, password)
	}, 180, 44)

	headerText := CreateHeaderText("PASSWORD GENERATOR", 18)
	headerSection := container.NewVBox(headerText, CreateGlowingDivider())

	passwordDisplayLabel := CreateLabel("GENERATED PASSWORD", 11, ColorPurple, true)
	passwordDisplayBox := CreateStyledInput(generatedPasswordDisplay, 650, 42)

	lengthLabel := CreateLabel("PASSWORD LENGTH", 11, ColorPurple, true)
	styledLengthInput := CreateStyledInput(lengthInput, 150, 42)

	optionsLabel := CreateLabel("OPTIONS", 11, ColorPurple, true)

	formContent := container.NewVBox(
		headerSection,
		widget.NewLabel(""),
		lengthLabel,
		container.NewCenter(styledLengthInput),
		widget.NewLabel(""),
		optionsLabel,
		uppercaseCheck,
		lowercaseCheck,
		numbersCheck,
		specialCharsCheck,
		ambiguousCheck,
		widget.NewLabel(""),
		container.NewCenter(generateBtn),
		widget.NewLabel(""),
		passwordDisplayLabel,
		container.NewCenter(passwordDisplayBox),
		widget.NewLabel(""),
		container.NewCenter(container.NewHBox(copyGeneratedBtn, saveToVaultBtn)),
	)

	formCard := CreateEnhancedCard(formContent, 750, 600)

	mainContent := container.NewVBox(
		widget.NewLabel(""),
		container.NewCenter(formCard),
	)

	return container.NewPadded(container.NewVScroll(mainContent))
}

// createCheckerView creates the password checker view
func (ns *NavigationState) createCheckerView() fyne.CanvasObject {
	passwordInput := widget.NewEntry()
	passwordInput.PlaceHolder = "Enter password to check"

	// Results container - will be updated
	resultsContainer := container.NewVBox(
		widget.NewLabel(""),
		CreateLabel("Results will appear here", 11, ColorTextSecondary, false),
		widget.NewLabel(""),
	)

	checkBtn := CreateNeonButton("CHECK PASSWORD", func() {
		password := passwordInput.Text

		if password == "" {
			return
		}

		// Perform validation
		vaultFile := GetVaultPath(ns.appState.currentVault)
		validation := ValidatePassword(password, vaultFile, ns.appState.encryptionKey, ns.appState.verificationKey, ns.appState.privateKey)

		// Check length
		hasLength := len(password) > 8
		lengthCheck := CreateCheckItem("Length > 8 characters", hasLength)

		// Check special characters
		specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
		hasSpecial := false
		for _, char := range password {
			for _, special := range specialChars {
				if char == special {
					hasSpecial = true
					break
				}
			}
			if hasSpecial {
				break
			}
		}
		specialCheck := CreateCheckItem("Contains special characters", hasSpecial)

		// Check duplicates
		isDuplicate := !validation.Valid && len(validation.ErrorMessage) > 0 && len(validation.ErrorMessage) >= 4 && validation.ErrorMessage[:4] == "This"
		duplicateCheck := CreateCheckItem("Not duplicate in vault", !isDuplicate)

		// Create status message
		var statusMsg string
		var statusColor = ColorSuccess

		if validation.Valid {
			statusMsg = "✓ Password is STRONG"
			statusColor = ColorSuccess
		} else {
			statusMsg = "✕ " + validation.ErrorMessage
			statusColor = ColorDanger
		}

		statusLabel := CreateLabel(statusMsg, 12, statusColor, true)

		// Update results container
		resultsContainer.Objects = []fyne.CanvasObject{
			CreateLabel("PASSWORD STRENGTH", 14, ColorAccentCyan, true),
			CreateDivider(),
			widget.NewLabel(""),
			lengthCheck,
			specialCheck,
			duplicateCheck,
			widget.NewLabel(""),
			container.NewCenter(statusLabel),
		}
		resultsContainer.Refresh()

	}, 200, 44)

	headerText := CreateHeaderText("PASSWORD CHECKER", 18)
	headerSection := container.NewVBox(headerText, CreateGlowingDivider())

	passwordLabel := CreateLabel("ENTER PASSWORD", 11, ColorPurple, true)
	styledPasswordInput := CreateStyledInput(passwordInput, 650, 42)

	formContent := container.NewVBox(
		headerSection,
		widget.NewLabel(""),
		passwordLabel,
		container.NewCenter(styledPasswordInput),
		widget.NewLabel(""),
		container.NewCenter(checkBtn),
		widget.NewLabel(""),
		resultsContainer,
	)

	formCard := CreateEnhancedCard(formContent, 750, 500)

	mainContent := container.NewVBox(
		widget.NewLabel(""),
		container.NewCenter(formCard),
	)

	return container.NewPadded(container.NewVScroll(mainContent))
}

// createSettingsView creates the settings view
func (ns *NavigationState) createSettingsView() fyne.CanvasObject {
	return buildCustomSettingsView(ns.window, ns.app, ns.appState)
}

// showSaveGeneratedPasswordDialog shows a dialog to save a generated password to a vault
func showSaveGeneratedPasswordDialog(w fyne.Window, fyneApp fyne.App, appState *AppState, password string) {
	// Get available vaults
	vaults := ListVaults()
	if len(vaults) == 0 {
		ShowAppError(fmt.Errorf("no vaults available. Create a vault first"), w)
		return
	}

	// Create input fields
	vaultSelect := widget.NewSelect(vaults, nil)
	vaultSelect.PlaceHolder = "Select a vault"
	if appState.currentVault != "" {
		vaultSelect.SetSelected(appState.currentVault)
	} else {
		vaultSelect.SetSelected(vaults[0])
	}

	serviceInput := widget.NewEntry()
	serviceInput.PlaceHolder = "Service name"

	usernameInput := widget.NewEntry()
	usernameInput.PlaceHolder = "Username or email"

	// Create styled labels
	vaultLabel := CreateLabel("Vault", 11, ColorAccentCyan, false)
	serviceLabel := CreateLabel("Service", 11, ColorAccentCyan, false)
	usernameLabel := CreateLabel("Username", 11, ColorAccentCyan, false)

	// Create styled containers for inputs
	vaultSelectBg := canvas.NewRectangle(ColorInputBg)
	vaultSelectBg.SetMinSize(fyne.NewSize(350, 40))
	vaultSelectBg.CornerRadius = BorderRadius
	styledVaultSelect := container.NewStack(vaultSelectBg, container.NewPadded(vaultSelect))

	serviceInputBg := canvas.NewRectangle(ColorInputBg)
	serviceInputBg.SetMinSize(fyne.NewSize(350, 40))
	serviceInputBg.CornerRadius = BorderRadius
	styledServiceInput := container.NewStack(serviceInputBg, container.NewPadded(serviceInput))

	usernameInputBg := canvas.NewRectangle(ColorInputBg)
	usernameInputBg.SetMinSize(fyne.NewSize(350, 40))
	usernameInputBg.CornerRadius = BorderRadius
	styledUsernameInput := container.NewStack(usernameInputBg, container.NewPadded(usernameInput))

	// Dialog content
	headerLabel := CreateLabel("ADD GENERATED PASSWORD", 13, ColorAccentCyan, true)

	formContent := container.NewVBox(
		headerLabel,
		CreateDivider(),
		widget.NewLabel(""),
		vaultLabel,
		styledVaultSelect,
		widget.NewLabel(""),
		serviceLabel,
		styledServiceInput,
		widget.NewLabel(""),
		usernameLabel,
		styledUsernameInput,
	)

	var customDialog *dialog.CustomDialog

	// Save button
	saveBtn := CreateNeonButton("✓ SAVE", func() {
		selectedVault := vaultSelect.Selected
		service := serviceInput.Text
		username := usernameInput.Text

		if selectedVault == "" {
			ShowAppError(fmt.Errorf("please select a vault"), w)
			return
		}

		if service == "" {
			ShowAppError(fmt.Errorf("service name cannot be empty"), w)
			return
		}

		// Close the form dialog first
		if customDialog != nil {
			customDialog.Hide()
		}

		// Now prompt for vault password and save
		OpenVault(w, fyneApp, appState, selectedVault, func() {
			// Vault is now unlocked, proceed with saving
			go func() {
				appState.mu.Lock()
				defer appState.mu.Unlock()

				vaultFile := GetVaultPath(selectedVault)
				entries, err := ReadVault(vaultFile, appState.encryptionKey, appState.verificationKey)
				if err != nil {
					fyne.Do(func() {
						ShowAppError(fmt.Errorf("failed to read vault: %w", err), w)
					})
					return
				}

				ct, ss, err := Encapsulate(appState.publicKey)
				if err != nil {
					fyne.Do(func() {
						ShowAppError(fmt.Errorf("encapsulation failed: %v", err), w)
					})
					return
				}

				nonce, ciphertext, err := EncryptAES256GCM(password, ss)
				if err != nil {
					fyne.Do(func() {
						ShowAppError(fmt.Errorf("encryption failed: %v", err), w)
					})
					return
				}

				entry := model.NewPasswordEntry()
				entry.KyberCiphertext = ct
				entry.Nonce = nonce
				entry.Ciphertext = ciphertext
				entry.Service = service
				entry.Username = username

				entries = append(entries, entry)

				err = WriteVault(entries, vaultFile, appState.encryptionKey, appState.verificationKey, appState.kdfParams)
				if err != nil {
					fyne.Do(func() {
						ShowAppError(fmt.Errorf("failed to save password: %v", err), w)
					})
					return
				}

				fyne.Do(func() {
					ShowAppInformation("Success", "✓ Password saved to vault successfully!", w)
				})
			}()
		})
	}, 120, 44)

	cancelBtn := CreateNeonButton("✕ CANCEL", func() {
		if customDialog != nil {
			customDialog.Hide()
		}
	}, 120, 44)

	buttonBox := container.NewHBox(cancelBtn, saveBtn)

	dialogContent := container.NewVBox(formContent, widget.NewLabel(""), container.NewCenter(buttonBox))
	customDialog = dialog.NewCustom("Add Generated Password", "Close", dialogContent, w)
	customDialog.Show()
}
