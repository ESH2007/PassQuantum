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
	path      string
	tempPath  string
	data      []byte
	kdfParams crypto.KDFParams
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
	loadBiometricFromProfile(appState)
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

	kdfParams := crypto.DefaultKDFParams()
	var err error
	kdfParams.Salt, err = crypto.GenerateSalt()
	if err != nil {
		return err
	}

	encryptionKey, verificationKey, err := crypto.DeriveKeys(appState.masterPassword, kdfParams)
	if err != nil {
		return err
	}

	if err := WriteVault([]*model.PasswordEntry{}, vaultFile, encryptionKey, verificationKey, kdfParams); err != nil {
		crypto.WipeBytes(encryptionKey)
		crypto.WipeBytes(verificationKey)
		return err
	}

	appState.storeCurrentVaultState(vaultName, kdfParams, encryptionKey, verificationKey)
	return nil
}

func openVaultWithUnlockedSession(appState *AppState, vaultName string) error {
	if !appState.isUnlocked || appState.masterPassword == "" {
		return fmt.Errorf("unlock the app before opening a vault")
	}

	vaultFile := GetVaultPath(vaultName)
	vaultData, err := os.ReadFile(vaultFile)
	if err != nil {
		return fmt.Errorf("failed to read vault file: %w", err)
	}

	vault, err := crypto.VaultFileDeserialize(vaultData)
	if err != nil {
		return fmt.Errorf("invalid vault file: %w", err)
	}

	encryptionKey, verificationKey, err := crypto.DeriveKeys(appState.masterPassword, vault.KDFParams)
	if err != nil {
		return fmt.Errorf("failed to derive vault keys: %w", err)
	}

	if _, err := storage.ReadVault(vaultFile, encryptionKey, verificationKey); err != nil {
		crypto.WipeBytes(encryptionKey)
		crypto.WipeBytes(verificationKey)
		return fmt.Errorf("failed to open vault with the unlocked global master password: %w", err)
	}

	appState.storeCurrentVaultState(vaultName, vault.KDFParams, encryptionKey, verificationKey)
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

		rotatedData, rotatedParams, err := storage.ReencryptVaultFile(vaultPath, currentPassword, newPassword)
		if err != nil {
			return fmt.Errorf("failed to rotate vault %s: %w", vaultName, err)
		}

		originalVaultData[vaultPath] = currentData
		preparedVaults = append(preparedVaults, preparedVaultRotation{
			path:      vaultPath,
			tempPath:  vaultPath + ".tmp",
			data:      rotatedData,
			kdfParams: rotatedParams,
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

	if appState.currentVault == "" {
		appState.mu.Lock()
		appState.kdfParams = crypto.KDFParams{}
		crypto.WipeBytes(appState.encryptionKey)
		crypto.WipeBytes(appState.verificationKey)
		appState.encryptionKey = nil
		appState.verificationKey = nil
		appState.mu.Unlock()
		return nil
	}

	for _, preparedVault := range preparedVaults {
		if preparedVault.path != GetVaultPath(appState.currentVault) {
			continue
		}

		currentEncryptionKey, currentVerificationKey, err := crypto.DeriveKeys(newPassword, preparedVault.kdfParams)
		if err != nil {
			return fmt.Errorf("failed to derive updated keys for current vault: %w", err)
		}

		appState.storeCurrentVaultState(appState.currentVault, preparedVault.kdfParams, currentEncryptionKey, currentVerificationKey)
		break
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

func (appState *AppState) storeCurrentVaultState(vaultName string, kdfParams crypto.KDFParams, encryptionKey []byte, verificationKey []byte) {
	appState.mu.Lock()
	defer appState.mu.Unlock()

	crypto.WipeBytes(appState.encryptionKey)
	crypto.WipeBytes(appState.verificationKey)
	appState.currentVault = vaultName
	appState.kdfParams = kdfParams
	appState.encryptionKey = append([]byte(nil), encryptionKey...)
	appState.verificationKey = append([]byte(nil), verificationKey...)
	appState.isUnlocked = true
	crypto.WipeBytes(encryptionKey)
	crypto.WipeBytes(verificationKey)
}

func (appState *AppState) clearSensitiveState() {
	appState.mu.Lock()
	defer appState.mu.Unlock()

	crypto.WipeBytes(appState.encryptionKey)
	crypto.WipeBytes(appState.verificationKey)
	crypto.WipeBytes(appState.sessionEncryptionKey)
	crypto.WipeBytes(appState.sessionVerificationKey)
	appState.encryptionKey = nil
	appState.verificationKey = nil
	appState.sessionEncryptionKey = nil
	appState.sessionVerificationKey = nil
	appState.masterPassword = ""
	appState.currentVault = ""
	appState.kdfParams = crypto.KDFParams{}
	appState.isUnlocked = false
	appState.securityProfile = nil
	appState.startupWarning = ""

	// Stop the continuous face-check goroutine and clear the in-memory template.
	// The pipeline (ONNX models) is kept alive to avoid reload overhead on the next unlock.
	if appState.biometricStopCheck != nil {
		appState.biometricStopCheck()
		appState.biometricStopCheck = nil
	}
	appState.biometricTemplate = nil
}

func showWindowError(window any, err error) {
	if err == nil {
		return
	}

	if fyneWindow, ok := window.(fyne.Window); ok && fyneWindow != nil {
		ShowAppError(err, fyneWindow)
	}
}
