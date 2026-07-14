package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/shyamborole/azure-ssh-tui/internal/azure"
)

type HistoryEntry struct {
	Name          string `json:"name"`
	ResourceGroup string `json:"resourceGroup"`
	PreferPrivate bool   `json:"preferPrivate"`
	UsedJumpHost  bool   `json:"usedJumpHost"`
}

type Config struct {
	JumpHostName          string `json:"jumpHostName"`
	JumpHostResourceGroup string `json:"jumpHostResourceGroup"`

	LastVMName          string `json:"lastVmName"`
	LastVMResourceGroup string `json:"lastVmResourceGroup"`
	LastVMPreferPrivate bool   `json:"lastVmPreferPrivate"`
	LastVMUsedJumpHost  bool   `json:"lastVmUsedJumpHost"`

	History []HistoryEntry `json:"history"`

	CachedSubscriptions []azure.Subscription          `json:"cachedSubscriptions"`
	CachedVMs           map[string][]azure.VM         `json:"cachedVMs"`
}

func (c *Config) AddHistory(entry HistoryEntry) {
	// Remove if already exists
	var newHistory []HistoryEntry
	for _, h := range c.History {
		if h.Name != entry.Name || h.ResourceGroup != entry.ResourceGroup {
			newHistory = append(newHistory, h)
		}
	}

	// Add to front
	c.History = append([]HistoryEntry{entry}, newHistory...)

	// Truncate to 5
	if len(c.History) > 5 {
		c.History = c.History[:5]
	}
}

func getConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(configDir, "azure-ssh-tui", "config.json")
}

func Load() Config {
	var c Config
	path := getConfigPath()
	if path == "" {
		c.CachedVMs = make(map[string][]azure.VM)
		return c
	}
	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, &c)
	}
	if c.CachedVMs == nil {
		c.CachedVMs = make(map[string][]azure.VM)
	}
	return c
}

func Save(c Config) error {
	path := getConfigPath()
	if path == "" {
		return nil
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
