package biometric

import (
	"encoding/binary"
	"fmt"
	"math"
)

const (
	// DefaultBlazeFaceModelPath is the repository-relative path to the BlazeFace ONNX model.
	DefaultBlazeFaceModelPath = "models/blazeface.onnx"

	// DefaultFaceMeshModelPath is the repository-relative path to the Face Mesh ONNX model.
	DefaultFaceMeshModelPath = "models/face_mesh.onnx"

	// DefaultSimilarityThreshold is the minimum cosine similarity required to accept the
	// same identity. Values below this cause a verification failure.
	DefaultSimilarityThreshold = float32(0.97)

	// ContinuousCheckIntervalMs is the number of milliseconds between consecutive
	// face verification checks while a session is live.
	ContinuousCheckIntervalMs = 1500

	// MaxConsecutiveFailures is the number of sequential failed checks before the
	// session is locked.
	MaxConsecutiveFailures = 3

	// MediaPipe Face Mesh landmark indices used for feature extraction.
	idxNoseTip    = 1
	idxNoseEnd    = 4
	idxLeftEye    = 33
	idxRightEye   = 263
	idxLeftMouth  = 61
	idxRightMouth = 291
	idxChin       = 152
	idxNoseBridge = 168
)

// Landmark is a single 3D face mesh point in frame-pixel space.
type Landmark struct {
	X, Y, Z float32
}

// RuntimeHandle is a backend-neutral biometric runtime used by the UI layer.
// The frame argument is backend-defined; current GoCV backend expects gocv.Mat.
type RuntimeHandle interface {
	RunFrame(frame any) ([]Landmark, bool, error)
	Close()
	BackendName() string
}

const BackendGoCV = "gocv"

// NewDefaultRuntime resolves model paths and initialises the default backend.
func NewDefaultRuntime(threshold float32) (RuntimeHandle, error) {
	blazeModelPath, meshModelPath, err := ResolveDefaultModelPaths()
	if err != nil {
		return nil, err
	}

	return NewPipeline(blazeModelPath, meshModelPath, threshold)
}

// ExtractFeatures computes an 11-element normalised feature vector from the
// distances between key MediaPipe landmark pairs. All distances are divided by
// the interocular distance for scale invariance.
//
// Returns nil when landmarks is too short or the interocular distance is zero.
func ExtractFeatures(landmarks []Landmark) []float32 {
	if len(landmarks) <= idxRightMouth {
		return nil
	}

	noseTip := landmarks[idxNoseTip]
	noseEnd := landmarks[idxNoseEnd]
	leftEye := landmarks[idxLeftEye]
	rightEye := landmarks[idxRightEye]
	leftMouth := landmarks[idxLeftMouth]
	rightMouth := landmarks[idxRightMouth]
	chin := landmarks[idxChin]
	noseBridge := landmarks[idxNoseBridge]

	ioD := dist2D(leftEye, rightEye)
	if ioD == 0 {
		return nil
	}

	return []float32{
		1.0,
		dist2D(leftMouth, rightMouth) / ioD,
		dist2D(noseBridge, noseEnd) / ioD,
		dist2D(leftEye, leftMouth) / ioD,
		dist2D(rightEye, rightMouth) / ioD,
		dist2D(noseTip, chin) / ioD,
		dist2D(noseBridge, chin) / ioD,
		dist2D(leftEye, noseTip) / ioD,
		dist2D(rightEye, noseTip) / ioD,
		dist2D(leftMouth, chin) / ioD,
		dist2D(rightMouth, chin) / ioD,
	}
}

// CosineSimilarity returns the cosine similarity in [-1, 1] of vectors a and b.
// Returns 0 for nil, empty, or length-mismatched inputs.
func CosineSimilarity(a, b []float32) float32 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}

	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0
	}
	return float32(dot / denom)
}

// SerializeFeatures converts a float32 feature vector to a compact byte slice
// using little-endian IEEE 754 encoding.
func SerializeFeatures(features []float32) []byte {
	buf := make([]byte, len(features)*4)
	for i, f := range features {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(f))
	}
	return buf
}

// DeserializeFeatures converts a byte slice produced by SerializeFeatures back to
// a float32 feature vector. Returns an error when the slice length is not a
// multiple of 4.
func DeserializeFeatures(data []byte) ([]float32, error) {
	if len(data)%4 != 0 {
		return nil, fmt.Errorf("biometric feature data length %d is not a multiple of 4", len(data))
	}
	features := make([]float32, len(data)/4)
	for i := range features {
		features[i] = math.Float32frombits(binary.LittleEndian.Uint32(data[i*4:]))
	}
	return features, nil
}

// EffectiveThreshold returns stored if it is above zero, otherwise DefaultSimilarityThreshold.
func EffectiveThreshold(stored float32) float32 {
	if stored > 0 {
		return stored
	}
	return DefaultSimilarityThreshold
}

// dist2D returns the 2-D Euclidean distance between Landmarks a and b.
func dist2D(a, b Landmark) float32 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return float32(math.Sqrt(float64(dx*dx + dy*dy)))
}
