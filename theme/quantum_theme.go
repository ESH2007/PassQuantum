package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	fynetheme "fyne.io/fyne/v2/theme"
)

// QuantumTheme implements fyne.Theme with the PassQuantum design system tokens.
type QuantumTheme struct{}

var _ fyne.Theme = (*QuantumTheme)(nil)

func (q *QuantumTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {
	// Surfaces
	case fynetheme.ColorNameBackground:
		return color.NRGBA{R: 0x0b, G: 0x0e, B: 0x13, A: 255} // bg0
	case fynetheme.ColorNameMenuBackground:
		return color.NRGBA{R: 0x0f, G: 0x13, B: 0x19, A: 255} // bg1
	case fynetheme.ColorNameOverlayBackground:
		return color.NRGBA{R: 0x13, G: 0x18, B: 0x22, A: 255} // bg2
	case fynetheme.ColorNameHeaderBackground:
		return color.NRGBA{R: 0x0f, G: 0x13, B: 0x19, A: 255} // bg1
	case fynetheme.ColorNameButton:
		return color.NRGBA{R: 0x1a, G: 0x20, B: 0x30, A: 255} // bg3
	case fynetheme.ColorNameDisabledButton:
		return color.NRGBA{R: 0x13, G: 0x18, B: 0x22, A: 255} // bg2
	case fynetheme.ColorNameInputBackground:
		return color.NRGBA{R: 0x0f, G: 0x13, B: 0x19, A: 255} // bg1
	case fynetheme.ColorNameInputBorder:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x1a} // line2

	// Text
	case fynetheme.ColorNameForeground:
		return color.NRGBA{R: 0xe7, G: 0xea, B: 0xf0, A: 255} // fg0
	case fynetheme.ColorNamePlaceHolder:
		return color.NRGBA{R: 0x55, G: 0x5c, B: 0x6c, A: 255} // fg3
	case fynetheme.ColorNameDisabled:
		return color.NRGBA{R: 0x7a, G: 0x82, B: 0x94, A: 255} // fg2

	// Accent
	case fynetheme.ColorNamePrimary:
		return color.NRGBA{R: 0x3b, G: 0x82, B: 0xf6, A: 255} // accent
	case fynetheme.ColorNameFocus:
		return color.NRGBA{R: 0x3b, G: 0x82, B: 0xf6, A: 0x66} // accentLine (40%)
	case fynetheme.ColorNameSelection:
		return color.NRGBA{R: 0x3b, G: 0x82, B: 0xf6, A: 0x24} // accentSoft (14%)
	case fynetheme.ColorNameHyperlink:
		return color.NRGBA{R: 0x3b, G: 0x82, B: 0xf6, A: 255} // accent
	case fynetheme.ColorNameForegroundOnPrimary:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 255} // accentFg

	// Semantic
	case fynetheme.ColorNameSuccess:
		return color.NRGBA{R: 0x2e, G: 0xa9, B: 0x6b, A: 255} // ok
	case fynetheme.ColorNameForegroundOnSuccess:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 255}
	case fynetheme.ColorNameWarning:
		return color.NRGBA{R: 0xd9, G: 0x90, B: 0x30, A: 255} // warn
	case fynetheme.ColorNameForegroundOnWarning:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 255}
	case fynetheme.ColorNameError:
		return color.NRGBA{R: 0xd0, G: 0x4a, B: 0x4a, A: 255} // danger
	case fynetheme.ColorNameForegroundOnError:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 255}

	// Chrome
	case fynetheme.ColorNameShadow:
		return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x66} // 40%
	case fynetheme.ColorNameSeparator:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x0f} // line1 (6%)
	case fynetheme.ColorNameHover:
		return color.NRGBA{R: 0x1a, G: 0x20, B: 0x30, A: 255} // bg3
	case fynetheme.ColorNamePressed:
		return color.NRGBA{R: 0x23, G: 0x2a, B: 0x3a, A: 255} // bg4
	case fynetheme.ColorNameScrollBar:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x29} // line3
	case fynetheme.ColorNameScrollBarBackground:
		return color.NRGBA{R: 0x0f, G: 0x13, B: 0x19, A: 255} // bg1

	default:
		return fynetheme.DefaultTheme().Color(name, fynetheme.VariantDark)
	}
}

func (q *QuantumTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Monospace {
		if style.Bold {
			return FontMonoMedium
		}
		return FontMonoRegular
	}
	if style.Bold {
		return FontSansSemiBold
	}
	return FontSansRegular
}

func (q *QuantumTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return fynetheme.DefaultTheme().Icon(name)
}

func (q *QuantumTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case fynetheme.SizeNameText:
		return 13
	case fynetheme.SizeNameHeadingText:
		return 22
	case fynetheme.SizeNameSubHeadingText:
		return 15
	case fynetheme.SizeNameCaptionText:
		return 11
	case fynetheme.SizeNameInnerPadding:
		return 8
	case fynetheme.SizeNamePadding:
		return 12
	case fynetheme.SizeNameInputBorder:
		return 1
	case fynetheme.SizeNameInputRadius:
		return 6
	case fynetheme.SizeNameSeparatorThickness:
		return 1
	case fynetheme.SizeNameScrollBar:
		return 10
	case fynetheme.SizeNameScrollBarSmall:
		return 6
	default:
		return fynetheme.DefaultTheme().Size(name)
	}
}
