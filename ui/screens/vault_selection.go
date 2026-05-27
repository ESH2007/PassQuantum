package screens

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"passquantum/app"
	"passquantum/theme"
	"passquantum/ui/widgets"
)

// ShowVaultSelection displays the vault management and selection screen
func ShowVaultSelection(w fyne.Window, fyneApp fyne.App, appState *app.AppState) {
	w.SetTitle("PassQuantum - Vault Selection")

	// Create navigation state
	navState := &NavigationState{
		currentView: NavViewVaults,
		window:      w,
		app:         fyneApp,
		appState:    appState,
	}

	// Create content container that will be dynamically updated
	navState.contentContainer = container.NewMax()
	navState.sidebarContainer = container.NewMax()

	// Build initial UI
	navState.rebuildUI()

	bgContainer := theme.CreateBackgroundContainer(navState.sidebarContainer)
	w.SetContent(bgContainer)
}

func createVaultCard(w fyne.Window, fyneApp fyne.App, appState *app.AppState, vaultName string) fyne.CanvasObject {
	isActive := vaultName == appState.CurrentVault

	avatar := theme.VaultAvatar(vaultName)

	titleTxt := canvas.NewText(vaultName, theme.ColorTextPrimary)
	titleTxt.TextSize = 13
	titleTxt.TextStyle = fyne.TextStyle{Bold: true}

	var statusPill fyne.CanvasObject
	if isActive {
		statusPill = theme.StatusPill("Active", theme.PillOk)
	} else {
		statusPill = theme.StatusPill("Locked", theme.PillMute)
	}
	titleRow := container.NewHBox(titleTxt, statusPill)

	vaultPath := app.GetVaultPath(vaultName)
	pathTxt := theme.MonoText(vaultPath, 11, theme.ColorFg2)

	openBtn := theme.CreateDefaultButton("Open", func() {
		if err := app.OpenVault(appState, vaultName, func() { ShowMainScreen(w, fyneApp, appState) }); err != nil {
			widgets.ShowAppError(err, w)
		}
	})

	deleteBtn := theme.CreateDangerButton("Delete", func() {
		showDeleteVaultDialog(w, fyneApp, appState, vaultName)
	})

	left := container.NewHBox(avatar, container.NewVBox(titleRow, pathTxt))
	buttons := container.NewHBox(openBtn, deleteBtn)
	row := container.NewBorder(nil, nil, left, buttons)

	return theme.CardWithHeader("", "", nil, row)
}

func showCreateVaultDialog(w fyne.Window, fyneApp fyne.App, appState *app.AppState) {
	vaultNameInput := widget.NewEntry()
	vaultNameInput.PlaceHolder = "Enter vault name"

	formContent := container.NewVBox(
		theme.SectionEyebrow("NEW VAULT"),
		theme.FieldLabel("VAULT NAME", nil),
		vaultNameInput,
		theme.MonoText("Uses the unlocked global master password automatically.", 11, theme.ColorFg2),
	)

	var customDialog *dialog.CustomDialog

	createBtn := theme.CreatePrimaryButton("Create vault", func() {
		vaultName := vaultNameInput.Text
		if vaultName == "" {
			widgets.ShowAppError(fmt.Errorf("vault name cannot be empty"), w)
			return
		}
		if err := app.CreateNewVault(appState, vaultName); err != nil {
			widgets.ShowAppError(err, w)
			return
		}
		if customDialog != nil {
			customDialog.Hide()
		}
		ShowVaultSelection(w, fyneApp, appState)
	})

	cancelBtn := theme.CreateGhostButton("Cancel", func() {
		if customDialog != nil {
			customDialog.Hide()
		}
	})

	buttonBox := container.NewHBox(cancelBtn, createBtn)
	dialogContent := container.NewVBox(formContent, container.NewCenter(buttonBox))
	customDialog = dialog.NewCustom("Create New Vault", "Close", dialogContent, w)
	customDialog.Show()
}

func showDeleteVaultDialog(w fyne.Window, fyneApp fyne.App, appState *app.AppState, vaultName string) {
	widgets.ShowAppConfirm("Delete Vault",
		fmt.Sprintf("Are you sure you want to delete '%s'? This cannot be undone.", vaultName),
		func(confirmed bool) {
			if confirmed {
				vaultPath := app.GetVaultPath(vaultName)
				err := os.Remove(vaultPath)
				if err != nil {
					widgets.ShowAppError(fmt.Errorf("failed to delete vault: %w", err), w)
					return
				}
				widgets.ShowAppInformation("Success", "Vault deleted successfully", w)
				ShowVaultSelection(w, fyneApp, appState)
			}
		}, w)
}
