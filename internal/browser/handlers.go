package browser

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// --- Request/Response types ---

type PairRequest struct {
	Token string `json:"token"`
}

type PairResponse struct {
	Secret  string `json:"secret,omitempty"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type StatusResponse struct {
	Unlocked    bool   `json:"unlocked"`
	AppUnlocked bool   `json:"app_unlocked"`
	Vault       string `json:"vault,omitempty"`
	Version     string `json:"version"`
}

type ExistsResponse struct {
	Found       bool                `json:"found"`
	Credentials []CredentialSummary `json:"credentials"`
}

type SaveRequest struct {
	Domain   string `json:"domain"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type SaveResponse struct {
	ID    uint64 `json:"id"`
	Saved bool   `json:"saved"`
}

type UpdateRequest struct {
	Password string `json:"password"`
	Username string `json:"username,omitempty"`
}

type UpdateResponse struct {
	Updated bool `json:"updated"`
}

type NeverSaveRequest struct {
	Domain string `json:"domain"`
}

type NeverSaveListResponse struct {
	Domains []string `json:"domains"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// --- Handlers ---

func (s *Server) handlePair(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req PairRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
	}

	if req.Token == "" {
		s.pairing.StartPairing()
		writeJSON(w, http.StatusOK, PairResponse{
			Status:  "pending",
			Message: "Enter the 6-digit code shown in PassQuantum",
		})
		return
	}

	if !s.pairing.ValidateToken(req.Token) {
		writeError(w, http.StatusUnauthorized, "invalid or expired token")
		return
	}

	secret, err := GenerateSecret()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate secret")
		return
	}

	s.config.SetPaired(secret)
	if err := s.config.Save(); err != nil {
		log.Printf("[Browser] WARNING: failed to save config after pairing: %v", err)
	}

	writeJSON(w, http.StatusOK, PairResponse{
		Status: "paired",
		Secret: secret,
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	vs := s.vault.Status()
	writeJSON(w, http.StatusOK, StatusResponse{
		Unlocked:    s.vault.IsReady(),
		AppUnlocked: vs.AppUnlocked,
		Vault:       vs.VaultName,
		Version:     "1.0.0",
	})
}

func (s *Server) handleExists(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	domain := r.URL.Query().Get("domain")
	if domain == "" {
		writeError(w, http.StatusBadRequest, "domain parameter required")
		return
	}

	creds, err := s.vault.FindCredentials(NormalizeDomain(domain))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "vault error")
		log.Printf("[Browser] FindCredentials error: %v", err)
		return
	}

	if creds == nil {
		creds = []CredentialSummary{}
	}

	writeJSON(w, http.StatusOK, ExistsResponse{
		Found:       len(creds) > 0,
		Credentials: creds,
	})
}

func (s *Server) handleSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req SaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Domain == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "domain and password are required")
		return
	}

	req.Domain = NormalizeDomain(req.Domain)

	id, err := s.vault.SaveCredential(req.Domain, req.Username, req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save credential")
		log.Printf("[Browser] SaveCredential error: %v", err)
		return
	}

	log.Printf("[Browser] Saved credential for %s (user: %s)", req.Domain, req.Username)

	writeJSON(w, http.StatusOK, SaveResponse{ID: id, Saved: true})
}

func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/vault/update/")
	if idStr == "" || idStr == r.URL.Path {
		writeError(w, http.StatusBadRequest, "entry ID required in path")
		return
	}

	entryID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid entry ID: %s", idStr))
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Password == "" {
		writeError(w, http.StatusBadRequest, "password is required")
		return
	}

	if err := s.vault.UpdatePassword(entryID, req.Password); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update credential")
		log.Printf("[Browser] UpdatePassword error: %v", err)
		return
	}

	log.Printf("[Browser] Updated credential ID %d", entryID)

	writeJSON(w, http.StatusOK, UpdateResponse{Updated: true})
}

func (s *Server) handleNeverSave(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, NeverSaveListResponse{
			Domains: s.config.GetNeverSaveList(),
		})

	case http.MethodPost:
		var req NeverSaveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.Domain == "" {
			writeError(w, http.StatusBadRequest, "domain is required")
			return
		}
		req.Domain = NormalizeDomain(req.Domain)
		s.config.AddNeverSave(req.Domain)
		if err := s.config.Save(); err != nil {
			log.Printf("[Browser] WARNING: failed to persist never-save list: %v", err)
		}
		writeJSON(w, http.StatusOK, map[string]bool{"success": true})

	case http.MethodDelete:
		domain := r.URL.Query().Get("domain")
		if domain == "" {
			writeError(w, http.StatusBadRequest, "domain parameter required")
			return
		}
		domain = NormalizeDomain(domain)
		s.config.RemoveNeverSave(domain)
		if err := s.config.Save(); err != nil {
			log.Printf("[Browser] WARNING: failed to persist never-save list: %v", err)
		}
		writeJSON(w, http.StatusOK, map[string]bool{"success": true})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}
