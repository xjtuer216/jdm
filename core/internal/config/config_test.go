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

func TestConfig_SetLegacyDownloadMirrorKey(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := NewConfig(tmpDir)

	const mirror = "https://mirror.example.com"
	if err := cfg.Set(LegacyProxyKey, mirror); err != nil {
		t.Fatalf("failed to set legacy key: %v", err)
	}

	if got, err := cfg.Get("proxy"); err != nil || got != mirror {
		t.Errorf("expected proxy %q, got %q (err: %v)", mirror, got, err)
	}
	if got, err := cfg.Get("download_mirror"); err != nil || got != mirror {
		t.Errorf("expected legacy key to resolve to %q, got %q (err: %v)", mirror, got, err)
	}

	// 配置文件中必须写入规范名称，而不是旧名称
	data, err := os.ReadFile(filepath.Join(tmpDir, "config.json"))
	if err != nil {
		t.Fatalf("failed to read saved config: %v", err)
	}
	if !strings.Contains(string(data), `"proxy": "`+mirror+`"`) {
		t.Errorf("expected saved config to contain proxy %q, got %s", mirror, data)
	}
	if strings.Contains(string(data), LegacyProxyKey) {
		t.Errorf("saved config must not contain the legacy key, got %s", data)
	}

	// 重新加载后仍然生效
	reloaded := NewConfig(tmpDir)
	if err := reloaded.Load(); err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}
	if got := reloaded.Proxy; got != mirror {
		t.Errorf("expected reloaded proxy %q, got %q", mirror, got)
	}
}

func TestConfig_SetUnknownKey(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := NewConfig(tmpDir)

	err := cfg.Set("mirorr", "https://mirror.example.com")
	if err == nil {
		t.Fatal("expected an error for an unknown config key")
	}
	if !strings.Contains(err.Error(), "settable:") {
		t.Errorf("expected error to list settable keys, got %v", err)
	}
	if strings.Contains(err.Error(), aliasesKey) {
		t.Errorf("unknown-key error must not advertise aliases as settable, got %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "config.json")); !os.IsNotExist(err) {
		t.Errorf("unknown key must not write a config file, stat error: %v", err)
	}
}

func TestConfig_GetAliases(t *testing.T) {
	cfg := NewConfig(t.TempDir())

	if got, err := cfg.Get("aliases"); err != nil || got != "{}" {
		t.Errorf("expected empty aliases as %q, got %q (err: %v)", "{}", got, err)
	}

	cfg.SetAlias("myjdk", "17.0.2+2")
	cfg.SetAlias("lts", "21")

	const want = `{"lts":"21","myjdk":"17.0.2+2"}`
	if got, err := cfg.Get("aliases"); err != nil || got != want {
		t.Errorf("expected aliases %s, got %s (err: %v)", want, got, err)
	}
}

func TestConfig_GetAllIncludesAliases(t *testing.T) {
	cfg := NewConfig(t.TempDir())
	cfg.SetAlias("myjdk", "17.0.2+2")

	all, err := cfg.GetAll()
	if err != nil {
		t.Fatalf("failed to get all config: %v", err)
	}
	const want = `{"myjdk":"17.0.2+2"}`
	if got := all["aliases"]; got != want {
		t.Errorf("expected aliases in config list %s, got %s", want, got)
	}
}

func TestConfig_SetAliasesRejected(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := NewConfig(tmpDir)

	err := cfg.Set("aliases", `{"myjdk":"17.0.2+2"}`)
	if err == nil {
		t.Fatal("expected setting aliases via config set to fail")
	}
	if !strings.Contains(err.Error(), "jdm alias") {
		t.Errorf("expected error to point at jdm alias, got %v", err)
	}

	if _, statErr := os.Stat(filepath.Join(tmpDir, "config.json")); !os.IsNotExist(statErr) {
		t.Errorf("rejected aliases set must not write a config file, stat error: %v", statErr)
	}
}

func TestCanonicalKey(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"mirror", mirrorKey},
		{"MIRROR", mirrorKey},
		{"Proxy", proxyKey},
		{LegacyProxyKey, proxyKey},
		{"Download_Mirror", proxyKey},
		{"aliases", aliasesKey},
		{"mirorr", "mirorr"},
	}

	for _, tt := range tests {
		if got := CanonicalKey(tt.key); got != tt.want {
			t.Errorf("CanonicalKey(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}

	if !IsDeprecatedKey("Download_Mirror") {
		t.Error("expected Download_Mirror to be reported as deprecated")
	}
	if IsDeprecatedKey(proxyKey) {
		t.Error("expected proxy not to be reported as deprecated")
	}
}

func TestConfig_SetKeyCaseInsensitive(t *testing.T) {
	cfg := NewConfig(t.TempDir())

	const proxy = "https://proxy.example.com"
	if err := cfg.Set("PROXY", proxy); err != nil {
		t.Fatalf("failed to set proxy with different case: %v", err)
	}
	if cfg.Proxy != proxy {
		t.Errorf("expected proxy %q, got %q", proxy, cfg.Proxy)
	}

	got, err := cfg.Get("Proxy")
	if err != nil {
		t.Fatalf("failed to get proxy with different case: %v", err)
	}
	if got != proxy {
		t.Errorf("expected proxy %q, got %q", proxy, got)
	}
}

func TestConfig_LoadLegacyDownloadMirror(t *testing.T) {
	tests := []struct {
		name   string
		config string
		want   string
	}{
		{
			name:   "legacy key only",
			config: `{"download_mirror": "https://legacy.example.com"}`,
			want:   "https://legacy.example.com",
		},
		{
			name:   "proxy wins over legacy key",
			config: `{"proxy": "https://proxy.example.com", "download_mirror": "https://legacy.example.com"}`,
			want:   "https://proxy.example.com",
		},
		{
			name:   "explicit empty proxy wins over legacy key",
			config: `{"proxy": "", "download_mirror": "https://legacy.example.com"}`,
			want:   "",
		},
		{
			name:   "proxy name matching is case-insensitive",
			config: `{"Proxy": "https://proxy.example.com", "download_mirror": "https://legacy.example.com"}`,
			want:   "https://proxy.example.com",
		},
		{
			name:   "legacy name matching is case-insensitive",
			config: `{"Download_Mirror": "https://legacy.example.com"}`,
			want:   "https://legacy.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			if err := os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte(tt.config), 0644); err != nil {
				t.Fatal(err)
			}

			cfg := NewConfig(tmpDir)
			if err := cfg.Load(); err != nil {
				t.Fatalf("failed to load config: %v", err)
			}
			if cfg.Proxy != tt.want {
				t.Errorf("expected proxy %q, got %q", tt.want, cfg.Proxy)
			}
		})
	}
}
