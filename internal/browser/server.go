package browser

import (
	"context"
	"crypto/subtle"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	listenAddr     = "127.0.0.1:8765"
	shutdownTimeout = 5 * time.Second
	rateBurst      = 10
	rateWindow     = time.Second
)

type Server struct {
	httpServer *http.Server
	vault      VaultService
	config     *Config
	pairing    *PairingState
	limiter    *rateLimiter
	mu         sync.Mutex
	running    bool
}

func NewServer(vault VaultService, config *Config) *Server {
	s := &Server{
		vault:   vault,
		config:  config,
		pairing: NewPairingState(nil),
		limiter: newRateLimiter(rateBurst, rateWindow),
	}
	return s
}

func (s *Server) SetPairingCallback(fn func(token string)) {
	s.pairing.onShowToken = fn
}

func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/vault/pair", s.handlePair)
	mux.HandleFunc("/vault/status", s.handleStatus)
	mux.HandleFunc("/vault/exists", s.handleExists)
	mux.HandleFunc("/vault/save", s.handleSave)
	mux.HandleFunc("/vault/update/", s.handleUpdate)
	mux.HandleFunc("/vault/never-save", s.handleNeverSave)

	s.httpServer = &http.Server{
		Addr:         listenAddr,
		Handler:      s.middleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}

	s.running = true
	log.Printf("[Browser] API server listening on %s", listenAddr)

	go func() {
		if err := s.httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("[Browser] server error: %v", err)
		}
	}()

	return nil
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running || s.httpServer == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	s.running = false
	log.Println("[Browser] API server shutting down")
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Localhost check
		remoteIP := extractIP(r.RemoteAddr)
		if remoteIP != "127.0.0.1" && remoteIP != "::1" {
			writeError(w, http.StatusForbidden, "localhost only")
			return
		}

		// 2. CORS
		origin := r.Header.Get("Origin")
		if origin != "" {
			if strings.HasPrefix(origin, "chrome-extension://") ||
				strings.HasPrefix(origin, "moz-extension://") {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "X-Secret, Content-Type")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// 3. Rate limit
		if !s.limiter.allow() {
			writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		// 4. Auth check (exempt: /vault/pair, /vault/status)
		if r.URL.Path != "/vault/pair" && r.URL.Path != "/vault/status" {
			secret := r.Header.Get("X-Secret")
			configSecret := s.config.Secret
			if configSecret == "" || subtle.ConstantTimeCompare([]byte(secret), []byte(configSecret)) != 1 {
				writeError(w, http.StatusUnauthorized, "invalid or missing X-Secret")
				return
			}
		}

		// 5. Unlock check (exempt: /vault/pair, /vault/status)
		if r.URL.Path != "/vault/pair" && r.URL.Path != "/vault/status" {
			if !s.vault.IsReady() {
				writeError(w, http.StatusLocked, "PassQuantum is locked")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func extractIP(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}

// Simple token bucket rate limiter — no external dependency.
type rateLimiter struct {
	mu       sync.Mutex
	tokens   int
	max      int
	window   time.Duration
	lastFill time.Time
}

func newRateLimiter(max int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		tokens:   max,
		max:      max,
		window:   window,
		lastFill: time.Now(),
	}
}

func (rl *rateLimiter) allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastFill)
	if elapsed >= rl.window {
		rl.tokens = rl.max
		rl.lastFill = now
	}

	if rl.tokens <= 0 {
		return false
	}
	rl.tokens--
	return true
}
