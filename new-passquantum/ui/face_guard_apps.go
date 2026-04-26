package main

// ==============================
// face_guard_apps.go — Kill-list helpers for the FaceGuard feature
// ==============================
// When the user's face is not detected for 5 seconds the app locks itself
// AND force-kills any companion processes the user has opted into via
// Settings > Security > Monitored Apps.
//
// Persistence: the list is stored as a JSON array of process-name strings in
// Fyne's built-in key-value preferences under the key "face_guard_kill_apps".

import (
	"encoding/json"
	"log"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
)

const killAppsPrefsKey = "face_guard_kill_apps"

// ─────────────────────────────────────────────────────────────────────────────
// Process listing
// ─────────────────────────────────────────────────────────────────────────────

// listRunningProcesses returns the names of currently running processes,
// deduplicated and sorted alphabetically.  PassQuantum itself is excluded.
func listRunningProcesses() []string {
	var names []string

	switch runtime.GOOS {
	case "windows":
		// tasklist /fo csv /nh  → "process.exe","PID","Session","#","Mem"
		out, err := exec.Command("tasklist", "/fo", "csv", "/nh").Output()
		if err != nil {
			log.Printf("[FaceGuardApps] tasklist error: %v", err)
			return nil
		}
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			// First CSV field is the exe name, quoted.
			fields := strings.SplitN(line, ",", 2)
			name := strings.Trim(fields[0], `"`)
			// Strip ".exe" suffix for display.
			name = strings.TrimSuffix(name, ".exe")
			names = append(names, name)
		}

	default: // Linux / macOS
		// ps -eo comm outputs just the command name, one per line.
		out, err := exec.Command("ps", "-eo", "comm").Output()
		if err != nil {
			log.Printf("[FaceGuardApps] ps error: %v", err)
			return nil
		}
		for _, line := range strings.Split(string(out), "\n") {
			name := strings.TrimSpace(line)
			if name == "" || name == "COMMAND" {
				continue
			}
			names = append(names, name)
		}
	}

	// Deduplicate and exclude the password manager itself.
	seen := make(map[string]struct{}, len(names))
	result := names[:0]
	for _, n := range names {
		lower := strings.ToLower(n)
		if _, dup := seen[lower]; dup {
			continue
		}
		if lower == "passquantum" || lower == "passquantum.exe" {
			continue
		}
		seen[lower] = struct{}{}
		result = append(result, n)
	}

	sort.Strings(result)
	return result
}

// ─────────────────────────────────────────────────────────────────────────────
// Process killing
// ─────────────────────────────────────────────────────────────────────────────

// killProcessesByName force-kills each process whose name is in the list.
// It is intentionally fire-and-forget: errors are logged but not returned.
func killProcessesByName(names []string) {
	if len(names) == 0 {
		return
	}

	for _, name := range names {
		var cmd *exec.Cmd

		switch runtime.GOOS {
		case "windows":
			// taskkill /F /IM <name>.exe
			exeName := name
			if !strings.HasSuffix(strings.ToLower(name), ".exe") {
				exeName = name + ".exe"
			}
			cmd = exec.Command("taskkill", "/F", "/IM", exeName)

		default: // Linux / macOS
			// pkill -9 -x <name>  — exact-match on process name
			cmd = exec.Command("pkill", "-9", "-x", name)
		}

		if err := cmd.Run(); err != nil {
			// pkill/taskkill return non-zero when no matching process is found;
			// that is not an error we need to surface.
			log.Printf("[FaceGuardApps] kill %q: %v", name, err)
		} else {
			log.Printf("[FaceGuardApps] killed %q", name)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Preference helpers
// ─────────────────────────────────────────────────────────────────────────────

// loadKillApps reads the persisted kill-list from Fyne preferences.
// Returns an empty slice when no list has been saved yet.
func loadKillApps(prefs fyne.Preferences) []string {
	raw := prefs.String(killAppsPrefsKey)
	if raw == "" {
		return nil
	}
	var list []string
	if err := json.Unmarshal([]byte(raw), &list); err != nil {
		log.Printf("[FaceGuardApps] loadKillApps: %v", err)
		return nil
	}
	return list
}

// saveKillApps persists the kill-list to Fyne preferences.
func saveKillApps(prefs fyne.Preferences, apps []string) {
	data, err := json.Marshal(apps)
	if err != nil {
		log.Printf("[FaceGuardApps] saveKillApps: %v", err)
		return
	}
	prefs.SetString(killAppsPrefsKey, string(data))
}
