//go:build !nobiometric

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gocv.io/x/gocv"
)

const biometricCameraEnvVar = "PASSQUANTUM_CAMERA_INDEX"

func openBiometricCamera() (*gocv.VideoCapture, int, error) {
	indices := candidateCameraIndices()
	var attempted []string

	for _, index := range indices {
		attempted = append(attempted, strconv.Itoa(index))

		cam, err := gocv.VideoCaptureDevice(index)
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
	indices := make([]int, 0, 6)

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
		}
	}

	for index := 0; index <= 5; index++ {
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
