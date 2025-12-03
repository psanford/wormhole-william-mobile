package wormhole

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewClient(t *testing.T) {
	dataDir := t.TempDir()
	client := NewClient(dataDir)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.dataDir != dataDir {
		t.Errorf("dataDir = %q, want %q", client.dataDir, dataDir)
	}
}

func TestClientConfiguration(t *testing.T) {
	client := NewClient(t.TempDir())

	// Test SetRendezvousURL
	url := "wss://example.com/relay"
	client.SetRendezvousURL(url)
	if got := client.GetRendezvousURL(); got != url {
		t.Errorf("GetRendezvousURL() = %q, want %q", got, url)
	}

	// Test SetCodeLength
	client.SetCodeLength(3)
	if got := client.GetCodeLength(); got != 3 {
		t.Errorf("GetCodeLength() = %d, want 3", got)
	}
}

func TestLoadConfigMissing(t *testing.T) {
	dataDir := t.TempDir()
	conf := LoadConfig(dataDir)

	if conf == nil {
		t.Fatal("LoadConfig returned nil")
	}

	// Should return defaults
	if conf.RendezvousURL != "" {
		t.Errorf("RendezvousURL = %q, want empty", conf.RendezvousURL)
	}
	if conf.CodeLength != 0 {
		t.Errorf("CodeLength = %d, want 0", conf.CodeLength)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	dataDir := t.TempDir()

	conf := &Config{
		RendezvousURL: "wss://test.example.com",
		CodeLength:    4,
	}

	err := SaveConfig(dataDir, conf)
	if err != nil {
		t.Fatalf("SaveConfig error: %v", err)
	}

	// Verify file exists
	filename := filepath.Join(dataDir, "wormhole-william.json")
	if _, err := os.Stat(filename); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	// Load it back
	loaded := LoadConfig(dataDir)
	if loaded.RendezvousURL != conf.RendezvousURL {
		t.Errorf("RendezvousURL = %q, want %q", loaded.RendezvousURL, conf.RendezvousURL)
	}
	if loaded.CodeLength != conf.CodeLength {
		t.Errorf("CodeLength = %d, want %d", loaded.CodeLength, conf.CodeLength)
	}
}

func TestCancelNoOp(t *testing.T) {
	client := NewClient(t.TempDir())

	// Cancel should not panic when no transfer is in progress
	client.Cancel()
}
