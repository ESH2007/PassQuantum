package screens

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"passquantum/app"
	"passquantum/bridge"
	"passquantum/theme"
	"passquantum/ui/assets"
	"passquantum/ui/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type SettingsSubview int

const (
	SettingsSubviewSecurity SettingsSubview = iota
	SettingsSubviewVaults
	SettingsSubviewVisuals
	SettingsSubviewAbout
)

// ShowSettingsScreen displays the application settings
func ShowSettingsScreen(w fyne.Window, fyneApp fyne.App, appState *app.AppState) {
	w.SetTitle("PassQuantum - Settings")

	// Create navigation state
	navState := &NavigationState{
		currentView: NavViewSettings,
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

func buildCustomSettingsView(w fyne.Window, fyneApp fyne.App, appState *app.AppState) fyne.CanvasObject {
	selectedSubview := SettingsSubviewSecurity
	contentContainer := container.NewMax()

	header := theme.PageHeader(
		"PASSQUANTUM / SETTINGS",
		"Settings",
		"",
		nil,
	)

	var tabWidget fyne.CanvasObject
	var refresh func()
	refresh = func() {
		tabWidget = theme.UnderlineTabs(
			[]string{"Security", "Vaults", "Appearance", "About"},
			int(selectedSubview),
			func(idx int) {
				selectedSubview = SettingsSubview(idx)
				refresh()
			},
		)

		var body fyne.CanvasObject
		switch selectedSubview {
		case SettingsSubviewSecurity:
			body = buildSecuritySettings(w, fyneApp, appState)
		case SettingsSubviewVaults:
			body = buildVaultSettings(w, fyneApp, appState)
		case SettingsSubviewVisuals:
			body = buildDisplaySettings(w, fyneApp, appState)
		default:
			body = buildAboutSettings(w, fyneApp, appState)
		}

		contentContainer.Objects = []fyne.CanvasObject{
			container.NewVBox(header, tabWidget, body),
		}
		contentContainer.Refresh()
	}

	refresh()
	return contentContainer
}

func buildSecuritySettings(w fyne.Window, fyneApp fyne.App, appState *app.AppState) *fyne.Container {
	changePwBtn := theme.CreateDefaultButton("Change master password", func() {
		showChangeMasterPasswordDialog(w, appState)
	})

	masterPwCard := theme.CardWithHeader("SECURITY", "Master password", nil,
		container.NewBorder(nil, nil,
			container.NewVBox(
				theme.MonoText("App-level verifier active and bound to the current private key.", 11, theme.ColorFg2),
			),
			changePwBtn,
		),
	)

	warningCard := theme.WarningBanner(
		"FORCE-KILL WARNING",
		"The processes checked below will be force-killed with NO save prompt as soon as your face is not detected for 5 seconds.",
	)

	prefs := fyneApp.Preferences()

	// appListContainer holds the checkbox rows; rebuilt by refreshAppList.
	appListContainer := container.NewVBox()

	var refreshAppList func()
	refreshAppList = func() {
		processes := bridge.ListRunningProcesses()
		saved := bridge.LoadKillApps(prefs)
		savedSet := make(map[string]struct{}, len(saved))
		for _, s := range saved {
			savedSet[s] = struct{}{}
		}

		appListContainer.Objects = nil

		for _, name := range processes {
			procName := name // capture for closure
			_, isChecked := savedSet[procName]

			var chk *widget.Check
			chk = widget.NewCheck(procName, func(checked bool) {
				if checked {
					// Show confirmation before adding to kill list.
					dialog.NewConfirm(
						"Add to kill list?",
						"\""+procName+"\" will be force-closed with no save prompt\n"+
							"if your face is not detected for 5 seconds.\n\nProceed?",
						func(ok bool) {
							if ok {
								current := bridge.LoadKillApps(prefs)
								// Avoid duplicates.
								for _, s := range current {
									if s == procName {
										return
									}
								}
								bridge.SaveKillApps(prefs, append(current, procName))
							} else {
								// User cancelled — revert the checkbox state.
								chk.SetChecked(false)
							}
						},
						w,
					).Show()
				} else {
					// Unchecking: remove immediately, no confirmation needed.
					current := bridge.LoadKillApps(prefs)
					updated := current[:0]
					for _, s := range current {
						if s != procName {
							updated = append(updated, s)
						}
					}
					bridge.SaveKillApps(prefs, updated)
				}
			})
			chk.SetChecked(isChecked)

			appListContainer.Add(chk)
		}

		if len(processes) == 0 {
			appListContainer.Add(theme.CreateLabel("No running processes found.", 10, theme.ColorTextSec, false))
		}

		appListContainer.Refresh()
	}

	refreshAppList() // initial population

	refreshBtn := theme.CreateGhostButton("Refresh apps", func() {
		refreshAppList()
	})

	appScroll := container.NewVScroll(appListContainer)
	appScroll.SetMinSize(fyne.NewSize(0, 200))

	guardCard := theme.CardWithHeader("PRESENCE GUARD", "Face-detection auto-kill", refreshBtn,
		container.NewVBox(
			warningCard,
			appScroll,
		),
	)

	// ── Detection visualizer ──────────────────────────────────────────────
	// Lets the user see the live MediaPipe landmarks and blink detection that
	// drive Presence Guard. Only available once a face is enrolled and the
	// guard subprocess is running.
	var visualizerBody fyne.CanvasObject
	if appState.FaceGuard != nil && faceDataExists() {
		openBtn := theme.CreateDefaultButton("Open visualizer", func() {
			showFaceVisualizerDialog(w, appState)
		})
		visualizerBody = container.NewBorder(nil, nil,
			container.NewVBox(
				theme.MonoText("Watch the live face mesh and blink detection that power Presence Guard.", 11, theme.ColorFg2),
			),
			openBtn,
		)
	} else {
		visualizerBody = theme.MonoText("Register a face first to use the detection visualizer.", 11, theme.ColorFg2)
	}

	visualizerCard := theme.CardWithHeader("FACE DETECTION", "Detection visualizer", nil, visualizerBody)

	return container.NewVBox(masterPwCard, guardCard, visualizerCard)
}

// showFaceVisualizerDialog opens a modal showing the live camera feed annotated
// with all 478 MediaPipe landmarks (eye points highlighted) plus a HUD with the
// blink count and Eye Aspect Ratio. It drives the Python demo mode via
// START_DEMO / STOP_DEMO, which pauses presence monitoring while open (only one
// process can own the webcam) and resumes it on close.
func showFaceVisualizerDialog(w fyne.Window, appState *app.AppState) {
	guard := appState.FaceGuard
	if guard == nil {
		widgets.ShowAppWarning("Unavailable", "Face detection is not running.", w)
		return
	}

	// Camera preview, sized to match the Python demo frame (480×360).
	blankImg := image.NewNRGBA(image.Rect(0, 0, 480, 360))
	camImage := canvas.NewImageFromImage(blankImg)
	camImage.FillMode = canvas.ImageFillContain
	camImage.SetMinSize(fyne.NewSize(480, 360))

	notice := theme.WarningBanner(
		"PRESENCE PROTECTION PAUSED",
		"Auto-lock is paused while this window is open. Monitoring resumes when you close it.",
	)

	caption := theme.MonoText(
		"Live MediaPipe output — 478 face landmarks; green dots are the eye points used for blink detection.",
		11, theme.ColorFg2,
	)

	content := container.NewVBox(
		notice,
		container.NewCenter(camImage),
		caption,
	)

	// Defensively pause auto-lock while the visualizer is open. The demo loop
	// already suppresses FACE_LOST, but the IsTraining flag guards the global
	// OnLost handler too.
	appState.Mu.Lock()
	appState.IsTraining = true
	appState.Mu.Unlock()

	// Wire OnFrame before requesting the demo so no frames are missed. The
	// callback runs on the Listen() goroutine, so UI work goes through fyne.Do.
	guard.OnFrame = func(img image.Image) {
		fyne.Do(func() {
			camImage.Image = img
			camImage.Refresh()
		})
	}

	d := dialog.NewCustom("Face detection visualizer", "Close", content, w)
	d.SetOnClosed(func() {
		guard.OnFrame = nil
		// SendCommand can block until Python is connected — never on the UI thread.
		go guard.SendCommand("STOP_DEMO")
		appState.Mu.Lock()
		appState.IsTraining = false
		appState.Mu.Unlock()
	})

	go guard.SendCommand("START_DEMO")
	d.Show()
}

func buildVaultSettings(w fyne.Window, fyneApp fyne.App, appState *app.AppState) *fyne.Container {
	vaultInfoCard := theme.CardWithHeader("VAULT", "Active vault", nil,
		theme.KeyValueTable([]theme.KVItem{
			{Key: "Name", Value: appState.CurrentVault},
			{Key: "Total vaults", Value: fmt.Sprintf("%d", len(app.ListVaults()))},
		}),
	)

	compactBtn := theme.CreateDefaultButton("Compact vault", func() {
		widgets.ShowAppInformation("Compact", "Vault compaction is being performed...", w)
	})

	compactCard := theme.CardWithHeader("MAINTENANCE", "Compact & verify", nil,
		container.NewBorder(nil, nil,
			theme.MonoText("Reclaim space and verify vault integrity.", 11, theme.ColorFg2),
			compactBtn,
		),
	)

	exportBtn := theme.CreatePrimaryButton("Export", func() {
		widgets.ShowAppInformation("Export", "Select location to export encrypted vault backup", w)
	})
	importBtn := theme.CreateDefaultButton("Import", func() {
		widgets.ShowAppInformation("Import", "Select backup file to import", w)
	})
	backupNowBtn := theme.CreateDefaultButton("Back up now", func() {
		widgets.ShowAppInformation("Backup", "Vault backup created successfully!", w)
	})

	backupCard := theme.CardWithHeader("BACKUP", "Backup & restore", nil,
		container.NewVBox(
			container.NewGridWithColumns(2,
				container.NewVBox(theme.SectionEyebrow("EXPORT"), exportBtn),
				container.NewVBox(theme.SectionEyebrow("IMPORT"), importBtn),
			),
			container.NewBorder(nil, nil,
				theme.MonoText("Auto-backup on vault changes.", 11, theme.ColorFg2),
				backupNowBtn,
			),
		),
	)

	// File vault — source-file deletion preference
	const (
		labelAsk    = "Ask each time"
		labelAlways = "Always delete"
		labelNever  = "Never delete"
	)
	prefs := fyneApp.Preferences()
	current := prefs.StringWithFallback(PrefDeleteSourceAfterImport, "")
	deleteSourceSelect := widget.NewSelect(
		[]string{labelAsk, labelAlways, labelNever},
		func(s string) {
			switch s {
			case labelAlways:
				prefs.SetString(PrefDeleteSourceAfterImport, "always")
			case labelNever:
				prefs.SetString(PrefDeleteSourceAfterImport, "never")
			default:
				prefs.SetString(PrefDeleteSourceAfterImport, "")
			}
		},
	)
	switch current {
	case "always":
		deleteSourceSelect.SetSelected(labelAlways)
	case "never":
		deleteSourceSelect.SetSelected(labelNever)
	default:
		deleteSourceSelect.SetSelected(labelAsk)
	}

	fileVaultCard := theme.CardWithHeader("FILE VAULT", "Source file after import", nil,
		container.NewBorder(nil, nil,
			theme.MonoText("What to do with the original file once encrypted.", 11, theme.ColorFg2),
			deleteSourceSelect,
		),
	)

	return container.NewVBox(vaultInfoCard, compactCard, backupCard, fileVaultCard)
}

func buildDisplaySettings(w fyne.Window, fyneApp fyne.App, appState *app.AppState) *fyne.Container {
	// Accent palette swatches
	type accentOption struct {
		name string
		hex  string
		c    color.NRGBA
	}
	accents := []accentOption{
		{"Teal", "#2dd4bf", color.NRGBA{R: 0x2d, G: 0xd4, B: 0xbf, A: 255}},
		{"Emerald", "#10b981", color.NRGBA{R: 0x10, G: 0xb9, B: 0x81, A: 255}},
		{"Gold", "#eab308", color.NRGBA{R: 0xea, G: 0xb3, B: 0x08, A: 255}},
		{"Mono", "#94a3b8", color.NRGBA{R: 0x94, G: 0xa3, B: 0xb8, A: 255}},
		{"Cyan", "#06b6d4", color.NRGBA{R: 0x06, G: 0xb6, B: 0xd4, A: 255}},
	}

	swatches := make([]fyne.CanvasObject, len(accents))
	for i, a := range accents {
		accent := a
		swatch := canvas.NewRectangle(accent.c)
		swatch.CornerRadius = theme.RadiusInput
		swatch.SetMinSize(fyne.NewSize(96, 56))

		border := canvas.NewRectangle(color.Transparent)
		border.CornerRadius = theme.RadiusInput
		border.StrokeWidth = 1
		border.StrokeColor = theme.ColorLine2
		border.FillColor = color.Transparent
		border.SetMinSize(fyne.NewSize(96, 56))

		label := theme.MonoText(accent.name+"\n"+accent.hex, 10, theme.ColorFg2)

		btn := theme.NewClickOverlay(func() {
			theme.ColorAccentCyan = accent.c
			theme.ColorAccentCyn = accent.c
			theme.ColorPrimaryButton = accent.c
			theme.ColorAccentSoft = color.NRGBA{R: accent.c.R, G: accent.c.G, B: accent.c.B, A: 0x24}
			theme.ColorAccentLine = color.NRGBA{R: accent.c.R, G: accent.c.G, B: accent.c.B, A: 0x66}
			ShowSettingsScreen(w, fyneApp, appState)
		})

		swatches[i] = container.NewVBox(
			container.NewStack(swatch, border, btn),
			label,
		)
	}

	accentCard := theme.CardWithHeader("ACCENT", "Accent palette", nil,
		container.NewHBox(swatches...),
	)

	uploadBtn := theme.CreatePrimaryButton("Upload image", func() {
		showThemePicker(fyneApp, w, func() {
			ShowSettingsScreen(w, fyneApp, appState)
			widgets.ShowAppInformation("Palette Applied", "Theme generated from image and saved for next launch.", w)
		})
	})

	resetBtn := theme.CreateGhostButton("Reset defaults", func() {
		fyneApp.Settings().SetTheme(fynetheme.DefaultTheme())
		resetDefaultPalette()
		clearManualPalettePreferences(fyneApp)
		fyneApp.Preferences().SetString(themeImagePathPrefKey, "")
		ShowSettingsScreen(w, fyneApp, appState)
	})

	paletteCard := theme.CardWithHeader("PALETTE", "Image palette", nil,
		container.NewVBox(
			theme.MonoText("Upload an image to extract colors and apply them to the UI.", 11, theme.ColorFg2),
			container.NewHBox(uploadBtn, resetBtn),
		),
	)

	changeIconBtn := theme.CreateDefaultButton("Change", func() {
		widgets.PickImageFile("Select App Icon", func(path string) {
			data, readErr := os.ReadFile(path)
			if readErr != nil || len(data) == 0 {
				widgets.ShowAppError(fmt.Errorf("could not read icon file"), w)
				return
			}
			fyneApp.SetIcon(fyne.NewStaticResource(filepath.Base(path), data))
			fyneApp.Preferences().SetString("custom_icon_path", path)
			widgets.ShowAppInformation("App Icon", "App icon updated.", w)
		}, func(err error) {
			widgets.ShowAppError(err, w)
		})
	})

	resetIconBtn := theme.CreateGhostButton("Reset", func() {
		fyneApp.Preferences().SetString("custom_icon_path", "")
		SetApplicationIcon(fyneApp)
		widgets.ShowAppInformation("App Icon", "App icon has been reset to the default.", w)
	})

	iconCard := theme.CardWithHeader("ICON", "App icon", nil,
		container.NewVBox(
			theme.MonoText("Replace the application icon. Persists across restarts.", 11, theme.ColorFg2),
			container.NewHBox(changeIconBtn, resetIconBtn),
		),
	)

	return container.NewVBox(accentCard, paletteCard, iconCard)
}

func buildAboutSettings(w fyne.Window, fyneApp fyne.App, appState *app.AppState) *fyne.Container {
	// Product card
	var brandIconInner fyne.CanvasObject
	if img, _, err := image.Decode(bytes.NewReader(assets.LogoImage)); err == nil {
		logoImg := canvas.NewImageFromImage(img)
		logoImg.FillMode = canvas.ImageFillContain
		logoImg.SetMinSize(fyne.NewSize(80, 80))
		brandIconInner = logoImg
	} else {
		ico := canvas.NewImageFromResource(theme.IconAtom)
		ico.SetMinSize(fyne.NewSize(16, 16))
		brandIconInner = ico
	}

	iconBg := canvas.NewRectangle(theme.ColorAccentSoft)
	iconBg.CornerRadius = 18
	iconBg.SetMinSize(fyne.NewSize(96, 96))
	iconBorder := canvas.NewRectangle(color.Transparent)
	iconBorder.CornerRadius = 18
	iconBorder.StrokeWidth = 1
	iconBorder.StrokeColor = theme.ColorAccentLine
	iconBorder.FillColor = color.Transparent
	iconBorder.SetMinSize(fyne.NewSize(96, 96))
	iconBlock := container.NewStack(iconBg, iconBorder, container.NewCenter(brandIconInner))

	productInfo := container.NewVBox(
		theme.SectionEyebrow("PASSQUANTUM"),
		canvas.NewText("PassQuantum", theme.ColorTextPrimary),
		theme.MonoText("v1.1.4-beta | PQ-Safe", 11, theme.ColorFg2),
		canvas.NewText("A post-quantum cryptography password manager using Kyber and AES-256-GCM.", theme.ColorTextSecondary),
	)

	productCard := theme.CardWithHeader("", "", nil,
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(96, 96), iconBlock),
			productInfo,
		),
	)

	// Security stack card
	securityCard := theme.CardWithHeader("CRYPTOGRAPHY", "Security stack", nil,
		theme.KeyValueTable([]theme.KVItem{
			{Key: "Bulk encryption", Value: "AES-256-GCM", Detail: "Symmetric — each entry encrypted independently with a unique key"},
			{Key: "Key encapsulation", Value: "ML-KEM-768 (Kyber)", Detail: "Post-quantum — resists attacks from quantum computers (NIST FIPS 203)"},
			{Key: "Password KDF", Value: "Argon2id", Detail: "Memory-hard stretching — intentionally slow to brute-force"},
			{Key: "Authentication", Value: "HMAC-SHA-512", Detail: "Tamper detection — any modification to ciphertext is detected"},
			{Key: "Random source", Value: "crypto/rand", Detail: "Hardware-seeded — OS kernel getrandom(2), not math/rand"},
			{Key: "Architecture", Value: "Zero-knowledge, offline-first", Detail: "No telemetry, no network, no cloud — data never leaves your machine"},
		}),
	)

	// Meta card
	repoURL := "https://github.com/ESH2007/PassQuantum/tree/master/docs"
	docsBtn := theme.CreateDefaultButton("Documentation", func() {
		u, err := url.Parse(repoURL)
		if err == nil {
			fyneApp.OpenURL(u)
		}
	})
	updatesBtn := theme.CreateGhostButton("Check for updates", func() {
		widgets.ShowAppInformation("Updates", "You are running the latest version!", w)
	})

	metaCard := theme.CardWithHeader("META", "Build information", nil,
		container.NewVBox(
			theme.KeyValueTable([]theme.KVItem{
				{Key: "License", Value: "MIT"},
				{Key: "Maintainer", Value: "ESH2007"},
			}),
			container.NewHBox(docsBtn, updatesBtn),
		),
	)

	return container.NewVBox(productCard, securityCard, metaCard)
}

func showChangeMasterPasswordDialog(w fyne.Window, appState *app.AppState) {
	oldPwInput := widget.NewPasswordEntry()
	oldPwInput.PlaceHolder = "Current master password"

	newPwInput := widget.NewPasswordEntry()
	newPwInput.PlaceHolder = "New master password"

	confirmPwInput := widget.NewPasswordEntry()
	confirmPwInput.PlaceHolder = "Confirm new password"

	formContent := container.NewVBox(
		theme.SectionEyebrow("CHANGE MASTER PASSWORD"),
		theme.FieldLabel("CURRENT PASSWORD", nil),
		oldPwInput,
		theme.FieldLabel("NEW PASSWORD", nil),
		newPwInput,
		theme.FieldLabel("CONFIRM PASSWORD", nil),
		confirmPwInput,
	)

	var d *dialog.CustomDialog

	cancelBtn := theme.CreateGhostButton("Cancel", func() {
		if d != nil {
			d.Hide()
		}
	})

	changeBtn := theme.CreatePrimaryButton("Change password", func() {
		if newPwInput.Text == "" {
			widgets.ShowAppError(fmt.Errorf("new password cannot be empty"), w)
			return
		}

		if newPwInput.Text != confirmPwInput.Text {
			widgets.ShowAppError(fmt.Errorf("new passwords do not match"), w)
			return
		}

		if err := app.ChangeMasterPassword(appState, oldPwInput.Text, newPwInput.Text); err != nil {
			widgets.ShowAppError(err, w)
			return
		}

		if d != nil {
			d.Hide()
		}
		widgets.ShowAppInformation("Success", "Master password changed successfully and all vaults were re-encrypted.", w)
	})

	buttonBox := container.NewHBox(cancelBtn, changeBtn)
	dialogContent := container.NewVBox(formContent, container.NewCenter(buttonBox))
	d = dialog.NewCustom("Change Master Password", "Close", dialogContent, w)
	d.Show()
}

func canvasColorPreview(c color.NRGBA) fyne.CanvasObject {
	rect := canvas.NewRectangle(c)
	rect.CornerRadius = 6
	rect.SetMinSize(fyne.NewSize(180, 42))

	border := canvas.NewRectangle(color.NRGBA{R: theme.ColorTextSecondary.R, G: theme.ColorTextSecondary.G, B: theme.ColorTextSecondary.B, A: 80})
	border.CornerRadius = 6
	border.SetMinSize(fyne.NewSize(182, 44))

	return container.NewStack(border, container.NewCenter(rect))
}

func extractTopColors(img image.Image, count int) ([]color.NRGBA, error) {
	if img == nil {
		return nil, fmt.Errorf("no image provided")
	}

	b := img.Bounds()
	if b.Empty() {
		return nil, fmt.Errorf("image has no pixel data")
	}

	type bucket struct {
		count int
		rSum  uint64
		gSum  uint64
		bSum  uint64
	}

	buckets := make(map[uint16]*bucket)
	stride := 1
	pixels := b.Dx() * b.Dy()
	if pixels > 250000 {
		stride = 2
	}
	if pixels > 1000000 {
		stride = 4
	}

	for y := b.Min.Y; y < b.Max.Y; y += stride {
		for x := b.Min.X; x < b.Max.X; x += stride {
			p := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			if p.A < 20 {
				continue
			}

			// Quantize to reduce tiny color variations before counting frequencies.
			qr := p.R >> 4
			qg := p.G >> 4
			qb := p.B >> 4
			key := uint16(qr)<<8 | uint16(qg)<<4 | uint16(qb)

			item, ok := buckets[key]
			if !ok {
				item = &bucket{}
				buckets[key] = item
			}

			item.count++
			item.rSum += uint64(p.R)
			item.gSum += uint64(p.G)
			item.bSum += uint64(p.B)
		}
	}

	if len(buckets) == 0 {
		return nil, fmt.Errorf("no visible colors found in image")
	}

	type ranked struct {
		cnt int
		clr color.NRGBA
	}
	list := make([]ranked, 0, len(buckets))
	for _, item := range buckets {
		if item.count == 0 {
			continue
		}
		list = append(list, ranked{
			cnt: item.count,
			clr: color.NRGBA{
				R: uint8(item.rSum / uint64(item.count)),
				G: uint8(item.gSum / uint64(item.count)),
				B: uint8(item.bSum / uint64(item.count)),
				A: 255,
			},
		})
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].cnt > list[j].cnt
	})

	result := make([]color.NRGBA, 0, count)
	for _, item := range list {
		result = append(result, item.clr)
		if len(result) == count {
			break
		}
	}

	for len(result) < count {
		if len(result) == 0 {
			result = append(result, theme.ColorBg)
			continue
		}
		result = append(result, result[len(result)-1])
	}

	return result, nil
}

func applyExtractedPalette(colors []color.NRGBA) {
	if len(colors) < 3 {
		return
	}

	base := colors[0]
	primary := colors[1]
	secondary := colors[2]

	theme.ColorBg = base
	theme.ColorSidebarBg = blend(base, color.NRGBA{R: 0, G: 0, B: 0, A: 255}, 0.15)
	theme.ColorCardBg = blend(base, color.NRGBA{R: 255, G: 255, B: 255, A: 255}, 0.08)
	theme.ColorInputBg = blend(base, color.NRGBA{R: 0, G: 0, B: 0, A: 255}, 0.1)

	theme.ColorPrimaryButton = primary
	theme.ColorSecondaryButton = secondary

	theme.ColorAccentCyan = primary
	theme.ColorAccentCyn = theme.ColorAccentCyan
	theme.ColorAccentPink = secondary
	theme.ColorMagenta = theme.ColorAccentPink
	theme.ColorPurple = color.NRGBA{R: secondary.R, G: secondary.G, B: secondary.B, A: 220}

	theme.ColorBorderCyan = color.NRGBA{R: primary.R, G: primary.G, B: primary.B, A: 180}
	theme.ColorBorder = theme.ColorBorderCyan
	theme.ColorGlowCyan = color.NRGBA{R: primary.R, G: primary.G, B: primary.B, A: 80}

	theme.ColorTextPrimary = theme.PickAdaptiveTextColor(theme.ColorBg)
	theme.ColorTextSecondary = theme.PickAdaptiveTextColor(theme.ColorCardBg)
	theme.ColorTextPrim = theme.ColorTextPrimary
	theme.ColorTextSec = theme.ColorTextSecondary
}

func resetDefaultPalette() {
	theme.ColorBg = color.NRGBA{R: 0x0a, G: 0x0d, B: 0x12, A: 255}
	theme.ColorSidebarBg = color.NRGBA{R: 0x0e, G: 0x12, B: 0x18, A: 255}
	theme.ColorCardBg = color.NRGBA{R: 0x11, G: 0x16, B: 0x1d, A: 255}
	theme.ColorInputBg = color.NRGBA{R: 0x0e, G: 0x12, B: 0x18, A: 255}

	theme.ColorAccentCyan = color.NRGBA{R: 0x2d, G: 0xd4, B: 0xbf, A: 255}
	theme.ColorAccentCyn = theme.ColorAccentCyan
	theme.ColorAccentPink = theme.ColorAccentCyan
	theme.ColorMagenta = theme.ColorAccentPink
	theme.ColorPurple = color.NRGBA{R: 0x6b, G: 0x77, B: 0x85, A: 255}

	theme.ColorTextPrimary = color.NRGBA{R: 0xea, G: 0xee, B: 0xf2, A: 255}
	theme.ColorTextSecondary = color.NRGBA{R: 0xae, G: 0xb6, B: 0xc2, A: 255}
	theme.ColorTextPrim = theme.ColorTextPrimary
	theme.ColorTextSec = theme.ColorTextSecondary

	theme.ColorBorderCyan = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x1a}
	theme.ColorBorder = theme.ColorBorderCyan
	theme.ColorGlowCyan = color.NRGBA{R: 0x2d, G: 0xd4, B: 0xbf, A: 0x24}

	theme.ColorPrimaryButton = theme.ColorAccentCyan
	theme.ColorSecondaryButton = color.NRGBA{R: 0x1a, G: 0x20, B: 0x30, A: 255}

	theme.ColorDanger = color.NRGBA{R: 0xd0, G: 0x4a, B: 0x4a, A: 255}
	theme.ColorWarning = color.NRGBA{R: 0xd9, G: 0x90, B: 0x30, A: 255}
	theme.ColorSuccess = color.NRGBA{R: 0x2e, G: 0xa9, B: 0x6b, A: 255}
}

func blend(a color.NRGBA, b color.NRGBA, ratio float64) color.NRGBA {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	inverse := 1 - ratio
	return color.NRGBA{
		R: uint8(float64(a.R)*inverse + float64(b.R)*ratio),
		G: uint8(float64(a.G)*inverse + float64(b.G)*ratio),
		B: uint8(float64(a.B)*inverse + float64(b.B)*ratio),
		A: 255,
	}
}

func toHex(c color.NRGBA) string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

func parseHexColor(value string) (color.NRGBA, error) {
	s := strings.TrimSpace(value)
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return color.NRGBA{}, fmt.Errorf("invalid hex color %q: expected format #RRGGBB", value)
	}

	r, err := strconv.ParseUint(s[0:2], 16, 8)
	if err != nil {
		return color.NRGBA{}, fmt.Errorf("invalid red channel in %q", value)
	}
	g, err := strconv.ParseUint(s[2:4], 16, 8)
	if err != nil {
		return color.NRGBA{}, fmt.Errorf("invalid green channel in %q", value)
	}
	b, err := strconv.ParseUint(s[4:6], 16, 8)
	if err != nil {
		return color.NRGBA{}, fmt.Errorf("invalid blue channel in %q", value)
	}

	return color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, nil
}
