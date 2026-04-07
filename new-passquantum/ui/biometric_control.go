//go:build !nobiometric

package main

// biometric_control.go — session-lifecycle helpers for face-recognition.
//
// Responsibilities:
//   - Load biometric settings from the security profile into AppState on unlock.
//   - Initialise (or reuse) the ONNX pipeline when biometric is enabled.
//   - Perform a one-shot face verification (used as an unlock gate and for enrolment).
//   - Start / stop the continuous background check that locks the session when the
//     enrolled face is no longer visible or the similarity drops below threshold.
//   - Save an updated biometric template back to the security profile on disk.
//
// Thread-safety: all AppState field mutations go through AppState.mu.
// Runtime objects are only ever accessed from a single goroutine at a time.

import (
	"context"
	"fmt"
	"image"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"gocv.io/x/gocv"

	"passquantum/core/biometric"
	"passquantum/core/storage"
)

// loadBiometricFromProfile copies biometric metadata from profile into AppState.
func loadBiometricFromProfile(appState *AppState) {
	appState.mu.Lock()
	profile := appState.securityProfile
	appState.mu.Unlock()

	if profile == nil {
		return
	}

	appState.mu.Lock()
	defer appState.mu.Unlock()

	appState.biometricEnabled = profile.Biometric.Enabled
	appState.biometricThreshold = profile.Biometric.Threshold
	if profile.Biometric.CameraIndex != nil {
		v := *profile.Biometric.CameraIndex
		appState.biometricCameraIndex = &v
	} else {
		appState.biometricCameraIndex = nil
	}

	if len(profile.BiometricTemplate) > 0 {
		template, err := biometric.DeserializeFeatures(profile.BiometricTemplate)
		if err == nil {
			appState.biometricTemplate = template
		}
	}
}

// saveBiometricToProfile persists any in-memory biometric changes back to the
// on-disk security profile. Must be called on the Fyne UI goroutine (or wrapped
// in fyne.Do) because it shows errors via ShowAppError.
func saveBiometricToProfile(appState *AppState) error {
	appState.mu.Lock()
	profile := appState.securityProfile
	if profile == nil {
		appState.mu.Unlock()
		return fmt.Errorf("no active security profile to update")
	}

	profile.Biometric.Enabled = appState.biometricEnabled
	profile.Biometric.Threshold = appState.biometricThreshold
	profile.Biometric.CameraIndex = appState.biometricCameraIndex
	if appState.biometricTemplate != nil {
		profile.BiometricTemplate = biometric.SerializeFeatures(appState.biometricTemplate)
	} else {
		profile.BiometricTemplate = nil
	}
	appState.mu.Unlock()

	if err := storage.SaveAppSecurityProfile(appSecurityMetadataPath, profile); err != nil {
		return fmt.Errorf("failed to save biometric settings: %w", err)
	}
	return nil
}

// -------------------------------------------------------------------
// Pipeline lifecycle
// -------------------------------------------------------------------

// ensureBiometricPipeline loads the ONNX models the first time biometric is used.
// Subsequent calls reuse the cached pipeline. Returns an error if the model files
// are not found or cannot be loaded.
func ensureBiometricPipeline(appState *AppState) error {
	appState.mu.Lock()
	runtime := appState.biometricRuntime
	threshold := biometric.EffectiveThreshold(appState.biometricThreshold)
	appState.mu.Unlock()

	if runtime != nil {
		return nil // already loaded
	}

	p, err := biometric.NewDefaultRuntime(threshold)
	if err != nil {
		return fmt.Errorf("failed to load biometric models: %w", err)
	}

	appState.mu.Lock()
	// Double-checked lock: another goroutine may have loaded in between.
	if appState.biometricRuntime == nil {
		appState.biometricRuntime = p
	} else {
		p.Close() // race-safe disposal of the extra pipeline
	}
	appState.mu.Unlock()
	return nil
}

// closeBiometricPipeline releases ONNX model resources. Call on app exit or when
// the user disables biometric in settings.
func closeBiometricPipeline(appState *AppState) {
	appState.mu.Lock()
	p := appState.biometricRuntime
	appState.biometricRuntime = nil
	appState.mu.Unlock()
	if p != nil {
		p.Close()
	}
}

// -------------------------------------------------------------------
// One-shot verification (unlock gate & enrolment helper)
// -------------------------------------------------------------------

// captureAndVerifyFace opens the default webcam, attempts up to maxAttempts frames
// to find a face that matches the enrolled template, and returns true if any
// frame passes. The caller must ensure biometricTemplate is set before calling.
//
// This function blocks the calling goroutine (run in a goroutine from the UI layer).
func captureAndVerifyFace(appState *AppState) (bool, error) {
	if err := ensureBiometricPipeline(appState); err != nil {
		return false, err
	}

	webcam, _, err := openBiometricCamera(appState)
	if err != nil {
		return false, fmt.Errorf("could not open camera: %w", err)
	}
	defer webcam.Close()

	appState.mu.Lock()
	runtime := appState.biometricRuntime
	template := appState.biometricTemplate
	threshold := biometric.EffectiveThreshold(appState.biometricThreshold)
	appState.mu.Unlock()

	if runtime == nil {
		return false, fmt.Errorf("biometric runtime not initialised")
	}

	frame := gocv.NewMat()
	defer frame.Close()

	const maxAttempts = 10
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if !webcam.Read(&frame) || frame.Empty() {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		landmarks, found, err := runtime.RunFrame(frame)
		if err != nil || !found {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		features := biometric.ExtractFeatures(landmarks)
		if features == nil {
			continue
		}

		if biometric.CosineSimilarity(features, template) >= threshold {
			return true, nil
		}
	}

	return false, nil
}

// captureEnrolmentFrame opens the webcam, grabs a single stable frame, runs the
// full pipeline, and returns the extracted feature vector. Used for initial
// enrolment and re-enrolment.
func captureEnrolmentFrame(appState *AppState) ([]float32, *image.NRGBA, error) {
	if err := ensureBiometricPipeline(appState); err != nil {
		return nil, nil, err
	}

	webcam, _, err := openBiometricCamera(appState)
	if err != nil {
		return nil, nil, fmt.Errorf("could not open camera: %w", err)
	}
	defer webcam.Close()

	appState.mu.Lock()
	runtime := appState.biometricRuntime
	appState.mu.Unlock()
	if runtime == nil {
		return nil, nil, fmt.Errorf("biometric runtime not initialised")
	}

	frame := gocv.NewMat()
	defer frame.Close()

	const warmUpFrames = 5 // discard early, often-dark camera frames
	for i := 0; i < warmUpFrames; i++ {
		webcam.Read(&frame)
		time.Sleep(40 * time.Millisecond)
	}

	// Capture the enrolment frame.
	if !webcam.Read(&frame) || frame.Empty() {
		return nil, nil, fmt.Errorf("failed to capture frame from webcam")
	}

	landmarks, found, err := runtime.RunFrame(frame)
	if err != nil {
		return nil, nil, fmt.Errorf("face mesh inference failed: %w", err)
	}
	if !found {
		return nil, nil, fmt.Errorf("no face detected in the captured frame")
	}

	features := biometric.ExtractFeatures(landmarks)
	if features == nil {
		return nil, nil, fmt.Errorf("could not extract biometric features: face landmarks are insufficient")
	}

	// Draw the mesh overlay on a preview image for the UI.
	rgb := gocv.NewMat()
	defer rgb.Close()
	gocv.CvtColor(frame, &rgb, gocv.ColorBGRToRGB)
	biometric.DrawMesh(&rgb, landmarks)

	preview, err := rgb.ToImage()
	if err != nil {
		// Non-fatal: return features without a preview image.
		return features, nil, nil
	}

	// Convert to *image.NRGBA for Fyne compatibility.
	bounds := preview.Bounds()
	nrgba := image.NewNRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			nrgba.Set(x, y, preview.At(x, y))
		}
	}
	return features, nrgba, nil
}

// -------------------------------------------------------------------
// Continuous background verification
// -------------------------------------------------------------------

// startContinuousCheck launches a background goroutine that periodically captures
// a camera frame, verifies the enrolled face, and locks the session after
// biometric.MaxConsecutiveFailures consecutive failures.
//
// Does nothing when biometric is disabled or the enrolled template is absent.
// Any previously running check is cancelled first.
func startContinuousCheck(appState *AppState, w fyne.Window, fyneApp fyne.App) {
	appState.mu.Lock()
	enabled := appState.biometricEnabled
	hasTemplate := len(appState.biometricTemplate) > 0
	appState.mu.Unlock()

	if !enabled || !hasTemplate {
		return
	}

	if err := ensureBiometricPipeline(appState); err != nil {
		log.Printf("biometric: cannot start continuous check — pipeline unavailable: %v", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	appState.mu.Lock()
	if appState.biometricStopCheck != nil {
		appState.biometricStopCheck() // stop the previous goroutine
	}
	appState.biometricStopCheck = cancel
	appState.mu.Unlock()

	go runContinuousCheck(ctx, appState, w, fyneApp)
}

// stopContinuousCheck cancels the background face-check goroutine, if any.
func stopContinuousCheck(appState *AppState) {
	appState.mu.Lock()
	defer appState.mu.Unlock()
	if appState.biometricStopCheck != nil {
		appState.biometricStopCheck()
		appState.biometricStopCheck = nil
	}
}

// runContinuousCheck is the body of the continuous verification goroutine.
// It opens a VideoCapture once, reads frames at ContinuousCheckIntervalMs,
// and calls the lock flow after MaxConsecutiveFailures consecutive failures.
func runContinuousCheck(ctx context.Context, appState *AppState, w fyne.Window, fyneApp fyne.App) {
	webcam, _, err := openBiometricCamera(appState)
	if err != nil {
		// Camera unavailable — fail closed: lock immediately.
		log.Printf("biometric: continuous check could not open camera: %v", err)
		fyne.Do(func() {
			lockSessionAndShowWarning(appState, w, fyneApp,
				"Face verification camera is unavailable. Session locked for security.")
		})
		return
	}
	defer webcam.Close()

	frame := gocv.NewMat()
	defer frame.Close()

	ticker := time.NewTicker(time.Duration(biometric.ContinuousCheckIntervalMs) * time.Millisecond)
	defer ticker.Stop()

	failures := 0

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			appState.mu.Lock()
			runtime := appState.biometricRuntime
			template := appState.biometricTemplate
			threshold := biometric.EffectiveThreshold(appState.biometricThreshold)
			stillEnabled := appState.biometricEnabled
			appState.mu.Unlock()

			if !stillEnabled || runtime == nil || template == nil {
				return // biometric was disabled mid-session
			}

			if !webcam.Read(&frame) || frame.Empty() {
				failures++
			} else {
				landmarks, found, err := runtime.RunFrame(frame)
				if err != nil || !found {
					failures++
				} else {
					features := biometric.ExtractFeatures(landmarks)
					if features == nil || biometric.CosineSimilarity(features, template) < threshold {
						failures++
					} else {
						failures = 0 // successful verification resets the counter
					}
				}
			}

			if failures >= biometric.MaxConsecutiveFailures {
				fyne.Do(func() {
					lockSessionAndShowWarning(appState, w, fyneApp,
						"Face verification failed. Session locked.")
				})
				return
			}
		}
	}
}

// lockSessionAndShowWarning clears sensitive state and navigates back to the
// unlock screen. Must be called on the Fyne UI goroutine.
func lockSessionAndShowWarning(appState *AppState, w fyne.Window, fyneApp fyne.App, message string) {
	appState.clearSensitiveState()
	ShowAppWarning("Session Locked", message, w)
	PromptMasterPassword(w, fyneApp, appState)
}
