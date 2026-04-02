//go:build nobiometric

package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

func startEnrollmentPreview(previewImg *canvas.Image, statusLabel *widget.Label) func() {
	fyne.Do(func() {
		statusLabel.SetText("Live camera preview is unavailable in this cross-compiled build.")
	})
	return func() {}
}
