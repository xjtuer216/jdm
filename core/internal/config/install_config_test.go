package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// readRawConfig 读取配置文件并保留每个键的原始内容
func readRawConfig(t *testing.T, path string) map[string]json.RawMessage {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to parse %s: %v", path, err)
	}
	return raw
}

// decodeStringConfigValue 断言配置值是可以解码为字符串的原始 JSON
func decodeStringConfigValue(t *testing.T, raw json.RawMessage) string {
	t.Helper()

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatalf("expected a JSON string, got %s: %v", raw, err)
	}
	return value
}

func TestEnsureInstallConfigCreatesFullTemplate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ConfigFileName)

	if err := EnsureInstallConfig(path); err != nil {
		t.Fatalf("EnsureInstallConfig failed: %v", err)
	}

	raw := readRawConfig(t, path)
	want := map[string]string{
		jdmHomeKey:   dir,
		jdkHomeKey:   filepath.Join(dir, "versions"),
		mirrorKey:    DefaultMirror,
		jdkMirrorKey: DefaultJDKMirror,
		proxyKey:     DefaultProxy,
		defaultKey:   "",
	}
	for key, value := range want {
		got, ok := raw[key]
		if !ok {
			t.Fatalf("expected key %q in generated template, got %s", key, raw)
		}
		if decoded := decodeStringConfigValue(t, got); decoded != value {
			t.Errorf("expected %s %q, got %q", key, value, decoded)
		}
	}

	var aliases map[string]string
	if err := json.Unmarshal(raw[aliasesKey], &aliases); err != nil {
		t.Fatalf("expected aliases to be an object, got %s: %v", raw[aliasesKey], err)
	}
	if len(aliases) != 0 {
		t.Errorf("expected empty aliases, got %v", aliases)
	}
}

func TestEnsureInstallConfigPreservesExistingValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ConfigFileName)
	existing := `{
  "jdm_home": "D:\\custom-home",
  "jdk_home": "D:\\custom-jdks",
  "mirror": "https://mirror.example.com",
  "proxy": "none",
  "default": "21",
  "aliases": {"lts": "21"},
  "extra": "keep-me"
}`
	if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	if err := EnsureInstallConfig(path); err != nil {
		t.Fatalf("EnsureInstallConfig failed: %v", err)
	}

	raw := readRawConfig(t, path)
	want := map[string]string{
		jdmHomeKey: `D:\custom-home`,
		jdkHomeKey: `D:\custom-jdks`,
		mirrorKey:  "https://mirror.example.com",
		proxyKey:   "none",
		defaultKey: "21",
		"extra":    "keep-me",
	}
	for key, value := range want {
		got, ok := raw[key]
		if !ok {
			t.Fatalf("expected key %q to be preserved, got %s", key, raw)
		}
		if decoded := decodeStringConfigValue(t, got); decoded != value {
			t.Errorf("expected %s to stay %q, got %q", key, value, decoded)
		}
	}

	var aliases map[string]string
	if err := json.Unmarshal(raw[aliasesKey], &aliases); err != nil {
		t.Fatalf("expected aliases to be preserved: %v", err)
	}
	if aliases["lts"] != "21" {
		t.Errorf("expected alias lts=21, got %v", aliases)
	}

	if _, ok := raw[jdkMirrorKey]; !ok {
		t.Error("expected missing jdk_mirror to be filled in")
	}
}

func TestEnsureInstallConfigTreatsLegacyProxyKeyAsPresent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ConfigFileName)
	existing := `{"download_mirror": "https://legacy.example.com"}`
	if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	if err := EnsureInstallConfig(path); err != nil {
		t.Fatalf("EnsureInstallConfig failed: %v", err)
	}

	raw := readRawConfig(t, path)
	for key := range raw {
		if isLegacyProxyKey(key) {
			continue
		}
		if CanonicalKey(key) == proxyKey {
			t.Errorf("proxy must not be added when the legacy key supplies it, found %q", key)
		}
	}
	if got := decodeStringConfigValue(t, raw[LegacyProxyKey]); got != "https://legacy.example.com" {
		t.Errorf("legacy value must be preserved, got %q", got)
	}
}

func TestEnsureInstallConfigMatchesExistingKeysCaseInsensitively(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ConfigFileName)
	existing := `{"MIRROR": "https://custom.example.com"}`
	if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	if err := EnsureInstallConfig(path); err != nil {
		t.Fatalf("EnsureInstallConfig failed: %v", err)
	}

	raw := readRawConfig(t, path)
	count := 0
	for key, value := range raw {
		if CanonicalKey(key) != mirrorKey {
			continue
		}
		count++
		if got := decodeStringConfigValue(t, value); got != "https://custom.example.com" {
			t.Errorf("expected existing mirror value to be preserved, got %q", got)
		}
	}
	if count != 1 {
		t.Errorf("expected exactly one mirror key, got %d: %s", count, raw)
	}
}

func TestEnsureInstallConfigLeavesCompleteConfigUntouched(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ConfigFileName)
	complete := `{"jdm_home":"D:\\h","jdk_home":"D:\\j","mirror":"https://m","jdk_mirror":"https://n","proxy":"none","default":"","aliases":{}}`
	if err := os.WriteFile(path, []byte(complete), 0644); err != nil {
		t.Fatal(err)
	}

	if err := EnsureInstallConfig(path); err != nil {
		t.Fatalf("EnsureInstallConfig failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != complete {
		t.Errorf("a complete config must not be rewritten, got %s", data)
	}
}

func TestEnsureInstallConfigTreatsNullFileAsEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ConfigFileName)
	if err := os.WriteFile(path, []byte("null"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := EnsureInstallConfig(path); err != nil {
		t.Fatalf("EnsureInstallConfig failed: %v", err)
	}

	raw := readRawConfig(t, path)
	if _, ok := raw[jdkMirrorKey]; !ok {
		t.Errorf("expected the template to be written for a null config, got %s", raw)
	}
}

func TestEnsureInstallConfigRejectsMalformedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ConfigFileName)
	original := []byte(`{"mirror": `)
	if err := os.WriteFile(path, original, 0644); err != nil {
		t.Fatal(err)
	}

	if err := EnsureInstallConfig(path); err == nil {
		t.Fatal("expected an error for a malformed config file")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(original) {
		t.Errorf("malformed config must be left untouched, got %s", data)
	}
}
