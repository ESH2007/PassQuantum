package screens

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
)

// SetApplicationIcon sets the application icon, preferring a user-configured custom
// icon path stored in preferences, then falling back to well-known file locations.
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

	iconCandidates := []string{
		"Icon.png",
		filepath.Join("..", "Icon.png"),
		filepath.Join("build", "windows", "Icon.png"),
	}

	for _, iconPath := range iconCandidates {
		data, err := os.ReadFile(iconPath)
		if err != nil || len(data) == 0 {
			continue
		}
		myApp.SetIcon(fyne.NewStaticResource("Icon.png", data))
		return
	}
}
