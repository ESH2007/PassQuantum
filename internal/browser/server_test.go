package browser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// --- Mock VaultService ---

type mockVaultService struct {
	ready       bool
	appUnlocked bool
	vaultName   string
	credentials []CredentialSummary
	lastSaved   *SaveRequest
	lastUpdated *struct {
		id       uint64
		password string
	}
	saveIDCounter uint64
}

func (m *mockVaultService) IsReady() bool { return m.ready }

func (m *mockVaultService) Status() VaultStatus {
	return VaultStatus{AppUnlocked: m.appUnlocked, VaultName: m.vaultName}
}

func (m *mockVaultService) FindCredentials(domain string) ([]CredentialSummary, error) {
	return m.credentials, nil
}

func (m *mockVaultService) SaveCredential(domain, username, password string) (uint64, error) {
	m.lastSaved = &SaveRequest{Domain: domain, Username: username, Password: password}
	m.saveIDCounter++
	return m.saveIDCounter, nil
}

func (m *mockVaultService) UpdatePassword(entryID uint64, newPassword string) error {
	m.lastUpdated = &struct {
		id       uint64
		password string
	}{entryID, newPassword}
	return nil
}

// --- Helpers ---

func newTestServer(vault *mockVaultService) (*Server, *httptest.Server) {
	cfg := &Config{
		Secret:    "test-secret-hex",
		NeverSave: []string{},
	}
	s := NewServer(vault, cfg)
	s.pairing = NewPairingState(nil)

	mux := http.NewServeMux()
	mux.HandleFunc("/vault/pair", s.handlePair)
	mux.HandleFunc("/vault/status", s.handleStatus)
	mux.HandleFunc("/vault/exists", s.handleExists)
	mux.HandleFunc("/vault/save", s.handleSave)
	mux.HandleFunc("/vault/update/", s.handleUpdate)
	mux.HandleFunc("/vault/never-save", s.handleNeverSave)

	ts := httptest.NewServer(s.middleware(mux))
	return s, ts
}

func doRequest(ts *httptest.Server, method, path, secret string, body interface{}) *http.Response {
	var reqBody *bytes.Buffer
	if body != nil {
		data, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(data)
	} else {
		reqBody = &bytes.Buffer{}
	}

	req, _ := http.NewRequest(method, ts.URL+path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if secret != "" {
		req.Header.Set("X-Secret", secret)
	}
	resp, _ := http.DefaultClient.Do(req)
	return resp
}

// --- Tests ---

func TestStatusEndpoint(t *testing.T) {
	vault := &mockVaultService{ready: true}
	_, ts := newTestServer(vault)
	defer ts.Close()

	resp := doRequest(ts, "GET", "/vault/status", "", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var status StatusResponse
	json.NewDecoder(resp.Body).Decode(&status)
	if !status.Unlocked {
		t.Fatal("expected unlocked=true")
	}
}

func TestAuthRequired(t *testing.T) {
	vault := &mockVaultService{ready: true}
	_, ts := newTestServer(vault)
	defer ts.Close()

	resp := doRequest(ts, "GET", "/vault/exists?domain=github.com", "", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without secret, got %d", resp.StatusCode)
	}

	resp = doRequest(ts, "GET", "/vault/exists?domain=github.com", "wrong-secret", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 with wrong secret, got %d", resp.StatusCode)
	}
}

func TestLockedVault(t *testing.T) {
	vault := &mockVaultService{ready: false}
	_, ts := newTestServer(vault)
	defer ts.Close()

	resp := doRequest(ts, "GET", "/vault/exists?domain=github.com", "test-secret-hex", nil)
	if resp.StatusCode != http.StatusLocked {
		t.Fatalf("expected 423, got %d", resp.StatusCode)
	}
}

func TestExistsEndpoint(t *testing.T) {
	vault := &mockVaultService{
		ready: true,
		credentials: []CredentialSummary{
			{ID: 1, Service: "github.com", Username: "user@test.com"},
		},
	}
	_, ts := newTestServer(vault)
	defer ts.Close()

	resp := doRequest(ts, "GET", "/vault/exists?domain=github.com", "test-secret-hex", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body ExistsResponse
	json.NewDecoder(resp.Body).Decode(&body)
	if !body.Found {
		t.Fatal("expected found=true")
	}
	if len(body.Credentials) != 1 {
		t.Fatalf("expected 1 credential, got %d", len(body.Credentials))
	}
}

func TestSaveEndpoint(t *testing.T) {
	vault := &mockVaultService{ready: true}
	_, ts := newTestServer(vault)
	defer ts.Close()

	resp := doRequest(ts, "POST", "/vault/save", "test-secret-hex", SaveRequest{
		Domain:   "github.com",
		Username: "user@test.com",
		Password: "secret123!",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body SaveResponse
	json.NewDecoder(resp.Body).Decode(&body)
	if !body.Saved {
		t.Fatal("expected saved=true")
	}

	if vault.lastSaved == nil {
		t.Fatal("SaveCredential was not called")
	}
	if vault.lastSaved.Domain != "github.com" {
		t.Fatalf("expected domain github.com, got %s", vault.lastSaved.Domain)
	}
}

func TestUpdateEndpoint(t *testing.T) {
	vault := &mockVaultService{ready: true}
	_, ts := newTestServer(vault)
	defer ts.Close()

	resp := doRequest(ts, "PUT", "/vault/update/12345", "test-secret-hex", UpdateRequest{
		Password: "newpass456!",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	if vault.lastUpdated == nil {
		t.Fatal("UpdatePassword was not called")
	}
	if vault.lastUpdated.id != 12345 {
		t.Fatalf("expected entry ID 12345, got %d", vault.lastUpdated.id)
	}
}

func TestNeverSaveEndpoints(t *testing.T) {
	vault := &mockVaultService{ready: true}
	s, ts := newTestServer(vault)
	defer ts.Close()

	// POST — add domain
	resp := doRequest(ts, "POST", "/vault/never-save", "test-secret-hex", NeverSaveRequest{
		Domain: "example.com",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST expected 200, got %d", resp.StatusCode)
	}
	if !s.config.IsNeverSave("example.com") {
		t.Fatal("expected example.com in never-save list")
	}

	// GET — list domains
	resp = doRequest(ts, "GET", "/vault/never-save", "test-secret-hex", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET expected 200, got %d", resp.StatusCode)
	}
	var list NeverSaveListResponse
	json.NewDecoder(resp.Body).Decode(&list)
	if len(list.Domains) != 1 || list.Domains[0] != "example.com" {
		t.Fatalf("expected [example.com], got %v", list.Domains)
	}

	// DELETE — remove domain
	resp = doRequest(ts, "DELETE", "/vault/never-save?domain=example.com", "test-secret-hex", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("DELETE expected 200, got %d", resp.StatusCode)
	}
	if s.config.IsNeverSave("example.com") {
		t.Fatal("expected example.com removed from never-save list")
	}
}

func TestPairingEndpoint(t *testing.T) {
	vault := &mockVaultService{ready: false}
	s, ts := newTestServer(vault)
	defer ts.Close()

	// Step 1: initiate pairing (no token)
	resp := doRequest(ts, "POST", "/vault/pair", "", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var pr PairResponse
	json.NewDecoder(resp.Body).Decode(&pr)
	if pr.Status != "pending" {
		t.Fatalf("expected status pending, got %s", pr.Status)
	}

	// Get the token from pairing state
	s.pairing.mu.Lock()
	token := s.pairing.token
	s.pairing.mu.Unlock()

	// Step 2: wrong token
	resp = doRequest(ts, "POST", "/vault/pair", "", PairRequest{Token: "000000"})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for wrong token, got %d", resp.StatusCode)
	}

	// Step 3: correct token
	resp = doRequest(ts, "POST", "/vault/pair", "", PairRequest{Token: token})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	json.NewDecoder(resp.Body).Decode(&pr)
	if pr.Status != "paired" || pr.Secret == "" {
		t.Fatalf("expected paired with secret, got status=%s secret=%q", pr.Status, pr.Secret)
	}
}

func TestCORSPreflight(t *testing.T) {
	vault := &mockVaultService{ready: true}
	_, ts := newTestServer(vault)
	defer ts.Close()

	req, _ := http.NewRequest("OPTIONS", ts.URL+"/vault/status", nil)
	req.Header.Set("Origin", "chrome-extension://abcdef123456")
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "chrome-extension://abcdef123456" {
		t.Fatalf("expected extension origin in CORS, got %q", got)
	}
}

func TestRateLimit(t *testing.T) {
	vault := &mockVaultService{ready: true}
	_, ts := newTestServer(vault)
	defer ts.Close()

	var lastStatus int
	for i := 0; i < rateBurst+5; i++ {
		resp := doRequest(ts, "GET", "/vault/status", "", nil)
		lastStatus = resp.StatusCode
		resp.Body.Close()
	}

	if lastStatus != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after exceeding rate limit, got %d", lastStatus)
	}
}

func TestSaveMissingFields(t *testing.T) {
	vault := &mockVaultService{ready: true}
	_, ts := newTestServer(vault)
	defer ts.Close()

	resp := doRequest(ts, "POST", "/vault/save", "test-secret-hex", SaveRequest{
		Domain: "github.com",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing password, got %d", resp.StatusCode)
	}
}

func TestExistsMissingDomain(t *testing.T) {
	vault := &mockVaultService{ready: true}
	_, ts := newTestServer(vault)
	defer ts.Close()

	resp := doRequest(ts, "GET", "/vault/exists", "test-secret-hex", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing domain, got %d", resp.StatusCode)
	}
}

func TestStatusEndpointNoAuth(t *testing.T) {
	vault := &mockVaultService{ready: false}
	_, ts := newTestServer(vault)
	defer ts.Close()

	// /vault/status should work without auth
	resp := doRequest(ts, "GET", "/vault/status", "", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var status StatusResponse
	json.NewDecoder(resp.Body).Decode(&status)
	if status.Unlocked {
		t.Fatal("expected unlocked=false")
	}
}

func TestFirefoxCORSOrigin(t *testing.T) {
	vault := &mockVaultService{ready: true}
	_, ts := newTestServer(vault)
	defer ts.Close()

	req, _ := http.NewRequest("OPTIONS", ts.URL+"/vault/status", nil)
	req.Header.Set("Origin", "moz-extension://abcdef-1234-5678")
	resp, _ := http.DefaultClient.Do(req)

	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "moz-extension://abcdef-1234-5678" {
		t.Fatalf("expected Firefox extension origin in CORS, got %q", got)
	}
}

func TestUpdateInvalidID(t *testing.T) {
	vault := &mockVaultService{ready: true}
	_, ts := newTestServer(vault)
	defer ts.Close()

	resp := doRequest(ts, "PUT", "/vault/update/not-a-number", "test-secret-hex", UpdateRequest{
		Password: "new!",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid ID, got %d", resp.StatusCode)
	}

	_ = fmt.Sprint("silence")
}
