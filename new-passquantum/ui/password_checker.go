package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ShowPasswordChecker displays a screen to check password strength
func ShowPasswordChecker(w fyne.Window, fyneApp fyne.App, appState *AppState) {
	w.SetTitle("Password Checker - " + appState.currentVault)
	w.Resize(fyne.NewSize(650, 450))

	passwordInput := widget.NewPasswordEntry()
	passwordInput.PlaceHolder = "Enter password to check"
	strengthBar := NewStrengthBar()
	BindStrengthBar(strengthBar, passwordInput, func() []string {
		return storedVaultPasswords(appState)
	})

	// Create input container
	inputBg := canvas.NewRectangle(ColorInputBg)
	inputBg.SetMinSize(fyne.NewSize(500, 50))
	inputBg.CornerRadius = BorderRadius
	inputContainer := container.NewMax(inputBg, container.NewPadded(passwordInput))

	backBtn := CreateNeonButton("← BACK", func() {
		ShowMainScreen(w, fyneApp, appState)
	}, 120, 44)

	// Initial content
	content := container.NewVBox(
		CreateLabel("PASSWORD CHECKER", 14, ColorAccentCyn, true),
		CreateDivider(),
		widget.NewLabel(""),
		CreateLabel("Enter Password:", 11, ColorPurple, true),
		inputContainer,
		widget.NewLabel(""),
		strengthBar,
		widget.NewLabel(""),
		widget.NewLabel(""),
		container.NewCenter(backBtn),
	)

	scrollBox := container.NewVScroll(content)
	scrollBox.SetMinSize(fyne.NewSize(700, 600))

	bgContainer := CreateBackgroundContainer(container.NewPadded(scrollBox))
	w.SetContent(bgContainer)
}
