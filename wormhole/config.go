package wormhole

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds user preferences.
// Use LoadConfig to load from disk and SaveConfig to persist.
type Config struct {
	RendezvousURL string `json:"rendezvous_url"`
	CodeLength    int    `json:"code_len"`
}

// LoadConfig loads configuration from the given data directory.
// Returns a default Config if the file doesn't exist.
func LoadConfig(dataDir string) *Config {
	filename := filepath.Join(dataDir, "wormhole-william.json")
	content, err := os.ReadFile(filename)
	if err != nil {
		return &Config{}
	}

	var conf Config
	json.Unmarshal(content, &conf)
	return &conf
}

// SaveConfig saves configuration to the given data directory.
func SaveConfig(dataDir string, conf *Config) error {
	filename := filepath.Join(dataDir, "wormhole-william.json")
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(conf)
}
