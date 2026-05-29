package browser

import (
	"testing"
	"time"
)

func TestPairingFlow(t *testing.T) {
	var shownToken string
	ps := NewPairingState(func(token string) {
		shownToken = token
	})

	token := ps.StartPairing()
	if token == "" || len(token) != 6 {
		t.Fatalf("expected 6-digit token, got %q", token)
	}
	if shownToken != token {
		t.Fatalf("callback received %q, expected %q", shownToken, token)
	}

	if !ps.IsActive() {
		t.Fatal("pairing should be active after StartPairing")
	}

	if ps.ValidateToken("000000") {
		t.Fatal("wrong token should not validate")
	}

	if !ps.ValidateToken(token) {
		t.Fatal("correct token should validate")
	}

	if ps.IsActive() {
		t.Fatal("pairing should be inactive after successful validation")
	}
}

func TestPairingExpiry(t *testing.T) {
	ps := NewPairingState(nil)
	token := ps.StartPairing()

	ps.mu.Lock()
	ps.expiresAt = time.Now().Add(-1 * time.Second)
	ps.mu.Unlock()

	if ps.ValidateToken(token) {
		t.Fatal("expired token should not validate")
	}
}

func TestPairingMaxAttempts(t *testing.T) {
	ps := NewPairingState(nil)
	token := ps.StartPairing()

	for i := 0; i < pairingMaxRetries; i++ {
		ps.ValidateToken("wrong!")
	}

	if ps.ValidateToken(token) {
		t.Fatal("token should be invalidated after max attempts")
	}
}
