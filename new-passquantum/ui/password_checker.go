package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ShowPasswordChecker displays a screen to check password strength
func ShowPasswordChecker(w fyne.Window, fyneApp fyne.App, appState *AppState) {
	w.SetTitle("Password Checker - " + appState.currentVault)
	w.Resize(fyne.NewSize(650, 450))

	passwordInput := widget.NewEntry()
	passwordInput.PlaceHolder = "Enter password to check"

	// Create input container
	inputBg := canvas.NewRectangle(ColorInputBg)
	inputBg.SetMinSize(fyne.NewSize(500, 50))
	inputBg.CornerRadius = BorderRadius
	inputContainer := container.NewMax(inputBg, container.NewPadded(passwordInput))

	// Results container - will be updated
	var resultsContainer *fyne.Container

	checkBtn := CreateNeonButton("CHECK PASSWORD", func() {
		password := passwordInput.Text

		if password == "" {
			return
		}

		// Perform validation
		vaultFile := GetVaultPath(appState.currentVault)
		validation := ValidatePassword(password, vaultFile, appState.encryptionKey, appState.verificationKey, appState.privateKey)

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
		var statusColor color.Color

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
			CreateLabel("PASSWORD STRENGTH", 14, ColorAccentCyn, true),
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

	backBtn := CreateNeonButton("← BACK", func() {
		ShowMainScreen(w, fyneApp, appState)
	}, 120, 44)

	// Initial results placeholder
	resultsContainer = container.NewVBox(
		widget.NewLabel(""),
		CreateLabel("Results will appear here", 11, ColorTextSecondary, false),
		widget.NewLabel(""),
	)

	// Initial content
	content := container.NewVBox(
		CreateLabel("PASSWORD CHECKER", 14, ColorAccentCyn, true),
		CreateDivider(),
		widget.NewLabel(""),
		CreateLabel("Enter Password:", 11, ColorPurple, true),
		inputContainer,
		widget.NewLabel(""),
		container.NewCenter(checkBtn),
		widget.NewLabel(""),
		resultsContainer,
		widget.NewLabel(""),
		widget.NewLabel(""),
		container.NewCenter(backBtn),
	)

	scrollBox := container.NewVScroll(content)
	scrollBox.SetMinSize(fyne.NewSize(700, 600))

	bgContainer := CreateBackgroundContainer(container.NewPadded(scrollBox))
	w.SetContent(bgContainer)
}
