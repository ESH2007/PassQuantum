//go:build nobiometric

package main

import (
	"context"
	"fmt"
	"image"

	"fyne.io/fyne/v2"

	"passquantum/core/storage"
)

func loadBiometricFromProfile(appState *AppState) {
	appState.mu.Lock()
	appState.biometricEnabled = false
	appState.biometricTemplate = nil
	appState.biometricThreshold = 0
	appState.biometricRuntime = nil
	appState.mu.Unlock()
}

func saveBiometricToProfile(appState *AppState) error {
	appState.mu.Lock()
	profile := appState.securityProfile
	if profile == nil {
		appState.mu.Unlock()
		return fmt.Errorf("no active security profile to update")
	}
	profile.Biometric.Enabled = false
	profile.BiometricTemplate = nil
	appState.mu.Unlock()

	if err := storage.SaveAppSecurityProfile(appSecurityMetadataPath, profile); err != nil {
		return fmt.Errorf("failed to save biometric settings: %w", err)
	}
	return nil
}

func ensureBiometricPipeline(appState *AppState) error {
	return fmt.Errorf("biometric support disabled in this build")
}

func closeBiometricPipeline(appState *AppState) {}

func captureAndVerifyFace(appState *AppState) (bool, error) {
	return false, fmt.Errorf("biometric support disabled in this build")
}

func captureEnrolmentFrame(appState *AppState) ([]float32, *image.NRGBA, error) {
	return nil, nil, fmt.Errorf("biometric support disabled in this build")
}

func startContinuousCheck(appState *AppState, w fyne.Window, fyneApp fyne.App) {}

func stopContinuousCheck(appState *AppState) {
	appState.mu.Lock()
	defer appState.mu.Unlock()
	if appState.biometricStopCheck != nil {
		appState.biometricStopCheck()
		appState.biometricStopCheck = nil
	}
}

func runContinuousCheck(ctx context.Context, appState *AppState, w fyne.Window, fyneApp fyne.App) {}

func lockSessionAndShowWarning(appState *AppState, w fyne.Window, fyneApp fyne.App, message string) {}
