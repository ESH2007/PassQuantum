package app

import (
	"sync"

	"github.com/cloudflare/circl/kem/kyber/kyber768"

	"passquantum/bridge"
	"passquantum/core/crypto"
)

const (
	PubKeyPath  = "public.key"
	PrivKeyPath = "private.key"
)

// AppState holds all in-memory, sensitive application state.
// Fields are intentionally unexported; screens access them through
// the exported methods and helpers in this package.
type AppState struct {
	PublicKey              *kyber768.PublicKey
	PrivateKey             *kyber768.PrivateKey
	MasterPassword         string
	SessionEncryptionKey   []byte
	SessionVerificationKey []byte
	SecurityProfile        *crypto.AppSecurityProfile
	Mu                     sync.Mutex
	IsUnlocked             bool
	IsTraining             bool
	CurrentVault           string
	StartupWarning         string
	FaceGuard              *bridge.FaceGuard
	// LockApp is called from any goroutine to lock the app immediately;
	// it clears sensitive state and returns the user to the login screen.
	LockApp func()
}

func (appState *AppState) StoreUnlockedSession(masterPassword string, profile *crypto.AppSecurityProfile, sessionEncryptionKey []byte, sessionVerificationKey []byte) {
	appState.Mu.Lock()
	defer appState.Mu.Unlock()

	crypto.WipeBytes(appState.SessionEncryptionKey)
	crypto.WipeBytes(appState.SessionVerificationKey)
	appState.MasterPassword = masterPassword
	appState.SessionEncryptionKey = append([]byte(nil), sessionEncryptionKey...)
	appState.SessionVerificationKey = append([]byte(nil), sessionVerificationKey...)
	appState.SecurityProfile = profile
	appState.IsUnlocked = true
	crypto.WipeBytes(sessionEncryptionKey)
	crypto.WipeBytes(sessionVerificationKey)
}

func (appState *AppState) StoreCurrentVaultState(vaultName string) {
	appState.Mu.Lock()
	defer appState.Mu.Unlock()

	appState.CurrentVault = vaultName
	appState.IsUnlocked = true
}

func (appState *AppState) ClearSensitiveState() {
	appState.Mu.Lock()
	defer appState.Mu.Unlock()

	crypto.WipeBytes(appState.SessionEncryptionKey)
	crypto.WipeBytes(appState.SessionVerificationKey)
	appState.SessionEncryptionKey = nil
	appState.SessionVerificationKey = nil
	appState.MasterPassword = ""
	appState.CurrentVault = ""
	appState.IsUnlocked = false
	appState.SecurityProfile = nil
	appState.StartupWarning = ""
}
