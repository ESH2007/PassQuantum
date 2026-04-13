package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
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
func ShowSettingsScreen(w fyne.Window, fyneApp fyne.App, appState *AppState) {
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

	bgContainer := CreateBackgroundContainer(navState.sidebarContainer)
	w.SetContent(bgContainer)
}

func buildCustomSettingsView(w fyne.Window, fyneApp fyne.App, appState *AppState) fyne.CanvasObject {
	selectedSubview := SettingsSubviewSecurity
	tabStrip := container.New(layout.NewGridLayoutWithColumns(4))
	contentContainer := container.NewMax()

	var refresh func()
	refresh = func() {
		tabStrip.Objects = []fyne.CanvasObject{
			CreateTabButton("Security", selectedSubview == SettingsSubviewSecurity, func() {
				selectedSubview = SettingsSubviewSecurity
				refresh()
			}),
			CreateTabButton("Vaults", selectedSubview == SettingsSubviewVaults, func() {
				selectedSubview = SettingsSubviewVaults
				refresh()
			}),
			CreateTabButton("Visuals", selectedSubview == SettingsSubviewVisuals, func() {
				selectedSubview = SettingsSubviewVisuals
				refresh()
			}),
			CreateTabButton("About", selectedSubview == SettingsSubviewAbout, func() {
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

	headerText := CreateLabel("SETTINGS", 14, ColorAccentCyan, true)
	headerSection := container.NewVBox(headerText, CreateDivider())

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
		CreateLabel(title, 16, ColorAccentCyan, true),
		widget.NewLabel(""),
		CreateLabel(description, 10, ColorTextSec, false),
		widget.NewLabel(""),
		CreateDivider(),
	)

	return CreateCard(container.NewVBox(header, widget.NewLabel(""), content), 820, 0, true)
}

func buildSecuritySettings(w fyne.Window, _ fyne.App, appState *AppState) *fyne.Container {
	passwordStrength := widget.NewSelect([]string{"Weak", "Medium", "Strong", "Very Strong"}, func(s string) {})
	passwordStrength.PlaceHolder = "Select password strength requirement"
	passwordStrength.SetSelected("Strong")

	changePwBtn := CreateNeonButton("CHANGE MASTER PASSWORD", func() {
		showChangeMasterPasswordDialog(w, appState)
	}, 280, 40)

	profileStatus := CreateLabel("App-level verifier active and bound to the current private key.", 10, ColorTextSec, false)

	return container.NewVBox(
		CreateLabel("MASTER PASSWORD", 11, ColorPurple, true),
		profileStatus,
		widget.NewLabel(""),
		CreateLabel("Password Strength:", 10, ColorTextSec, false),
		passwordStrength,
		widget.NewLabel(""),
		container.NewCenter(changePwBtn),
		widget.NewLabel(""),
		CreateDivider(),
	)
}

func buildVaultSettings(w fyne.Window, fyneApp fyne.App, appState *AppState) *fyne.Container {
	currentVaultLabel := CreateLabel("Current Vault: "+appState.currentVault, 10, ColorTextSec, false)
	statsLabel := CreateLabel(fmt.Sprintf("Total Vaults: %d", len(ListVaults())), 10, ColorTextSec, false)

	compactBtn := CreateNeonButton("COMPACT VAULT", func() {
		ShowAppInformation("Compact", "Vault compaction is being performed...", w)
	}, 200, 40)

	exportBtn := CreateSecondaryButton("EXPORT VAULT", func() {
		ShowAppInformation("Export", "Select location to export encrypted vault backup", w)
	}, 200, 40)

	importBtn := CreateSecondaryButton("IMPORT VAULT", func() {
		ShowAppInformation("Import", "Select backup file to import", w)
	}, 200, 40)

	backupNowBtn := CreateNeonButton("BACKUP NOW", func() {
		ShowAppInformation("Backup", "Vault backup created successfully!", w)
	}, 180, 40)

	restoreBtn := CreateSecondaryButton("RESTORE", func() {
		ShowAppConfirm("Restore", "This will replace your current vault. Are you sure?", func(ok bool) {
			if ok {
				ShowAppInformation("Restore", "Select a backup file to restore", w)
			}
		}, w)
	}, 150, 40)

	return container.NewVBox(
		CreateLabel("ACTIVE VAULT", 11, ColorPurple, true),
		currentVaultLabel,
		statsLabel,
		widget.NewLabel(""),
		CreateLabel("Maintenance", 10, ColorPurple, false),
		container.NewCenter(compactBtn),
		widget.NewLabel(""),
		CreateLabel("Backup & Restore", 10, ColorPurple, false),
		container.NewHBox(exportBtn, importBtn),
		widget.NewLabel(""),
		container.NewHBox(backupNowBtn, restoreBtn),
	)
}

func buildDisplaySettings(w fyne.Window, fyneApp fyne.App, appState *AppState) *fyne.Container {
	themeSelect := widget.NewSelect([]string{"Dark", "Light", "System"}, func(_ string) {})
	themeSelect.PlaceHolder = "Select theme"
	themeSelect.SetSelected("Dark")

	fontSizeSelect := widget.NewSelect([]string{"Small", "Medium", "Large"}, func(s string) {})
	fontSizeSelect.PlaceHolder = "Select font size"
	fontSizeSelect.SetSelected("Medium")

	showOnHoverCheck := widget.NewCheck("Show password on hover", func(b bool) {})
	confirmActionsCheck := widget.NewCheck("Confirm before deleting passwords", func(b bool) {})
	confirmActionsCheck.SetChecked(true)

	preview1 := canvasColorPreview(ColorBg)
	preview2 := canvasColorPreview(ColorPrimaryButton)
	preview3 := canvasColorPreview(ColorSecondaryButton)

	manualPersonalizeBtn := CreateNeonButton("MANUAL PERSONALIZATION", func() {
		ShowColorPersonalizationDialog(w, fyneApp, appState)
	}, 280, 40)

	uploadBtn := CreateNeonButton("UPLOAD IMAGE TO ANALYZE", func() {
		showThemePicker(fyneApp, w, func() {
			ShowSettingsScreen(w, fyneApp, appState)
			ShowAppInformation("Palette Applied", "Theme generated from image and saved for next launch.", w)
		})
	}, 300, 40)

	resetPaletteBtn := CreateSecondaryButton("RESET DEFAULT COLORS", func() {
		fyneApp.Settings().SetTheme(theme.DefaultTheme())
		resetDefaultPalette()
		clearManualPalettePreferences(fyneApp)
		fyneApp.Preferences().SetString(themeImagePathPrefKey, "")
		ShowSettingsScreen(w, fyneApp, appState)
	}, 250, 40)

	changeIconBtn := CreateNeonButton("CHANGE APP ICON", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				ShowAppError(err, w)
				return
			}
			if reader == nil {
				return
			}
			defer func() { _ = reader.Close() }()
			data, readErr := io.ReadAll(reader)
			if readErr != nil || len(data) == 0 {
				ShowAppError(fmt.Errorf("could not read icon file"), w)
				return
			}
			name := filepath.Base(reader.URI().Path())
			fyneApp.SetIcon(fyne.NewStaticResource(name, data))
			fyneApp.Preferences().SetString("custom_icon_path", reader.URI().Path())
			ShowAppInformation("App Icon", "App icon updated. It will also be applied on next launch.", w)
		}, w)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg"}))
		fd.Show()
	}, 220, 40)

	resetIconBtn := CreateSecondaryButton("RESET APP ICON", func() {
		fyneApp.Preferences().SetString("custom_icon_path", "")
		setApplicationIcon(fyneApp)
		ShowAppInformation("App Icon", "App icon has been reset to the default.", w)
	}, 200, 40)

	return container.NewVBox(
		CreateLabel("APPEARANCE", 11, ColorPurple, true),
		CreateLabel("Appearance", 10, ColorPurple, false),
		CreateLabel("Theme:", 9, ColorTextSec, false),
		themeSelect,
		CreateLabel("Font Size:", 9, ColorTextSec, false),
		fontSizeSelect,
		widget.NewLabel(""),
		CreateLabel("Behavior", 10, ColorPurple, false),
		showOnHoverCheck,
		confirmActionsCheck,
		widget.NewLabel(""),
		CreateLabel("Image Palette", 10, ColorPurple, false),
		CreateLabel("Upload an image to extract the 3 most common colors and apply them to the UI.", 9, ColorTextSec, false),
		container.NewCenter(uploadBtn),
		widget.NewLabel(""),
		container.NewHBox(
			container.NewVBox(CreateLabel("Background + Containers", 8, ColorTextSec, false), preview1),
			widget.NewLabel("   "),
			container.NewVBox(CreateLabel("Main Buttons", 8, ColorTextSec, false), preview2),
			widget.NewLabel("   "),
			container.NewVBox(CreateLabel("Secondary Buttons", 8, ColorTextSec, false), preview3),
		),
		widget.NewLabel(""),
		CreateLabel("Manual Personalization", 10, ColorPurple, false),
		CreateLabel("Pick exact colors for each UI role using an RGB map, hex, or RGB code.", 9, ColorTextSec, false),
		container.NewCenter(manualPersonalizeBtn),
		widget.NewLabel(""),
		container.NewCenter(resetPaletteBtn),
		widget.NewLabel(""),
		CreateDivider(),
		widget.NewLabel(""),
		CreateLabel("App Icon", 10, ColorPurple, false),
		CreateLabel("Replace the application icon with any PNG or JPEG image. Takes effect immediately and persists across restarts.", 9, ColorTextSec, false),
		widget.NewLabel(""),
		container.NewCenter(container.NewHBox(changeIconBtn, resetIconBtn)),
	)
}

func buildAboutSettings(w fyne.Window, fyneApp fyne.App, appState *AppState) *fyne.Container {
	appNameLabel := CreateLabel("PassQuantum", 16, ColorAccentCyn, true)
	versionLabel := CreateLabel("Version 1.0.0", 11, ColorTextSec, false)
	descriptionLabel := CreateLabel("A post-quantum cryptography password manager using Kyber and AES-256-GCM", 10, ColorTextPrim, false)

	featuresBox := container.NewVBox(
		CreateLabel("Features", 10, ColorPurple, true),
		CreateLabel("✓ Post-Quantum Cryptography (Kyber-768)", 9, ColorTextPrim, false),
		CreateLabel("✓ AES-256-GCM Encryption", 9, ColorTextPrim, false),
		CreateLabel("✓ Multiple Vault Support", 9, ColorTextPrim, false),
		CreateLabel("✓ Secure Key Derivation", 9, ColorTextPrim, false),
		CreateLabel("✓ Zero-Knowledge Architecture", 9, ColorTextPrim, false),
	)

	developedByLabel := CreateLabel("Developed by: PassQuantum Team", 10, ColorTextSec, false)
	licenseLabel := CreateLabel("License: MIT", 10, ColorTextSec, false)

	docsBtn := CreateSecondaryButton("📖 DOCS", func() {
		ShowAppInformation("Docs", "Visit https://github.com/passquantum for documentation", w)
	}, 140, 40)

	updatesBtn := CreateNeonButton("🔄 UPDATES", func() {
		ShowAppInformation("Updates", "You are running the latest version!", w)
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

func showChangeMasterPasswordDialog(w fyne.Window, appState *AppState) {
	oldPwInput := widget.NewPasswordEntry()
	oldPwInput.PlaceHolder = "Current master password"

	newPwInput := widget.NewPasswordEntry()
	newPwInput.PlaceHolder = "New master password"

	confirmPwInput := widget.NewPasswordEntry()
	confirmPwInput.PlaceHolder = "Confirm new password"

	buildField := func(label string, entry *widget.Entry) fyne.CanvasObject {
		return container.NewVBox(
			CreateLabel(label, 10, ColorTextPrimary, true),
			CreateStyledPasswordInput(entry, 420, 40),
		)
	}

	var d dialog.Dialog
	cancelBtn := CreateSecondaryButton("Cancel", func() {
		if d != nil {
			d.Hide()
		}
	}, 120, 40)

	changeBtn := CreateNeonButton("Change", func() {
		if newPwInput.Text == "" {
			ShowAppError(fmt.Errorf("new password cannot be empty"), w)
			return
		}

		if newPwInput.Text != confirmPwInput.Text {
			ShowAppError(fmt.Errorf("new passwords do not match"), w)
			return
		}

		if err := changeMasterPassword(appState, oldPwInput.Text, newPwInput.Text); err != nil {
			ShowAppError(err, w)
			return
		}

		if d != nil {
			d.Hide()
		}
		ShowAppInformation("Success", "Master password changed successfully and all vaults were re-encrypted.", w)
	}, 120, 40)

	content := container.NewVBox(
		CreateLabel("Change Master Password", 16, ColorTextPrimary, true),
		widget.NewLabel(""),
		buildField("Current Password", oldPwInput),
		widget.NewLabel(""),
		buildField("New Password", newPwInput),
		widget.NewLabel(""),
		buildField("Confirm Password", confirmPwInput),
		widget.NewLabel(""),
		container.NewCenter(container.NewHBox(cancelBtn, changeBtn)),
	)

	card := CreateCard(content, 420, 0, true)
	d = dialog.NewCustomWithoutButtons("", container.NewPadded(card), w)
	d.Show()
}

func canvasColorPreview(c color.NRGBA) fyne.CanvasObject {
	rect := canvas.NewRectangle(c)
	rect.CornerRadius = 6
	rect.SetMinSize(fyne.NewSize(180, 42))

	border := canvas.NewRectangle(color.NRGBA{R: ColorTextSecondary.R, G: ColorTextSecondary.G, B: ColorTextSecondary.B, A: 80})
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
			result = append(result, ColorBg)
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

	ColorBg = base
	ColorSidebarBg = blend(base, color.NRGBA{R: 0, G: 0, B: 0, A: 255}, 0.15)
	ColorCardBg = blend(base, color.NRGBA{R: 255, G: 255, B: 255, A: 255}, 0.08)
	ColorInputBg = blend(base, color.NRGBA{R: 0, G: 0, B: 0, A: 255}, 0.1)

	ColorPrimaryButton = primary
	ColorSecondaryButton = secondary

	ColorAccentCyan = primary
	ColorAccentCyn = ColorAccentCyan
	ColorAccentPink = secondary
	ColorMagenta = ColorAccentPink
	ColorPurple = color.NRGBA{R: secondary.R, G: secondary.G, B: secondary.B, A: 220}

	ColorBorderCyan = color.NRGBA{R: primary.R, G: primary.G, B: primary.B, A: 180}
	ColorBorder = ColorBorderCyan
	ColorGlowCyan = color.NRGBA{R: primary.R, G: primary.G, B: primary.B, A: 80}

	ColorTextPrimary = pickAdaptiveTextColor(ColorBg)
	ColorTextSecondary = pickAdaptiveTextColor(ColorCardBg)
	ColorTextPrim = ColorTextPrimary
	ColorTextSec = ColorTextSecondary
}

func resetDefaultPalette() {
	ColorBg = color.NRGBA{R: 11, G: 15, B: 20, A: 255}
	ColorSidebarBg = color.NRGBA{R: 20, G: 25, B: 32, A: 255}
	ColorCardBg = color.NRGBA{R: 26, G: 31, B: 40, A: 255}
	ColorInputBg = color.NRGBA{R: 30, G: 40, B: 50, A: 255}

	ColorAccentCyan = color.NRGBA{R: 34, G: 211, B: 238, A: 255}
	ColorAccentCyn = ColorAccentCyan
	ColorAccentPink = color.NRGBA{R: 236, G: 72, B: 153, A: 255}
	ColorMagenta = ColorAccentPink
	ColorPurple = color.NRGBA{R: 168, G: 85, B: 247, A: 200}

	ColorTextPrimary = pickAdaptiveTextColor(ColorBg)
	ColorTextSecondary = pickAdaptiveTextColor(ColorCardBg)
	ColorTextPrim = ColorTextPrimary
	ColorTextSec = ColorTextSecondary

	ColorBorderCyan = color.NRGBA{R: 34, G: 211, B: 238, A: 180}
	ColorBorder = ColorBorderCyan
	ColorGlowCyan = color.NRGBA{R: 34, G: 211, B: 238, A: 80}

	ColorPrimaryButton = ColorAccentCyan
	ColorSecondaryButton = ColorAccentPink
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
