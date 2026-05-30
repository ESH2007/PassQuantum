//go:build windows

package main

import (
	"log"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const windowTitle = "PassQuantum - Post-Quantum Safe Password Manager"

var (
	user32          = windows.NewLazySystemDLL("user32.dll")
	procFindWindowW = user32.NewProc("FindWindowW")
	procShowWindow  = user32.NewProc("ShowWindow")
)

// maximizeWindow asks Windows to maximize the PassQuantum window with
// decorations preserved (standard SW_MAXIMIZE behavior: title bar and
// taskbar remain visible). Fyne v2 exposes no native maximize API, so we
// call user32.dll directly through golang.org/x/sys/windows.
func maximizeWindow() {
	time.Sleep(300 * time.Millisecond)

	title, err := windows.UTF16PtrFromString(windowTitle)
	if err != nil {
		log.Printf("[maximize] could not encode window title: %v", err)
		return
	}

	var hwnd uintptr
	for i := 0; i < 5; i++ {
		h, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(title)))
		if h != 0 {
			hwnd = h
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if hwnd == 0 {
		log.Printf("[maximize] could not locate window %q after retries", windowTitle)
		return
	}

	procShowWindow.Call(hwnd, uintptr(windows.SW_MAXIMIZE))
}
