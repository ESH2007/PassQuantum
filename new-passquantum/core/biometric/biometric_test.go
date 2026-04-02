package biometric

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// -------------------------------------------------------------------
// CosineSimilarity — pure math, no GoCV or model files required
// -------------------------------------------------------------------

func TestCosineSimilarityIdentical(t *testing.T) {
	v := []float32{1, 2, 3, 4, 5}
	sim := CosineSimilarity(v, v)
	if math.Abs(float64(sim)-1.0) > 1e-5 {
		t.Fatalf("CosineSimilarity(v, v) = %v, want 1.0", sim)
	}
}

func TestCosineSimilarityOrthogonal(t *testing.T) {
	a := []float32{1, 0}
	b := []float32{0, 1}
	sim := CosineSimilarity(a, b)
	if math.Abs(float64(sim)) > 1e-5 {
		t.Fatalf("CosineSimilarity(orthogonal) = %v, want ~0.0", sim)
	}
}

func TestCosineSimilarityOpposite(t *testing.T) {
	a := []float32{1, 2, 3}
	b := []float32{-1, -2, -3}
	sim := CosineSimilarity(a, b)
	if math.Abs(float64(sim)+1.0) > 1e-5 {
		t.Fatalf("CosineSimilarity(opposite) = %v, want -1.0", sim)
	}
}

func TestCosineSimilarityNilInputs(t *testing.T) {
	if sim := CosineSimilarity(nil, nil); sim != 0 {
		t.Fatalf("CosineSimilarity(nil, nil) = %v, want 0", sim)
	}
}

func TestCosineSimilarityLengthMismatch(t *testing.T) {
	a := []float32{1, 2, 3}
	b := []float32{1, 2}
	if sim := CosineSimilarity(a, b); sim != 0 {
		t.Fatalf("CosineSimilarity(length mismatch) = %v, want 0", sim)
	}
}

func TestCosineSimilarityZeroVectors(t *testing.T) {
	a := []float32{0, 0, 0}
	b := []float32{0, 0, 0}
	if sim := CosineSimilarity(a, b); sim != 0 {
		t.Fatalf("CosineSimilarity(zero, zero) = %v, want 0", sim)
	}
}

func TestCosineSimilarityAboveThreshold(t *testing.T) {
	base := []float32{1.0, 0.9, 1.1, 0.95, 1.05, 0.98, 1.02, 0.91, 1.03, 0.94, 0.97}
	// Tiny per-element perturbation — the same person.
	perturbed := make([]float32, len(base))
	for i, v := range base {
		perturbed[i] = v * (1 + float32(i)*0.001)
	}
	sim := CosineSimilarity(base, perturbed)
	if sim < DefaultSimilarityThreshold {
		t.Fatalf("CosineSimilarity(similar vectors) = %v, want >= %v (DefaultSimilarityThreshold)",
			sim, DefaultSimilarityThreshold)
	}
}

func TestCosineSimilarityBelowThreshold(t *testing.T) {
	a := []float32{1.0, 0.5, 2.0, 0.3, 1.8, 0.6, 2.1, 0.4, 1.9, 0.7, 2.2}
	b := []float32{2.0, 2.0, 0.2, 2.0, 0.1, 2.1, 0.3, 2.2, 0.2, 1.9, 0.4} // very different face
	sim := CosineSimilarity(a, b)
	if sim >= DefaultSimilarityThreshold {
		t.Fatalf("CosineSimilarity(different faces) = %v, want < %v", sim, DefaultSimilarityThreshold)
	}
}

// -------------------------------------------------------------------
// SerializeFeatures / DeserializeFeatures
// -------------------------------------------------------------------

func TestSerializeDeserializeRoundTrip(t *testing.T) {
	original := []float32{1.0, 0.85, 0.92, 0.78, 0.80, 1.2, 1.3, 0.7, 0.71, 0.60, 0.65}

	serialized := SerializeFeatures(original)
	if len(serialized) != len(original)*4 {
		t.Fatalf("SerializeFeatures() byte len = %d, want %d", len(serialized), len(original)*4)
	}

	recovered, err := DeserializeFeatures(serialized)
	if err != nil {
		t.Fatalf("DeserializeFeatures() error = %v", err)
	}
	if len(recovered) != len(original) {
		t.Fatalf("DeserializeFeatures() element count = %d, want %d", len(recovered), len(original))
	}
	for i := range original {
		if math.Abs(float64(recovered[i]-original[i])) > 1e-6 {
			t.Fatalf("DeserializeFeatures() element[%d] = %v, want %v", i, recovered[i], original[i])
		}
	}
}

func TestDeserializeFeaturesInvalidLength(t *testing.T) {
	if _, err := DeserializeFeatures([]byte{1, 2, 3}); err == nil {
		t.Fatal("DeserializeFeatures(3 bytes) expected error, got nil")
	}
}

func TestDeserializeFeaturesEmpty(t *testing.T) {
	features, err := DeserializeFeatures([]byte{})
	if err != nil {
		t.Fatalf("DeserializeFeatures(empty) error = %v", err)
	}
	if len(features) != 0 {
		t.Fatalf("DeserializeFeatures(empty) len = %d, want 0", len(features))
	}
}

// -------------------------------------------------------------------
// ExtractFeatures
// -------------------------------------------------------------------

func TestExtractFeaturesValidLandmarks(t *testing.T) {
	lm := makeSyntheticLandmarks(1.0)
	features := ExtractFeatures(lm)
	if features == nil {
		t.Fatal("ExtractFeatures() = nil for valid landmarks")
	}
	if len(features) != 11 {
		t.Fatalf("ExtractFeatures() len = %d, want 11", len(features))
	}
}

func TestExtractFeaturesAnchorIsOne(t *testing.T) {
	lm := makeSyntheticLandmarks(1.0)
	features := ExtractFeatures(lm)
	if features == nil {
		t.Fatal("ExtractFeatures() = nil")
	}
	if math.Abs(float64(features[0])-1.0) > 1e-5 {
		t.Fatalf("ExtractFeatures()[0] (anchor) = %v, want 1.0", features[0])
	}
}

func TestExtractFeaturesScaleInvariance(t *testing.T) {
	lm1 := makeSyntheticLandmarks(1.0)
	lm2 := makeSyntheticLandmarks(3.5) // 3.5× scaled

	f1 := ExtractFeatures(lm1)
	f2 := ExtractFeatures(lm2)

	if f1 == nil || f2 == nil {
		t.Fatal("ExtractFeatures() returned nil for valid landmarks")
	}
	if len(f1) != len(f2) {
		t.Fatalf("ExtractFeatures() length mismatch: %d vs %d", len(f1), len(f2))
	}
	for i := range f1 {
		diff := math.Abs(float64(f1[i] - f2[i]))
		if diff > 1e-4 {
			t.Fatalf("ExtractFeatures() element[%d] differs after 3.5× scale: %v vs %v (diff=%v)",
				i, f1[i], f2[i], diff)
		}
	}
}

func TestExtractFeaturesNilOnTooFewLandmarks(t *testing.T) {
	if result := ExtractFeatures(make([]Landmark, 5)); result != nil {
		t.Fatalf("ExtractFeatures(5 landmarks) = %v, want nil", result)
	}
}

func TestExtractFeaturesNilOnZeroInterocular(t *testing.T) {
	lm := make([]Landmark, 468)
	// Left and right eye at same position → interocular == 0
	lm[idxLeftEye] = Landmark{X: 50, Y: 50}
	lm[idxRightEye] = Landmark{X: 50, Y: 50}
	if result := ExtractFeatures(lm); result != nil {
		t.Fatalf("ExtractFeatures(zero interocular) = %v, want nil", result)
	}
}

// -------------------------------------------------------------------
// EffectiveThreshold
// -------------------------------------------------------------------

func TestEffectiveThresholdUsesDefaultWhenZero(t *testing.T) {
	if got := EffectiveThreshold(0); got != DefaultSimilarityThreshold {
		t.Fatalf("EffectiveThreshold(0) = %v, want %v", got, DefaultSimilarityThreshold)
	}
}

func TestEffectiveThresholdUsesDefaultWhenNegative(t *testing.T) {
	if got := EffectiveThreshold(-1); got != DefaultSimilarityThreshold {
		t.Fatalf("EffectiveThreshold(-1) = %v, want %v", got, DefaultSimilarityThreshold)
	}
}

func TestEffectiveThresholdUsesStoredWhenPositive(t *testing.T) {
	const stored = float32(0.95)
	if got := EffectiveThreshold(stored); got != stored {
		t.Fatalf("EffectiveThreshold(%v) = %v, want %v", stored, got, stored)
	}
}

// -------------------------------------------------------------------
// Model path resolution
// -------------------------------------------------------------------

func TestResolveDefaultModelPathsFromEnv(t *testing.T) {
	t.Setenv(modelDirEnvVar, t.TempDir())
	modelDir := os.Getenv(modelDirEnvVar)

	blazePath := filepath.Join(modelDir, filepath.Base(DefaultBlazeFaceModelPath))
	meshPath := filepath.Join(modelDir, filepath.Base(DefaultFaceMeshModelPath))

	blazeData := append([]byte{0x08, 0x01, 0x12, 0x07}, make([]byte, 2048)...)
	meshData := append([]byte{0x08, 0x01, 0x12, 0x07}, make([]byte, 2048)...)

	if err := os.WriteFile(blazePath, blazeData, 0o600); err != nil {
		t.Fatalf("WriteFile(blaze) error = %v", err)
	}
	if err := os.WriteFile(meshPath, meshData, 0o600); err != nil {
		t.Fatalf("WriteFile(mesh) error = %v", err)
	}

	resolvedBlaze, resolvedMesh, err := ResolveDefaultModelPaths()
	if err != nil {
		t.Fatalf("ResolveDefaultModelPaths() error = %v", err)
	}
	if resolvedBlaze != blazePath {
		t.Fatalf("ResolveDefaultModelPaths() blaze = %q, want %q", resolvedBlaze, blazePath)
	}
	if resolvedMesh != meshPath {
		t.Fatalf("ResolveDefaultModelPaths() mesh = %q, want %q", resolvedMesh, meshPath)
	}
}

func TestCandidateModelPathsIncludesExecutableLayouts(t *testing.T) {
	paths := candidateModelPaths(DefaultBlazeFaceModelPath)

	if len(paths) == 0 {
		t.Fatal("candidateModelPaths() returned no candidates")
	}

	want := []string{
		filepath.Clean(DefaultBlazeFaceModelPath),
		filepath.Clean(filepath.Join("new-passquantum", DefaultBlazeFaceModelPath)),
	}

	for _, expected := range want {
		if !containsPath(paths, expected) {
			t.Fatalf("candidateModelPaths() missing %q from %v", expected, paths)
		}
	}
}

func TestValidateModelFileRejectsHTML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "blazeface.onnx")
	html := []byte("<!DOCTYPE html><html><body>not-a-model</body></html>")
	if err := os.WriteFile(path, append(html, make([]byte, 2048)...), 0o600); err != nil {
		t.Fatalf("WriteFile(html) error = %v", err)
	}

	err := ValidateModelFile(path)
	if err == nil {
		t.Fatal("ValidateModelFile(html) expected error, got nil")
	}
}

func TestValidateModelFileRejectsLFSPointer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "blazeface.onnx")
	lfs := []byte("version https://git-lfs.github.com/spec/v1\noid sha256:abc\nsize 123\n")
	if err := os.WriteFile(path, append(lfs, make([]byte, 2048)...), 0o600); err != nil {
		t.Fatalf("WriteFile(lfs) error = %v", err)
	}

	err := ValidateModelFile(path)
	if err == nil {
		t.Fatal("ValidateModelFile(lfs) expected error, got nil")
	}
}

func TestResolveModelPathSkipsInvalidCandidate(t *testing.T) {
	modelDir := t.TempDir()
	t.Setenv(modelDirEnvVar, modelDir)

	invalidPreferred := filepath.Join(modelDir, filepath.Base(DefaultBlazeFaceModelPath))
	validFallback := filepath.Join(modelDir, DefaultBlazeFaceModelPath)

	if err := os.MkdirAll(filepath.Dir(validFallback), 0o755); err != nil {
		t.Fatalf("MkdirAll(valid fallback) error = %v", err)
	}
	if err := os.WriteFile(invalidPreferred, append([]byte("<!DOCTYPE html>"), make([]byte, 2048)...), 0o600); err != nil {
		t.Fatalf("WriteFile(invalid preferred) error = %v", err)
	}
	if err := os.WriteFile(validFallback, append([]byte{0x08, 0x01, 0x12, 0x07}, make([]byte, 2048)...), 0o600); err != nil {
		t.Fatalf("WriteFile(valid fallback) error = %v", err)
	}

	resolved, err := resolveModelPath(DefaultBlazeFaceModelPath)
	if err != nil {
		t.Fatalf("resolveModelPath() error = %v", err)
	}
	if resolved != filepath.Clean(validFallback) {
		t.Fatalf("resolveModelPath() = %q, want %q", resolved, filepath.Clean(validFallback))
	}
}

func TestResolveModelPathInvalidCandidatesError(t *testing.T) {
	modelDir := t.TempDir()
	t.Setenv(modelDirEnvVar, modelDir)

	invalidPreferred := filepath.Join(modelDir, filepath.Base(DefaultBlazeFaceModelPath))
	invalidFallback := filepath.Join(modelDir, DefaultBlazeFaceModelPath)

	if err := os.MkdirAll(filepath.Dir(invalidFallback), 0o755); err != nil {
		t.Fatalf("MkdirAll(invalid fallback) error = %v", err)
	}
	if err := os.WriteFile(invalidPreferred, append([]byte("<!DOCTYPE html>"), make([]byte, 2048)...), 0o600); err != nil {
		t.Fatalf("WriteFile(invalid preferred) error = %v", err)
	}
	if err := os.WriteFile(invalidFallback, append([]byte("version https://git-lfs.github.com/spec/v1\n"), make([]byte, 2048)...), 0o600); err != nil {
		t.Fatalf("WriteFile(invalid fallback) error = %v", err)
	}

	_, err := resolveModelPath(DefaultBlazeFaceModelPath)
	if err == nil {
		t.Fatal("resolveModelPath() expected error for invalid candidates, got nil")
	}
	if !errors.Is(err, os.ErrNotExist) && err.Error() == "" {
		t.Fatal("resolveModelPath() returned empty error")
	}
}

// -------------------------------------------------------------------
// Model-dependent tests (skipped when model files are absent)
// -------------------------------------------------------------------

func TestNewPipelineSkipWithoutModels(t *testing.T) {
	if _, err := os.Stat(DefaultBlazeFaceModelPath); os.IsNotExist(err) {
		t.Skip("BlazeFace model not found at " + DefaultBlazeFaceModelPath + "; skipping model-load test")
	}
	if _, err := os.Stat(DefaultFaceMeshModelPath); os.IsNotExist(err) {
		t.Skip("Face Mesh model not found at " + DefaultFaceMeshModelPath + "; skipping model-load test")
	}

	p, err := NewPipeline(DefaultBlazeFaceModelPath, DefaultFaceMeshModelPath, DefaultSimilarityThreshold)
	if err != nil {
		t.Fatalf("NewPipeline() error = %v", err)
	}
	defer p.Close()

	if p.Threshold != DefaultSimilarityThreshold {
		t.Fatalf("Pipeline.Threshold = %v, want %v", p.Threshold, DefaultSimilarityThreshold)
	}
}

func TestNewFaceDetectorMissingFile(t *testing.T) {
	_, err := NewFaceDetector("/nonexistent/blazeface.onnx")
	if err == nil {
		t.Fatal("NewFaceDetector(missing) expected error, got nil")
	}
}

func TestNewFaceMeshMissingFile(t *testing.T) {
	_, err := NewFaceMesh("/nonexistent/face_mesh.onnx")
	if err == nil {
		t.Fatal("NewFaceMesh(missing) expected error, got nil")
	}
}

// -------------------------------------------------------------------
// Test fixture
// -------------------------------------------------------------------

// makeSyntheticLandmarks returns a 468-element slice with key landmark indices
// placed at known positions, all scaled by the given factor.
func makeSyntheticLandmarks(scale float32) []Landmark {
	lm := make([]Landmark, 468)
	lm[idxNoseTip] = Landmark{X: 100 * scale, Y: 150 * scale}
	lm[idxNoseEnd] = Landmark{X: 100 * scale, Y: 170 * scale}
	lm[idxLeftEye] = Landmark{X: 70 * scale, Y: 100 * scale}
	lm[idxRightEye] = Landmark{X: 130 * scale, Y: 100 * scale}
	lm[idxLeftMouth] = Landmark{X: 80 * scale, Y: 200 * scale}
	lm[idxRightMouth] = Landmark{X: 120 * scale, Y: 200 * scale}
	lm[idxChin] = Landmark{X: 100 * scale, Y: 240 * scale}
	lm[idxNoseBridge] = Landmark{X: 100 * scale, Y: 120 * scale}
	return lm
}

func containsPath(paths []string, target string) bool {
	for _, path := range paths {
		if filepath.Clean(path) == filepath.Clean(target) {
			return true
		}
	}
	return false
}
