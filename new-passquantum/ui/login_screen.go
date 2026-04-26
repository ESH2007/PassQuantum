package main

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func PromptMasterPassword(w fyne.Window, fyneApp fyne.App, appState *AppState) {
	accessState, err := ResolveStartupAccessState(appState)
	if err != nil {
		ShowAppError(fmt.Errorf("failed to load startup access state: %w", err), w)
		return
	}

	if accessState.RequiresSetup {
		showCreateMasterPasswordScreen(w, fyneApp, appState, accessState.Warning)
		return
	}

	showUnlockScreen(w, fyneApp, appState)
}

func showCreateMasterPasswordScreen(w fyne.Window, fyneApp fyne.App, appState *AppState, warning string) {
	passwordInput := widget.NewPasswordEntry()
	passwordInput.PlaceHolder = "Create your global master password"

	confirmInput := widget.NewPasswordEntry()
	confirmInput.PlaceHolder = "Confirm the master password"

	actionBtn := CreateNeonButton("[  CREATE MASTER PASSWORD  ]", func() {
		if passwordInput.Text == "" {
			ShowAppError(fmt.Errorf("master password cannot be empty"), w)
			return
		}

		if passwordInput.Text != confirmInput.Text {
			ShowAppError(fmt.Errorf("passwords do not match"), w)
			return
		}

		if err := CreateMasterPasswordProfile(appState, passwordInput.Text); err != nil {
			ShowAppError(fmt.Errorf("failed to create master password profile: %w", err), w)
			return
		}

		if len(ListVaults()) == 0 {
			if !CreateNewVault(w, appState, "Default") {
				return
			}
		}

		// First-time setup: if the face guard is active and no face data exists
		// yet, show the training screen.  onComplete proceeds to vault selection.
		if appState.faceGuard != nil && !faceDataExists() {
			ShowTrainingScreen(w, appState.faceGuard, func() {
				ShowVaultSelection(w, fyneApp, appState)
			})
			return
		}
		ShowVaultSelection(w, fyneApp, appState)
	}, 320, 50)

	screen := buildAccessScreen(
		"CREATE MASTER PASSWORD",
		"Set one global password for the app. It is verified against metadata bound to the current private key.",
		warning,
		[]fyne.CanvasObject{
			CreateLabel("[ MASTER PASSWORD ]", 12, ColorPurple, true),
			container.NewCenter(CreateStyledInput(passwordInput, 450, 45)),
			widget.NewLabel(""),
			CreateLabel("[ CONFIRM PASSWORD ]", 12, ColorPurple, true),
			container.NewCenter(CreateStyledInput(confirmInput, 450, 45)),
			widget.NewLabel(""),
			container.NewCenter(actionBtn),
		},
	)

	w.SetContent(screen)
	w.Resize(fyne.NewSize(820, 640))
}

func showUnlockScreen(w fyne.Window, fyneApp fyne.App, appState *AppState) {
	passwordInput := widget.NewPasswordEntry()
	passwordInput.PlaceHolder = "Enter your global master password"

	actionBtn := CreateNeonButton("[  UNLOCK  ]", func() {
		if passwordInput.Text == "" {
			ShowAppError(fmt.Errorf("master password cannot be empty"), w)
			return
		}

		if !UnlockVault(w, appState, passwordInput.Text) {
			return
		}
		ShowVaultSelection(w, fyneApp, appState)
	}, 280, 50)

	screen := buildAccessScreen(
		"UNLOCK PASSQUANTUM",
		"Unlock the app before any vault screen becomes available.",
		"",
		[]fyne.CanvasObject{
			CreateLabel("[ MASTER PASSWORD ]", 12, ColorPurple, true),
			container.NewCenter(CreateStyledInput(passwordInput, 450, 45)),
			widget.NewLabel(""),
			container.NewCenter(actionBtn),
		},
	)

	w.SetContent(screen)
	w.Resize(fyne.NewSize(800, 600))
}

func buildAccessScreen(title string, subtitle string, warning string, fields []fyne.CanvasObject) fyne.CanvasObject {
	var logo fyne.CanvasObject
	possible := []string{"PM.png", "../PM.png", "ui/PM.png", "new-passquantum/PM.png", "./PM.png"}
	for _, path := range possible {
		if _, err := os.Stat(path); err == nil {
			image := canvas.NewImageFromFile(path)
			image.FillMode = canvas.ImageFillContain
			image.SetMinSize(fyne.NewSize(120, 120))
			logo = image
			break
		}
	}

	if logo == nil {
		logo = CreateHeaderText("PassQuantum", 36)
	}

	noticeText := CreateLabel(subtitle, 12, ColorTextSecondary, false)
	cardObjects := []fyne.CanvasObject{
		widget.NewLabel(""),
		container.NewCenter(logo),
		widget.NewLabel(""),
		container.NewCenter(CreateHeaderText(title, 26)),
		container.NewCenter(noticeText),
		widget.NewLabel(""),
		CreateGlowingDivider(),
		widget.NewLabel(""),
	}

	if warning != "" {
		warningBox := CreateCardWithBorderColor(
			container.NewVBox(
				CreateLabel("KEY MISMATCH DETECTED", 11, ColorWarning, true),
				widget.NewLabel(""),
				CreateLabel(warning, 10, ColorTextPrimary, false),
			),
			480,
			120,
			ColorWarning,
		)
		cardObjects = append(cardObjects, container.NewCenter(warningBox), widget.NewLabel(""))
	}

	cardObjects = append(cardObjects, fields...)

	card := CreateEnhancedCard(container.NewVBox(cardObjects...), 560, 0)
	return CreateBackgroundContainer(container.NewCenter(card))
}

// faceDataExists reports whether the face_guard.py training output file is present
// in the current working directory (same directory as the running binary).
func faceDataExists() bool {
	_, err := os.Stat("face_data.pkl")
	return err == nil
}
