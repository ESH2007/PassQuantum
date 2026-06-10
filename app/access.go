package app

import (
"bytes"
"fmt"
"os"
"path/filepath"
"runtime"
"strings"
"sync"

"passquantum/core/crypto"
"passquantum/core/model"
"passquantum/core/storage"
)

const appSecurityMetadataPath = storage.DefaultAppSecurityMetadataPath

// maxRotationWorkers caps how many vaults are re-encrypted in parallel during a
// master-password change. Each ReencryptVaultFile runs Argon2id (64 MB) for both
// the old and new password, so this bounds peak memory regardless of core count.
const maxRotationWorkers = 4

// StartupAccessState describes what the login screen should present on startup.
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
appState.SecurityProfile = nil
appState.StartupWarning = ""
return StartupAccessState{RequiresSetup: true}, nil
}

profile, err := storage.LoadAppSecurityProfile(appSecurityMetadataPath)
if err != nil {
return StartupAccessState{}, err
}

fingerprint, err := crypto.PrivateKeyFingerprint(appState.PrivateKey)
if err != nil {
return StartupAccessState{}, err
}

if !bytes.Equal(profile.PrivateKeyFingerprint, fingerprint) {
warning := "The stored master-password profile belongs to a different private.key. Existing vaults may require manual migration before they can be opened again."
appState.SecurityProfile = nil
appState.StartupWarning = warning
return StartupAccessState{RequiresSetup: true, Warning: warning}, nil
}

appState.SecurityProfile = profile
appState.StartupWarning = ""
return StartupAccessState{RequiresSetup: false}, nil
}

func CreateMasterPasswordProfile(appState *AppState, masterPassword string) error {
profile, sessionEncryptionKey, sessionVerificationKey, err := crypto.CreateAppSecurityProfile(masterPassword, appState.PrivateKey)
if err != nil {
return err
}

if err := storage.SaveAppSecurityProfile(appSecurityMetadataPath, profile); err != nil {
crypto.WipeBytes(sessionEncryptionKey)
crypto.WipeBytes(sessionVerificationKey)
return err
}

appState.StoreUnlockedSession(masterPassword, profile, sessionEncryptionKey, sessionVerificationKey)
appState.StartupWarning = ""
return nil
}

func unlockAppSession(appState *AppState, masterPassword string) error {
profile := appState.SecurityProfile
if profile == nil {
var err error
profile, err = storage.LoadAppSecurityProfile(appSecurityMetadataPath)
if err != nil {
return fmt.Errorf("failed to load app security profile: %w", err)
}
}

sessionEncryptionKey, sessionVerificationKey, verified, fingerprintMatches, err := crypto.VerifyAppSecurityProfile(profile, masterPassword, appState.PrivateKey)
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

appState.StoreUnlockedSession(masterPassword, profile, sessionEncryptionKey, sessionVerificationKey)
return nil
}

func createVaultWithUnlockedSession(appState *AppState, vaultName string) error {
vaultName = strings.TrimSpace(vaultName)
if vaultName == "" {
return fmt.Errorf("vault name cannot be empty")
}

if !appState.IsUnlocked || appState.MasterPassword == "" {
return fmt.Errorf("unlock the app before creating a vault")
}

vaultFile := GetVaultPath(vaultName)
if _, err := os.Stat(vaultFile); err == nil {
return fmt.Errorf("vault '%s' already exists", vaultName)
}

if err := WriteVault([]*model.VaultEntry{}, vaultFile, appState.MasterPassword); err != nil {
return err
}

appState.StoreCurrentVaultState(vaultName)
return nil
}

func openVaultWithUnlockedSession(appState *AppState, vaultName string) error {
if !appState.IsUnlocked || appState.MasterPassword == "" {
return fmt.Errorf("unlock the app before opening a vault")
}

vaultFile := GetVaultPath(vaultName)
if _, err := storage.ReadVault(vaultFile, appState.MasterPassword); err != nil {
return fmt.Errorf("failed to open vault with the unlocked global master password: %w", err)
}

appState.StoreCurrentVaultState(vaultName)
return nil
}

func ChangeMasterPassword(appState *AppState, currentPassword string, newPassword string) error {
if !appState.IsUnlocked {
return fmt.Errorf("unlock the app before changing the master password")
}

if strings.TrimSpace(newPassword) == "" {
return fmt.Errorf("new master password cannot be empty")
}

profile := appState.SecurityProfile
if profile == nil {
var err error
profile, err = storage.LoadAppSecurityProfile(appSecurityMetadataPath)
if err != nil {
return fmt.Errorf("failed to load app security profile: %w", err)
}
}

_, _, verified, fingerprintMatches, err := crypto.VerifyAppSecurityProfile(profile, currentPassword, appState.PrivateKey)
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
preparedVaults := make([]preparedVaultRotation, len(vaultNames))

// Re-encrypt vaults concurrently. This step is purely in-memory (no vault
// file is written until staging below), so an error from any worker simply
// aborts the whole change with nothing persisted. Parallelism is bounded by
// maxRotationWorkers to keep Argon2id's memory cost in check.
type rotationOutcome struct {
path         string
originalData []byte
prepared     preparedVaultRotation
err          error
}

if len(vaultNames) > 0 {
workers := runtime.NumCPU()
if workers > maxRotationWorkers {
workers = maxRotationWorkers
}
if workers > len(vaultNames) {
workers = len(vaultNames)
}

results := make([]rotationOutcome, len(vaultNames))
sem := make(chan struct{}, workers)
var wg sync.WaitGroup

for i, vaultName := range vaultNames {
wg.Add(1)
go func(i int, vaultName string) {
defer wg.Done()
sem <- struct{}{}
defer func() { <-sem }()

vaultPath := GetVaultPath(vaultName)
currentData, err := os.ReadFile(vaultPath)
if err != nil {
results[i] = rotationOutcome{err: fmt.Errorf("failed to read vault %s: %w", vaultName, err)}
return
}

rotatedData, err := storage.ReencryptVaultFile(vaultPath, currentPassword, newPassword)
if err != nil {
results[i] = rotationOutcome{err: fmt.Errorf("failed to rotate vault %s: %w", vaultName, err)}
return
}

results[i] = rotationOutcome{
path:         vaultPath,
originalData: currentData,
prepared: preparedVaultRotation{
path:     vaultPath,
tempPath: vaultPath + ".tmp",
data:     rotatedData,
},
}
}(i, vaultName)
}
wg.Wait()

for _, r := range results {
if r.err != nil {
return r.err
}
}
for i, r := range results {
originalVaultData[r.path] = r.originalData
preparedVaults[i] = r.prepared
}
}

newProfile, sessionEncryptionKey, sessionVerificationKey, err := crypto.CreateAppSecurityProfile(newPassword, appState.PrivateKey)
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

appState.StoreUnlockedSession(newPassword, newProfile, sessionEncryptionKey, sessionVerificationKey)

if appState.CurrentVault != "" {
appState.StoreCurrentVaultState(appState.CurrentVault)
}

return nil
}
