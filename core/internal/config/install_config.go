package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// EnsureInstallConfig 初始化安装目录的配置模板，默认值以 path 所在目录为准。
// 文件不存在时写入全部默认项；文件已存在时只补充缺失的键，
// 保留已有值与未知键，旧版的 download_mirror 视为已提供 proxy。
func EnsureInstallConfig(path string) error {
	existing := make(map[string]json.RawMessage)

	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		if err := json.Unmarshal(data, &existing); err != nil {
			return fmt.Errorf("failed to parse install config %s: %w", path, err)
		}
		if existing == nil {
			// 文件内容是 JSON 字面量 null：合法但没有键，按空配置处理
			existing = make(map[string]json.RawMessage)
		}
	case errors.Is(err, os.ErrNotExist):
		// 新装环境：从空配置开始写入完整模板
	default:
		return fmt.Errorf("failed to read install config %s: %w", path, err)
	}

	present := make(map[string]bool, len(existing))
	for key := range existing {
		present[CanonicalKey(key)] = true
	}

	defaultValues, err := installConfigValues(NewConfig(filepath.Dir(path)))
	if err != nil {
		return err
	}
	added := false
	for key, value := range defaultValues {
		if present[key] {
			continue
		}
		existing[key] = value
		added = true
	}
	if !added {
		return nil
	}

	data, err = json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode install config %s: %w", path, err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write install config %s: %w", path, err)
	}
	return nil
}

// installConfigValues 把配置默认值序列化为原始 JSON，供补齐模板使用
func installConfigValues(c *Config) (map[string]json.RawMessage, error) {
	aliases := c.Aliases
	if aliases == nil {
		aliases = map[string]string{}
	}
	values := map[string]any{
		jdmHomeKey:   c.JDMHome,
		jdkHomeKey:   c.JDKHome,
		mirrorKey:    c.Mirror,
		jdkMirrorKey: c.JDKMirror,
		proxyKey:     c.Proxy,
		defaultKey:   c.Default,
		aliasesKey:   aliases,
	}

	raw := make(map[string]json.RawMessage, len(values))
	for key, value := range values {
		data, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("failed to encode default %s: %w", key, err)
		}
		raw[key] = data
	}
	return raw, nil
}
