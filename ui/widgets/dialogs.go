package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"passquantum/theme"
)

type appDialogAction struct {
	label   string
	primary bool
	onTap   func()
}

func ShowAppInformation(title, message string, w fyne.Window) {
	showAppDialog(w, title, message, "i", theme.ColorAccentCyan, []appDialogAction{
		{label: "OK", primary: true},
	})
}

func ShowAppWarning(title, message string, w fyne.Window) {
	showAppDialog(w, title, message, "!", theme.ColorWarning, []appDialogAction{
		{label: "OK", primary: true},
	})
}

func ShowAppError(err error, w fyne.Window) {
	if err == nil {
		return
	}
	showAppDialog(w, "Error", err.Error(), "!", theme.ColorDanger, []appDialogAction{
		{label: "OK", primary: true},
	})
}

func ShowAppConfirm(title, message string, onResult func(bool), w fyne.Window) {
	showAppDialog(w, title, message, "?", theme.ColorWarning, []appDialogAction{
		{
			label:   "No",
			primary: false,
			onTap: func() {
				if onResult != nil {
					onResult(false)
				}
			},
		},
		{
			label:   "Yes",
			primary: true,
			onTap: func() {
				if onResult != nil {
					onResult(true)
				}
			},
		},
	})
}

func showAppDialog(w fyne.Window, title, message, glyph string, tone color.NRGBA, actions []appDialogAction) {
	if w == nil {
		return
	}

	titleLabel := theme.CreateLabel(title, 16, theme.ColorTextPrimary, true)
	messageLabel := widget.NewLabel(message)
	messageLabel.Wrapping = fyne.TextWrapWord

	iconCircle := canvas.NewCircle(color.NRGBA{R: tone.R, G: tone.G, B: tone.B, A: 35})
	iconSymbol := theme.CreateLabel(glyph, 44, color.NRGBA{R: tone.R, G: tone.G, B: tone.B, A: 170}, true)
	iconStack := container.NewGridWrap(fyne.NewSize(90, 90), container.NewStack(iconCircle, container.NewCenter(iconSymbol)))

	var d dialog.Dialog
	buttons := make([]fyne.CanvasObject, 0, len(actions))
	for _, action := range actions {
		a := action
		var btn fyne.CanvasObject
		onTap := func() {
			if a.onTap != nil {
				a.onTap()
			}
			if d != nil {
				d.Hide()
			}
		}
		if a.primary {
			btn = theme.CreateNeonButton(a.label, onTap, 95, 38)
		} else {
			btn = theme.CreateSecondaryButton(a.label, onTap, 95, 38)
		}
		buttons = append(buttons, btn)
	}

	buttonBar := container.NewCenter(container.NewHBox(buttons...))

	content := container.NewVBox(
		titleLabel,
		widget.NewLabel(""),
		container.NewCenter(iconStack),
		widget.NewLabel(""),
		messageLabel,
		widget.NewLabel(""),
		buttonBar,
	)

	card := theme.CreateCard(content, 440, 0, true)
	d = dialog.NewCustomWithoutButtons("", container.NewPadded(card), w)
	d.Show()
}
