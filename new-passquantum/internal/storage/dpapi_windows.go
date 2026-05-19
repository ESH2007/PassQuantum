//go:build windows

package storage

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// DPAPIEncrypt encrypts data for the current Windows user using DPAPI.
func DPAPIEncrypt(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data is required")
	}

	input := bytesToDataBlob(data)
	var output windows.DataBlob

	if err := windows.CryptProtectData(&input, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &output); err != nil {
		return nil, fmt.Errorf("CryptProtectData failed: %w", err)
	}
	defer func() {
		_, _ = windows.LocalFree(windows.Handle(unsafe.Pointer(output.Data)))
	}()

	return dataBlobToBytes(&output), nil
}

// DPAPIDecrypt decrypts DPAPI-encrypted data for the current Windows user.
func DPAPIDecrypt(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data is required")
	}

	input := bytesToDataBlob(data)
	var output windows.DataBlob

	if err := windows.CryptUnprotectData(&input, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &output); err != nil {
		return nil, fmt.Errorf("CryptUnprotectData failed: %w", err)
	}
	defer func() {
		_, _ = windows.LocalFree(windows.Handle(unsafe.Pointer(output.Data)))
	}()

	return dataBlobToBytes(&output), nil
}

func bytesToDataBlob(data []byte) windows.DataBlob {
	if len(data) == 0 {
		return windows.DataBlob{}
	}

	return windows.DataBlob{
		Size: uint32(len(data)),
		Data: &data[0],
	}
}

func dataBlobToBytes(blob *windows.DataBlob) []byte {
	if blob == nil || blob.Size == 0 || blob.Data == nil {
		return []byte{}
	}

	data := unsafe.Slice(blob.Data, int(blob.Size))
	cloned := make([]byte, len(data))
	copy(cloned, data)
	return cloned
}
