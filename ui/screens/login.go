package screens

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"passquantum/app"
	"passquantum/theme"
	"passquantum/ui/widgets"
)

func PromptMasterPassword(w fyne.Window, fyneApp fyne.App, appState *app.AppState) {
	accessState, err := app.ResolveStartupAccessState(appState)
	if err != nil {
		widgets.ShowAppError(fmt.Errorf("failed to load startup access state: %w", err), w)
		return
	}

	if accessState.RequiresSetup {
		showCreateMasterPasswordScreen(w, fyneApp, appState, accessState.Warning)
		return
	}

	showUnlockScreen(w, fyneApp, appState)
}

func showCreateMasterPasswordScreen(w fyne.Window, fyneApp fyne.App, appState *app.AppState, warning string) {
	passwordInput := widget.NewPasswordEntry()
	passwordInput.PlaceHolder = "Create your global master password"

	confirmInput := widget.NewPasswordEntry()
	confirmInput.PlaceHolder = "Confirm the master password"

	actionBtn := theme.CreateNeonButton("[  CREATE MASTER PASSWORD  ]", func() {
		if passwordInput.Text == "" {
			widgets.ShowAppError(fmt.Errorf("master password cannot be empty"), w)
			return
		}

		if passwordInput.Text != confirmInput.Text {
			widgets.ShowAppError(fmt.Errorf("passwords do not match"), w)
			return
		}

		if err := app.CreateMasterPasswordProfile(appState, passwordInput.Text); err != nil {
			widgets.ShowAppError(fmt.Errorf("failed to create master password profile: %w", err), w)
			return
		}

		if len(app.ListVaults()) == 0 {
			if err := app.CreateNewVault(appState, "Default"); err != nil {
				widgets.ShowAppError(err, w)
				return
			}
		}

		// First-time setup: if the face guard is active and no face data exists
		// yet, show the training screen.  onComplete proceeds to vault selection.
		if appState.FaceGuard != nil && !faceDataExists() {
			ShowTrainingScreen(w, appState.FaceGuard, appState, func() {
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
			theme.CreateLabel("[ MASTER PASSWORD ]", 12, theme.ColorPurple, true),
			container.NewCenter(theme.CreateStyledInput(passwordInput, 450, 45)),
			widget.NewLabel(""),
			theme.CreateLabel("[ CONFIRM PASSWORD ]", 12, theme.ColorPurple, true),
			container.NewCenter(theme.CreateStyledInput(confirmInput, 450, 45)),
			widget.NewLabel(""),
			container.NewCenter(actionBtn),
		},
	)

	w.SetContent(screen)
	w.Resize(fyne.NewSize(820, 640))
}

func showUnlockScreen(w fyne.Window, fyneApp fyne.App, appState *app.AppState) {
	passwordInput := widget.NewPasswordEntry()
	passwordInput.PlaceHolder = "Enter your global master password"

	actionBtn := theme.CreateNeonButton("[  UNLOCK  ]", func() {
		if passwordInput.Text == "" {
			widgets.ShowAppError(fmt.Errorf("master password cannot be empty"), w)
			return
		}

		if err := app.UnlockVault(appState, passwordInput.Text); err != nil {
			widgets.ShowAppError(err, w)
			return
		}

		// Start continuous face monitoring as soon as the app is unlocked.
		// This covers the vault-selection screen and any subsequent screen.
		if appState.FaceGuard != nil {
			go appState.FaceGuard.SendCommand("START_MONITOR")
		}

		ShowVaultSelection(w, fyneApp, appState)
	}, 280, 50)

	screen := buildAccessScreen(
		"UNLOCK PASSQUANTUM",
		"Unlock the app before any vault screen becomes available.",
		"",
		[]fyne.CanvasObject{
			theme.CreateLabel("[ MASTER PASSWORD ]", 12, theme.ColorPurple, true),
			container.NewCenter(theme.CreateStyledInput(passwordInput, 450, 45)),
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
		logo = theme.CreateHeaderText("PassQuantum", 36)
	}

	noticeText := theme.CreateLabel(subtitle, 12, theme.ColorTextSecondary, false)
	cardObjects := []fyne.CanvasObject{
		widget.NewLabel(""),
		container.NewCenter(logo),
		widget.NewLabel(""),
		container.NewCenter(theme.CreateHeaderText(title, 26)),
		container.NewCenter(noticeText),
		widget.NewLabel(""),
		theme.CreateGlowingDivider(),
		widget.NewLabel(""),
	}

	if warning != "" {
		warningBox := theme.CreateCardWithBorderColor(
			container.NewVBox(
				theme.CreateLabel("KEY MISMATCH DETECTED", 11, theme.ColorWarning, true),
				widget.NewLabel(""),
				theme.CreateLabel(warning, 10, theme.ColorTextPrimary, false),
			),
			480,
			120,
			theme.ColorWarning,
		)
		cardObjects = append(cardObjects, container.NewCenter(warningBox), widget.NewLabel(""))
	}

	cardObjects = append(cardObjects, fields...)

	card := theme.CreateEnhancedCard(container.NewVBox(cardObjects...), 560, 0)
	return theme.CreateBackgroundContainer(container.NewCenter(card))
}

// faceDataExists reports whether the face_guard.py training output file is present
// in the current working directory (same directory as the running binary).
func faceDataExists() bool {
	_, err := os.Stat("face_data.pkl")
	return err == nil
}
