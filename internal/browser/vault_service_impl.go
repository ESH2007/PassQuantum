package browser

import (
	"fmt"
	"strings"

	pqapp "passquantum/app"
	"passquantum/core/model"
)

type appVaultService struct {
	state     *pqapp.AppState
	domainMap *DomainMap
}

func NewAppVaultService(state *pqapp.AppState, domainMap *DomainMap) VaultService {
	return &appVaultService{state: state, domainMap: domainMap}
}

func (s *appVaultService) IsReady() bool {
	s.state.Mu.Lock()
	defer s.state.Mu.Unlock()
	return s.state.IsUnlocked && s.state.CurrentVault != ""
}

func (s *appVaultService) Status() VaultStatus {
	s.state.Mu.Lock()
	defer s.state.Mu.Unlock()
	return VaultStatus{
		AppUnlocked: s.state.IsUnlocked,
		VaultName:   s.state.CurrentVault,
	}
}

func (s *appVaultService) FindCredentials(domain string) ([]CredentialSummary, error) {
	s.state.Mu.Lock()
	defer s.state.Mu.Unlock()

	if !s.state.IsUnlocked || s.state.CurrentVault == "" {
		return nil, fmt.Errorf("vault is locked")
	}

	vaultFile := pqapp.GetVaultPath(s.state.CurrentVault)
	entries, err := pqapp.ReadVault(vaultFile, s.state.MasterPassword)
	if err != nil {
		return nil, fmt.Errorf("read vault: %w", err)
	}

	passwords := pqapp.EntriesByType(entries, model.EntryTypePassword)

	normalized := NormalizeDomain(domain)
	ids := s.domainMap.Lookup(domain)
	idSet := make(map[uint64]bool, len(ids))
	for _, id := range ids {
		idSet[id] = true
	}

	var results []CredentialSummary
	seen := make(map[uint64]bool)

	for _, entry := range passwords {
		if idSet[entry.ID] {
			results = append(results, CredentialSummary{
				ID:       entry.ID,
				Service:  entry.Service,
				Username: entry.Username,
			})
			seen[entry.ID] = true
		}
	}

	// Fallback: fuzzy match on Service name
	if normalized != "" {
		baseDomain := strings.Split(normalized, ".")[0]
		for _, entry := range passwords {
			if seen[entry.ID] {
				continue
			}
			serviceLower := strings.ToLower(entry.Service)
			if strings.Contains(serviceLower, strings.ToLower(baseDomain)) ||
				strings.Contains(serviceLower, strings.ToLower(normalized)) {
				results = append(results, CredentialSummary{
					ID:       entry.ID,
					Service:  entry.Service,
					Username: entry.Username,
				})
			}
		}
	}

	return results, nil
}

func (s *appVaultService) SaveCredential(domain, username, password string) (uint64, error) {
	s.state.Mu.Lock()
	defer s.state.Mu.Unlock()

	if !s.state.IsUnlocked || s.state.CurrentVault == "" {
		return 0, fmt.Errorf("vault is locked")
	}

	vaultFile := pqapp.GetVaultPath(s.state.CurrentVault)
	entries, err := pqapp.ReadVault(vaultFile, s.state.MasterPassword)
	if err != nil {
		return 0, fmt.Errorf("read vault: %w", err)
	}

	ct, ss, err := pqapp.Encapsulate(s.state.PublicKey)
	if err != nil {
		return 0, fmt.Errorf("encapsulate: %w", err)
	}

	nonce, ciphertext, err := pqapp.EncryptAES256GCM(password, ss)
	if err != nil {
		return 0, fmt.Errorf("encrypt: %w", err)
	}

	entry := model.NewVaultEntry()
	entry.Type = model.EntryTypePassword
	entry.Service = domain
	entry.Username = username
	entry.KyberCiphertext = ct
	entry.Nonce = nonce
	entry.Ciphertext = ciphertext

	entries = append(entries, entry)

	if err := pqapp.WriteVault(entries, vaultFile, s.state.MasterPassword); err != nil {
		return 0, fmt.Errorf("write vault: %w", err)
	}

	_ = s.domainMap.Associate(domain, entry.ID)

	return entry.ID, nil
}

func (s *appVaultService) UpdatePassword(entryID uint64, newPassword string) error {
	s.state.Mu.Lock()
	defer s.state.Mu.Unlock()

	if !s.state.IsUnlocked || s.state.CurrentVault == "" {
		return fmt.Errorf("vault is locked")
	}

	vaultFile := pqapp.GetVaultPath(s.state.CurrentVault)
	entries, err := pqapp.ReadVault(vaultFile, s.state.MasterPassword)
	if err != nil {
		return fmt.Errorf("read vault: %w", err)
	}

	var target *model.VaultEntry
	for _, e := range entries {
		if e.ID == entryID {
			target = e
			break
		}
	}
	if target == nil {
		return fmt.Errorf("entry %d not found", entryID)
	}

	ct, ss, err := pqapp.Encapsulate(s.state.PublicKey)
	if err != nil {
		return fmt.Errorf("encapsulate: %w", err)
	}

	nonce, ciphertext, err := pqapp.EncryptAES256GCM(newPassword, ss)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	target.KyberCiphertext = ct
	target.Nonce = nonce
	target.Ciphertext = ciphertext

	return pqapp.WriteVault(entries, vaultFile, s.state.MasterPassword)
}
