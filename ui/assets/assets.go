package assets

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed "Icons/passquantum-mark-512.png"
var DefaultIcon []byte

//go:embed "Icons/passquantum-mark-512.png"
var LogoImage []byte

// DefaultIconResource returns a fyne.Resource backed by the embedded icon.
func DefaultIconResource() fyne.Resource {
	return fyne.NewStaticResource("./Icons/passquantum-favicon-512.png", DefaultIcon)
}
