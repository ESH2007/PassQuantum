package palette

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type PassQuantumTheme struct {
	palette ThemePalette
}

func NewPassQuantumTheme(p ThemePalette) *PassQuantumTheme {
	return &PassQuantumTheme{palette: p}
}

func (t *PassQuantumTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return t.palette.Background
	case theme.ColorNameButton:
		return t.palette.Primary
	case theme.ColorNameHover:
		return t.palette.Hover
	case theme.ColorNamePressed:
		return t.palette.Pressed
	case theme.ColorNameForeground:
		return t.palette.TextOnBg
	case theme.ColorNameInputBackground:
		return t.palette.Surface
	case theme.ColorNameOverlayBackground:
		return t.palette.Overlay
	case theme.ColorNameSeparator:
		return t.palette.Border
	case theme.ColorNamePrimary:
		return t.palette.Primary
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t *PassQuantumTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *PassQuantumTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *PassQuantumTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
