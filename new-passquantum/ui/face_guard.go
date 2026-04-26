package main

// ==============================
// face_guard.go — PassQuantum Face Guard
// ==============================
// Manages the face_guard.py child process and the TCP server that it connects to.
// All camera capture and ML inference runs in Python; this file handles the
// Go side: process lifecycle, TCP message parsing, and callback dispatch.
//
// Protocol messages received from Python:
//   FRAME:<base64 JPEG>      — live camera frame (training only)
//   PROGRESS:<n>/<total>     — training progress
//   TRAINING_DONE            — all face samples saved
//   FACE_OK                  — recognised face reappeared after FACE_LOST
//   FACE_LOST                — recognised face absent for grace period
//
// Commands sent to Python:
//   START_TRAINING
//   START_MONITOR

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg" // register JPEG decoder
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"
)

// ==============================
// Constants
// ==============================

const (
	faceGuardAddr   = "127.0.0.1:9876"
	faceGuardScript = "face_guard.py"
	scannerBufSize  = 1 << 20 // 1 MB — large enough for base64-encoded JPEG frames
)

// ==============================
// FaceGuard Struct
// ==============================

// FaceGuard manages the face recognition subprocess and the TCP connection it uses.
type FaceGuard struct {
	listener net.Listener
	conn     net.Conn

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
	return &FaceGuard{listener: ln}, nil
}

// ==============================
// Process Launch
// ==============================

// Launch starts the face_guard.py child process.
// It tries "python3" first; if that fails it tries "python".
// The process is started detached (Stdout/Stderr are inherited so logs appear
// in the terminal where the app is run from).
// Launch is non-blocking — it does not wait for the process to finish.
func (g *FaceGuard) Launch() error {
	cmd, err := buildPythonCommand(faceGuardScript)
	if err != nil {
		return fmt.Errorf("face guard: no Python interpreter found: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("face guard: failed to start %s: %w", faceGuardScript, err)
	}
	log.Printf("[FaceGuard] Launched %s (pid %d)", faceGuardScript, cmd.Process.Pid)
	return nil
}

// buildPythonCommand constructs an *exec.Cmd for face_guard.py, preferring python3.
func buildPythonCommand(script string) (*exec.Cmd, error) {
	for _, interp := range []string{"python3", "python"} {
		if path, err := exec.LookPath(interp); err == nil {
			cmd := exec.Command(path, script)
			return cmd, nil
		}
	}
	return nil, fmt.Errorf("neither python3 nor python found in PATH")
}

// ==============================
// Command Sending
// ==============================

// SendCommand writes a newline-terminated command to the connected Python process.
// It is safe to call before Listen() has accepted a connection only if conn is non-nil.
func (g *FaceGuard) SendCommand(cmd string) {
	if g.conn == nil {
		log.Printf("[FaceGuard] SendCommand(%q): no connection yet", cmd)
		return
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
		return
	}
	g.conn = conn
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
