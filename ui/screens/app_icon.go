package screens

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"

	"passquantum/ui/assets"
)

// SetApplicationIcon sets the application icon, preferring a user-configured custom
// icon path stored in preferences, then falling back to the icon embedded at
// compile time. The embedded fallback is always available regardless of the
// working directory, so the icon is visible on first launch without any manual
// configuration.
func SetApplicationIcon(myApp fyne.App) {
	if customPath := myApp.Preferences().StringWithFallback("custom_icon_path", ""); customPath != "" {
		data, err := os.ReadFile(customPath)
		if err == nil && len(data) > 0 {
			myApp.SetIcon(fyne.NewStaticResource(filepath.Base(customPath), data))
			return
		}
		// Stale path (file moved/deleted) — clear it and fall through to bundled default.
		myApp.Preferences().SetString("custom_icon_path", "")
	}

	// Always available: icon embedded into the binary at build time.
	myApp.SetIcon(assets.DefaultIconResource())
}
