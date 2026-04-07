//go:build !nobiometric

// Package biometric provides 3D face mesh recognition using MediaPipe Face Mesh
// (via ONNX + GoCV) for continuous identity verification inside PassQuantum.
//
// Pipeline overview:
//  1. BlazeFace (blazeface.onnx) detects and crops the face region from a frame.
//  2. MediaPipe Face Mesh (face_mesh.onnx) predicts 468 3D landmarks from the crop.
//  3. ExtractFeatures computes an interocular-normalised distance vector.
//  4. CosineSimilarity compares two feature vectors; ≥ DefaultSimilarityThreshold
//     means the same identity.
//
// Both ONNX models run CPU-only (NetBackendDefault / NetTargetCPU). Target throughput
// on a modern laptop: 15–30 fps for the full pipeline.
package biometric

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"runtime"
	"sync"

	"gocv.io/x/gocv"
)

const (
	faceDetectorInputSize = 128 // BlazeFace short-range input resolution (px)
	faceMeshInputSize     = 192 // MediaPipe Face Mesh input resolution (px)
	faceMeshLandmarkCount = 468
)

var blazeFaceBlobShapeLogOnce sync.Once

// FaceDetector wraps a BlazeFace ONNX net for face detection and ROI cropping.
type FaceDetector struct {
	net gocv.Net
}

// FaceMesh wraps a MediaPipe Face Mesh ONNX net for 468-point 3D landmark prediction.
type FaceMesh struct {
	net    gocv.Net
	Points []Landmark
}

// Pipeline combines face detection and mesh inference into a single end-to-end operation.
type Pipeline struct {
	Detector  *FaceDetector
	Mesh      *FaceMesh
	Threshold float32
}

// -------------------------------------------------------------------
// Constructors
// -------------------------------------------------------------------

// NewFaceDetector loads a BlazeFace ONNX model from modelPath and configures
// it for CPU-only inference.
func NewFaceDetector(modelPath string) (*FaceDetector, error) {
	if _, err := os.Stat(modelPath); err != nil {
		return nil, fmt.Errorf("BlazeFace model not accessible at %q: %w", modelPath, err)
	}
	if err := ValidateModelFile(modelPath); err != nil {
		return nil, fmt.Errorf("BlazeFace model at %q is invalid: %w", modelPath, err)
	}
	net := gocv.ReadNetFromONNX(modelPath)
	if net.Empty() {
		return nil, fmt.Errorf("failed to load BlazeFace model from %q: verify the file exists and is a valid ONNX model", modelPath)
	}
	net.SetPreferableBackend(gocv.NetBackendDefault)
	net.SetPreferableTarget(gocv.NetTargetCPU)
	return &FaceDetector{net: net}, nil
}

// Close releases native resources held by the FaceDetector.
func (fd *FaceDetector) Close() {
	fd.net.Close()
}

// NewFaceMesh loads a MediaPipe Face Mesh ONNX model from modelPath and configures
// it for CPU-only inference.
func NewFaceMesh(modelPath string) (*FaceMesh, error) {
	if _, err := os.Stat(modelPath); err != nil {
		return nil, fmt.Errorf("Face Mesh model not accessible at %q: %w", modelPath, err)
	}
	if err := ValidateModelFile(modelPath); err != nil {
		return nil, fmt.Errorf("Face Mesh model at %q is invalid: %w", modelPath, err)
	}
	opsetLabel, err := validateFaceMeshOpenCVCompatibility(modelPath)
	if err != nil {
		return nil, err
	}
	net := gocv.ReadNetFromONNX(modelPath)
	if net.Empty() {
		return nil, fmt.Errorf("failed to load Face Mesh model from %q: verify the file exists and is a valid ONNX model", modelPath)
	}
	net.SetPreferableBackend(gocv.NetBackendDefault)
	net.SetPreferableTarget(gocv.NetTargetCPU)
	log.Printf("biometric: loaded Face Mesh model %q (opset=%s)", modelPath, opsetLabel)
	return &FaceMesh{net: net}, nil
}

func validateFaceMeshOpenCVCompatibility(modelPath string) (string, error) {
	opencvVersion := gocv.OpenCVVersion()

	data, err := os.ReadFile(modelPath)
	if err != nil {
		return "unknown", fmt.Errorf("failed to read Face Mesh model for compatibility check: %w", err)
	}

	opsetLabel := "unknown"
	if opsetVersion, ok := detectDefaultONNXOpsetVersion(data); ok {
		opsetLabel = fmt.Sprintf("%d", opsetVersion)
	}

	if runtime.GOOS != "windows" {
		return opsetLabel, nil
	}

	// This model signature is known to trigger native OpenCV DNN crashes on some
	// Windows builds instead of returning a regular Go error.
	if bytes.Contains(data, []byte("Split")) {
		return opsetLabel, fmt.Errorf(
			"Face Mesh model %q is incompatible with OpenCV %s on Windows (Split-node ONNX parser issue that can crash the process). Replace models with PINTO OpenCV-compatible exports: face_mesh.onnx=%s blazeface.onnx=%s",
			modelPath,
			opencvVersion,
			"https://github.com/PINTO0309/PINTO_model_zoo/raw/main/032_FaceMesh/01_float32/face_mesh.onnx",
			"https://github.com/PINTO0309/PINTO_model_zoo/raw/main/030_BlazeFace/01_float32/blazeface.onnx",
		)
	}

	return opsetLabel, nil
}

func detectDefaultONNXOpsetVersion(data []byte) (int64, bool) {
	idx := 0
	var fallbackVersion int64
	hasFallback := false

	for idx < len(data) {
		tag, n := binary.Uvarint(data[idx:])
		if n <= 0 {
			return 0, false
		}
		idx += n

		fieldNum := int(tag >> 3)
		wireType := int(tag & 0x7)

		switch wireType {
		case 0:
			_, n = binary.Uvarint(data[idx:])
			if n <= 0 {
				return 0, false
			}
			idx += n
		case 1:
			if idx+8 > len(data) {
				return 0, false
			}
			idx += 8
		case 2:
			msgLen, n := binary.Uvarint(data[idx:])
			if n <= 0 {
				return 0, false
			}
			idx += n
			end := idx + int(msgLen)
			if end > len(data) {
				return 0, false
			}

			if fieldNum == 8 {
				version, isDefault, ok := parseOpsetImportEntry(data[idx:end])
				if ok {
					if isDefault {
						return version, true
					}
					if !hasFallback {
						fallbackVersion = version
						hasFallback = true
					}
				}
			}

			idx = end
		case 5:
			if idx+4 > len(data) {
				return 0, false
			}
			idx += 4
		default:
			return 0, false
		}
	}

	return fallbackVersion, hasFallback
}

func parseOpsetImportEntry(data []byte) (int64, bool, bool) {
	idx := 0
	domain := ""
	var version int64
	hasVersion := false

	for idx < len(data) {
		tag, n := binary.Uvarint(data[idx:])
		if n <= 0 {
			return 0, false, false
		}
		idx += n

		fieldNum := int(tag >> 3)
		wireType := int(tag & 0x7)

		switch wireType {
		case 0:
			value, n := binary.Uvarint(data[idx:])
			if n <= 0 {
				return 0, false, false
			}
			idx += n
			if fieldNum == 2 {
				version = int64(value)
				hasVersion = true
			}
		case 1:
			if idx+8 > len(data) {
				return 0, false, false
			}
			idx += 8
		case 2:
			fieldLen, n := binary.Uvarint(data[idx:])
			if n <= 0 {
				return 0, false, false
			}
			idx += n
			end := idx + int(fieldLen)
			if end > len(data) {
				return 0, false, false
			}
			if fieldNum == 1 {
				domain = string(data[idx:end])
			}
			idx = end
		case 5:
			if idx+4 > len(data) {
				return 0, false, false
			}
			idx += 4
		default:
			return 0, false, false
		}
	}

	if !hasVersion {
		return 0, false, false
	}

	return version, domain == "", true
}

// Close releases native resources held by the FaceMesh.
func (fm *FaceMesh) Close() {
	fm.net.Close()
}

// NewPipeline loads both ONNX models and returns a ready Pipeline with the given
// similarity threshold. On error both the partial pipeline and the error are returned;
// the caller need not close on error.
func NewPipeline(blazeFaceModelPath, faceMeshModelPath string, threshold float32) (*Pipeline, error) {
	detector, err := NewFaceDetector(blazeFaceModelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise face detector: %w", err)
	}
	mesh, err := NewFaceMesh(faceMeshModelPath)
	if err != nil {
		detector.Close()
		return nil, fmt.Errorf("failed to initialise face mesh: %w", err)
	}
	return &Pipeline{Detector: detector, Mesh: mesh, Threshold: threshold}, nil
}

// Close releases all native resources held by the Pipeline.
func (p *Pipeline) Close() {
	p.Detector.Close()
	p.Mesh.Close()
}

// -------------------------------------------------------------------
// Detection
// -------------------------------------------------------------------

// DetectAndCrop runs BlazeFace on frame and returns a cloned crop of the face ROI.
//
// Expected ONNX output format:
//   - A single tensor of shape [1, N, 5] where each detection row is
//     [confidence, y_min, x_min, y_max, x_max] in [0,1]-normalised coordinates.
//
// If no high-confidence face is found, or the output format is not recognised,
// the function returns a square centre-crop of frame as a fallback so that the
// mesh stage can still proceed. The returned Mat is always a Clone — the caller
// must call Close() on it.
func (fd *FaceDetector) DetectAndCrop(frame gocv.Mat) (gocv.Mat, error) {
	if frame.Empty() {
		return gocv.NewMat(), fmt.Errorf("cannot detect face in an empty frame")
	}

	frameH := frame.Rows()
	frameW := frame.Cols()

	resized := gocv.NewMat()
	defer resized.Close()
	gocv.Resize(frame, &resized,
		image.Pt(faceDetectorInputSize, faceDetectorInputSize),
		0, 0, gocv.InterpolationLinear)

	blob := gocv.BlobFromImage(
		resized,
		1.0/127.5,
		image.Pt(faceDetectorInputSize, faceDetectorInputSize),
		gocv.NewScalar(127.5, 127.5, 127.5, 0),
		true, false,
	)
	defer blob.Close()

	blobShape := blob.Size()
	if len(blobShape) != 4 ||
		blobShape[0] != 1 ||
		blobShape[1] != 3 ||
		blobShape[2] != faceDetectorInputSize ||
		blobShape[3] != faceDetectorInputSize {
		return gocv.NewMat(), fmt.Errorf(
			"unexpected BlazeFace blob shape %v; expected [1 3 %d %d] (NCHW)",
			blobShape,
			faceDetectorInputSize,
			faceDetectorInputSize,
		)
	}

	fd.net.SetInput(blob, "")
	detections := fd.net.Forward("")
	defer detections.Close()

	blazeFaceBlobShapeLogOnce.Do(func() {
		log.Printf("biometric: BlazeFace forward OK, blob shape=%v", blobShape)
	})

	data, err := detections.DataPtrFloat32()
	if err != nil || len(data) < 5 {
		// Output format not recognised — use full-frame centre crop.
		return fullFrameCrop(frame), nil
	}

	var bestScore, bestY1, bestX1, bestY2, bestX2 float32
	numDets := len(data) / 5
	for i := 0; i < numDets; i++ {
		if score := data[i*5]; score > bestScore {
			bestScore = score
			bestY1 = data[i*5+1]
			bestX1 = data[i*5+2]
			bestY2 = data[i*5+3]
			bestX2 = data[i*5+4]
		}
	}

	const detConfThreshold = float32(0.5)
	if bestScore < detConfThreshold {
		return fullFrameCrop(frame), nil
	}

	x1 := clampInt(int(bestX1*float32(frameW)), 0, frameW)
	y1 := clampInt(int(bestY1*float32(frameH)), 0, frameH)
	x2 := clampInt(int(bestX2*float32(frameW)), 0, frameW)
	y2 := clampInt(int(bestY2*float32(frameH)), 0, frameH)

	if x2 <= x1 || y2 <= y1 {
		return fullFrameCrop(frame), nil
	}

	roi := frame.Region(image.Rect(x1, y1, x2, y2))
	crop := roi.Clone()
	roi.Close()
	return crop, nil
}

// fullFrameCrop returns a square centre-crop of frame (Clone). Used as fallback ROI.
func fullFrameCrop(frame gocv.Mat) gocv.Mat {
	h, w := frame.Rows(), frame.Cols()
	side := h
	if w < side {
		side = w
	}
	x := (w - side) / 2
	y := (h - side) / 2
	roi := frame.Region(image.Rect(x, y, x+side, y+side))
	crop := roi.Clone()
	roi.Close()
	return crop
}

// -------------------------------------------------------------------
// Mesh prediction
// -------------------------------------------------------------------

// Predict runs Face Mesh inference on frame and returns 468 3D landmarks in
// frame-pixel coordinates. The frame should be a cropped face region.
//
// Expected ONNX output format:
//   - A tensor containing at least 1404 float32 values (468 × [x, y, z]).
//     Coordinates are in [0, faceMeshInputSize] pixel space and are re-scaled
//     back to the original frame dimensions.
//
// The method also caches the landmarks in fm.Points.
func (fm *FaceMesh) Predict(frame gocv.Mat) ([]Landmark, error) {
	if frame.Empty() {
		return nil, fmt.Errorf("cannot run face mesh on an empty frame")
	}

	resized := gocv.NewMat()
	defer resized.Close()
	gocv.Resize(frame, &resized,
		image.Point{X: faceMeshInputSize, Y: faceMeshInputSize},
		0, 0, gocv.InterpolationLinear)

	// Preprocess: 1×3×192×192 float32 blob normalised to [-1, 1].
	blob := gocv.BlobFromImage(
		resized,
		1.0/127.5,
		image.Point{X: faceMeshInputSize, Y: faceMeshInputSize},
		gocv.NewScalar(127.5, 127.5, 127.5, 0),
		true, false,
	)
	defer blob.Close()

	fm.net.SetInput(blob, "")
	output := fm.net.Forward("")
	defer output.Close()

	data, err := output.DataPtrFloat32()
	if err != nil {
		return nil, fmt.Errorf("failed to read face mesh output: %w", err)
	}

	if len(data) < faceMeshLandmarkCount*3 {
		return nil, fmt.Errorf("unexpected face mesh output: got %d float32 values, want at least %d (%d landmarks × 3)",
			len(data), faceMeshLandmarkCount*3, faceMeshLandmarkCount)
	}

	scaleX := float32(frame.Cols()) / float32(faceMeshInputSize)
	scaleY := float32(frame.Rows()) / float32(faceMeshInputSize)

	landmarks := make([]Landmark, faceMeshLandmarkCount)
	for i := range landmarks {
		landmarks[i] = Landmark{
			X: data[i*3] * scaleX,
			Y: data[i*3+1] * scaleY,
			Z: data[i*3+2],
		}
	}

	fm.Points = landmarks
	return landmarks, nil
}

// -------------------------------------------------------------------
// Full pipeline
// -------------------------------------------------------------------

// Run executes the full detect → crop → mesh pipeline on frame.
// Returns (landmarks, true, nil) on success, (nil, false, nil) when frame is
// valid but the mesh stage returned no data, and (nil, false, err) on error.
func (p *Pipeline) Run(frame gocv.Mat) ([]Landmark, bool, error) {
	crop, err := p.Detector.DetectAndCrop(frame)
	if err != nil {
		return nil, false, fmt.Errorf("face detection failed: %w", err)
	}
	defer crop.Close()

	landmarks, err := p.Mesh.Predict(crop)
	if err != nil {
		return nil, false, fmt.Errorf("face mesh prediction failed: %w", err)
	}

	return landmarks, true, nil
}

// RunFrame adapts the concrete GoCV frame path to the backend-neutral runtime API.
func (p *Pipeline) RunFrame(frame any) ([]Landmark, bool, error) {
	mat, ok := frame.(gocv.Mat)
	if !ok {
		return nil, false, fmt.Errorf("invalid frame type %T; expected gocv.Mat", frame)
	}
	return p.Run(mat)
}

// BackendName identifies the inference backend used by this runtime.
func (p *Pipeline) BackendName() string {
	return BackendGoCV
}

// DrawMesh draws all 468 face mesh landmarks on frame as 1-pixel green circles.
// Intended for debug / enrolment UI preview only.
func DrawMesh(frame *gocv.Mat, landmarks []Landmark) {
	green := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	for i := range landmarks {
		pt := image.Point{X: int(landmarks[i].X), Y: int(landmarks[i].Y)}
		gocv.Circle(frame, pt, 1, green, -1)
	}
}

// -------------------------------------------------------------------
// Internal helpers
// -------------------------------------------------------------------

// clampInt clamps v to the closed interval [lo, hi].
func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
