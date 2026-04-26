package main

// ==============================
// training_screen.go — PassQuantum Face Training UI
// ==============================
// ShowTrainingScreen renders a Fyne-native face registration flow entirely
// inside the existing app window.  No external windows or cv2.imshow is used.
//
// Layout (top-to-bottom, centred):
//   1. Bold title: "Facial Registration"
//   2. 320×240 canvas.Image   — live camera preview (updated via OnFrame)
//   3. ProgressBar (Max = 20) — updated via OnProgress
//   4. Status label           — shows current step / completion text
//   5. "Start Registration" button

import (
	"fmt"
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// captureSamples mirrors the Python CAPTURE_SAMPLES constant.
const captureSamples = 20

// ShowTrainingScreen replaces the window content with the face registration UI.
// onComplete is called (on the Fyne goroutine) once training finishes and
// START_MONITOR has been dispatched to the Python process.
func ShowTrainingScreen(w fyne.Window, guard *FaceGuard, onComplete func()) {
	// ── Title ──────────────────────────────────────────────────────
	title := CreateLabel("FACIAL REGISTRATION", 18, ColorAccentCyan, true)

	// ── Camera preview ─────────────────────────────────────────────
	// Starts with a blank image; updated on every FRAME message.
	blankImg := image.NewNRGBA(image.Rect(0, 0, captureSamples*16, 240))
	camImage := canvas.NewImageFromImage(blankImg)
	camImage.FillMode = canvas.ImageFillContain
	camImage.SetMinSize(fyne.NewSize(320, 240))

	// ── Progress bar ───────────────────────────────────────────────
	progressBar := widget.NewProgressBar()
	progressBar.Max = float64(captureSamples)
	progressBar.Min = 0

	// ── Status label ───────────────────────────────────────────────
	statusLabel := CreateLabel("Press the button below to begin.", 12, ColorTextPrimary, false)

	// ── Start button ───────────────────────────────────────────────
	startBtn := CreateNeonButton("[ START REGISTRATION ]", nil, 280, 48)

	// Re-create the button with a real tap handler (CreateNeonButton requires
	// the handler at construction time, so we rebuild it once we have all refs).
	startBtn = CreateNeonButton("[ START REGISTRATION ]", func() {
		// Disable the button immediately so it cannot be tapped twice.
		disableNeonButton(startBtn)

		// Wire callbacks before sending START_TRAINING so no messages are missed.
		// All three callbacks are invoked from the Listen() goroutine, so every
		// Fyne UI operation must be dispatched via fyne.Do().
		guard.OnFrame = func(img image.Image) {
			fyne.Do(func() {
				camImage.Image = img
				camImage.Refresh()
			})
		}

		guard.OnProgress = func(cur, total int) {
			fyne.Do(func() {
				progressBar.SetValue(float64(cur))
				updateCanvasText(statusLabel, fmt.Sprintf("Capturing sample %d / %d …", cur, total))
			})
		}

		guard.OnDone = func() {
			// SendCommand is safe to call here — Python is already connected
			// (it just sent TRAINING_DONE), so this write is non-blocking.
			guard.SendCommand("START_MONITOR")
			fyne.Do(func() {
				updateCanvasText(statusLabel, "✓ Registration complete")
				guard.OnFrame = nil
				onComplete()
			})
		}

		updateCanvasText(statusLabel, "Starting webcam …  (may take a few seconds)")
		// Run in a goroutine: SendCommand blocks until Python connects, which can
		// take several seconds while face_recognition imports.  Blocking here
		// would freeze the Fyne UI thread.
		go guard.SendCommand("START_TRAINING")
	}, 280, 48)

	// ── Layout ─────────────────────────────────────────────────────
	content := container.NewVBox(
		container.NewCenter(title),
		widget.NewSeparator(),
		container.NewCenter(camImage),
		container.NewCenter(progressBar),
		container.NewCenter(statusLabel),
		container.NewCenter(startBtn),
	)

	card := CreateCard(content, 480, 480, true)
	screen := container.NewCenter(card)

	w.SetContent(screen)
	w.Resize(fyne.NewSize(640, 560))
}

// ==============================
// Helpers
// ==============================

// updateCanvasText sets the text of a *canvas.Text wrapped inside the
// fyne.CanvasObject returned by CreateLabel.
func updateCanvasText(obj fyne.CanvasObject, text string) {
	if txt, ok := obj.(*canvas.Text); ok {
		txt.Text = text
		txt.Refresh()
	}
}

// disableNeonButton walks the CanvasObject returned by CreateNeonButton
// (which is a container.Max wrapping a *widget.Button) and disables the button.
func disableNeonButton(obj fyne.CanvasObject) {
	type widgetDisabler interface {
		Disable()
	}
	if d, ok := obj.(widgetDisabler); ok {
		d.Disable()
		return
	}
	// Walk container children.
	if c, ok := obj.(*fyne.Container); ok {
		for _, child := range c.Objects {
			disableNeonButton(child)
		}
	}
}
