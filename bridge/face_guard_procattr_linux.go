//go:build linux

package bridge

import (
	"os/exec"
	"syscall"
)

// setParentDeathSignal asks the kernel to deliver SIGKILL to the child process
// if the parent (this Go process) dies for any reason — including abrupt
// termination (SIGKILL, crash) where neither w.SetOnClosed nor the signal
// handler can run. Without this, face_guard.py would be orphaned and keep the
// webcam held open.
func setParentDeathSignal(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Pdeathsig = syscall.SIGKILL
}
