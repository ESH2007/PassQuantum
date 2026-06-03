package bridge

// ==============================
// face_guard.go — PassQuantum Face Guard
// ==============================
// Manages the face_guard.py child process and the TCP server that it connects to.
// All camera capture and ML inference runs in Python; this file handles the
// Go side: process lifecycle, TCP message parsing, and callback dispatch.
//
// Protocol messages received from Python:
//   FRAME:<base64 JPEG>      — live camera frame (training and demo modes)
//   PROGRESS:<n>/<total>     — training progress
//   TRAINING_DONE            — all face samples saved
//   FACE_OK                  — recognised face reappeared after FACE_LOST
//   FACE_LOST                — recognised face absent for grace period
//
// Commands sent to Python:
//   START_TRAINING
//   START_MONITOR
//   START_DEMO               — pause monitoring; stream annotated landmark/blink
//                             frames for the Security-settings visualizer
//   STOP_DEMO                — leave demo mode and resume monitoring

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg" // register JPEG decoder
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ==============================
// Constants
// ==============================

const (
	faceGuardAddr   = "127.0.0.1:9876"
	faceGuardScript = "python/face_guard.py"
	scannerBufSize  = 1 << 20 // 1 MB — large enough for base64-encoded JPEG frames
)

// ==============================
// FaceGuard Struct
// ==============================

// FaceGuard manages the face recognition subprocess and the TCP connection it uses.
type FaceGuard struct {
	listener  net.Listener
	conn      net.Conn
	connReady chan struct{} // closed by Listen() once Python has connected
	cmd       *exec.Cmd     // the running Python process; nil until Launch() succeeds

	// OnFrame is called on the Go main goroutine with each decoded JPEG frame
	// received during training.  May be nil.
	OnFrame func(img image.Image)

	// OnProgress is called with (current, total) during training.  May be nil.
	OnProgress func(current, total int)

	// OnDone is called when Python sends TRAINING_DONE.  May be nil.
	OnDone func()

	// OnLost is called when Python sends FACE_LOST.  May be nil.
	OnLost func()

	// OnOK is called when Python sends FACE_OK.  May be nil.
	OnOK func()
}

// ==============================
// Constructor
// ==============================

// NewFaceGuard creates a new FaceGuard and starts a TCP listener on faceGuardAddr.
// The caller must call Launch() and then go guard.Listen() to start the subprocess.
func NewFaceGuard() (*FaceGuard, error) {
	ln, err := net.Listen("tcp", faceGuardAddr)
	if err != nil {
		return nil, fmt.Errorf("face guard: failed to listen on %s: %w", faceGuardAddr, err)
	}
	return &FaceGuard{listener: ln, connReady: make(chan struct{})}, nil
}

// ==============================
// Process Launch
// ==============================

// Launch starts the face_guard.py child process.
// It tries "python3" first; if that fails it tries "python".
//
// Python's stderr is piped to Go's logger so import errors, webcam failures,
// and any other crash output are visible in the app log rather than silently
// discarded.  A background goroutine also calls cmd.Wait() and logs a clear
// message if the process exits unexpectedly.
//
// Launch is non-blocking — it does not wait for the process to finish.
func (g *FaceGuard) Launch() error {
	cmd, err := buildPythonCommand(faceGuardScript)
	if err != nil {
		return fmt.Errorf("face guard: no Python interpreter found: %w", err)
	}

	// Pipe Python's stderr into Go's logger so errors (ImportError, webcam
	// failures, tracebacks) are always visible.
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("face guard: could not create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("face guard: failed to start %s: %w", faceGuardScript, err)
	}

	g.cmd = cmd
	log.Printf("[FaceGuard] Launched %s (pid %d)", faceGuardScript, cmd.Process.Pid)

	// Forward every line of Python stderr to Go's logger.
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			log.Printf("[face_guard.py] %s", scanner.Text())
		}
	}()

	// Watch for unexpected process exit and emit a clear error log.
	go func() {
		err := cmd.Wait()
		if err != nil {
			log.Printf("[FaceGuard] ERROR: face_guard.py exited unexpectedly: %v", err)
			log.Printf("[FaceGuard] Hint: check that python3 is installed with cv2, face_recognition, and numpy.")
		} else {
			log.Printf("[FaceGuard] face_guard.py exited cleanly (process finished).")
		}
		// Drain the stderrPipe scanner above will reach EOF on its own.
		_ = io.Discard // reference io so the import is used
	}()

	return nil
}

// buildPythonCommand constructs an *exec.Cmd for face_guard.py, preferring python3.
// The script is resolved relative to the executable first (production), then cwd (dev).
//
// When a PyInstaller bundle was embedded at build time, python_bundle.go's init()
// sets PASSQUANTUM_FACE_GUARD_BUNDLE to the extracted executable path.  In that case
// the bundle is run directly — no Python interpreter is required on the target machine.
func buildPythonCommand(script string) (*exec.Cmd, error) {
	// ── Embedded PyInstaller bundle (set by python_bundle.go init) ──────────
	if bundlePath := os.Getenv("PASSQUANTUM_FACE_GUARD_BUNDLE"); bundlePath != "" {
		if _, err := os.Stat(bundlePath); err == nil {
			cmd := exec.Command(bundlePath)
			// Store face_data.npy next to the PassQuantum executable, not in /tmp.
			if workDir := os.Getenv("PASSQUANTUM_WORK_DIR"); workDir != "" {
				cmd.Dir = workDir
			} else {
				cmd.Dir = filepath.Dir(bundlePath)
			}
			setParentDeathSignal(cmd)
			return cmd, nil
		}
	}

	// ── Fallback: call Python interpreter with face_guard.py ─────────────────
	scriptPath := resolveScript(script)

	for _, interp := range []string{"python3", "python"} {
		if interpPath, err := exec.LookPath(interp); err == nil {
			cmd := exec.Command(interpPath, scriptPath)
			// Store face_data.npy in the vault directory when available,
			// otherwise fall back to the script's own directory.
			if workDir := os.Getenv("PASSQUANTUM_WORK_DIR"); workDir != "" {
				cmd.Dir = workDir
			} else {
				cmd.Dir = filepath.Dir(scriptPath)
			}
			setParentDeathSignal(cmd)
			return cmd, nil
		}
	}
	return nil, fmt.Errorf("neither python3 nor python found in PATH")
}

// resolveScript finds the absolute path to script by probing:
//  1. Same directory as the running executable (production installs, CI builds).
//  2. Current working directory (development / go run).
func resolveScript(script string) string {
	// 1. Next to the executable.
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), script)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	// 2. Relative to cwd.
	if cwd, err := os.Getwd(); err == nil {
		candidate := filepath.Join(cwd, script)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return script // last resort — let the OS try
}

// ==============================
// Shutdown
// ==============================

// Shutdown terminates the face_guard.py child process and closes all network
// connections, releasing the camera immediately.  Safe to call multiple times.
func (g *FaceGuard) Shutdown() {
	if g.conn != nil {
		_ = g.conn.Close()
		g.conn = nil
	}
	if g.listener != nil {
		_ = g.listener.Close()
		g.listener = nil
	}
	if g.cmd != nil && g.cmd.Process != nil {
		if err := g.cmd.Process.Kill(); err != nil {
			log.Printf("[FaceGuard] Shutdown: kill error: %v", err)
		} else {
			log.Printf("[FaceGuard] Shutdown: face_guard.py (pid %d) killed", g.cmd.Process.Pid)
		}
		g.cmd = nil
	}
}

// ==============================
// Command Sending
// ==============================

// SendCommand writes a newline-terminated command to the connected Python process.
// If Python has not yet connected, it blocks until the connection is ready (up to
// 30 seconds) so that callers don't need to poll.  Always invoke from a goroutine
// when calling from a Fyne tap handler to avoid blocking the UI thread.
func (g *FaceGuard) SendCommand(cmd string) {
	if g.conn == nil {
		select {
		case <-g.connReady:
			// Python connected — fall through to send
		case <-time.After(30 * time.Second):
			log.Printf("[FaceGuard] SendCommand(%q): timed out waiting for Python connection", cmd)
			return
		}
	}
	if _, err := fmt.Fprintf(g.conn, "%s\n", cmd); err != nil {
		log.Printf("[FaceGuard] SendCommand(%q) error: %v", cmd, err)
	}
}

// ==============================
// TCP Accept + Message Loop
// ==============================

// Listen blocks until the Python process connects, then continuously reads
// newline-delimited messages and dispatches them to the registered callbacks.
//
// The bufio.Scanner buffer is set to scannerBufSize (1 MB) because FRAME messages
// contain base64-encoded JPEGs that exceed the default 64 KB scanner limit.
//
// This method is intended to be called in its own goroutine:
//
//	go guard.Listen()
func (g *FaceGuard) Listen() {
	log.Printf("[FaceGuard] Waiting for Python to connect on %s ...", faceGuardAddr)

	conn, err := g.listener.Accept()
	if err != nil {
		log.Printf("[FaceGuard] Accept error: %v", err)
		close(g.connReady) // unblock any SendCommand waiting on connection
		return
	}
	g.conn = conn
	close(g.connReady) // signal: Python is connected, SendCommand may proceed
	log.Printf("[FaceGuard] Python connected from %s", conn.RemoteAddr())

	scanner := bufio.NewScanner(conn)
	buf := make([]byte, scannerBufSize)
	scanner.Buffer(buf, scannerBufSize)

	for scanner.Scan() {
		line := scanner.Text()
		g.dispatch(line)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[FaceGuard] Scanner error: %v", err)
	}
	log.Printf("[FaceGuard] Connection from Python closed.")
}

// ==============================
// Message Dispatch
// ==============================

// dispatch parses one message line and invokes the corresponding callback.
func (g *FaceGuard) dispatch(line string) {
	switch {
	case strings.HasPrefix(line, "FRAME:"):
		g.handleFrame(strings.TrimPrefix(line, "FRAME:"))

	case strings.HasPrefix(line, "PROGRESS:"):
		g.handleProgress(strings.TrimPrefix(line, "PROGRESS:"))

	case line == "TRAINING_DONE":
		if g.OnDone != nil {
			g.OnDone()
		}

	case line == "FACE_LOST":
		if g.OnLost != nil {
			g.OnLost()
		}

	case line == "FACE_OK":
		if g.OnOK != nil {
			g.OnOK()
		}

	default:
		log.Printf("[FaceGuard] Unknown message: %q", line)
	}
}

// handleFrame decodes a base64 JPEG and calls OnFrame if set.
func (g *FaceGuard) handleFrame(b64data string) {
	if g.OnFrame == nil {
		return
	}
	raw, err := base64.StdEncoding.DecodeString(b64data)
	if err != nil {
		log.Printf("[FaceGuard] FRAME: base64 decode error: %v", err)
		return
	}
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		log.Printf("[FaceGuard] FRAME: image decode error: %v", err)
		return
	}
	g.OnFrame(img)
}

// handleProgress parses "current/total" and calls OnProgress if set.
func (g *FaceGuard) handleProgress(payload string) {
	if g.OnProgress == nil {
		return
	}
	parts := strings.SplitN(payload, "/", 2)
	if len(parts) != 2 {
		log.Printf("[FaceGuard] PROGRESS: unexpected format %q", payload)
		return
	}
	cur, err1 := strconv.Atoi(parts[0])
	total, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		log.Printf("[FaceGuard] PROGRESS: parse error in %q", payload)
		return
	}
	g.OnProgress(cur, total)
}
