package browser

// VaultService abstracts vault operations for the browser API.
// All methods are safe for concurrent use.
type VaultService interface {
	IsReady() bool
	Status() VaultStatus
	FindCredentials(domain string) ([]CredentialSummary, error)
	SaveCredential(domain, username, password string) (uint64, error)
	UpdatePassword(entryID uint64, newPassword string) error
}

type VaultStatus struct {
	AppUnlocked bool
	VaultName   string
}

type CredentialSummary struct {
	ID       uint64 `json:"id"`
	Service  string `json:"service"`
	Username string `json:"username"`
}
