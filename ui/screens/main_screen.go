package screens

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"passquantum/app"
	"passquantum/core/model"
	"passquantum/core/totp"
	"passquantum/strength"
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
		logoImg.SetMinSize(fyne.NewSize(20, 20))
		brandIconInner = logoImg
	} else {
		ico := canvas.NewImageFromResource(theme.IconAtom)
		ico.SetMinSize(fyne.NewSize(16, 16))
		brandIconInner = ico
	}

	brandIcon := container.NewStack(brandIconBg, brandIconBorder, container.NewCenter(brandIconInner))
	brandIconWrap := container.NewGridWrap(fyne.NewSize(28, 28), brandIcon)

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

func (ns *NavigationState) createPasswordsView() fyne.CanvasObject {
	type cardPayload struct {
		Subtype string `json:"subtype"`
		Holder  string `json:"holder"`
		Number  string `json:"number"`
		Expiry  string `json:"expiry"`
		CVV     string `json:"cvv"`
	}

	backBtn := theme.CreateGhostButton("All items", func() {
		ns.switchView(NavViewItems)
	})

	header := theme.PageHeader(
		"PASSQUANTUM / "+ns.appState.CurrentVault,
		"Add vault item",
		"Create a new encrypted entry in this vault.",
		backBtn,
	)

	passwordInput := widget.NewPasswordEntry()
	passwordInput.PlaceHolder = "Enter password"

	// Segmented strength meter
	strengthMeter := theme.SegmentedStrengthMeterFlex(0)
	strengthLabel := canvas.NewText("", theme.ColorTextSecondary)
	strengthLabel.TextSize = 11
	strengthLabel.TextStyle = fyne.TextStyle{Monospace: true}
	strengthCrack := canvas.NewText("", theme.ColorFg2)
	strengthCrack.TextSize = 11
	strengthCrack.TextStyle = fyne.TextStyle{Monospace: true}

	strengthIssues := container.NewVBox(
		newStrengthText("Start typing to analyze password strength.", theme.ColorFg2, 11, false),
	)

	strengthContainer := container.NewVBox(
		container.NewBorder(nil, nil, nil, strengthLabel, strengthMeter),
		strengthCrack,
		strengthIssues,
	)

	updateAddStrength := func(value string) {
		result := strength.Analyze(value, storedVaultPasswords(ns.appState))

		if result.EasterEggMode {
			level := int(result.Score) + 1
			if level > 5 {
				level = 5
			}
			strengthMeter = theme.SegmentedStrengthMeterFlex(level)
			strengthLabel.Text = "Password Game Mode"
			strengthLabel.Refresh()
			strengthCrack.Text = "neal.fun trigger detected"
			strengthCrack.Refresh()
			strengthIssues.Objects = []fyne.CanvasObject{NewEasterEggPanel(result.EasterEggRules)}
			strengthIssues.Refresh()
			return
		}

		level := 0
		if value != "" {
			level = int(result.Score) + 1
			if level > 5 {
				level = 5
			}
		}

		newMeter := theme.SegmentedStrengthMeterFlex(level)
		strengthContainer.Objects[0] = container.NewBorder(nil, nil, nil, strengthLabel, newMeter)
		strengthContainer.Refresh()

		if value == "" {
			strengthLabel.Text = ""
			strengthLabel.Refresh()
			strengthCrack.Text = ""
			strengthCrack.Refresh()
			strengthIssues.Objects = []fyne.CanvasObject{
				newStrengthText("Start typing to analyze password strength.", theme.ColorFg2, 11, false),
			}
		} else {
			strengthLabel.Text = result.ScoreLabel
			strengthLabel.Refresh()
			strengthCrack.Text = "Crack time: " + result.CrackTime
			strengthCrack.Refresh()
			strengthIssues.Objects = []fyne.CanvasObject{NewIssuesList(result.Issues)}
		}
		strengthIssues.Refresh()
	}

	passwordInput.OnChanged = updateAddStrength
	updateAddStrength(passwordInput.Text)

	serviceInput := widget.NewEntry()
	serviceInput.PlaceHolder = "e.g. github.com"

	usernameInput := widget.NewEntry()
	usernameInput.PlaceHolder = "Username or email"

	noteTitleInput := widget.NewEntry()
	noteTitleInput.PlaceHolder = "Note title"
	noteContentInput := widget.NewMultiLineEntry()
	noteContentInput.SetMinRowsVisible(5)
	noteContentInput.PlaceHolder = "Plaintext is encrypted at rest. Markdown is preserved."

	cardTypeSelect := widget.NewSelect([]string{"Credit", "Debit", "Prepaid"}, nil)
	cardTypeSelect.SetSelected("Credit")
	cardNameInput := widget.NewEntry()
	cardNameInput.PlaceHolder = "Card nickname"
	cardHolderInput := widget.NewEntry()
	cardHolderInput.PlaceHolder = "Card holder"
	cardNumberInput := widget.NewEntry()
	cardNumberInput.PlaceHolder = "Card number"
	cardExpiryInput := widget.NewEntry()
	cardExpiryInput.PlaceHolder = "MM/YY"
	cardCVVInput := widget.NewPasswordEntry()
	cardCVVInput.PlaceHolder = "CVV"

	// TOTP fields
	totpIssuerInput := widget.NewEntry()
	totpIssuerInput.PlaceHolder = "e.g. GitHub"
	totpAccountInput := widget.NewEntry()
	totpAccountInput.PlaceHolder = "user@example.com"
	totpSecretInput := widget.NewEntry()
	totpSecretInput.PlaceHolder = "Base32 secret (e.g. JBSWY3DPEHPK3PXP)"
	totpAlgorithmSelect := widget.NewSelect([]string{"SHA1", "SHA256", "SHA512"}, nil)
	totpAlgorithmSelect.SetSelected("SHA1")
	totpDigitsSelect := widget.NewSelect([]string{"6", "7", "8"}, nil)
	totpDigitsSelect.SetSelected("6")
	totpPeriodSelect := widget.NewSelect([]string{"30", "60", "90"}, nil)
	totpPeriodSelect.SetSelected("30")

	itemTypes := []string{"Password", "Cyphered Note", "Card", "TOTP"}
	activeItemType := 0 // 0=Password, 1=Note, 2=Card, 3=TOTP

	passwordSection := container.NewVBox(
		theme.FieldLabel("SERVICE", nil),
		serviceInput,
		theme.FieldLabel("USERNAME / EMAIL", nil),
		usernameInput,
		theme.FieldLabel("PASSWORD", nil),
		passwordInput,
		strengthContainer,
	)

	noteSection := container.NewVBox(
		theme.FieldLabel("NOTE TITLE", nil),
		noteTitleInput,
		theme.FieldLabel("CYPHERED NOTE", nil),
		noteContentInput,
	)

	cardSection := container.NewVBox(
		theme.FieldLabel("CARD TYPE", nil),
		cardTypeSelect,
		theme.FieldLabel("CARD NICKNAME", nil),
		cardNameInput,
		theme.FieldLabel("CARD HOLDER", nil),
		cardHolderInput,
		theme.FieldLabel("CARD NUMBER", nil),
		cardNumberInput,
		container.NewGridWithColumns(2,
			container.NewVBox(theme.FieldLabel("EXPIRY", nil), cardExpiryInput),
			container.NewVBox(theme.FieldLabel("CVV", nil), cardCVVInput),
		),
	)

	totpSection := container.NewVBox(
		theme.FieldLabel("ISSUER", nil),
		totpIssuerInput,
		theme.FieldLabel("ACCOUNT", nil),
		totpAccountInput,
		theme.FieldLabel("SECRET (BASE32)", nil),
		totpSecretInput,
		container.NewGridWithColumns(3,
			container.NewVBox(theme.FieldLabel("ALGORITHM", nil), totpAlgorithmSelect),
			container.NewVBox(theme.FieldLabel("DIGITS", nil), totpDigitsSelect),
			container.NewVBox(theme.FieldLabel("PERIOD (s)", nil), totpPeriodSelect),
		),
	)

	var formContent *fyne.Container
	var typeTabs fyne.CanvasObject
	var buildTypeTabs func()

	refreshSections := func(idx int) {
		activeItemType = idx
		passwordSection.Hide()
		noteSection.Hide()
		cardSection.Hide()
		totpSection.Hide()

		switch idx {
		case 1:
			noteSection.Show()
		case 2:
			cardSection.Show()
		case 3:
			totpSection.Show()
		default:
			passwordSection.Show()
		}

		if formContent != nil {
			formContent.Refresh()
		}

		// Rebuild tabs to update active indicator
		buildTypeTabs()
		if formContent != nil {
			formContent.Objects[0] = typeTabs
			formContent.Refresh()
		}
	}

	buildTypeTabs = func() {
		typeTabs = theme.UnderlineTabs(itemTypes, activeItemType, refreshSections)
	}
	buildTypeTabs()
	refreshSections(0)

	saveBtn := theme.CreatePrimaryButton("Save item", func() {
		itemType := itemTypes[activeItemType]
		service := serviceInput.Text
		username := usernameInput.Text
		secret := passwordInput.Text
		entryType := model.EntryTypePassword
		cardSubtype := ""

		switch itemType {
		case "Cyphered Note":
			entryType = model.EntryTypeNote
			title := noteTitleInput.Text
			content := noteContentInput.Text
			if title == "" || content == "" {
				widgets.ShowAppError(fmt.Errorf("note title and content cannot be empty"), ns.window)
				return
			}
			notePayload, _ := json.Marshal(map[string]string{
				"type":    "note",
				"title":   title,
				"content": content,
			})
			service = "NOTE:" + title
			username = "note"
			secret = string(notePayload)
		case "Card":
			entryType = model.EntryTypeCard
			cardSubtype = cardTypeSelect.Selected
			if cardNameInput.Text == "" || cardHolderInput.Text == "" || cardNumberInput.Text == "" {
				widgets.ShowAppError(fmt.Errorf("card name, holder and number cannot be empty"), ns.window)
				return
			}
			cp := cardPayload{
				Subtype: cardTypeSelect.Selected,
				Holder:  cardHolderInput.Text,
				Number:  cardNumberInput.Text,
				Expiry:  cardExpiryInput.Text,
				CVV:     cardCVVInput.Text,
			}
			cardJSON, _ := json.Marshal(cp)
			service = "CARD:" + cardNameInput.Text
			username = cardTypeSelect.Selected
			secret = string(cardJSON)
		case "TOTP":
			entryType = model.EntryTypeTOTP
			if totpSecretInput.Text == "" || totpIssuerInput.Text == "" {
				widgets.ShowAppError(fmt.Errorf("TOTP issuer and secret are required"), ns.window)
				return
			}
			digits, _ := strconv.Atoi(totpDigitsSelect.Selected)
			period, _ := strconv.Atoi(totpPeriodSelect.Selected)
			params := &totp.TOTPParams{
				Secret:    totpSecretInput.Text,
				Algorithm: totp.Algorithm(totpAlgorithmSelect.Selected),
				Digits:    digits,
				Period:    period,
				Issuer:    totpIssuerInput.Text,
				Account:   totpAccountInput.Text,
			}
			if err := totp.Validate(params); err != nil {
				widgets.ShowAppError(fmt.Errorf("invalid TOTP params: %w", err), ns.window)
				return
			}
			totpJSON, _ := totp.Serialize(params)
			service = "TOTP:" + totpIssuerInput.Text
			username = totpAccountInput.Text
			secret = string(totpJSON)
		default:
			if secret == "" {
				widgets.ShowAppError(fmt.Errorf("password cannot be empty"), ns.window)
				return
			}
			if service == "" {
				widgets.ShowAppError(fmt.Errorf("service name cannot be empty"), ns.window)
				return
			}
		}

		clearInputs := func() {
			serviceInput.SetText("")
			usernameInput.SetText("")
			passwordInput.SetText("")
			noteTitleInput.SetText("")
			noteContentInput.SetText("")
			cardNameInput.SetText("")
			cardHolderInput.SetText("")
			cardNumberInput.SetText("")
			cardExpiryInput.SetText("")
			cardCVVInput.SetText("")
			totpIssuerInput.SetText("")
			totpAccountInput.SetText("")
			totpSecretInput.SetText("")
		}

		// writeEntry encrypts the secret and either appends a new entry or
		// replaces the crypto fields of an existing one (keeping its ID).
		writeEntry := func(replaceID uint64, replaceExisting bool, successMsg string) {
			ns.appState.Mu.Lock()
			defer ns.appState.Mu.Unlock()

			vaultFile := app.GetVaultPath(ns.appState.CurrentVault)
			entries, err := app.ReadVault(vaultFile, ns.appState.MasterPassword)
			if err != nil {
				fyne.Do(func() {
					widgets.ShowAppError(fmt.Errorf("failed to read vault: %w", err), ns.window)
				})
				return
			}

			ct, ss, err := app.Encapsulate(ns.appState.PublicKey)
			if err != nil {
				fyne.Do(func() {
					widgets.ShowAppError(fmt.Errorf("encapsulation failed: %v", err), ns.window)
				})
				return
			}

			nonce, ciphertext, err := app.EncryptAES256GCM(secret, ss)
			if err != nil {
				fyne.Do(func() {
					widgets.ShowAppError(fmt.Errorf("encryption failed: %v", err), ns.window)
				})
				return
			}

			if replaceExisting {
				var target *model.VaultEntry
				for _, e := range entries {
					if e != nil && e.ID == replaceID {
						target = e
						break
					}
				}
				if target == nil {
					fyne.Do(func() {
						widgets.ShowAppError(fmt.Errorf("entry no longer exists; please retry"), ns.window)
					})
					return
				}
				target.KyberCiphertext = ct
				target.Nonce = nonce
				target.Ciphertext = ciphertext
			} else {
				entry := model.NewVaultEntry()
				entry.KyberCiphertext = ct
				entry.Nonce = nonce
				entry.Ciphertext = ciphertext
				entry.Type = entryType
				entry.CardSubtype = cardSubtype
				entry.Service = service
				entry.Username = username
				entries = append(entries, entry)
			}

			if err := app.WriteVault(entries, vaultFile, ns.appState.MasterPassword); err != nil {
				fyne.Do(func() {
					widgets.ShowAppError(fmt.Errorf("failed to save vault item: %v", err), ns.window)
				})
				return
			}

			fyne.Do(func() {
				clearInputs()
				widgets.ShowAppInformation("Success", successMsg, ns.window)
			})
		}

		go func() {
			ns.appState.Mu.Lock()
			vaultFile := app.GetVaultPath(ns.appState.CurrentVault)
			entries, err := app.ReadVault(vaultFile, ns.appState.MasterPassword)
			if err != nil {
				ns.appState.Mu.Unlock()
				fyne.Do(func() {
					widgets.ShowAppError(fmt.Errorf("failed to read vault: %w", err), ns.window)
				})
				return
			}

			dup := app.FindDuplicateEntry(entries, entryType, service, username)
			ns.appState.Mu.Unlock()

			if dup != nil {
				dupID := dup.ID
				displayService := strings.TrimPrefix(service, "TOTP:")
				fyne.Do(func() {
					widgets.ShowAppConfirm(
						"Replace existing entry?",
						fmt.Sprintf("An entry for '%s' / '%s' already exists. Replace it?", displayService, username),
						func(confirmed bool) {
							if !confirmed {
								return
							}
							go writeEntry(dupID, true, "Entry replaced successfully!")
						},
						ns.window,
					)
				})
				return
			}

			writeEntry(0, false, "Item saved successfully!")
		}()
	})

	cancelBtn := theme.CreateGhostButton("Cancel", func() {
		ns.switchView(NavViewItems)
	})

	formContent = container.NewVBox(
		typeTabs,
		passwordSection,
		noteSection,
		cardSection,
		totpSection,
	)

	footer := theme.FormFooter("Encrypted on save: AES-256-GCM", cancelBtn, saveBtn)

	formCard := theme.CardWithHeader("", "", nil, formContent)

	return container.NewVBox(header, formCard, footer)
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

func (ns *NavigationState) createGeneratorView() fyne.CanvasObject {
	generatedPasswordDisplay := widget.NewEntry()
	generatedPasswordDisplay.PlaceHolder = "Generated password will appear here"
	generatedPasswordDisplay.MultiLine = false

	gc := newGeneratorControls(ns.window,
		func() string { return generatedPasswordDisplay.Text },
		func(s string) { generatedPasswordDisplay.SetText(s) },
	)

	// Auto-generate on load
	if initial, err := GeneratePassword(*gc.Settings); err == nil {
		generatedPasswordDisplay.SetText(initial)
	}

	header := theme.PageHeader(
		"PASSQUANTUM / GENERATOR",
		"Password generator",
		"Cryptographically secure random password generation.",
		theme.StatusPill("CSPRNG: crypto/rand", theme.PillAccent),
	)

	// Output card
	regenerateBtn := theme.CreateGhostButton("Regenerate", func() {
		password, err := GeneratePassword(*gc.Settings)
		if err != nil {
			widgets.ShowAppError(err, ns.window)
			return
		}
		generatedPasswordDisplay.SetText(password)
	})

	passwordDisplay := canvas.NewText("", theme.ColorTextPrimary)
	passwordDisplay.TextSize = 22
	passwordDisplay.TextStyle = fyne.TextStyle{Monospace: true}

	generatedPasswordDisplay.OnChanged = func(s string) {
		passwordDisplay.Text = s
		passwordDisplay.Refresh()
	}
	// Trigger initial display update
	passwordDisplay.Text = generatedPasswordDisplay.Text

	displayBg := canvas.NewRectangle(theme.ColorSidebarBg)
	displayBg.CornerRadius = theme.Space2
	displayBorder := canvas.NewRectangle(color.Transparent)
	displayBorder.CornerRadius = theme.Space2
	displayBorder.StrokeWidth = 1
	displayBorder.StrokeColor = theme.ColorLine2
	displayBorder.FillColor = color.Transparent

	displayBox := container.NewStack(
		displayBg, displayBorder,
		container.NewPadded(container.NewPadded(passwordDisplay)),
	)

	copyBtn := theme.CreateDefaultButton("Copy", func() {
		if generatedPasswordDisplay.Text != "" {
			ns.window.Clipboard().SetContent(generatedPasswordDisplay.Text)
			widgets.ShowAppInformation("Copied", "Password copied to clipboard!", ns.window)
		}
	})

	saveToVaultBtn := theme.CreatePrimaryButton("Save to vault", func() {
		password := generatedPasswordDisplay.Text
		if password == "" {
			widgets.ShowAppInformation("Empty", "Generate a password first!", ns.window)
			return
		}
		showSaveGeneratedPasswordDialog(ns.window, ns.app, ns.appState, password)
	})

	outputBody := container.NewVBox(
		displayBox,
		container.NewHBox(copyBtn, saveToVaultBtn),
	)

	outputCard := theme.CardWithHeader("OUTPUT", "Generated password", regenerateBtn, outputBody)

	// Length row: slider fills available space, entry shows numeric value on right
	lengthRow := container.NewBorder(nil, nil, nil,
		container.NewGridWrap(fyne.NewSize(52, 0), gc.LengthInput),
		gc.LengthSlider,
	)

	// Options card
	optionsBody := container.NewVBox(
		theme.FieldLabel("LENGTH", nil),
		lengthRow,
		theme.FieldLabel("CHARACTER CLASSES", nil),
		container.NewGridWithColumns(2,
			gc.UppercaseCheck,
			gc.LowercaseCheck,
			gc.NumbersCheck,
			gc.SpecialCharsCheck,
		),
		gc.AmbiguousCheck,
	)

	optionsCard := theme.CardWithHeader("CONFIGURATION", "Parameters", nil, optionsBody)

	return container.NewVBox(header, outputCard, optionsCard)
}

func (ns *NavigationState) createCheckerView() fyne.CanvasObject {
	header := theme.PageHeader(
		"PASSQUANTUM / ANALYZER",
		"Password strength analyzer",
		"Offline analysis. Nothing leaves your machine.",
		theme.StatusPill("Offline: local only", theme.PillOk),
	)

	passwordInput := widget.NewPasswordEntry()
	passwordInput.PlaceHolder = "Enter password to check"

	strengthMeter := theme.SegmentedStrengthMeterFlex(0)
	strengthLabel := canvas.NewText("", theme.ColorTextSecondary)
	strengthLabel.TextSize = 11
	strengthLabel.TextStyle = fyne.TextStyle{Monospace: true}

	lengthVal := canvas.NewText("-", theme.ColorTextPrimary)
	lengthVal.TextSize = 13
	charsetVal := canvas.NewText("-", theme.ColorTextPrimary)
	charsetVal.TextSize = 13
	entropyVal := canvas.NewText("-", theme.ColorTextPrimary)
	entropyVal.TextSize = 13
	crackVal := canvas.NewText("-", theme.ColorTextPrimary)
	crackVal.TextSize = 13

	issuesBox := container.NewVBox()

	analysisContainer := container.NewVBox(
		container.NewBorder(nil, nil, nil, strengthLabel, strengthMeter),
	)

	divider := canvas.NewRectangle(theme.ColorLine1)
	divider.SetMinSize(fyne.NewSize(0, 1))

	kvTable := theme.KeyValueTable([]theme.KVItem{
		{Key: "Length", Value: "-"},
		{Key: "Character set", Value: "-"},
		{Key: "Estimated entropy", Value: "-"},
		{Key: "Brute-force estimate", Value: "-"},
	})

	hintText := canvas.NewText("Start typing to analyze password strength.", theme.ColorFg2)
	hintText.TextSize = 12
	hintText.TextStyle = fyne.TextStyle{Monospace: true}

	updateStrength := func(value string) {
		result := strength.Analyze(value, storedVaultPasswords(ns.appState))

		if result.EasterEggMode {
			level := int(result.Score) + 1
			if level > 5 {
				level = 5
			}
			newMeter := theme.SegmentedStrengthMeterFlex(level)
			strengthLabel.Text = "Password Game Mode"
			strengthLabel.Refresh()

			analysisContainer.Objects = []fyne.CanvasObject{
				container.NewBorder(nil, nil, nil, strengthLabel, newMeter),
				divider,
				NewEasterEggPanel(result.EasterEggRules),
			}
			analysisContainer.Refresh()
			return
		}

		level := 0
		if value != "" {
			level = int(result.Score) + 1
			if level > 5 {
				level = 5
			}
		}

		newMeter := theme.SegmentedStrengthMeterFlex(level)

		if value == "" {
			strengthLabel.Text = ""
			strengthLabel.Refresh()
			analysisContainer.Objects = []fyne.CanvasObject{
				container.NewBorder(nil, nil, nil, strengthLabel, newMeter),
				hintText,
			}
		} else {
			strengthLabel.Text = result.ScoreLabel
			strengthLabel.Refresh()
			// Build character set description
			var sets []string
			for _, r := range value {
				if r >= 'A' && r <= 'Z' {
					sets = append(sets, "A-Z")
					break
				}
			}
			for _, r := range value {
				if r >= 'a' && r <= 'z' {
					sets = append(sets, "a-z")
					break
				}
			}
			for _, r := range value {
				if r >= '0' && r <= '9' {
					sets = append(sets, "0-9")
					break
				}
			}
			hasSpecial := false
			for _, r := range value {
				if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
					hasSpecial = true
					break
				}
			}
			if hasSpecial {
				sets = append(sets, "symbols")
			}

			charsetStr := "-"
			if len(sets) > 0 {
				charsetStr = ""
				for i, s := range sets {
					if i > 0 {
						charsetStr += " : "
					}
					charsetStr += s
				}
			}

			lengthVal.Text = fmt.Sprintf("%d characters", len(value))
			lengthVal.Refresh()
			charsetVal.Text = charsetStr
			charsetVal.Refresh()
			entropyVal.Text = result.CrackTime
			entropyVal.Refresh()
			crackVal.Text = result.CrackTime
			crackVal.Refresh()

			kvTable = theme.KeyValueTable([]theme.KVItem{
				{Key: "Length", Value: fmt.Sprintf("%d characters", len(value))},
				{Key: "Character set", Value: charsetStr},
				{Key: "Estimated entropy", Value: fmt.Sprintf("%.0f bits", result.Entropy)},
				{Key: "Brute-force estimate", Value: result.CrackTime},
			})

			issuesBox.Objects = nil
			if len(result.Issues) > 0 {
				issuesBox.Objects = []fyne.CanvasObject{NewIssuesList(result.Issues)}
			}
			issuesBox.Refresh()

			analysisContainer.Objects = []fyne.CanvasObject{
				container.NewBorder(nil, nil, nil, strengthLabel, newMeter),
				divider,
				kvTable,
				issuesBox,
			}
		}
		analysisContainer.Refresh()
	}

	passwordInput.OnChanged = updateStrength
	updateStrength(passwordInput.Text)

	inputCard := theme.CardWithHeader("INPUT", "Password to analyze", nil,
		container.NewVBox(passwordInput),
	)

	analysisCard := theme.CardWithHeader("ANALYSIS", "Strength result", nil, analysisContainer)

	return container.NewVBox(header, inputCard, analysisCard)
}

func (ns *NavigationState) createItemsView() fyne.CanvasObject {
	addItemBtn := theme.CreatePrimaryButtonWithIcon("Add item", theme.IconPlus, func() {
		ns.switchView(NavViewAddItem)
	})

	countLabel := canvas.NewText("Loading…", theme.ColorTextSecondary)
	countLabel.TextSize = 13

	// Track all decrypted cards for filtering
	type cardEntry struct {
		service  string
		username string
		card     fyne.CanvasObject
	}
	var allCards []cardEntry

	itemsContainer := container.NewVBox()
	loadingText := canvas.NewText("Loading vault items...", theme.ColorFg2)
	loadingText.TextSize = 13
	itemsContainer.Objects = []fyne.CanvasObject{container.NewCenter(loadingText)}

	// Search bar
	searchEntry := widget.NewEntry()
	searchEntry.PlaceHolder = "Search items…"
	searchEntry.OnChanged = func(query string) {
		q := strings.ToLower(query)
		var filtered []fyne.CanvasObject
		for _, ce := range allCards {
			if q == "" || strings.Contains(strings.ToLower(ce.service), q) || strings.Contains(strings.ToLower(ce.username), q) {
				filtered = append(filtered, ce.card)
			}
		}
		if len(filtered) == 0 && query != "" {
			noMatch := canvas.NewText("No items match “"+query+"”", theme.ColorFg2)
			noMatch.TextSize = 13
			itemsContainer.Objects = []fyne.CanvasObject{container.NewCenter(noMatch)}
		} else {
			itemsContainer.Objects = filtered
		}
		itemsContainer.Refresh()
	}

	go func() {
		ns.appState.Mu.Lock()
		defer ns.appState.Mu.Unlock()

		vaultFile := app.GetVaultPath(ns.appState.CurrentVault)
		entries, err := app.ReadVault(vaultFile, ns.appState.MasterPassword)
		if err != nil {
			fyne.Do(func() {
				errText := canvas.NewText("Failed to read vault: "+err.Error(), theme.ColorDanger)
				errText.TextSize = 13
				itemsContainer.Objects = []fyne.CanvasObject{errText}
				itemsContainer.Refresh()
				countLabel.Text = "Error"
				countLabel.Refresh()
			})
			return
		}

		fyne.Do(func() {
			if len(entries) == 0 {
				// Rich empty state
				vaultIco := canvas.NewImageFromResource(theme.IconVault)
				vaultIco.SetMinSize(fyne.NewSize(40, 40))
				emptyTitle := canvas.NewText("No items yet", theme.ColorTextPrimary)
				emptyTitle.TextSize = 15
				emptyTitle.TextStyle = fyne.TextStyle{Bold: true}
				emptySubtitle := canvas.NewText("Add your first credential to this vault.", theme.ColorFg2)
				emptySubtitle.TextSize = 12
				addFirstBtn := theme.CreatePrimaryButton("Add item", func() {
					ns.switchView(NavViewAddItem)
				})
				emptyState := container.NewCenter(container.NewVBox(
					container.NewCenter(vaultIco),
					container.NewCenter(emptyTitle),
					container.NewCenter(emptySubtitle),
					container.NewCenter(addFirstBtn),
				))
				itemsContainer.Objects = []fyne.CanvasObject{emptyState}
				itemsContainer.Refresh()
				countLabel.Text = "No items"
				countLabel.Refresh()
				return
			}

			allCards = nil
			var cards []fyne.CanvasObject
			for _, entry := range entries {
				ss, err := app.Decapsulate(entry.KyberCiphertext, ns.appState.PrivateKey)
				if err != nil {
					continue
				}
				plaintext, err := app.DecryptAES256GCM(entry.Nonce, entry.Ciphertext, ss)
				if err != nil {
					continue
				}
				card := createVaultItemCard(0, entry, plaintext, ns.window, ns.app, ns.appState)
				allCards = append(allCards, cardEntry{
					service:  entry.Service,
					username: entry.Username,
					card:     card,
				})
				cards = append(cards, card)
			}

			itemsContainer.Objects = cards
			itemsContainer.Refresh()

			n := len(allCards)
			if n == 1 {
				countLabel.Text = "1 item"
			} else {
				countLabel.Text = fmt.Sprintf("%d items", n)
			}
			countLabel.Refresh()
		})
	}()

	// Build header manually to include a live countLabel
	eyebrow := canvas.NewText("PASSQUANTUM / "+ns.appState.CurrentVault, theme.ColorFg2)
	eyebrow.TextSize = 10
	eyebrow.TextStyle = fyne.TextStyle{Monospace: true}
	headerTitle := canvas.NewText("Vault items", theme.ColorTextPrimary)
	headerTitle.TextSize = 22
	headerTitle.TextStyle = fyne.TextStyle{Bold: true}
	headerLeft := container.NewVBox(eyebrow, headerTitle, countLabel)
	headerRow := container.NewBorder(nil, nil, headerLeft, container.NewCenter(addItemBtn))
	headerDivider := canvas.NewRectangle(theme.ColorLine1)
	headerDivider.SetMinSize(fyne.NewSize(0, 1))
	header := container.NewVBox(
		container.New(layout.NewCustomPaddedLayout(0, theme.Space4, 0, 0), headerRow),
		headerDivider,
	)

	searchBg := canvas.NewRectangle(theme.ColorSidebarBg)
	searchBg.CornerRadius = theme.RadiusInput
	searchBorder := canvas.NewRectangle(color.Transparent)
	searchBorder.CornerRadius = theme.RadiusInput
	searchBorder.StrokeWidth = 1
	searchBorder.StrokeColor = theme.ColorLine2
	searchBorder.FillColor = color.Transparent
	searchRow := container.NewStack(searchBg, searchBorder, container.NewPadded(searchEntry))

	return container.NewVBox(header, searchRow, itemsContainer)
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
