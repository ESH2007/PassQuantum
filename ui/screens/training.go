package screens

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
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"passquantum/app"
	"passquantum/bridge"
	"passquantum/theme"
)

// captureSamples mirrors the Python CAPTURE_SAMPLES constant.
const captureSamples = 20

// ShowTrainingScreen replaces the window content with the face registration UI.
// onComplete is called (on the Fyne goroutine) once training finishes and
// START_MONITOR has been dispatched to the Python process.
func ShowTrainingScreen(w fyne.Window, guard *bridge.FaceGuard, appState *app.AppState, onComplete func()) {
	// Mark training as active so the global OnLost handler does not lock the
	// app while the user is deliberately repositioning their face.
	appState.Mu.Lock()
	appState.IsTraining = true
	appState.Mu.Unlock()
	// ── Title ──────────────────────────────────────────────────────
	title := canvas.NewText("FACIAL REGISTRATION", theme.ColorTextPrimary)
	title.TextSize = 22
	title.TextStyle = fyne.TextStyle{Bold: true}

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
	statusLabel := canvas.NewText("Press the button below to begin.", theme.ColorTextSecondary)
	statusLabel.TextSize = 13

	// ── Start button ───────────────────────────────────────────────
	var started bool
	startBtn := theme.CreatePrimaryButton("Start registration", func() {
		if started {
			return
		}
		started = true

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

		trainingDone := make(chan struct{})
		guard.OnDone = func() {
			close(trainingDone)
			// SendCommand is safe to call here — Python is already connected
			// (it just sent TRAINING_DONE), so this write is non-blocking.
			guard.SendCommand("START_MONITOR")
			fyne.Do(func() {
				updateCanvasText(statusLabel, "✓ Registration complete")
				guard.OnFrame = nil
				// Training finished — re-enable face-loss locking.
				appState.Mu.Lock()
				appState.IsTraining = false
				appState.Mu.Unlock()
				onComplete()
			})
		}

		updateCanvasText(statusLabel, "Starting webcam …  (may take a few seconds)")
		// Run in a goroutine: SendCommand blocks until Python connects, which can
		// take several seconds while face_recognition imports.  Blocking here
		// would freeze the Fyne UI thread.
		go guard.SendCommand("START_TRAINING")

		time.AfterFunc(2*time.Minute, func() {
			select {
			case <-trainingDone:
				return
			default:
			}
			log.Println("[FaceGuard] WARNING: training timed out after 2 minutes")
			appState.Mu.Lock()
			appState.IsTraining = false
			appState.Mu.Unlock()
			fyne.Do(func() {
				started = false
				updateCanvasText(statusLabel, "Training timed out. You can try again.")
			})
		})
	})

	// ── Layout ─────────────────────────────────────────────────────
	content := container.NewVBox(
		container.NewCenter(title),
		widget.NewSeparator(),
		container.NewCenter(camImage),
		container.NewCenter(progressBar),
		container.NewCenter(statusLabel),
		container.NewCenter(startBtn),
	)

	card := theme.CardWithHeader("", "Facial Registration", nil, content)
	screen := container.NewCenter(card)

	w.SetContent(screen)
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

