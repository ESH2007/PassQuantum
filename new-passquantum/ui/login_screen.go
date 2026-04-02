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

		// When biometric is enabled and an enrolled template exists, verify the
		// user's face before granting access to the vault selection screen.
		appState.mu.Lock()
		biometricEnabled := appState.biometricEnabled
		hasTemplate := len(appState.biometricTemplate) > 0
		appState.mu.Unlock()

		if biometricEnabled && hasTemplate {
			showFaceVerificationGate(w, fyneApp, appState)
			return
		}

		// No biometric required — proceed directly.
		startContinuousCheck(appState, w, fyneApp)
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

// showFaceVerificationGate displays a brief "Scanning face…" screen and runs the
// one-shot biometric check. On success it starts continuous verification and
// navigates to vault selection. On failure it shows an error and resets to the
// unlock screen so the user can try again.
func showFaceVerificationGate(w fyne.Window, fyneApp fyne.App, appState *AppState) {
	statusLabel := CreateLabel("SCANNING FACE...", 14, ColorAccentCyan, true)
	subLabel := CreateLabel("Please look at the camera.", 11, ColorTextSec, false)

	screen := buildAccessScreen(
		"FACE VERIFICATION",
		"Verifying identity using the enrolled face template.",
		"",
		[]fyne.CanvasObject{
			widget.NewLabel(""),
			container.NewCenter(statusLabel),
			container.NewCenter(subLabel),
			widget.NewLabel(""),
		},
	)
	w.SetContent(screen)
	w.Resize(fyne.NewSize(800, 600))

	// Run face verification off the UI goroutine to avoid blocking Fyne.
	go func() {
		matched, err := captureAndVerifyFace(appState)
		fyne.Do(func() {
			if err != nil {
				appState.clearSensitiveState()
				ShowAppError(fmt.Errorf("face verification error: %w", err), w)
				showUnlockScreen(w, fyneApp, appState)
				return
			}
			if !matched {
				appState.clearSensitiveState()
				ShowAppWarning("Face Not Recognised",
					"The captured face does not match the enrolled template. Please unlock again.", w)
				showUnlockScreen(w, fyneApp, appState)
				return
			}
			// Face verified — begin continuous check and proceed.
			startContinuousCheck(appState, w, fyneApp)
			ShowVaultSelection(w, fyneApp, appState)
		})
	}()
}
