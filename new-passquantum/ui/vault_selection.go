package main

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowVaultSelection displays the vault management and selection screen
func ShowVaultSelection(w fyne.Window, fyneApp fyne.App, appState *AppState) {
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

	bgContainer := CreateBackgroundContainer(navState.sidebarContainer)
	w.SetContent(bgContainer)
}

func createVaultCard(w fyne.Window, fyneApp fyne.App, appState *AppState, vaultName string) fyne.CanvasObject {
	nameLabel := CreateLabel(vaultName, 13, ColorAccentCyn, true)
	vaultPath := filepath.Join("vaults", vaultName+".pqdb")
	infoLabel := CreateLabel("Location: "+vaultPath, 10, ColorTextSec, false)

	openBtn := CreateNeonButton("OPEN", func() {
		OpenVault(w, fyneApp, appState, vaultName, nil)
	}, 80, 36)

	deleteBtn := CreateNeonButton("DELETE", func() {
		showDeleteVaultDialog(w, fyneApp, appState, vaultName)
	}, 100, 36)

	btnContainer := container.NewHBox(openBtn, deleteBtn)
	content := container.NewVBox(nameLabel, infoLabel, btnContainer)

	return CreateCard(content, 850, 70, true)
}

func showCreateVaultDialog(w fyne.Window, fyneApp fyne.App, appState *AppState) {
	vaultNameInput := widget.NewEntry()
	vaultNameInput.PlaceHolder = "Enter vault name"

	// Create styled input containers
	createInputContainer := func(input fyne.CanvasObject) fyne.CanvasObject {
		bg := canvas.NewRectangle(color.NRGBA{R: 30, G: 40, B: 50, A: 255})
		bg.SetMinSize(fyne.NewSize(350, 50))
		bg.CornerRadius = BorderRadius
		return container.NewMax(bg, container.NewPadded(input))
	}

	// Build form content
	formContent := container.NewVBox(
		CreateLabel("CREATE NEW VAULT", 14, ColorAccentCyn, true),
		CreateDivider(),
		widget.NewLabel(""),
		CreateLabel("Vault Name", 11, ColorPurple, true),
		createInputContainer(vaultNameInput),
		widget.NewLabel(""),
		CreateLabel("Vaults now use the unlocked global master password automatically.", 10, ColorTextSec, false),
		widget.NewLabel(""),
	)

	// Declare customDialog first so it can be referenced in button closures
	var customDialog *dialog.CustomDialog

	createBtn := CreateNeonButton("✓ CREATE", func() {
		vaultName := vaultNameInput.Text

		if vaultName == "" {
			ShowAppError(fmt.Errorf("vault name cannot be empty"), w)
			return
		}

		if CreateNewVault(w, appState, vaultName) {
			if customDialog != nil {
				customDialog.Hide()
			}
			ShowVaultSelection(w, fyneApp, appState)
		}
	}, 120, 44)

	cancelBtn := CreateNeonButton("✕ CANCEL", func() {
		if customDialog != nil {
			customDialog.Hide()
		}
	}, 120, 44)

	buttonBox := container.NewHBox(cancelBtn, createBtn)

	dialogContent := container.NewVBox(formContent, container.NewCenter(buttonBox))
	customDialog = dialog.NewCustom("Create New Vault", "Close", dialogContent, w)
	customDialog.Show()
}

func showDeleteVaultDialog(w fyne.Window, fyneApp fyne.App, appState *AppState, vaultName string) {
	ShowAppConfirm("Delete Vault",
		fmt.Sprintf("Are you sure you want to delete '%s'? This cannot be undone.", vaultName),
		func(confirmed bool) {
			if confirmed {
				vaultPath := filepath.Join("vaults", vaultName+".pqdb")
				err := os.Remove(vaultPath)
				if err != nil {
					ShowAppError(fmt.Errorf("failed to delete vault: %w", err), w)
					return
				}
				ShowAppInformation("Success", "Vault deleted successfully", w)
				ShowVaultSelection(w, fyneApp, appState)
			}
		}, w)
}
