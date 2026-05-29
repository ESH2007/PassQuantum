package browser

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	securestorage "passquantum/internal/storage"
)

const (
	configFileName = "browser_config.json"
	secretBytes    = 32
)

type Config struct {
	mu        sync.RWMutex
	Secret    string   `json:"secret"`
	PairedAt  string   `json:"paired_at,omitempty"`
	NeverSave []string `json:"never_save"`
	filePath  string
}

func ConfigPath() (string, error) {
	return securestorage.GetSecureFilePath(configFileName)
}

func LoadConfig() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, fmt.Errorf("config path: %w", err)
	}

	cfg := &Config{filePath: path, NeverSave: []string{}}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.NeverSave == nil {
		cfg.NeverSave = []string{}
	}
	cfg.filePath = path
	return cfg, nil
}

func (c *Config) Save() error {
	c.mu.RLock()
	data, err := json.MarshalIndent(c, "", "  ")
	c.mu.RUnlock()
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	path := c.filePath
	if path == "" {
		p, err := ConfigPath()
		if err != nil {
			return err
		}
		path = p
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func (c *Config) IsPaired() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Secret != ""
}

func (c *Config) SetPaired(secret string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Secret = secret
	c.PairedAt = time.Now().UTC().Format(time.RFC3339)
}

func (c *Config) IsNeverSave(domain string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, d := range c.NeverSave {
		if d == domain {
			return true
		}
	}
	return false
}

func (c *Config) AddNeverSave(domain string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, d := range c.NeverSave {
		if d == domain {
			return
		}
	}
	c.NeverSave = append(c.NeverSave, domain)
}

func (c *Config) RemoveNeverSave(domain string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, d := range c.NeverSave {
		if d == domain {
			c.NeverSave = append(c.NeverSave[:i], c.NeverSave[i+1:]...)
			return
		}
	}
}

func (c *Config) GetNeverSaveList() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]string, len(c.NeverSave))
	copy(out, c.NeverSave)
	return out
}

func GenerateSecret() (string, error) {
	b := make([]byte, secretBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate secret: %w", err)
	}
	return hex.EncodeToString(b), nil
}
