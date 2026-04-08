//go:build !nobiometric && cgo

package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gocv.io/x/gocv"
)

const biometricCameraEnvVar = "PASSQUANTUM_CAMERA_INDEX"
const biometricCameraMaxIndexEnvVar = "PASSQUANTUM_CAMERA_MAX_INDEX"

func openBiometricCamera(appState *AppState) (*gocv.VideoCapture, int, error) {
	if appState != nil {
		appState.mu.Lock()
		preferred := appState.biometricCameraIndex
		appState.mu.Unlock()

		if preferred != nil {
			cam, err := OpenCamera(*preferred)
			if err == nil {
				if cameraProducesFrames(cam) {
					return cam, *preferred, nil
				}
				cam.Close()
			}
		}
	}

	if runtime.GOOS == "windows" {
		if configured := strings.TrimSpace(os.Getenv(biometricCameraEnvVar)); configured == "" {
			cam, idx, err := OpenCameraWindows()
			if err == nil {
				return cam, idx, nil
			}
		}
	}

	indices := candidateCameraIndices()
	var attempted []string

	for _, index := range indices {
		attempted = append(attempted, strconv.Itoa(index))

		cam, err := OpenCamera(index)
		if err != nil {
			continue
		}

		if cameraProducesFrames(cam) {
			return cam, index, nil
		}

		cam.Close()
	}

	return nil, -1, fmt.Errorf("could not open a working camera; tried indices %s", strings.Join(attempted, ", "))
}

func candidateCameraIndices() []int {
	seen := make(map[int]struct{})
	indices := make([]int, 0, 3)

	add := func(index int) {
		if index < 0 {
			return
		}
		if _, ok := seen[index]; ok {
			return
		}
		seen[index] = struct{}{}
		indices = append(indices, index)
	}

	if configured := strings.TrimSpace(os.Getenv(biometricCameraEnvVar)); configured != "" {
		if index, err := strconv.Atoi(configured); err == nil {
			add(index)
			return indices
		}
	}

	maxIndex := 2
	if runtime.GOOS == "windows" {
		// On Windows, probing multiple indexes often triggers noisy DSHOW warnings.
		// Default to camera 0 unless the user explicitly overrides max index.
		maxIndex = 0
	}
	if configuredMax := strings.TrimSpace(os.Getenv(biometricCameraMaxIndexEnvVar)); configuredMax != "" {
		if parsedMax, err := strconv.Atoi(configuredMax); err == nil && parsedMax >= 0 {
			maxIndex = parsedMax
		}
	}

	for index := 0; index <= maxIndex; index++ {
		add(index)
	}

	return indices
}

func cameraProducesFrames(cam *gocv.VideoCapture) bool {
	frame := gocv.NewMat()
	defer frame.Close()

	for attempt := 0; attempt < 8; attempt++ {
		if cam.Read(&frame) && !frame.Empty() {
			return true
		}
		time.Sleep(60 * time.Millisecond)
	}

	return false
}
