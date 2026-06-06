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
	// Surfaces — dark ramp anchored to the brand ink #11161D
	case fynetheme.ColorNameBackground:
		return color.NRGBA{R: 0x0a, G: 0x0d, B: 0x12, A: 255} // bg0
	case fynetheme.ColorNameMenuBackground:
		return color.NRGBA{R: 0x0e, G: 0x12, B: 0x18, A: 255} // bg1
	case fynetheme.ColorNameOverlayBackground:
		return color.NRGBA{R: 0x11, G: 0x16, B: 0x1d, A: 255} // bg2 (brand ink)
	case fynetheme.ColorNameHeaderBackground:
		return color.NRGBA{R: 0x0e, G: 0x12, B: 0x18, A: 255} // bg1
	case fynetheme.ColorNameButton:
		return color.NRGBA{R: 0x1a, G: 0x21, B: 0x2b, A: 255} // bg3
	case fynetheme.ColorNameDisabledButton:
		return color.NRGBA{R: 0x11, G: 0x16, B: 0x1d, A: 255} // bg2
	case fynetheme.ColorNameInputBackground:
		return color.NRGBA{R: 0x0e, G: 0x12, B: 0x18, A: 255} // bg1
	case fynetheme.ColorNameInputBorder:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x1a} // line2

	// Text
	case fynetheme.ColorNameForeground:
		return color.NRGBA{R: 0xea, G: 0xee, B: 0xf2, A: 255} // fg0 (brand near-white)
	case fynetheme.ColorNamePlaceHolder:
		return color.NRGBA{R: 0x4d, G: 0x56, B: 0x62, A: 255} // fg3
	case fynetheme.ColorNameDisabled:
		return color.NRGBA{R: 0x6b, G: 0x77, B: 0x85, A: 255} // fg2 (brand muted slate)

	// Accent — brand teal
	case fynetheme.ColorNamePrimary:
		return color.NRGBA{R: 0x2d, G: 0xd4, B: 0xbf, A: 255} // accent
	case fynetheme.ColorNameFocus:
		return color.NRGBA{R: 0x2d, G: 0xd4, B: 0xbf, A: 0x66} // accentLine (40%)
	case fynetheme.ColorNameSelection:
		return color.NRGBA{R: 0x2d, G: 0xd4, B: 0xbf, A: 0x24} // accentSoft (14%)
	case fynetheme.ColorNameHyperlink:
		return color.NRGBA{R: 0x0e, G: 0x9c, B: 0x86, A: 255} // deep teal
	case fynetheme.ColorNameForegroundOnPrimary:
		return color.NRGBA{R: 0x11, G: 0x16, B: 0x1d, A: 255} // accentFg — brand ink on teal

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
		return color.NRGBA{R: 0x1a, G: 0x21, B: 0x2b, A: 255} // bg3
	case fynetheme.ColorNamePressed:
		return color.NRGBA{R: 0x24, G: 0x2c, B: 0x38, A: 255} // bg4
	case fynetheme.ColorNameScrollBar:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x29} // line3
	case fynetheme.ColorNameScrollBarBackground:
		return color.NRGBA{R: 0x0e, G: 0x12, B: 0x18, A: 255} // bg1

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
