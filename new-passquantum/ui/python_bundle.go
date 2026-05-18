//go:build with_face_bundle && !windows

package main

// python_bundle.go — embeds the PyInstaller face_guard bundle into the binary.
//
// This file is compiled only when the build tag "with_face_bundle" is set
// (i.e. only by build.sh after PyInstaller has produced ui/face_guard_bundle).
// Regular `go build ./ui` without the tag compiles cleanly without this file.
//
// At startup the init() function below extracts the embedded bundle to a
// per-user temp directory and tells buildPythonCommand (face_guard.go) where
// to find it via the PASSQUANTUM_FACE_GUARD_BUNDLE environment variable.

import (
	_ "embed"
	"log"
	"os"
	"path/filepath"
)

//go:embed face_guard_bundle
var faceGuardBundleData []byte

func init() {
	// Choose a stable extraction path so we only write the file once per
	// binary update.  Using os.TempDir() is safe on all platforms.
	dir := filepath.Join(os.TempDir(), "passquantum-face-guard")
	if err := os.MkdirAll(dir, 0700); err != nil {
		log.Printf("[FaceGuard] WARNING: could not create bundle dir %s: %v", dir, err)
		return
	}

	bundlePath := filepath.Join(dir, "face_guard")

	// Overwrite unconditionally so the binary always matches the embedded version.
	if err := os.WriteFile(bundlePath, faceGuardBundleData, 0700); err != nil {
		log.Printf("[FaceGuard] WARNING: could not extract face_guard bundle: %v", err)
		return
	}

	// Expose the path and the work directory (next to the PassQuantum executable)
	// so that face_data.npy is stored in a predictable persistent location.
	os.Setenv("PASSQUANTUM_FACE_GUARD_BUNDLE", bundlePath)

	if exe, err := os.Executable(); err == nil {
		os.Setenv("PASSQUANTUM_WORK_DIR", filepath.Dir(exe))
	}

	log.Printf("[FaceGuard] Embedded Python bundle extracted to %s", bundlePath)
}
