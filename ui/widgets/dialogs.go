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

	titleTxt := canvas.NewText(title, theme.ColorTextPrimary)
	titleTxt.TextSize = 15
	titleTxt.TextStyle = fyne.TextStyle{Bold: true}

	messageLabel := widget.NewLabel(message)
	messageLabel.Wrapping = fyne.TextWrapWord

	iconCircle := canvas.NewCircle(color.NRGBA{R: tone.R, G: tone.G, B: tone.B, A: 0x24})
	iconCircle.StrokeWidth = 1
	iconCircle.StrokeColor = color.NRGBA{R: tone.R, G: tone.G, B: tone.B, A: 0x66}
	glyphTxt := canvas.NewText(glyph, color.NRGBA{R: tone.R, G: tone.G, B: tone.B, A: 0xcc})
	glyphTxt.TextSize = 36
	glyphTxt.TextStyle = fyne.TextStyle{Bold: true}
	iconStack := container.NewGridWrap(fyne.NewSize(145, 72), container.NewStack(iconCircle, container.NewCenter(glyphTxt)))

	var d dialog.Dialog
	buttons := make([]fyne.CanvasObject, 0, len(actions))
	for _, action := range actions {
		a := action
		onTap := func() {
			if a.onTap != nil {
				a.onTap()
			}
			if d != nil {
				d.Hide()
			}
		}
		var btn fyne.CanvasObject
		if a.primary {
			btn = theme.CreatePrimaryButton(a.label, onTap)
		} else {
			btn = theme.CreateGhostButton(a.label, onTap)
		}
		buttons = append(buttons, btn)
	}

	buttonBar := container.NewCenter(container.NewHBox(buttons...))

	divider := canvas.NewRectangle(theme.ColorLine1)
	divider.SetMinSize(fyne.NewSize(0, 1))

	content := container.NewVBox(
		container.NewCenter(iconStack),
		container.NewCenter(titleTxt),
		messageLabel,
		divider,
		buttonBar,
	)

	card := theme.CardWithHeader("", "", nil, content)
	d = dialog.NewCustomWithoutButtons("", container.NewPadded(card), w)
	d.Show()
}
