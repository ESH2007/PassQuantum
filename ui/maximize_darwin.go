//go:build darwin

package main

import "log"

// maximizeWindow is a no-op on macOS. macOS has no Win32-style maximize
// primitive; the green-button "zoom" is per-app and would require AppleScript
// or a CGO Cocoa call. The window opens at Fyne's default size for now.
func maximizeWindow() {
	log.Printf("[maximize] not implemented on darwin; window opens at Fyne default size")
}
