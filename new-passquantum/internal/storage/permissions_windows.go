//go:build windows

package storage

// ValidateVaultPermissions is a no-op on Windows.
func ValidateVaultPermissions() error {
	return nil
}
