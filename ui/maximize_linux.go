//go:build linux

package main

import (
	"log"
	"os/exec"
	"strings"
	"time"
)

// maximizeWindow asks the window manager to maximize the PassQuantum window.
// Fyne v2 has no native maximize API, so we use wmctrl (X11/XWayland) with
// xdotool as a fallback. A short sleep lets the window become visible first.
func maximizeWindow() {
	time.Sleep(300 * time.Millisecond)

	// wmctrl: available on most Linux desktops, works on X11 and XWayland
	if err := exec.Command("wmctrl", "-r", "PassQuantum", "-b", "add,maximized_vert,maximized_horz").Run(); err == nil {
		return
	}

	// xdotool fallback
	out, err := exec.Command("xdotool", "search", "--name", "PassQuantum").Output()
	if err != nil {
		log.Printf("[maximize] no wmctrl or xdotool found; install one to auto-maximize: %v", err)
		return
	}
	for _, wid := range strings.Fields(string(out)) {
		exec.Command("xdotool", "windowmaximize", wid).Run()
	}
}
