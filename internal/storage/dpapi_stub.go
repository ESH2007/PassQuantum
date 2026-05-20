//go:build !windows

package storage

import "fmt"

func DPAPIEncrypt(_ []byte) ([]byte, error) {
	return nil, fmt.Errorf("DPAPI is only available on Windows")
}

func DPAPIDecrypt(_ []byte) ([]byte, error) {
	return nil, fmt.Errorf("DPAPI is only available on Windows")
}
