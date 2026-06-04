//go:build darwin

package main

// python_bundle_darwin.go — locates the face_guard helper that ships INSIDE the
// macOS .app bundle.
//
// macOS is different from Linux/Windows on purpose. Linux/Windows embed the
// PyInstaller bundle via go:embed and extract it to a temp directory at startup
// (see python_bundle.go / python_bundle_windows.go). That does NOT work on
// macOS: the camera (AVFoundation / TCC privacy) system grants access based on
// the app's code signature + NSCameraUsageDescription, and it only does so for
// a helper that lives INSIDE the signed .app bundle. A helper copied to /tmp is
// unsigned and outside the bundle, so macOS denies the camera (and the
// permission prompt shows up inconsistently).
//
// build-mac-native.sh therefore builds the helper with PyInstaller --onedir and
// copies it into the bundle, then codesign --deep signs it together with the app:
//
//   PassQuantum.app/Contents/Resources/faceguard/face_guard_bundle
//   PassQuantum.app/Contents/Resources/faceguard/_internal/...
//
// At startup init() resolves that path relative to the running executable and,
// when present, exports PASSQUANTUM_FACE_GUARD_BUNDLE so buildPythonCommand
// (bridge/face_guard.go) runs it directly — in place, no extraction. In a plain
// `go build` / `go run` dev build the helper is absent, so the bridge falls back
// to running `python3 face_guard.py`.
//
// This file is compiled for every darwin build (no build tag), so the macOS Go
// binary is built WITHOUT -tags with_face_bundle.

import (
	"log"
	"os"
	"path/filepath"
)

func init() {
	exe, err := os.Executable()
	if err != nil {
		log.Printf("[FaceGuard] WARNING: could not resolve executable path: %v", err)
		return
	}

	// exe = <App>.app/Contents/MacOS/<bin>; the helper lives under ../Resources.
	helper := filepath.Join(filepath.Dir(exe), "..", "Resources", "faceguard", "face_guard_bundle")
	if abs, err := filepath.Abs(helper); err == nil {
		helper = abs
	}

	if _, err := os.Stat(helper); err != nil {
		// Not packaged (dev build) — buildPythonCommand falls back to python3.
		return
	}

	os.Setenv("PASSQUANTUM_FACE_GUARD_BUNDLE", helper)
	log.Printf("[FaceGuard] Using in-bundle face guard helper at %s", helper)
}
