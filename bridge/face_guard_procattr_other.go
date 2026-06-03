//go:build !linux

package bridge

import "os/exec"

// setParentDeathSignal is a no-op on platforms without a parent-death signal.
// There, releasing the webcam on exit relies on FaceGuard.Shutdown() being
// called from the window-close handler or the SIGINT/SIGTERM handler.
func setParentDeathSignal(cmd *exec.Cmd) {}
