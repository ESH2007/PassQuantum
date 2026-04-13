package storage

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"passquantum/core/model"
)

var vaultPlaintextMagic = []byte("PQV2")

func serializeEntries(entries []*model.VaultEntry) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 256))
	buf.Write(vaultPlaintextMagic)

	_ = binary.Write(buf, binary.BigEndian, uint32(len(entries)))
	for _, entry := range entries {
		if entry == nil {
			continue
		}
		encoded := entry.SerializeV2()
		_ = binary.Write(buf, binary.BigEndian, uint32(len(encoded)))
		buf.Write(encoded)
	}

	return buf.Bytes()
}

func deserializeEntries(plaintext []byte) ([]*model.VaultEntry, error) {
	if len(plaintext) == 0 {
		return []*model.VaultEntry{}, nil
	}

	if bytes.HasPrefix(plaintext, vaultPlaintextMagic) {
		return deserializeEntriesV2(plaintext)
	}

	return deserializeLegacyEntries(plaintext)
}

func deserializeEntriesV2(plaintext []byte) ([]*model.VaultEntry, error) {
	if len(plaintext) < len(vaultPlaintextMagic)+4 {
		return nil, fmt.Errorf("invalid typed vault payload: missing header")
	}

	idx := len(vaultPlaintextMagic)
	entryCount := int(binary.BigEndian.Uint32(plaintext[idx : idx+4]))
	idx += 4

	entries := make([]*model.VaultEntry, 0, entryCount)
	for idx < len(plaintext) {
		if idx+4 > len(plaintext) {
			return nil, fmt.Errorf("invalid typed vault payload: truncated entry length")
		}

		entryLen := int(binary.BigEndian.Uint32(plaintext[idx : idx+4]))
		idx += 4
		if entryLen < 0 || idx+entryLen > len(plaintext) {
			return nil, fmt.Errorf("invalid typed vault payload: truncated entry")
		}

		entry, err := model.DeserializeV2(plaintext[idx : idx+entryLen])
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
		idx += entryLen
	}

	if entryCount > 0 && len(entries) != entryCount {
		return nil, fmt.Errorf("invalid typed vault payload: entry count mismatch")
	}

	return entries, nil
}

func deserializeLegacyEntries(plaintext []byte) ([]*model.VaultEntry, error) {
	entries := make([]*model.VaultEntry, 0)
	idx := 0

	for idx < len(plaintext) {
		if idx+30 > len(plaintext) {
			break
		}

		serviceLenPos := idx + 8
		serviceLen := int(plaintext[serviceLenPos])<<8 | int(plaintext[serviceLenPos+1])

		servicePosStart := idx + 10
		servicePos := servicePosStart + serviceLen
		if servicePos > len(plaintext) {
			break
		}

		usernameLenPos := servicePos
		if usernameLenPos+2 > len(plaintext) {
			break
		}
		usernameLen := int(plaintext[usernameLenPos])<<8 | int(plaintext[usernameLenPos+1])

		usernamePosStart := usernameLenPos + 2
		usernamePos := usernamePosStart + usernameLen
		if usernamePos+2 > len(plaintext) {
			break
		}

		kyberLenPos := usernamePos
		kyberLen := int(plaintext[kyberLenPos])<<8 | int(plaintext[kyberLenPos+1])

		kyberPosStart := kyberLenPos + 2
		kyberPos := kyberPosStart + kyberLen
		if kyberPos+12+2 > len(plaintext) {
			break
		}

		nonceEnd := kyberPos + 12
		ciphertextLenPos := nonceEnd
		ciphertextLen := int(plaintext[ciphertextLenPos])<<8 | int(plaintext[ciphertextLenPos+1])

		ciphertextPosStart := ciphertextLenPos + 2
		ciphertextPos := ciphertextPosStart + ciphertextLen
		if ciphertextPos > len(plaintext) {
			break
		}

		entry, err := model.Deserialize(plaintext[idx:ciphertextPos])
		if err != nil {
			break
		}

		entries = append(entries, entry)
		idx = ciphertextPos
	}

	return entries, nil
}
