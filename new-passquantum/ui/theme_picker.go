package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	"passquantum/palette"
)

const (
	themeImagePathPrefKey  = "theme_image_path"
	manualBgPrefKey        = "manual_color_bg"
	manualPrimaryPrefKey   = "manual_color_primary"
	manualSecondaryPrefKey = "manual_color_secondary"
)

func ShowThemePicker(a fyne.App, w fyne.Window) {
	showThemePicker(a, w, nil)
}

func showThemePicker(a fyne.App, w fyne.Window, onApplied func()) {
	pickImageFile("Select Theme Image", func(path string) {
		f, err := os.Open(path)
		if err != nil {
			ShowAppError(fmt.Errorf("could not open image: %w", err), w)
			return
		}
		defer func() { _ = f.Close() }()

		img, _, decodeErr := image.Decode(f)
		if decodeErr != nil {
			ShowAppError(fmt.Errorf("could not decode image: %w", decodeErr), w)
			return
		}

		samples := palette.SamplePixels(img, 2000)
		clusters := palette.KMeans(samples, 6, 10)
		if len(clusters) == 0 {
			ShowAppError(fmt.Errorf("could not extract dominant colors from image"), w)
			return
		}

		built := palette.BuildThemePalette(clusters)
		a.Settings().SetTheme(palette.NewPassQuantumTheme(built))
		applyThemePaletteToGlobals(built, clusters)
		clearManualPalettePreferences(a)
		a.Preferences().SetString(themeImagePathPrefKey, path)

		if onApplied != nil {
			onApplied()
		} else if w.Content() != nil {
			w.Content().Refresh()
		}
	}, func(err error) {
		ShowAppError(err, w)
	})
}

func RestoreThemeOnLaunch(a fyne.App, w fyne.Window) {
	if manual, ok := loadManualPalettePreferences(a); ok {
		a.Settings().SetTheme(theme.DefaultTheme())
		applyExtractedPalette(manual)
		return
	}

	themePath := a.Preferences().StringWithFallback(themeImagePathPrefKey, "")
	if themePath == "" {
		return
	}

	f, err := os.Open(themePath)
	if err != nil {
		a.Preferences().SetString(themeImagePathPrefKey, "")
		return
	}
	defer func() {
		_ = f.Close()
	}()

	img, _, decodeErr := image.Decode(f)
	if decodeErr != nil {
		a.Preferences().SetString(themeImagePathPrefKey, "")
		return
	}

	samples := palette.SamplePixels(img, 2000)
	clusters := palette.KMeans(samples, 6, 10)
	if len(clusters) == 0 {
		a.Preferences().SetString(themeImagePathPrefKey, "")
		return
	}

	built := palette.BuildThemePalette(clusters)
	a.Settings().SetTheme(palette.NewPassQuantumTheme(built))
	applyThemePaletteToGlobals(built, clusters)
	if w.Content() != nil {
		w.Content().Refresh()
	}
}

func persistManualPalettePreferences(a fyne.App, colors []color.NRGBA) {
	if len(colors) < 3 {
		return
	}
	a.Preferences().SetString(manualBgPrefKey, toHex(colors[0]))
	a.Preferences().SetString(manualPrimaryPrefKey, toHex(colors[1]))
	a.Preferences().SetString(manualSecondaryPrefKey, toHex(colors[2]))
	a.Preferences().SetString(themeImagePathPrefKey, "")
}

func clearManualPalettePreferences(a fyne.App) {
	a.Preferences().SetString(manualBgPrefKey, "")
	a.Preferences().SetString(manualPrimaryPrefKey, "")
	a.Preferences().SetString(manualSecondaryPrefKey, "")
}

func loadManualPalettePreferences(a fyne.App) ([]color.NRGBA, bool) {
	bgHex := a.Preferences().StringWithFallback(manualBgPrefKey, "")
	primaryHex := a.Preferences().StringWithFallback(manualPrimaryPrefKey, "")
	secondaryHex := a.Preferences().StringWithFallback(manualSecondaryPrefKey, "")
	if bgHex == "" || primaryHex == "" || secondaryHex == "" {
		return nil, false
	}

	bg, err := parseHexColor(bgHex)
	if err != nil {
		return nil, false
	}
	primary, err := parseHexColor(primaryHex)
	if err != nil {
		return nil, false
	}
	secondary, err := parseHexColor(secondaryHex)
	if err != nil {
		return nil, false
	}

	return []color.NRGBA{bg, primary, secondary}, true
}

func applyThemePaletteToGlobals(tp palette.ThemePalette, clusters []palette.ColorCluster) {
	base := nrgbaFromRGBA(tp.Background)
	primary := nrgbaFromRGBA(tp.Primary)
	secondary := pickSecondaryColor(clusters, tp)

	applyExtractedPalette([]color.NRGBA{base, primary, secondary})
}

func pickSecondaryColor(clusters []palette.ColorCluster, tp palette.ThemePalette) color.NRGBA {
	primary := tp.Primary
	background := tp.Background

	for _, c := range clusters {
		if c.Color == primary || c.Color == background {
			continue
		}
		return nrgbaFromRGBA(c.Color)
	}

	return nrgbaFromRGBA(tp.Hover)
}

func nrgbaFromRGBA(c color.RGBA) color.NRGBA {
	return color.NRGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}
