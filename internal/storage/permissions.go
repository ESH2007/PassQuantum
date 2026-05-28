//go:build darwin || linux

package storage

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// ValidateVaultPermissions validates and hardens vault directory and file permissions.
func ValidateVaultPermissions() error {
	vaultDir, err := getVaultDir()
	if err != nil {
		return fmt.Errorf("failed to resolve vault directory: %w", err)
	}

	return validateVaultPermissions(vaultDir)
}

func validateVaultPermissions(vaultDir string) error {
	if vaultDir == "" {
		return fmt.Errorf("vault directory is required")
	}

	info, err := os.Stat(vaultDir)
	if err != nil {
		if os.IsNotExist(err) {
			if mkErr := os.MkdirAll(vaultDir, 0700); mkErr != nil {
				return fmt.Errorf("failed to create vault directory %q: %w", vaultDir, mkErr)
			}
			log.Printf("WARNING: created missing vault directory %q with 0700 permissions", vaultDir)
			info, err = os.Stat(vaultDir)
		}
		if err != nil {
			return fmt.Errorf("failed to stat vault directory %q: %w", vaultDir, err)
		}
	}

	if !info.IsDir() {
		return fmt.Errorf("vault path %q is not a directory", vaultDir)
	}

	if info.Mode().Perm() != 0700 {
		if err := os.Chmod(vaultDir, 0700); err != nil {
			return fmt.Errorf("failed to set vault directory permissions on %q: %w", vaultDir, err)
		}
		log.Printf("WARNING: corrected vault directory permissions to 0700: %q", vaultDir)
	}

	entries, err := os.ReadDir(vaultDir)
	if err != nil {
		return fmt.Errorf("failed to list vault directory %q: %w", vaultDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if entry.Name() == "files" {
				validateFilesSubdirectory(filepath.Join(vaultDir, "files"))
			}
			continue
		}

		ext := filepath.Ext(entry.Name())
		if ext != ".enc" && ext != ".key" && ext != ".pqmeta" && ext != ".npy" && ext != ".bin" {
			continue
		}

		fullPath := filepath.Join(vaultDir, entry.Name())
		fileInfo, err := entry.Info()
		if err != nil {
			return fmt.Errorf("failed to stat vault file %q: %w", fullPath, err)
		}

		if fileInfo.Mode().Perm() != 0600 {
			if err := os.Chmod(fullPath, 0600); err != nil {
				return fmt.Errorf("failed to set vault file permissions on %q: %w", fullPath, err)
			}
			log.Printf("WARNING: corrected vault file permissions to 0600: %q", fullPath)
		}
	}

	return nil
}

// validateFilesSubdirectory hardens the files/ subdirectory and its contents.
func validateFilesSubdirectory(filesDir string) {
	info, err := os.Stat(filesDir)
	if err != nil || !info.IsDir() {
		return
	}

	if info.Mode().Perm() != 0700 {
		if err := os.Chmod(filesDir, 0700); err == nil {
			log.Printf("WARNING: corrected files directory permissions to 0700: %q", filesDir)
		}
	}

	// Recurse into per-vault subdirectories
	vaultDirs, err := os.ReadDir(filesDir)
	if err != nil {
		return
	}
	for _, vaultEntry := range vaultDirs {
		if !vaultEntry.IsDir() {
			continue
		}
		vaultSubdir := filepath.Join(filesDir, vaultEntry.Name())
		if vInfo, err := os.Stat(vaultSubdir); err == nil && vInfo.Mode().Perm() != 0700 {
			if err := os.Chmod(vaultSubdir, 0700); err == nil {
				log.Printf("WARNING: corrected vault files subdirectory permissions to 0700: %q", vaultSubdir)
			}
		}

		files, err := os.ReadDir(vaultSubdir)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			ext := filepath.Ext(f.Name())
			if ext != ".bin" && ext != ".enc" {
				continue
			}
			fPath := filepath.Join(vaultSubdir, f.Name())
			if fInfo, err := f.Info(); err == nil && fInfo.Mode().Perm() != 0600 {
				if err := os.Chmod(fPath, 0600); err == nil {
					log.Printf("WARNING: corrected file vault permissions to 0600: %q", fPath)
				}
			}
		}
	}
}
