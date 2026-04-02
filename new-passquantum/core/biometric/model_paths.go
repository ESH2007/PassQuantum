package biometric

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const modelDirEnvVar = "PASSQUANTUM_MODEL_DIR"

// ResolveDefaultModelPaths resolves the default BlazeFace and Face Mesh model paths
// from the runtime environment, packaged app layout, or repository layout.
func ResolveDefaultModelPaths() (string, string, error) {
	blazePath, err := resolveModelPath(DefaultBlazeFaceModelPath)
	if err != nil {
		return "", "", err
	}

	meshPath, err := resolveModelPath(DefaultFaceMeshModelPath)
	if err != nil {
		return "", "", err
	}

	return blazePath, meshPath, nil
}

func resolveModelPath(defaultRelativePath string) (string, error) {
	candidates := candidateModelPaths(defaultRelativePath)
	invalid := make([]string, 0)
	for _, candidate := range candidates {
		if !fileExists(candidate) {
			continue
		}
		if err := ValidateModelFile(candidate); err == nil {
			return candidate, nil
		} else {
			invalid = append(invalid, fmt.Sprintf("%s (%v)", candidate, err))
		}
	}

	if len(invalid) > 0 {
		return "", fmt.Errorf("could not locate a valid biometric model %q; invalid candidates: %s", defaultRelativePath, strings.Join(invalid, "; "))
	}

	return "", fmt.Errorf("could not locate biometric model %q; checked: %s", defaultRelativePath, strings.Join(candidates, ", "))
}

func candidateModelPaths(defaultRelativePath string) []string {
	seen := make(map[string]struct{})
	var paths []string

	add := func(path string) {
		if path == "" {
			return
		}
		cleaned := filepath.Clean(path)
		if _, ok := seen[cleaned]; ok {
			return
		}
		seen[cleaned] = struct{}{}
		paths = append(paths, cleaned)
	}

	if modelDir := strings.TrimSpace(os.Getenv(modelDirEnvVar)); modelDir != "" {
		add(filepath.Join(modelDir, filepath.Base(defaultRelativePath)))
		add(filepath.Join(modelDir, defaultRelativePath))
	}

	add(defaultRelativePath)
	add(filepath.Join("new-passquantum", defaultRelativePath))

	workingDir, err := os.Getwd()
	if err == nil {
		add(filepath.Join(workingDir, defaultRelativePath))
		add(filepath.Join(workingDir, "new-passquantum", defaultRelativePath))
	}

	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		add(filepath.Join(execDir, defaultRelativePath))
		add(filepath.Join(execDir, "models", filepath.Base(defaultRelativePath)))
		add(filepath.Join(execDir, "..", defaultRelativePath))
		add(filepath.Join(execDir, "..", "models", filepath.Base(defaultRelativePath)))
		add(filepath.Join(execDir, "..", "Resources", "models", filepath.Base(defaultRelativePath)))
	}

	return paths
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// ValidateModelFile rejects common non-ONNX placeholders that otherwise crash
// OpenCV's ONNX loader inside C++ code paths.
func ValidateModelFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("not accessible: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("path is a directory")
	}
	if info.Size() < 1024 {
		return fmt.Errorf("file too small to be a model (%d bytes)", info.Size())
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open model: %w", err)
	}
	defer f.Close()

	header := make([]byte, 256)
	n, err := io.ReadFull(f, header)
	if err != nil && err != io.ErrUnexpectedEOF {
		return fmt.Errorf("failed to read model header: %w", err)
	}
	header = header[:n]
	lower := bytes.ToLower(bytes.TrimSpace(header))

	if bytes.HasPrefix(lower, []byte("version https://git-lfs.github.com/spec/v1")) {
		return fmt.Errorf("git-lfs pointer detected instead of ONNX binary")
	}
	if bytes.HasPrefix(lower, []byte("<!doctype html")) || bytes.HasPrefix(lower, []byte("<html")) {
		return fmt.Errorf("html content detected instead of ONNX binary")
	}

	return nil
}
