package screens

import (
	"fmt"
	"image/color"
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
	w.Resize(fyne.NewSize(1100, 700))

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
	nameLabel := theme.CreateLabel(vaultName, 13, theme.ColorAccentCyn, true)
	vaultPath := app.GetVaultPath(vaultName)
	infoLabel := theme.CreateLabel("Location: "+vaultPath, 10, theme.ColorTextSec, false)

	openBtn := theme.CreateNeonButton("OPEN", func() {
		if err := app.OpenVault(appState, vaultName, func() { ShowMainScreen(w, fyneApp, appState) }); err != nil {
			widgets.ShowAppError(err, w)
		}
	}, 80, 36)

	deleteBtn := theme.CreateNeonButton("DELETE", func() {
		showDeleteVaultDialog(w, fyneApp, appState, vaultName)
	}, 100, 36)

	btnContainer := container.NewHBox(openBtn, deleteBtn)
	content := container.NewVBox(nameLabel, infoLabel, btnContainer)

	return theme.CreateCard(content, 850, 70, true)
}

func showCreateVaultDialog(w fyne.Window, fyneApp fyne.App, appState *app.AppState) {
	vaultNameInput := widget.NewEntry()
	vaultNameInput.PlaceHolder = "Enter vault name"

	// Create styled input containers
	createInputContainer := func(input fyne.CanvasObject) fyne.CanvasObject {
		bg := canvas.NewRectangle(color.NRGBA{R: 30, G: 40, B: 50, A: 255})
		bg.SetMinSize(fyne.NewSize(350, 50))
		bg.CornerRadius = theme.BorderRadius
		return container.NewMax(bg, container.NewPadded(input))
	}

	// Build form content
	formContent := container.NewVBox(
		theme.CreateLabel("CREATE NEW VAULT", 14, theme.ColorAccentCyn, true),
		theme.CreateDivider(),
		widget.NewLabel(""),
		theme.CreateLabel("Vault Name", 11, theme.ColorPurple, true),
		createInputContainer(vaultNameInput),
		widget.NewLabel(""),
		theme.CreateLabel("Vaults now use the unlocked global master password automatically.", 10, theme.ColorTextSec, false),
		widget.NewLabel(""),
	)

	// Declare customDialog first so it can be referenced in button closures
	var customDialog *dialog.CustomDialog

	createBtn := theme.CreateNeonButton("✓ CREATE", func() {
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
	}, 120, 44)

	cancelBtn := theme.CreateNeonButton("✕ CANCEL", func() {
		if customDialog != nil {
			customDialog.Hide()
		}
	}, 120, 44)

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
