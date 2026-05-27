package assets

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed Icon.png
var DefaultIcon []byte

//go:embed PM.png
var LogoImage []byte

// DefaultIconResource returns a fyne.Resource backed by the embedded icon.
func DefaultIconResource() fyne.Resource {
	return fyne.NewStaticResource("Icon.png", DefaultIcon)
}
