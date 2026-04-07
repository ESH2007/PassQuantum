//go:build !nobiometric

package main

import (
	"context"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"gocv.io/x/gocv"
)

func startEnrollmentPreview(appState *AppState, previewImg *canvas.Image, statusLabel *widget.Label) func() {
	previewCtx, cancelPreview := context.WithCancel(context.Background())

	go func() {
		cam, _, err := openBiometricCamera(appState)
		if err != nil {
			fyne.Do(func() { statusLabel.SetText("Camera unavailable.") })
			return
		}
		defer cam.Close()

		fyne.Do(func() { statusLabel.SetText("Camera ready. Position your face.") })

		frame := gocv.NewMat()
		defer frame.Close()

		for {
			select {
			case <-previewCtx.Done():
				return
			case <-time.After(100 * time.Millisecond):
				if !cam.Read(&frame) || frame.Empty() {
					continue
				}
				rgb := gocv.NewMat()
				gocv.CvtColor(frame, &rgb, gocv.ColorBGRToRGB)
				goImg, convErr := rgb.ToImage()
				rgb.Close()
				if convErr != nil {
					continue
				}
				imgCopy := goImg
				fyne.Do(func() {
					previewImg.Image = imgCopy
					previewImg.Refresh()
				})
			}
		}
	}()

	return cancelPreview
}
