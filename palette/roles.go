package palette

import (
	"image/color"
	"sort"
)

type ThemePalette struct {
	Background    color.RGBA
	Surface       color.RGBA
	Primary       color.RGBA
	Hover         color.RGBA
	Pressed       color.RGBA
	Border        color.RGBA
	TextOnBg      color.RGBA
	TextOnPrimary color.RGBA
	Overlay       color.RGBA
}

func BuildThemePalette(clusters []ColorCluster) ThemePalette {
	if len(clusters) == 0 {
		fallback := color.RGBA{R: 0x0a, G: 0x0d, B: 0x12, A: 255}
		primary := color.RGBA{R: 0x2d, G: 0xd4, B: 0xbf, A: 255}
		return ThemePalette{
			Background:    fallback,
			Surface:       Lighten(fallback, 0.08),
			Primary:       primary,
			Hover:         Lighten(primary, 0.15),
			Pressed:       Darken(primary, 0.15),
			Border:        WithAlpha(ContrastColor(fallback), 40),
			TextOnBg:      ContrastColor(fallback),
			TextOnPrimary: ContrastColor(primary),
			Overlay:       WithAlpha(fallback, 180),
		}
	}

	byLuminance := append([]ColorCluster(nil), clusters...)
	sort.Slice(byLuminance, func(i, j int) bool {
		return Luminance(byLuminance[i].Color) < Luminance(byLuminance[j].Color)
	})
	background := byLuminance[0].Color

	bySaturation := append([]ColorCluster(nil), clusters...)
	sort.Slice(bySaturation, func(i, j int) bool {
		return Saturation(bySaturation[i].Color) > Saturation(bySaturation[j].Color)
	})
	primary := bySaturation[0].Color

	return ThemePalette{
		Background:    background,
		Surface:       Lighten(background, 0.08),
		Primary:       primary,
		Hover:         Lighten(primary, 0.15),
		Pressed:       Darken(primary, 0.15),
		Border:        WithAlpha(ContrastColor(background), 40),
		TextOnBg:      ContrastColor(background),
		TextOnPrimary: ContrastColor(primary),
		Overlay:       WithAlpha(background, 180),
	}
}
