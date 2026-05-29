package browser

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"
)

const (
	pairingTokenTTL  = 60 * time.Second
	pairingMaxRetries = 5
)

type PairingState struct {
	mu          sync.Mutex
	token       string
	expiresAt   time.Time
	attempts    int
	onShowToken func(token string)
}

func NewPairingState(onShowToken func(token string)) *PairingState {
	return &PairingState{onShowToken: onShowToken}
}

// StartPairing generates a 6-digit token and displays it via the callback.
// Returns the token for testing purposes.
func (ps *PairingState) StartPairing() string {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	token := generate6DigitToken()
	ps.token = token
	ps.expiresAt = time.Now().Add(pairingTokenTTL)
	ps.attempts = 0

	if ps.onShowToken != nil {
		ps.onShowToken(token)
	}

	return token
}

// ValidateToken checks the provided token. Returns true on match.
// Clears the token on success or after max attempts.
func (ps *PairingState) ValidateToken(token string) bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.token == "" {
		return false
	}

	if time.Now().After(ps.expiresAt) {
		ps.token = ""
		return false
	}

	ps.attempts++
	if ps.attempts > pairingMaxRetries {
		ps.token = ""
		return false
	}

	if token != ps.token {
		return false
	}

	ps.token = ""
	return true
}

func (ps *PairingState) IsActive() bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.token != "" && time.Now().Before(ps.expiresAt)
}

func generate6DigitToken() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "000000"
	}
	return fmt.Sprintf("%06d", n.Int64())
}
