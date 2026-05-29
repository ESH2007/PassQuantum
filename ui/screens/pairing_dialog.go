package screens

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func ShowPairingDialog(w fyne.Window, token string) {
	content := widget.NewLabel(
		"A browser extension is requesting to pair with PassQuantum.\n\n" +
			"Enter this code in your browser extension:\n\n" +
			"    " + token + "\n\n" +
			"This code expires in 60 seconds.",
	)
	content.Wrapping = fyne.TextWrapWord

	d := dialog.NewCustom("Browser Extension Pairing", "Close", content, w)
	d.Resize(fyne.NewSize(400, 250))
	d.Show()
}
