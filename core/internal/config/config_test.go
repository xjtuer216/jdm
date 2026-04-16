package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfig_Load(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "jdm-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")

	// Use forward slashes for JSON paths (they work on all platforms)
	tmpDirSlash := strings.ReplaceAll(tmpDir, "\\", "/")
	versionsDir := strings.ReplaceAll(filepath.Join(tmpDir, "versions"), "\\", "/")

	// Write test config with paths based on tmpDir
	testConfig := `{
		"jdm_home": "` + tmpDirSlash + `",
		"jdk_home": "` + versionsDir + `",
		"mirror": "https://api.adoptium.net/v3",
		"default": "",
		"aliases": {}
	}`
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatal(err)
	}

	// Test loading
	cfg := NewConfig(tmpDir)
	if err := cfg.Load(); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.JDMHome != tmpDirSlash {
		t.Errorf("expected jdm_home %s, got %s", tmpDirSlash, cfg.JDMHome)
	}

	if cfg.Mirror != "https://api.adoptium.net/v3" {
		t.Errorf("expected mirror https://api.adoptium.net/v3, got %s", cfg.Mirror)
	}
}

func TestConfig_Save(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jdm-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := NewConfig(tmpDir)
	cfg.Mirror = "https://api.adoptium.net/v3"
	cfg.Default = ""
	cfg.Aliases = make(map[string]string)

	if err := cfg.Save(); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify file exists
	configPath := filepath.Join(tmpDir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}
}