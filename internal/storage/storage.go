package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const vaultDirName = "passquantum"

// GetVaultDir resolves and creates the secure vault directory.
func GetVaultDir() (string, error) {
	return getVaultDir()
}

// ResolveVaultPath resolves a vault filename into the secure vault directory.
func ResolveVaultPath(vaultPath string) (string, error) {
	if strings.TrimSpace(vaultPath) == "" {
		return "", fmt.Errorf("vault path is required")
	}

	vaultDir, err := getVaultDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve vault directory: %w", err)
	}

	fileName := filepath.Base(vaultPath)
	if fileName == "." || fileName == string(filepath.Separator) {
		return "", fmt.Errorf("invalid vault path: %q", vaultPath)
	}

	return filepath.Join(vaultDir, fileName), nil
}

// WriteVaultFile writes vault data to the secure vault directory.
func WriteVaultFile(vaultPath string, data []byte) error {
	return writeVaultFile(vaultPath, data)
}

// ReadVaultFile reads vault data from the secure vault directory.
func ReadVaultFile(vaultPath string) ([]byte, error) {
	return readVaultFile(vaultPath)
}

// GetSecureFilePath returns the full path for a named file inside the vault directory.
func GetSecureFilePath(name string) (string, error) {
	vaultDir, err := GetVaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(vaultDir, name), nil
}

func getVaultDir() (string, error) {
	baseConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve user config directory: %w", err)
	}

	vaultDir := filepath.Join(baseConfigDir, vaultDirName)
	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create vault directory %q: %w", vaultDir, err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(vaultDir, 0700); err != nil {
			return "", fmt.Errorf("failed to set vault directory permissions on %q: %w", vaultDir, err)
		}
	}

	return vaultDir, nil
}

func writeVaultFile(vaultPath string, data []byte) error {
	resolvedPath, err := ResolveVaultPath(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to resolve vault path: %w", err)
	}

	file, err := os.OpenFile(resolvedPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open vault file %q: %w", resolvedPath, err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write vault file %q: %w", resolvedPath, err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(resolvedPath, 0600); err != nil {
			return fmt.Errorf("failed to set vault file permissions on %q: %w", resolvedPath, err)
		}
	}

	return nil
}

func readVaultFile(vaultPath string) ([]byte, error) {
	resolvedPath, err := ResolveVaultPath(vaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve vault path: %w", err)
	}

	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read vault file %q: %w", resolvedPath, err)
	}

	return data, nil
}
