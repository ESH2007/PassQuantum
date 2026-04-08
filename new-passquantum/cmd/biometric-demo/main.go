//go:build !nobiometric && cgo

package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"runtime"

	"gocv.io/x/gocv"

	"passquantum/core/biometric"
)

func main() {
	cameraIndex := flag.Int("camera", 0, "camera index")
	threshold := flag.Float64("threshold", float64(biometric.DefaultSimilarityThreshold), "match threshold")
	flag.Parse()

	runtimeHandle, err := biometric.NewDefaultRuntime(float32(*threshold))
	if err != nil {
		log.Fatalf("failed to initialize biometric runtime: %v", err)
	}
	defer runtimeHandle.Close()

	cam, err := openCamera(*cameraIndex)
	if err != nil {
		log.Fatalf("failed to open camera: %v", err)
	}
	defer cam.Close()

	window := gocv.NewWindow("PassQuantum Face Recognition Demo")
	defer window.Close()

	frame := gocv.NewMat()
	defer frame.Close()

	var template []float32
	var lastFeatures []float32
	status := "Press E to enroll, C to clear, Q to quit"

	for {
		if ok := cam.Read(&frame); !ok || frame.Empty() {
			continue
		}

		landmarks, found, runErr := runtimeHandle.RunFrame(frame)
		if runErr != nil {
			status = fmt.Sprintf("Inference error: %v", runErr)
			drawOverlay(&frame, status, 0.0, false, template != nil)
			window.IMShow(frame)
			if shouldQuit(window.WaitKey(1)) {
				break
			}
			continue
		}

		similarity := float32(0)
		matched := false
		if found {
			biometric.DrawMesh(&frame, landmarks)
			features := biometric.ExtractFeatures(landmarks)
			if features != nil {
				lastFeatures = cloneFeatures(features)
				if template != nil {
					similarity = biometric.CosineSimilarity(features, template)
					matched = similarity >= float32(*threshold)
				}
			}
		}

		if !found {
			status = "No face detected"
		} else if template == nil {
			status = "Face detected. Press E to enroll current face"
		} else if matched {
			status = "Match"
		} else {
			status = "No match"
		}

		drawOverlay(&frame, status, similarity, matched, template != nil)
		window.IMShow(frame)

		key := window.WaitKey(1)
		switch {
		case shouldQuit(key):
			return
		case key == int('e') || key == int('E'):
			if lastFeatures == nil {
				status = "Cannot enroll: no valid face features"
			} else {
				template = cloneFeatures(lastFeatures)
				status = "Template enrolled from current face"
			}
		case key == int('c') || key == int('C'):
			template = nil
			status = "Template cleared"
		}
	}
}

func shouldQuit(key int) bool {
	return key == int('q') || key == int('Q') || key == 27
}

func cloneFeatures(features []float32) []float32 {
	if features == nil {
		return nil
	}
	out := make([]float32, len(features))
	copy(out, features)
	return out
}

func openCamera(index int) (*gocv.VideoCapture, error) {
	if runtime.GOOS == "windows" {
		cam, err := gocv.OpenVideoCaptureWithAPI(index, gocv.VideoCaptureDshow)
		if err == nil && cam != nil && cam.IsOpened() {
			return cam, nil
		}
		if cam != nil {
			cam.Close()
		}
	}

	cam, err := gocv.OpenVideoCapture(index)
	if err != nil {
		return nil, err
	}
	if cam == nil || !cam.IsOpened() {
		if cam != nil {
			cam.Close()
		}
		return nil, fmt.Errorf("camera %d is not ready", index)
	}
	return cam, nil
}

func drawOverlay(frame *gocv.Mat, status string, similarity float32, matched bool, hasTemplate bool) {
	line1 := status
	line2 := ""
	if hasTemplate {
		line2 = fmt.Sprintf("Similarity: %.4f", similarity)
	}

	white := colorRGB(255, 255, 255)
	green := colorRGB(40, 220, 80)
	red := colorRGB(220, 70, 70)

	gocv.PutText(frame, line1, imagePoint(20, 30), gocv.FontHersheySimplex, 0.7, white, 2)
	if line2 != "" {
		lineColor := red
		if matched {
			lineColor = green
		}
		gocv.PutText(frame, line2, imagePoint(20, 60), gocv.FontHersheySimplex, 0.7, lineColor, 2)
	}
	gocv.PutText(frame, "Keys: E=enroll C=clear Q=quit", imagePoint(20, frame.Rows()-20), gocv.FontHersheySimplex, 0.6, white, 2)
}

func imagePoint(x, y int) image.Point {
	return image.Point{X: x, Y: y}
}

func colorRGB(r, g, b uint8) color.RGBA {
	return color.RGBA{R: r, G: g, B: b, A: 255}
}
