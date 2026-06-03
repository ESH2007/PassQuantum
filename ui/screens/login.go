package screens

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"passquantum/app"
	"passquantum/theme"
	"passquantum/ui/assets"
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

	actionBtn := theme.CreatePrimaryButton("Create master password", func() {
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

		if appState.FaceGuard != nil && !faceDataExists() {
			ShowTrainingScreen(w, appState.FaceGuard, appState, func() {
				ShowVaultSelection(w, fyneApp, appState)
			})
			return
		}
		ShowVaultSelection(w, fyneApp, appState)
	})

	screen := buildAccessScreen(
		"CREATE MASTER PASSWORD",
		"Set one global password for the app. It is verified against metadata bound to the current private key.",
		warning,
		[]fyne.CanvasObject{
			theme.FieldLabel("MASTER PASSWORD", nil),
			passwordInput,
			theme.FieldLabel("CONFIRM PASSWORD", nil),
			confirmInput,
			container.NewCenter(actionBtn),
		},
	)

	w.SetContent(screen)
}

func showUnlockScreen(w fyne.Window, fyneApp fyne.App, appState *app.AppState) {
	passwordInput := widget.NewPasswordEntry()
	passwordInput.PlaceHolder = "Enter your global master password"

	actionBtn := theme.CreatePrimaryButton("Unlock", func() {
		if passwordInput.Text == "" {
			widgets.ShowAppError(fmt.Errorf("master password cannot be empty"), w)
			return
		}

		if err := app.UnlockVault(appState, passwordInput.Text); err != nil {
			widgets.ShowAppError(err, w)
			return
		}

		if appState.FaceGuard != nil {
			go appState.FaceGuard.SendCommand("START_MONITOR")
		}

		ShowVaultSelection(w, fyneApp, appState)
	})

	screen := buildAccessScreen(
		"UNLOCK PASSQUANTUM",
		"Unlock the app before any vault screen becomes available.",
		"",
		[]fyne.CanvasObject{
			theme.FieldLabel("MASTER PASSWORD", nil),
			passwordInput,
			container.NewCenter(actionBtn),
		},
	)

	w.SetContent(screen)
}

func buildAccessScreen(title string, subtitle string, warning string, fields []fyne.CanvasObject) fyne.CanvasObject {
	var logo fyne.CanvasObject
	if img, _, err := image.Decode(bytes.NewReader(assets.LogoImage)); err == nil {
		logoImg := canvas.NewImageFromImage(img)
		logoImg.FillMode = canvas.ImageFillContain
		logoImg.SetMinSize(fyne.NewSize(96, 96))
		logo = logoImg
	} else {
		atomIco := canvas.NewImageFromResource(theme.IconAtom)
		atomIco.SetMinSize(fyne.NewSize(48, 48))
		iconBg := canvas.NewRectangle(theme.ColorAccentSoft)
		iconBg.CornerRadius = 18
		iconBg.SetMinSize(fyne.NewSize(96, 96))
		logo = container.NewStack(iconBg, container.NewCenter(atomIco))
	}

	titleTxt := canvas.NewText(title, theme.ColorTextPrimary)
	titleTxt.TextSize = 22
	titleTxt.TextStyle = fyne.TextStyle{Bold: true}

	subTxt := canvas.NewText(subtitle, theme.ColorTextSecondary)
	subTxt.TextSize = 13

	divider := canvas.NewRectangle(theme.ColorLine1)
	divider.SetMinSize(fyne.NewSize(0, 1))

	cardObjects := []fyne.CanvasObject{
		container.NewCenter(logo),
		container.NewCenter(titleTxt),
		container.NewCenter(subTxt),
		divider,
	}

	if warning != "" {
		cardObjects = append(cardObjects, theme.WarningBanner("KEY MISMATCH DETECTED", warning))
	}

	cardObjects = append(cardObjects, fields...)

	// Width anchor forces the card to be at least 480px wide without collapsing height.
	widthAnchor := canvas.NewRectangle(theme.ColorBg)
	widthAnchor.SetMinSize(fyne.NewSize(480, 0))
	cardObjects = append([]fyne.CanvasObject{widthAnchor}, cardObjects...)

	card := theme.CardWithHeader("", "", nil, container.NewVBox(cardObjects...))

	bg := canvas.NewRectangle(theme.ColorBg)
	return container.NewStack(bg, container.NewCenter(card))
}

// faceDataExists reports whether the face_guard.py training output file is present
// in the current working directory (same directory as the running binary).
func faceDataExists() bool {
	workDir := os.Getenv("PASSQUANTUM_WORK_DIR")
	if workDir == "" {
		return false
	}
	_, err := os.Stat(filepath.Join(workDir, "face_data.npy"))
	return err == nil
}
