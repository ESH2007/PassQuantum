package browser

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	securestorage "passquantum/internal/storage"
)

const domainMapFileName = "domain_map.json"

type DomainMap struct {
	mu       sync.RWMutex
	entries  map[string][]uint64
	filePath string
}

func NewDomainMap() (*DomainMap, error) {
	path, err := securestorage.GetSecureFilePath(domainMapFileName)
	if err != nil {
		return nil, fmt.Errorf("domain map path: %w", err)
	}

	dm := &DomainMap{
		entries:  make(map[string][]uint64),
		filePath: path,
	}

	if err := dm.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return dm, nil
}

func (dm *DomainMap) Lookup(domain string) []uint64 {
	normalized := NormalizeDomain(domain)
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	ids, ok := dm.entries[normalized]
	if !ok {
		return nil
	}
	out := make([]uint64, len(ids))
	copy(out, ids)
	return out
}

func (dm *DomainMap) Associate(domain string, entryID uint64) error {
	normalized := NormalizeDomain(domain)
	dm.mu.Lock()
	defer dm.mu.Unlock()

	for _, id := range dm.entries[normalized] {
		if id == entryID {
			return nil
		}
	}

	dm.entries[normalized] = append(dm.entries[normalized], entryID)
	return dm.save()
}

func (dm *DomainMap) Dissociate(domain string, entryID uint64) error {
	normalized := NormalizeDomain(domain)
	dm.mu.Lock()
	defer dm.mu.Unlock()

	ids := dm.entries[normalized]
	for i, id := range ids {
		if id == entryID {
			dm.entries[normalized] = append(ids[:i], ids[i+1:]...)
			if len(dm.entries[normalized]) == 0 {
				delete(dm.entries, normalized)
			}
			return dm.save()
		}
	}
	return nil
}

func (dm *DomainMap) Load() error {
	data, err := os.ReadFile(dm.filePath)
	if err != nil {
		return err
	}

	dm.mu.Lock()
	defer dm.mu.Unlock()

	return json.Unmarshal(data, &dm.entries)
}

func (dm *DomainMap) save() error {
	data, err := json.MarshalIndent(dm.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal domain map: %w", err)
	}
	if err := os.WriteFile(dm.filePath, data, 0600); err != nil {
		return fmt.Errorf("write domain map: %w", err)
	}
	return nil
}
