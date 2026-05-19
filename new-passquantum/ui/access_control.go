package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"

	"passquantum/core/crypto"
	"passquantum/core/model"
	"passquantum/core/storage"
)

const appSecurityMetadataPath = storage.DefaultAppSecurityMetadataPath

type StartupAccessState struct {
	RequiresSetup bool
	Warning       string
}

type preparedVaultRotation struct {
	path     string
	tempPath string
	data     []byte
}

func ResolveStartupAccessState(appState *AppState) (StartupAccessState, error) {
	if !storage.AppSecurityProfileExists(appSecurityMetadataPath) {
		appState.securityProfile = nil
		appState.startupWarning = ""
		return StartupAccessState{RequiresSetup: true}, nil
	}

	profile, err := storage.LoadAppSecurityProfile(appSecurityMetadataPath)
	if err != nil {
		return StartupAccessState{}, err
	}

	fingerprint, err := crypto.PrivateKeyFingerprint(appState.privateKey)
	if err != nil {
		return StartupAccessState{}, err
	}

	if !bytes.Equal(profile.PrivateKeyFingerprint, fingerprint) {
		warning := "The stored master-password profile belongs to a different private.key. Existing vaults may require manual migration before they can be opened again."
		appState.securityProfile = nil
		appState.startupWarning = warning
		return StartupAccessState{RequiresSetup: true, Warning: warning}, nil
	}

	appState.securityProfile = profile
	appState.startupWarning = ""
	return StartupAccessState{RequiresSetup: false}, nil
}

func CreateMasterPasswordProfile(appState *AppState, masterPassword string) error {
	profile, sessionEncryptionKey, sessionVerificationKey, err := crypto.CreateAppSecurityProfile(masterPassword, appState.privateKey)
	if err != nil {
		return err
	}

	if err := storage.SaveAppSecurityProfile(appSecurityMetadataPath, profile); err != nil {
		crypto.WipeBytes(sessionEncryptionKey)
		crypto.WipeBytes(sessionVerificationKey)
		return err
	}

	appState.storeUnlockedSession(masterPassword, profile, sessionEncryptionKey, sessionVerificationKey)
	appState.startupWarning = ""
	return nil
}

func unlockAppSession(appState *AppState, masterPassword string) error {
	profile := appState.securityProfile
	if profile == nil {
		var err error
		profile, err = storage.LoadAppSecurityProfile(appSecurityMetadataPath)
		if err != nil {
			return fmt.Errorf("failed to load app security profile: %w", err)
		}
	}

	sessionEncryptionKey, sessionVerificationKey, verified, fingerprintMatches, err := crypto.VerifyAppSecurityProfile(profile, masterPassword, appState.privateKey)
	if err != nil {
		return err
	}

	if !fingerprintMatches {
		return fmt.Errorf("the stored master-password profile does not match the loaded private.key")
	}

	if !verified {
		crypto.WipeBytes(sessionEncryptionKey)
		crypto.WipeBytes(sessionVerificationKey)
		return fmt.Errorf("incorrect master password")
	}

	appState.storeUnlockedSession(masterPassword, profile, sessionEncryptionKey, sessionVerificationKey)
	return nil
}

func createVaultWithUnlockedSession(appState *AppState, vaultName string) error {
	vaultName = strings.TrimSpace(vaultName)
	if vaultName == "" {
		return fmt.Errorf("vault name cannot be empty")
	}

	if !appState.isUnlocked || appState.masterPassword == "" {
		return fmt.Errorf("unlock the app before creating a vault")
	}

	vaultFile := GetVaultPath(vaultName)
	if _, err := os.Stat(vaultFile); err == nil {
		return fmt.Errorf("vault '%s' already exists", vaultName)
	}

	if err := WriteVault([]*model.VaultEntry{}, vaultFile, appState.masterPassword); err != nil {
		return err
	}

	appState.storeCurrentVaultState(vaultName)
	return nil
}

func openVaultWithUnlockedSession(appState *AppState, vaultName string) error {
	if !appState.isUnlocked || appState.masterPassword == "" {
		return fmt.Errorf("unlock the app before opening a vault")
	}

	vaultFile := GetVaultPath(vaultName)
	if _, err := storage.ReadVault(vaultFile, appState.masterPassword); err != nil {
		return fmt.Errorf("failed to open vault with the unlocked global master password: %w", err)
	}

	appState.storeCurrentVaultState(vaultName)
	return nil
}

func changeMasterPassword(appState *AppState, currentPassword string, newPassword string) error {
	if !appState.isUnlocked {
		return fmt.Errorf("unlock the app before changing the master password")
	}

	if strings.TrimSpace(newPassword) == "" {
		return fmt.Errorf("new master password cannot be empty")
	}

	profile := appState.securityProfile
	if profile == nil {
		var err error
		profile, err = storage.LoadAppSecurityProfile(appSecurityMetadataPath)
		if err != nil {
			return fmt.Errorf("failed to load app security profile: %w", err)
		}
	}

	_, _, verified, fingerprintMatches, err := crypto.VerifyAppSecurityProfile(profile, currentPassword, appState.privateKey)
	if err != nil {
		return err
	}

	if !fingerprintMatches {
		return fmt.Errorf("the stored master-password profile no longer matches the loaded private.key")
	}

	if !verified {
		return fmt.Errorf("current password is incorrect")
	}

	vaultNames := ListVaults()
	originalVaultData := make(map[string][]byte, len(vaultNames))
	preparedVaults := make([]preparedVaultRotation, 0, len(vaultNames))

	for _, vaultName := range vaultNames {
		vaultPath := GetVaultPath(vaultName)
		currentData, err := os.ReadFile(vaultPath)
		if err != nil {
			return fmt.Errorf("failed to read vault %s: %w", vaultName, err)
		}

		rotatedData, err := storage.ReencryptVaultFile(vaultPath, currentPassword, newPassword)
		if err != nil {
			return fmt.Errorf("failed to rotate vault %s: %w", vaultName, err)
		}

		originalVaultData[vaultPath] = currentData
		preparedVaults = append(preparedVaults, preparedVaultRotation{
			path:     vaultPath,
			tempPath: vaultPath + ".tmp",
			data:     rotatedData,
		})
	}

	newProfile, sessionEncryptionKey, sessionVerificationKey, err := crypto.CreateAppSecurityProfile(newPassword, appState.privateKey)
	if err != nil {
		return err
	}

	metadataExisted := false
	originalMetadataData, readMetadataErr := os.ReadFile(appSecurityMetadataPath)
	if readMetadataErr == nil {
		metadataExisted = true
	} else if !os.IsNotExist(readMetadataErr) {
		return fmt.Errorf("failed to read existing app security metadata: %w", readMetadataErr)
	}

	metadataTempPath := appSecurityMetadataPath + ".tmp"

	cleanupTempFiles := func() {
		for _, preparedVault := range preparedVaults {
			_ = os.Remove(preparedVault.tempPath)
		}
		_ = os.Remove(metadataTempPath)
	}

	restoreVaults := func(paths []string) {
		for _, path := range paths {
			if data, exists := originalVaultData[path]; exists {
				_ = os.WriteFile(path, data, 0600)
			}
		}
	}

	restoreMetadata := func() {
		if metadataExisted {
			_ = os.WriteFile(appSecurityMetadataPath, originalMetadataData, 0600)
			return
		}
		_ = os.Remove(appSecurityMetadataPath)
	}

	for _, preparedVault := range preparedVaults {
		if err := os.WriteFile(preparedVault.tempPath, preparedVault.data, 0600); err != nil {
			cleanupTempFiles()
			return fmt.Errorf("failed to stage rotated vault %s: %w", filepath.Base(preparedVault.path), err)
		}
	}

	if err := storage.SaveAppSecurityProfile(metadataTempPath, newProfile); err != nil {
		cleanupTempFiles()
		return fmt.Errorf("failed to stage app security metadata: %w", err)
	}

	replacedVaults := make([]string, 0, len(preparedVaults))
	for _, preparedVault := range preparedVaults {
		if err := os.Rename(preparedVault.tempPath, preparedVault.path); err != nil {
			cleanupTempFiles()
			restoreVaults(replacedVaults)
			return fmt.Errorf("failed to activate rotated vault %s: %w", filepath.Base(preparedVault.path), err)
		}
		replacedVaults = append(replacedVaults, preparedVault.path)
	}

	if err := os.Rename(metadataTempPath, appSecurityMetadataPath); err != nil {
		cleanupTempFiles()
		restoreVaults(replacedVaults)
		restoreMetadata()
		return fmt.Errorf("failed to activate app security metadata: %w", err)
	}

	appState.storeUnlockedSession(newPassword, newProfile, sessionEncryptionKey, sessionVerificationKey)

	// If a vault was open, update masterPassword in state so subsequent reads
	// use the new password. storeCurrentVaultState keeps currentVault name.
	if appState.currentVault != "" {
		appState.storeCurrentVaultState(appState.currentVault)
	}

	return nil
}

func (appState *AppState) storeUnlockedSession(masterPassword string, profile *crypto.AppSecurityProfile, sessionEncryptionKey []byte, sessionVerificationKey []byte) {
	appState.mu.Lock()
	defer appState.mu.Unlock()

	crypto.WipeBytes(appState.sessionEncryptionKey)
	crypto.WipeBytes(appState.sessionVerificationKey)
	appState.masterPassword = masterPassword
	appState.sessionEncryptionKey = append([]byte(nil), sessionEncryptionKey...)
	appState.sessionVerificationKey = append([]byte(nil), sessionVerificationKey...)
	appState.securityProfile = profile
	appState.isUnlocked = true
	crypto.WipeBytes(sessionEncryptionKey)
	crypto.WipeBytes(sessionVerificationKey)
}

func (appState *AppState) storeCurrentVaultState(vaultName string) {
	appState.mu.Lock()
	defer appState.mu.Unlock()

	appState.currentVault = vaultName
	appState.isUnlocked = true
}

func (appState *AppState) clearSensitiveState() {
	appState.mu.Lock()
	defer appState.mu.Unlock()

	crypto.WipeBytes(appState.sessionEncryptionKey)
	crypto.WipeBytes(appState.sessionVerificationKey)
	appState.sessionEncryptionKey = nil
	appState.sessionVerificationKey = nil
	appState.masterPassword = ""
	appState.currentVault = ""
	appState.isUnlocked = false
	appState.securityProfile = nil
	appState.startupWarning = ""
}

func showWindowError(window any, err error) {
	if err == nil {
		return
	}

	if fyneWindow, ok := window.(fyne.Window); ok && fyneWindow != nil {
		ShowAppError(err, fyneWindow)
	}
}
