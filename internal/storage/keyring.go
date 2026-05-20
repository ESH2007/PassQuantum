package storage

import (
	"encoding/base64"
	"fmt"
	"runtime"

	"github.com/zalando/go-keyring"
)

const (
	keyringService = "passquantum"
	keyringAccount = "master-key"
)

// StoreMasterKey stores the master key in the OS keyring.
func StoreMasterKey(key []byte) error {
	if len(key) == 0 {
		return fmt.Errorf("master key is required")
	}

	value := append([]byte(nil), key...)

	if runtime.GOOS == "windows" {
		encrypted, err := DPAPIEncrypt(value)
		if err != nil {
			return fmt.Errorf("failed to protect master key with DPAPI: %w", err)
		}
		value = encrypted
	}

	encoded := base64.StdEncoding.EncodeToString(value)
	if err := keyring.Set(keyringService, keyringAccount, encoded); err != nil {
		return fmt.Errorf("failed to store master key in keyring: %w", err)
	}

	return nil
}

// LoadMasterKey loads the master key from the OS keyring.
func LoadMasterKey() ([]byte, error) {
	encoded, err := keyring.Get(keyringService, keyringAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to load master key from keyring: %w", err)
	}

	value, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode stored master key: %w", err)
	}

	if runtime.GOOS == "windows" {
		decrypted, err := DPAPIDecrypt(value)
		if err != nil {
			return nil, fmt.Errorf("failed to unprotect master key with DPAPI: %w", err)
		}
		return decrypted, nil
	}

	return value, nil
}

// DeleteMasterKey removes the master key from the OS keyring.
func DeleteMasterKey() error {
	if err := keyring.Delete(keyringService, keyringAccount); err != nil {
		return fmt.Errorf("failed to delete master key from keyring: %w", err)
	}
	return nil
}
