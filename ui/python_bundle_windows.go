//go:build with_face_bundle && windows

package main

// python_bundle_windows.go — embeds the PyInstaller face_guard bundle on Windows.
//
// Compiled only when both build tags are set:
//   -tags with_face_bundle    (set by Build-FaceBundle.ps1 after PyInstaller succeeds)
//   GOOS=windows              (automatic for any Windows build)
//
// At startup init() extracts face_guard_bundle.exe to a per-user temp directory
// and tells buildPythonCommand (face_guard.go) where to find it via
// PASSQUANTUM_FACE_GUARD_BUNDLE. No Python installation is required on the
// target machine — all dependencies are baked into the bundle.

import (
	_ "embed"
	"log"
	"os"
	"path/filepath"
)

//go:embed face_guard_bundle.exe
var faceGuardBundleData []byte

func init() {
	dir := filepath.Join(os.TempDir(), "passquantum-face-guard")
	if err := os.MkdirAll(dir, 0700); err != nil {
		log.Printf("[FaceGuard] WARNING: could not create bundle dir %s: %v", dir, err)
		return
	}

	// On Windows the extracted file must carry the .exe extension so the OS
	// can execute it directly without a shell wrapper.
	bundlePath := filepath.Join(dir, "face_guard.exe")

	// Always overwrite so the bundled version stays in sync with the binary.
	if err := os.WriteFile(bundlePath, faceGuardBundleData, 0700); err != nil {
		log.Printf("[FaceGuard] WARNING: could not extract face_guard bundle: %v", err)
		return
	}

	os.Setenv("PASSQUANTUM_FACE_GUARD_BUNDLE", bundlePath)

	log.Printf("[FaceGuard] Embedded Python bundle extracted to %s", bundlePath)
}
