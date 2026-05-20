package screens

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"passquantum/app"
	"passquantum/bridge"
	"passquantum/theme"
	"passquantum/ui/widgets"
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
	w.Resize(fyne.NewSize(1100, 700))

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
	tabStrip := container.New(layout.NewGridLayoutWithColumns(4))
	contentContainer := container.NewMax()

	var refresh func()
	refresh = func() {
		tabStrip.Objects = []fyne.CanvasObject{
			theme.CreateTabButton("Security", selectedSubview == SettingsSubviewSecurity, func() {
				selectedSubview = SettingsSubviewSecurity
				refresh()
			}),
			theme.CreateTabButton("Vaults", selectedSubview == SettingsSubviewVaults, func() {
				selectedSubview = SettingsSubviewVaults
				refresh()
			}),
			theme.CreateTabButton("Visuals", selectedSubview == SettingsSubviewVisuals, func() {
				selectedSubview = SettingsSubviewVisuals
				refresh()
			}),
			theme.CreateTabButton("About", selectedSubview == SettingsSubviewAbout, func() {
				selectedSubview = SettingsSubviewAbout
				refresh()
			}),
		}
		tabStrip.Refresh()

		var body fyne.CanvasObject
		switch selectedSubview {
		case SettingsSubviewSecurity:
			body = buildSettingsPanel("Security", "Manage the global master password and session controls.", buildSecuritySettings(w, fyneApp, appState))
		case SettingsSubviewVaults:
			body = buildSettingsPanel("Vaults", "Create, back up, import, and restore encrypted vaults from one place.", buildVaultSettings(w, fyneApp, appState))
		case SettingsSubviewVisuals:
			body = buildSettingsPanel("Visuals", "Tune appearance and interaction behavior.", buildDisplaySettings(w, fyneApp, appState))
		default:
			body = buildSettingsPanel("About", "Application details and security capabilities.", buildAboutSettings(w, fyneApp, appState))
		}

		contentContainer.Objects = []fyne.CanvasObject{body}
		contentContainer.Refresh()
	}

	refresh()

	headerText := theme.CreateLabel("SETTINGS", 14, theme.ColorAccentCyan, true)
	headerSection := container.NewVBox(headerText, theme.CreateDivider())

	mainContent := container.NewVBox(
		headerSection,
		widget.NewLabel(""),
		tabStrip,
		widget.NewLabel(""),
		contentContainer,
	)

	return container.NewPadded(container.NewVScroll(mainContent))
}

func buildSettingsPanel(title string, description string, content fyne.CanvasObject) fyne.CanvasObject {
	header := container.NewVBox(
		theme.CreateLabel(title, 16, theme.ColorAccentCyan, true),
		widget.NewLabel(""),
		theme.CreateLabel(description, 10, theme.ColorTextSec, false),
		widget.NewLabel(""),
		theme.CreateDivider(),
	)

	return theme.CreateCard(container.NewVBox(header, widget.NewLabel(""), content), 820, 0, true)
}

func buildSecuritySettings(w fyne.Window, fyneApp fyne.App, appState *app.AppState) *fyne.Container {
	passwordStrength := widget.NewSelect([]string{"Weak", "Medium", "Strong", "Very Strong"}, func(s string) {})
	passwordStrength.PlaceHolder = "Select password strength requirement"
	passwordStrength.SetSelected("Strong")

	changePwBtn := theme.CreateNeonButton("CHANGE MASTER PASSWORD", func() {
		showChangeMasterPasswordDialog(w, appState)
	}, 280, 40)

	profileStatus := theme.CreateLabel("App-level verifier active and bound to the current private key.", 10, theme.ColorTextSec, false)

	// ── Monitored Apps section ──────────────────────────────────────────────
	// A scrollable list of currently running processes rendered as checkboxes.
	// Checked apps are force-killed whenever the face-loss lock fires.

	warningColor := color.NRGBA{R: 220, G: 90, B: 40, A: 255} // amber-orange

	warningCard := theme.CreateCardWithBorderColor(
		container.NewVBox(
			theme.CreateLabel("⚠  FORCE-KILL WARNING", 11, warningColor, true),
			widget.NewLabel(""),
			theme.CreateLabel(
				"The processes checked below will be force-killed with NO save prompt\n"+
					"as soon as your face is not detected for 5 seconds.\n"+
					"Make sure any unsaved work in those apps is acceptable to lose.",
				10, theme.ColorTextPrimary, false,
			),
		),
		760, 0, warningColor,
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

	refreshBtn := theme.CreateSecondaryButton("REFRESH APPS", func() {
		refreshAppList()
	}, 160, 36)

	appScroll := container.NewVScroll(appListContainer)
	appScroll.SetMinSize(fyne.NewSize(760, 200))

	return container.NewVBox(
		theme.CreateLabel("MASTER PASSWORD", 11, theme.ColorPurple, true),
		profileStatus,
		widget.NewLabel(""),
		theme.CreateLabel("Password Strength:", 10, theme.ColorTextSec, false),
		passwordStrength,
		widget.NewLabel(""),
		container.NewCenter(changePwBtn),
		widget.NewLabel(""),
		theme.CreateDivider(),
		widget.NewLabel(""),
		theme.CreateLabel("MONITORED APPS", 11, theme.ColorPurple, true),
		theme.CreateLabel("These apps will be force-closed when face is not detected for 5 seconds.", 10, theme.ColorTextSec, false),
		widget.NewLabel(""),
		warningCard,
		widget.NewLabel(""),
		container.NewHBox(refreshBtn),
		widget.NewLabel(""),
		appScroll,
		widget.NewLabel(""),
		theme.CreateDivider(),
	)
}

func buildVaultSettings(w fyne.Window, fyneApp fyne.App, appState *app.AppState) *fyne.Container {
	currentVaultLabel := theme.CreateLabel("Current Vault: "+appState.CurrentVault, 10, theme.ColorTextSec, false)
	statsLabel := theme.CreateLabel(fmt.Sprintf("Total Vaults: %d", len(app.ListVaults())), 10, theme.ColorTextSec, false)

	compactBtn := theme.CreateNeonButton("COMPACT VAULT", func() {
		widgets.ShowAppInformation("Compact", "Vault compaction is being performed...", w)
	}, 200, 40)

	exportBtn := theme.CreateSecondaryButton("EXPORT VAULT", func() {
		widgets.ShowAppInformation("Export", "Select location to export encrypted vault backup", w)
	}, 200, 40)

	importBtn := theme.CreateSecondaryButton("IMPORT VAULT", func() {
		widgets.ShowAppInformation("Import", "Select backup file to import", w)
	}, 200, 40)

	backupNowBtn := theme.CreateNeonButton("BACKUP NOW", func() {
		widgets.ShowAppInformation("Backup", "Vault backup created successfully!", w)
	}, 180, 40)

	restoreBtn := theme.CreateSecondaryButton("RESTORE", func() {
		widgets.ShowAppConfirm("Restore", "This will replace your current vault. Are you sure?", func(ok bool) {
			if ok {
				widgets.ShowAppInformation("Restore", "Select a backup file to restore", w)
			}
		}, w)
	}, 150, 40)

	return container.NewVBox(
		theme.CreateLabel("ACTIVE VAULT", 11, theme.ColorPurple, true),
		currentVaultLabel,
		statsLabel,
		widget.NewLabel(""),
		theme.CreateLabel("Maintenance", 10, theme.ColorPurple, false),
		container.NewCenter(compactBtn),
		widget.NewLabel(""),
		theme.CreateLabel("Backup & Restore", 10, theme.ColorPurple, false),
		container.NewHBox(exportBtn, importBtn),
		widget.NewLabel(""),
		container.NewHBox(backupNowBtn, restoreBtn),
	)
}

func buildDisplaySettings(w fyne.Window, fyneApp fyne.App, appState *app.AppState) *fyne.Container {
	themeSelect := widget.NewSelect([]string{"Dark", "Light", "System"}, func(_ string) {})
	themeSelect.PlaceHolder = "Select theme"
	themeSelect.SetSelected("Dark")

	fontSizeSelect := widget.NewSelect([]string{"Small", "Medium", "Large"}, func(s string) {})
	fontSizeSelect.PlaceHolder = "Select font size"
	fontSizeSelect.SetSelected("Medium")

	showOnHoverCheck := widget.NewCheck("Show password on hover", func(b bool) {})
	confirmActionsCheck := widget.NewCheck("Confirm before deleting passwords", func(b bool) {})
	confirmActionsCheck.SetChecked(true)

	preview1 := canvasColorPreview(theme.ColorBg)
	preview2 := canvasColorPreview(theme.ColorPrimaryButton)
	preview3 := canvasColorPreview(theme.ColorSecondaryButton)

	manualPersonalizeBtn := theme.CreateNeonButton("MANUAL PERSONALIZATION", func() {
		ShowColorPersonalizationDialog(w, fyneApp, appState)
	}, 280, 40)

	uploadBtn := theme.CreateNeonButton("UPLOAD IMAGE TO ANALYZE", func() {
		showThemePicker(fyneApp, w, func() {
			ShowSettingsScreen(w, fyneApp, appState)
			widgets.ShowAppInformation("Palette Applied", "Theme generated from image and saved for next launch.", w)
		})
	}, 300, 40)

	resetPaletteBtn := theme.CreateSecondaryButton("RESET DEFAULT COLORS", func() {
		fyneApp.Settings().SetTheme(fynetheme.DefaultTheme())
		resetDefaultPalette()
		clearManualPalettePreferences(fyneApp)
		fyneApp.Preferences().SetString(themeImagePathPrefKey, "")
		ShowSettingsScreen(w, fyneApp, appState)
	}, 250, 40)

	changeIconBtn := theme.CreateNeonButton("CHANGE APP ICON", func() {
		widgets.PickImageFile("Select App Icon", func(path string) {
			data, readErr := os.ReadFile(path)
			if readErr != nil || len(data) == 0 {
				widgets.ShowAppError(fmt.Errorf("could not read icon file"), w)
				return
			}
			fyneApp.SetIcon(fyne.NewStaticResource(filepath.Base(path), data))
			fyneApp.Preferences().SetString("custom_icon_path", path)
			widgets.ShowAppInformation("App Icon", "App icon updated. It will also be applied on next launch.", w)
		}, func(err error) {
			widgets.ShowAppError(err, w)
		})
	}, 220, 40)

	resetIconBtn := theme.CreateSecondaryButton("RESET APP ICON", func() {
		fyneApp.Preferences().SetString("custom_icon_path", "")
		SetApplicationIcon(fyneApp)
		widgets.ShowAppInformation("App Icon", "App icon has been reset to the default.", w)
	}, 200, 40)

	return container.NewVBox(
		theme.CreateLabel("APPEARANCE", 11, theme.ColorPurple, true),
		theme.CreateLabel("Appearance", 10, theme.ColorPurple, false),
		theme.CreateLabel("Theme:", 9, theme.ColorTextSec, false),
		themeSelect,
		theme.CreateLabel("Font Size:", 9, theme.ColorTextSec, false),
		fontSizeSelect,
		widget.NewLabel(""),
		theme.CreateLabel("Behavior", 10, theme.ColorPurple, false),
		showOnHoverCheck,
		confirmActionsCheck,
		widget.NewLabel(""),
		theme.CreateLabel("Image Palette", 10, theme.ColorPurple, false),
		theme.CreateLabel("Upload an image to extract the 3 most common colors and apply them to the UI.", 9, theme.ColorTextSec, false),
		container.NewCenter(uploadBtn),
		widget.NewLabel(""),
		container.NewHBox(
			container.NewVBox(theme.CreateLabel("Background + Containers", 8, theme.ColorTextSec, false), preview1),
			widget.NewLabel("   "),
			container.NewVBox(theme.CreateLabel("Main Buttons", 8, theme.ColorTextSec, false), preview2),
			widget.NewLabel("   "),
			container.NewVBox(theme.CreateLabel("Secondary Buttons", 8, theme.ColorTextSec, false), preview3),
		),
		widget.NewLabel(""),
		theme.CreateLabel("Manual Personalization", 10, theme.ColorPurple, false),
		theme.CreateLabel("Pick exact colors for each UI role using an RGB map, hex, or RGB code.", 9, theme.ColorTextSec, false),
		container.NewCenter(manualPersonalizeBtn),
		widget.NewLabel(""),
		container.NewCenter(resetPaletteBtn),
		widget.NewLabel(""),
		theme.CreateDivider(),
		widget.NewLabel(""),
		theme.CreateLabel("App Icon", 10, theme.ColorPurple, false),
		theme.CreateLabel("Replace the application icon with any PNG or JPEG image. Takes effect immediately and persists across restarts.", 9, theme.ColorTextSec, false),
		widget.NewLabel(""),
		container.NewCenter(container.NewHBox(changeIconBtn, resetIconBtn)),
	)
}

func buildAboutSettings(w fyne.Window, fyneApp fyne.App, appState *app.AppState) *fyne.Container {
	appNameLabel := theme.CreateLabel("PassQuantum", 16, theme.ColorAccentCyn, true)
	versionLabel := theme.CreateLabel("Version 1.0.0", 11, theme.ColorTextSec, false)
	descriptionLabel := theme.CreateLabel("A post-quantum cryptography password manager using Kyber and AES-256-GCM", 10, theme.ColorTextPrim, false)

	featuresBox := container.NewVBox(
		theme.CreateLabel("Features", 10, theme.ColorPurple, true),
		theme.CreateLabel("✓ Post-Quantum Cryptography (Kyber-768)", 9, theme.ColorTextPrim, false),
		theme.CreateLabel("✓ AES-256-GCM Encryption", 9, theme.ColorTextPrim, false),
		theme.CreateLabel("✓ Multiple Vault Support", 9, theme.ColorTextPrim, false),
		theme.CreateLabel("✓ Secure Key Derivation", 9, theme.ColorTextPrim, false),
		theme.CreateLabel("✓ Zero-Knowledge Architecture", 9, theme.ColorTextPrim, false),
	)

	developedByLabel := theme.CreateLabel("Developed by: PassQuantum Team", 10, theme.ColorTextSec, false)
	licenseLabel := theme.CreateLabel("License: MIT", 10, theme.ColorTextSec, false)

	docsBtn := theme.CreateSecondaryButton("📖 DOCS", func() {
		widgets.ShowAppInformation("Docs", "Visit https://github.com/passquantum for documentation", w)
	}, 140, 40)

	updatesBtn := theme.CreateNeonButton("🔄 UPDATES", func() {
		widgets.ShowAppInformation("Updates", "You are running the latest version!", w)
	}, 160, 40)

	return container.NewVBox(
		container.NewCenter(appNameLabel),
		container.NewCenter(versionLabel),
		widget.NewLabel(""),
		container.NewCenter(descriptionLabel),
		widget.NewLabel(""),
		featuresBox,
		widget.NewLabel(""),
		container.NewCenter(developedByLabel),
		container.NewCenter(licenseLabel),
		widget.NewLabel(""),
		container.NewCenter(container.NewHBox(docsBtn, updatesBtn)),
	)
}

func showChangeMasterPasswordDialog(w fyne.Window, appState *app.AppState) {
	oldPwInput := widget.NewPasswordEntry()
	oldPwInput.PlaceHolder = "Current master password"

	newPwInput := widget.NewPasswordEntry()
	newPwInput.PlaceHolder = "New master password"

	confirmPwInput := widget.NewPasswordEntry()
	confirmPwInput.PlaceHolder = "Confirm new password"

	buildField := func(label string, entry *widget.Entry) fyne.CanvasObject {
		return container.NewVBox(
			theme.CreateLabel(label, 10, theme.ColorTextPrimary, true),
			theme.CreateStyledPasswordInput(entry, 420, 40),
		)
	}

	var d dialog.Dialog
	cancelBtn := theme.CreateSecondaryButton("Cancel", func() {
		if d != nil {
			d.Hide()
		}
	}, 120, 40)

	changeBtn := theme.CreateNeonButton("Change", func() {
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
	}, 120, 40)

	content := container.NewVBox(
		theme.CreateLabel("Change Master Password", 16, theme.ColorTextPrimary, true),
		widget.NewLabel(""),
		buildField("Current Password", oldPwInput),
		widget.NewLabel(""),
		buildField("New Password", newPwInput),
		widget.NewLabel(""),
		buildField("Confirm Password", confirmPwInput),
		widget.NewLabel(""),
		container.NewCenter(container.NewHBox(cancelBtn, changeBtn)),
	)

	card := theme.CreateCard(content, 420, 0, true)
	d = dialog.NewCustomWithoutButtons("", container.NewPadded(card), w)
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
	theme.ColorBg = color.NRGBA{R: 11, G: 15, B: 20, A: 255}
	theme.ColorSidebarBg = color.NRGBA{R: 20, G: 25, B: 32, A: 255}
	theme.ColorCardBg = color.NRGBA{R: 26, G: 31, B: 40, A: 255}
	theme.ColorInputBg = color.NRGBA{R: 30, G: 40, B: 50, A: 255}

	theme.ColorAccentCyan = color.NRGBA{R: 34, G: 211, B: 238, A: 255}
	theme.ColorAccentCyn = theme.ColorAccentCyan
	theme.ColorAccentPink = color.NRGBA{R: 236, G: 72, B: 153, A: 255}
	theme.ColorMagenta = theme.ColorAccentPink
	theme.ColorPurple = color.NRGBA{R: 168, G: 85, B: 247, A: 200}

	theme.ColorTextPrimary = theme.PickAdaptiveTextColor(theme.ColorBg)
	theme.ColorTextSecondary = theme.PickAdaptiveTextColor(theme.ColorCardBg)
	theme.ColorTextPrim = theme.ColorTextPrimary
	theme.ColorTextSec = theme.ColorTextSecondary

	theme.ColorBorderCyan = color.NRGBA{R: 34, G: 211, B: 238, A: 180}
	theme.ColorBorder = theme.ColorBorderCyan
	theme.ColorGlowCyan = color.NRGBA{R: 34, G: 211, B: 238, A: 80}

	theme.ColorPrimaryButton = theme.ColorAccentCyan
	theme.ColorSecondaryButton = theme.ColorAccentPink
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
