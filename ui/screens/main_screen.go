package screens

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/png"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"passquantum/app"
	"passquantum/core/model"
	"passquantum/theme"
	"passquantum/ui/assets"
	"passquantum/ui/widgets"
)

// Current active view in the navigation
type NavView int

const NavViewNone NavView = -1

const (
	NavViewVaults NavView = iota
	NavViewAddItem
	NavViewItems
	NavViewGenerator
	NavViewChecker
	NavViewSettings
	NavViewTOTP
	NavViewFiles
	NavViewImport
)

// NavigationState tracks the current view
type NavigationState struct {
	currentView      NavView
	window           fyne.Window
	app              fyne.App
	appState         *app.AppState
	contentContainer *fyne.Container
	sidebarContainer *fyne.Container
	sidebarCollapsed bool
	viewCleanup      func()
}

func ShowMainScreen(w fyne.Window, fyneApp fyne.App, appState *app.AppState) {
	w.SetTitle("PassQuantum - " + appState.CurrentVault)

	navState := &NavigationState{
		currentView: NavViewItems,
		window:      w,
		app:         fyneApp,
		appState:    appState,
	}

	navState.contentContainer = container.NewMax()
	navState.sidebarContainer = container.NewMax()

	navState.rebuildUI()

	bg := canvas.NewRectangle(theme.ColorBg)
	w.SetContent(container.NewStack(bg, navState.sidebarContainer))

	if guard := appState.FaceGuard; guard != nil {
		guard.SendCommand("START_MONITOR")
	}
}

func (ns *NavigationState) breadcrumbs() []string {
	base := "PassQuantum"
	switch ns.currentView {
	case NavViewVaults:
		return []string{base, "Vaults"}
	case NavViewAddItem:
		return []string{base, ns.appState.CurrentVault, "Add item"}
	case NavViewItems:
		return []string{base, ns.appState.CurrentVault, "Items"}
	case NavViewGenerator:
		return []string{base, "Generator"}
	case NavViewChecker:
		return []string{base, "Analyzer"}
	case NavViewSettings:
		return []string{base, "Settings"}
	case NavViewTOTP:
		return []string{base, ns.appState.CurrentVault, "Authenticator"}
	case NavViewFiles:
		return []string{base, ns.appState.CurrentVault, "Files"}
	case NavViewImport:
		return []string{base, ns.appState.CurrentVault, "Import"}
	default:
		return []string{base}
	}
}

func (ns *NavigationState) rebuildUI() {
	sidebarBg := canvas.NewRectangle(theme.ColorSidebarBg)

	// Brand section
	brandIconBg := canvas.NewRectangle(theme.ColorAccentSoft)
	brandIconBg.CornerRadius = theme.RadiusInput
	brandIconBg.SetMinSize(fyne.NewSize(28, 28))

	brandIconBorder := canvas.NewRectangle(color.Transparent)
	brandIconBorder.CornerRadius = theme.RadiusInput
	brandIconBorder.StrokeWidth = 1
	brandIconBorder.StrokeColor = theme.ColorAccentLine
	brandIconBorder.FillColor = color.Transparent
	brandIconBorder.SetMinSize(fyne.NewSize(28, 28))

	var brandIconInner fyne.CanvasObject
	if img, _, err := image.Decode(bytes.NewReader(assets.LogoImage)); err == nil {
		logoImg := canvas.NewImageFromImage(img)
		logoImg.FillMode = canvas.ImageFillContain
		logoImg.SetMinSize(fyne.NewSize(56, 56))
		brandIconInner = logoImg
	} else {
		ico := canvas.NewImageFromResource(theme.IconAtom)
		ico.SetMinSize(fyne.NewSize(16, 16))
		brandIconInner = ico
	}

	brandIcon := container.NewStack(brandIconBg, brandIconBorder, container.NewCenter(brandIconInner))
	brandIconWrap := container.NewGridWrap(fyne.NewSize(64, 64), brandIcon)

	brandName := canvas.NewText("PassQuantum", theme.ColorTextPrimary)
	brandName.TextSize = 13
	brandName.TextStyle = fyne.TextStyle{Bold: true}

	brandMeta := canvas.NewText("PQ-SAFE", theme.ColorFg2)
	brandMeta.TextSize = 10
	brandMeta.TextStyle = fyne.TextStyle{Monospace: true}

	brandText := container.NewVBox(brandName, brandMeta)
	brandRow := container.NewHBox(brandIconWrap, brandText)
	brandSection := container.NewPadded(brandRow)

	// Nav items per section
	type navEntry struct {
		icon   *fyne.StaticResource
		label  string
		view   NavView
		action func()
	}

	vaultSection := []navEntry{
		{theme.IconVault, "Vaults", NavViewVaults, nil},
		{theme.IconPlus, "Add item", NavViewAddItem, nil},
		{theme.IconKey, "Items", NavViewItems, nil},
		{theme.IconClock, "Authenticator", NavViewTOTP, nil},
		{theme.IconFolder, "Files", NavViewFiles, nil},
		{theme.IconDownload, "Import", NavViewImport, nil},
	}
	toolsSection := []navEntry{
		{theme.IconWand, "Generate", NavViewGenerator, nil},
		{theme.IconShieldCheck, "Analyze", NavViewChecker, nil},
	}

	collapseIcon := theme.IconPanelLeftClose
	collapseLabel := "Collapse"
	if ns.sidebarCollapsed {
		collapseIcon = theme.IconPanelLeftOpen
		collapseLabel = "Expand"
	}

	systemSection := []navEntry{
		{theme.IconSettings, "Settings", NavViewSettings, nil},
		{collapseIcon, collapseLabel, NavViewNone, func() {
			ns.sidebarCollapsed = !ns.sidebarCollapsed
			ns.rebuildUI()
		}},
		{theme.IconLock, "Lock vault", NavViewNone, func() {
			ns.appState.ClearSensitiveState()
			ns.app.Quit()
		}},
	}

	buildSection := func(eyebrow string, entries []navEntry) fyne.CanvasObject {
		items := []fyne.CanvasObject{
			container.NewPadded(theme.SectionEyebrow(eyebrow)),
		}
		for _, e := range entries {
			entry := e
			isActive := entry.view != NavViewNone && ns.currentView == entry.view
			navItem := theme.NewNavItem(entry.icon, entry.label, isActive, func() {
				if entry.action != nil {
					entry.action()
				} else {
					ns.switchView(entry.view)
				}
			})
			items = append(items, navItem)
		}
		return container.NewVBox(items...)
	}

	vaultNav := buildSection("VAULT", vaultSection)
	toolsNav := buildSection("TOOLS", toolsSection)
	systemNav := buildSection("SYSTEM", systemSection)

	divider1 := canvas.NewRectangle(theme.ColorLine1)
	divider1.SetMinSize(fyne.NewSize(0, 0.5))
	divider2 := canvas.NewRectangle(theme.ColorLine1)
	divider2.SetMinSize(fyne.NewSize(0, 0.5))
	divider3 := canvas.NewRectangle(theme.ColorLine1)
	divider3.SetMinSize(fyne.NewSize(0, 0.5))

	var sidebarFixed *fyne.Container

	if ns.sidebarCollapsed {
		// Collapsed: icon-only narrow sidebar (56px)
		makeIconBtn := func(icon *fyne.StaticResource, view NavView, action func()) fyne.CanvasObject {
			isActive := view != NavViewNone && ns.currentView == view
			return theme.NewNavItem(icon, "", isActive, func() {
				if action != nil {
					action()
				} else {
					ns.switchView(view)
				}
			})
		}

		collapseDivider := canvas.NewRectangle(theme.ColorLine1)
		collapseDivider.SetMinSize(fyne.NewSize(0, 0.5))

		iconNav := container.NewVBox(
			container.NewCenter(brandIconWrap),
			divider1,
			makeIconBtn(theme.IconVault, NavViewVaults, nil),
			makeIconBtn(theme.IconPlus, NavViewAddItem, nil),
			makeIconBtn(theme.IconKey, NavViewItems, nil),
			makeIconBtn(theme.IconClock, NavViewTOTP, nil),
			makeIconBtn(theme.IconFolder, NavViewFiles, nil),
			makeIconBtn(theme.IconDownload, NavViewImport, nil),
			divider2,
			makeIconBtn(theme.IconWand, NavViewGenerator, nil),
			makeIconBtn(theme.IconShieldCheck, NavViewChecker, nil),
			divider3,
			makeIconBtn(theme.IconSettings, NavViewSettings, nil),
			collapseDivider,
			makeIconBtn(collapseIcon, NavViewNone, func() {
				ns.sidebarCollapsed = !ns.sidebarCollapsed
				ns.rebuildUI()
			}),
			makeIconBtn(theme.IconLock, NavViewNone, func() {
				ns.appState.ClearSensitiveState()
				ns.app.Quit()
			}),
		)

		collapsedInner := container.NewBorder(iconNav, nil, nil, nil)
		collapsedPadded := container.NewPadded(collapsedInner)
		sidebar := container.NewStack(sidebarBg, collapsedPadded)
		sidebarFixed = container.NewGridWrap(fyne.NewSize(56, 0), sidebar)
	} else {
		// Expanded: full sidebar with labels
		navContent := container.NewVBox(
			brandSection,
			divider1,
			vaultNav,
			divider2,
			toolsNav,
			divider3,
			systemNav,
		)

		sidebarInner := container.NewBorder(navContent, nil, nil, nil)
		sidebarPadded := container.NewPadded(sidebarInner)
		sidebar := container.NewStack(sidebarBg, sidebarPadded)
		sidebarFixed = container.NewGridWrap(fyne.NewSize(theme.SidebarWidth, 0), sidebar)
	}

	// Topbar
	pills := []fyne.CanvasObject{
		theme.StatusPill("Vault: "+ns.appState.CurrentVault, theme.PillAccent),
	}
	if ns.appState.FaceGuard != nil {
		pills = append(pills, theme.StatusPill("Watching: ON", theme.PillOk))
	}

	//topbar := theme.Topbar(ns.breadcrumbs(), pills)

	// Content area
	ns.updateContent()

	contentScroll := container.NewVScroll(ns.contentContainer)

	mainLayout := container.NewBorder(nil, nil, sidebarFixed, nil, contentScroll)

	ns.sidebarContainer.Objects = []fyne.CanvasObject{mainLayout}
	ns.sidebarContainer.Refresh()
}

func (ns *NavigationState) switchView(view NavView) {
	if ns.viewCleanup != nil {
		ns.viewCleanup()
		ns.viewCleanup = nil
	}
	ns.currentView = view
	crumbs := ns.breadcrumbs()
	ns.window.SetTitle("PassQuantum — " + crumbs[len(crumbs)-1])
	ns.rebuildUI()
}

func (ns *NavigationState) updateContent() {
	var content fyne.CanvasObject

	switch ns.currentView {
	case NavViewVaults:
		content = ns.createVaultsView()
	case NavViewAddItem:
		content = ns.createPasswordsView()
	case NavViewItems:
		content = ns.createItemsView()
	case NavViewGenerator:
		content = ns.createGeneratorView()
	case NavViewChecker:
		content = ns.createCheckerView()
	case NavViewSettings:
		content = ns.createSettingsView()
	case NavViewTOTP:
		content = ns.createTOTPView()
	case NavViewFiles:
		content = ns.createFilesView()
	case NavViewImport:
		content = ns.createImportView()
	default:
		content = ns.createItemsView()
	}

	padded := container.NewPadded(content)

	ns.contentContainer.Objects = []fyne.CanvasObject{padded}
	ns.contentContainer.Refresh()
}

func (ns *NavigationState) createVaultsView() fyne.CanvasObject {
	vaults := app.ListVaults()

	newVaultBtn := theme.CreatePrimaryButtonWithIcon("New vault", theme.IconPlus, func() {
		showCreateVaultDialog(ns.window, ns.app, ns.appState)
	})

	header := theme.PageHeader(
		"PASSQUANTUM / VAULTS",
		"Your vaults",
		"Encrypted containers for passwords, cards, and notes.",
		newVaultBtn,
	)

	var vaultItems []fyne.CanvasObject
	if len(vaults) == 0 {
		emptyMsg := canvas.NewText("No vaults found. Create one to get started.", theme.ColorFg2)
		emptyMsg.TextSize = 13
		vaultItems = append(vaultItems, container.NewCenter(emptyMsg))
	} else {
		for _, vaultName := range vaults {
			vaultCard := createVaultCard(ns.window, ns.app, ns.appState, vaultName)
			vaultItems = append(vaultItems, vaultCard)
		}
	}

	encryptionInfo := theme.CollapsibleCardWithHeader("ENCRYPTION", "Vault security", nil,
		theme.KeyValueTable([]theme.KVItem{
			{Key: "Algorithm", Value: "AES-256-GCM", Detail: "96-bit IV, 128-bit auth tag"},
			{Key: "Key encapsulation", Value: "ML-KEM-768 (Kyber)", Detail: "NIST FIPS 203"},
			{Key: "Password KDF", Value: "Argon2id", Detail: "m=64 MiB, t=3, p=4, 16-byte salt"},
			{Key: "Random source", Value: "crypto/rand", Detail: "OS CSPRNG"},
		}),
	)

	return container.NewVBox(
		header,
		container.NewVBox(vaultItems...),
		encryptionInfo,
	)
}

func (ns *NavigationState) createSettingsView() fyne.CanvasObject {
	return buildCustomSettingsView(ns.window, ns.app, ns.appState)
}

// showSaveGeneratedPasswordDialog shows a dialog to save a generated password to a vault
func showSaveGeneratedPasswordDialog(w fyne.Window, fyneApp fyne.App, appState *app.AppState, password string) {
	// Get available vaults
	vaults := app.ListVaults()
	if len(vaults) == 0 {
		widgets.ShowAppError(fmt.Errorf("no vaults available. Create a vault first"), w)
		return
	}

	// Create input fields
	vaultSelect := widget.NewSelect(vaults, nil)
	vaultSelect.PlaceHolder = "Select a vault"
	if appState.CurrentVault != "" {
		vaultSelect.SetSelected(appState.CurrentVault)
	} else {
		vaultSelect.SetSelected(vaults[0])
	}

	serviceInput := widget.NewEntry()
	serviceInput.PlaceHolder = "Service name"

	usernameInput := widget.NewEntry()
	usernameInput.PlaceHolder = "Username or email"

	vaultLabel := theme.SectionEyebrow("VAULT")
	serviceLabel := theme.SectionEyebrow("SERVICE")
	usernameLabel := theme.SectionEyebrow("USERNAME")

	// Create styled containers for inputs
	vaultSelectBg := canvas.NewRectangle(theme.ColorInputBg)
	vaultSelectBg.SetMinSize(fyne.NewSize(350, 40))
	vaultSelectBg.CornerRadius = theme.BorderRadius
	styledVaultSelect := container.NewStack(vaultSelectBg, container.NewPadded(vaultSelect))

	serviceInputBg := canvas.NewRectangle(theme.ColorInputBg)
	serviceInputBg.SetMinSize(fyne.NewSize(350, 40))
	serviceInputBg.CornerRadius = theme.BorderRadius
	styledServiceInput := container.NewStack(serviceInputBg, container.NewPadded(serviceInput))

	usernameInputBg := canvas.NewRectangle(theme.ColorInputBg)
	usernameInputBg.SetMinSize(fyne.NewSize(350, 40))
	usernameInputBg.CornerRadius = theme.BorderRadius
	styledUsernameInput := container.NewStack(usernameInputBg, container.NewPadded(usernameInput))

	formContent := container.NewVBox(
		theme.SectionEyebrow("SAVE TO VAULT"),
		vaultLabel,
		styledVaultSelect,
		serviceLabel,
		styledServiceInput,
		usernameLabel,
		styledUsernameInput,
	)

	var customDialog *dialog.CustomDialog

	saveBtn := theme.CreatePrimaryButton("Save entry", func() {
		selectedVault := vaultSelect.Selected
		service := serviceInput.Text
		username := usernameInput.Text

		if selectedVault == "" {
			widgets.ShowAppError(fmt.Errorf("please select a vault"), w)
			return
		}

		if service == "" {
			widgets.ShowAppError(fmt.Errorf("service name cannot be empty"), w)
			return
		}

		// Close the form dialog first
		if customDialog != nil {
			customDialog.Hide()
		}

		// Now prompt for vault password and save
		if err := app.OpenVault(appState, selectedVault, func() {
			// Vault is now unlocked, proceed with saving
			go func() {
				appState.Mu.Lock()
				defer appState.Mu.Unlock()

				vaultFile := app.GetVaultPath(selectedVault)
				entries, err := app.ReadVault(vaultFile, appState.MasterPassword)
				if err != nil {
					fyne.Do(func() {
						widgets.ShowAppError(fmt.Errorf("failed to read vault: %w", err), w)
					})
					return
				}

				ct, ss, err := app.Encapsulate(appState.PublicKey)
				if err != nil {
					fyne.Do(func() {
						widgets.ShowAppError(fmt.Errorf("encapsulation failed: %v", err), w)
					})
					return
				}

				nonce, ciphertext, err := app.EncryptAES256GCM(password, ss)
				if err != nil {
					fyne.Do(func() {
						widgets.ShowAppError(fmt.Errorf("encryption failed: %v", err), w)
					})
					return
				}

				entry := model.NewVaultEntry()
				entry.Type = model.EntryTypePassword
				entry.KyberCiphertext = ct
				entry.Nonce = nonce
				entry.Ciphertext = ciphertext
				entry.Service = service
				entry.Username = username

				entries = append(entries, entry)

				err = app.WriteVault(entries, vaultFile, appState.MasterPassword)
				if err != nil {
					fyne.Do(func() {
						widgets.ShowAppError(fmt.Errorf("failed to save vault item: %v", err), w)
					})
					return
				}

				fyne.Do(func() {
					widgets.ShowAppInformation("Success", "✓ Vault item saved to vault successfully!", w)
				})
			}()
		}); err != nil {
			widgets.ShowAppError(err, w)
		}
	})

	cancelBtn := theme.CreateGhostButton("Cancel", func() {
		if customDialog != nil {
			customDialog.Hide()
		}
	})

	buttonBox := container.NewHBox(cancelBtn, saveBtn)

	dialogContent := container.NewVBox(formContent, container.NewCenter(buttonBox))
	customDialog = dialog.NewCustomWithoutButtons("Save to Vault", dialogContent, w)
	customDialog.Show()
}
