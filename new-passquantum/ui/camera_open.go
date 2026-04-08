//go:build !nobiometric && cgo

package main

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"

	"gocv.io/x/gocv"
)

const windowsCameraProbeMaxIndex = 3

// OpenCameraWindowsByName tries to open a Windows camera using the device's friendly name.
// It attempts three strategies in order: DSHOW by device name, GStreamer by device name,
// then CAP_ANY by numeric index. Returns the first camera that produces a real frame.
func OpenCameraWindowsByName(deviceName string) (*gocv.VideoCapture, error) {
	// Attempt 1 — DSHOW by device name
	dshowStr := fmt.Sprintf("video=%s", deviceName)
	cam, err := gocv.OpenVideoCapture(dshowStr)
	if err == nil && cam != nil && cam.IsOpened() {
		f := gocv.NewMat()
		cam.Read(&f)
		empty := f.Empty()
		f.Close()
		if !empty {
			warmupCamera(cam)
			log.Printf("camera: opened %q via DSHOW device-name string (attempt 1)", deviceName)
			return cam, nil
		}
		cam.Close()
	} else if cam != nil {
		cam.Close()
	}

	// Attempt 2 — GStreamer by device name
	gstStr := fmt.Sprintf(`ksvideosrc device-name="%s" ! videoconvert ! appsink`, deviceName)
	cam, err = gocv.OpenVideoCapture(gstStr)
	if err == nil && cam != nil && cam.IsOpened() {
		f := gocv.NewMat()
		cam.Read(&f)
		empty := f.Empty()
		f.Close()
		if !empty {
			warmupCamera(cam)
			log.Printf("camera: opened %q via GStreamer device-name string (attempt 2)", deviceName)
			return cam, nil
		}
		cam.Close()
	} else if cam != nil {
		cam.Close()
	}

	// Attempt 3 — CAP_ANY by numeric index with frame validation
	for i := 0; i < 3; i++ {
		cam, err = gocv.OpenVideoCapture(i)
		if err == nil && cam != nil && cam.IsOpened() {
			f := gocv.NewMat()
			cam.Read(&f)
			empty := f.Empty()
			f.Close()
			if !empty {
				warmupCamera(cam)
				log.Printf("camera: opened index %d via CAP_ANY (attempt 3, device name=%q)", i, deviceName)
				return cam, nil
			}
			cam.Close()
		} else if cam != nil {
			cam.Close()
		}
	}

	return nil, fmt.Errorf("camera: could not open %q on Windows (tried DSHOW name, GStreamer name, CAP_ANY indices 0-2)", deviceName)
}

func OpenCamera(deviceID int) (*gocv.VideoCapture, error) {
	if runtime.GOOS == "windows" {
		cam, err := gocv.OpenVideoCaptureWithAPI(deviceID, gocv.VideoCaptureDshow)
		if err == nil && cam != nil && cam.IsOpened() {
			warmupCamera(cam)
			log.Printf("camera: opened device %d with backend DSHOW", deviceID)
			return cam, nil
		}
		if cam != nil {
			cam.Close()
		}

		return nil, fmt.Errorf("camera: could not open device %d on Windows (tried DSHOW)", deviceID)
	}

	cam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		return nil, fmt.Errorf("camera: could not open device %d: %w", deviceID, err)
	}
	if cam == nil || !cam.IsOpened() {
		if cam != nil {
			cam.Close()
		}
		return nil, fmt.Errorf("camera: opened handle for device %d but camera is not ready", deviceID)
	}

	log.Printf("camera: opened device %d with default backend", deviceID)
	return cam, nil
}

func GetCameraDevicePaths() []string {
	if runtime.GOOS != "windows" {
		return nil
	}

	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		`Get-PnpDevice -Class Camera -ErrorAction SilentlyContinue | Where-Object {$_.Status -eq 'OK'} | Select-Object -ExpandProperty FriendlyName`)
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var names []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		name := strings.TrimSpace(line)
		if name != "" {
			names = append(names, name)
		}
	}

	return names
}

func OpenCameraWindows() (*gocv.VideoCapture, int, error) {
	names := GetCameraDevicePaths()
	if len(names) > 0 {
		log.Printf("camera: Windows detected devices: %s", strings.Join(names, ", "))
		// Try each detected device by name first (most reliable on DSHOW-only builds)
		for i, name := range names {
			cam, err := OpenCameraWindowsByName(name)
			if err == nil {
				return cam, i, nil
			}
			log.Printf("camera: device-name open failed for %q: %v", name, err)
		}
	}

	// Fallback: probe by numeric index
	for idx := 0; idx <= windowsCameraProbeMaxIndex; idx++ {
		cam, err := OpenCamera(idx)
		if err != nil {
			continue
		}

		frame := gocv.NewMat()
		ok := cam.Read(&frame) && !frame.Empty()
		frame.Close()
		if ok {
			return cam, idx, nil
		}

		cam.Close()
	}

	return nil, -1, fmt.Errorf("camera: no functional Windows camera found (tried device names + indices 0-%d)", windowsCameraProbeMaxIndex)
}

func warmupCamera(cam *gocv.VideoCapture) {
	if runtime.GOOS != "windows" {
		return
	}
	warmup := gocv.NewMat()
	defer warmup.Close()
	for i := 0; i < 5; i++ {
		cam.Read(&warmup)
	}
}
