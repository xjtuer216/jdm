package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// DefaultMirror 是默认的 Adoptium 镜像地址
	DefaultMirror = "https://api.adoptium.net"
	// DefaultProxy 是默认的 JDK 下载代理地址
	// 使用 ghproxy.net 代理 GitHub Releases 下载
	DefaultProxy = "https://ghproxy.net"
	// ConfigFileName 是配置文件名
	ConfigFileName = "config.json"
	// LegacyProxyKey 是 proxy 的旧名称（v0.1.1 前），仍作为别名接受
	LegacyProxyKey = "download_mirror"
)

// supportedKeys 是 config get/set 可识别的配置项
var supportedKeys = []string{"jdm_home", "jdk_home", "mirror", "proxy", "default", "aliases"}

// SupportedKeys 返回 config get/set 支持的配置项名称
func SupportedKeys() []string {
	return append([]string(nil), supportedKeys...)
}

// CanonicalKey 把旧配置项名称映射为当前名称，其他名称原样返回
// 与 encoding/json 一致，名称匹配不区分大小写
func CanonicalKey(key string) string {
	if strings.EqualFold(key, LegacyProxyKey) {
		return "proxy"
	}
	return key
}

type Config struct {
	JDMHome        string            `json:"jdm_home"`
	JDKHome        string            `json:"jdk_home"`
	Mirror         string            `json:"mirror"`
	Proxy          string            `json:"proxy"`
	Default        string            `json:"default"`
	Aliases        map[string]string `json:"aliases"`
}

// 配置来源标识
const (
	// SourceUser 表示用户配置文件（~/.jdm/config.json）
	SourceUser = "user"
	// SourceInstall 表示安装目录配置文件（exe 同目录/config.json）
	SourceInstall = "install"
)

// ConfigError 表示配置文件解析失败
// 携带出错文件的来源与路径，便于用户定位并修复
type ConfigError struct {
	// Source 是配置来源：SourceUser 或 SourceInstall
	Source string
	// Path 是解析失败的配置文件路径
	Path string
	// Err 是底层的解析错误
	Err error
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("failed to parse %s config %s: %v", e.Source, e.Path, e.Err)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// Hint 返回针对该来源的修复建议
func (e *ConfigError) Hint() string {
	if e.Source == SourceInstall {
		return fmt.Sprintf("the install config is only a template; fix or delete %s, then run \"jdm config init\" to generate your user config", e.Path)
	}
	return fmt.Sprintf("fix or delete %s, then run \"jdm config init\" to regenerate it", e.Path)
}

// GetConfigPath 获取用户配置文件的路径
func (c *Config) GetConfigPath() string {
	return filepath.Join(c.JDMHome, ConfigFileName)
}

// GetInstallConfigPath 获取安装目录配置文件的路径
func (c *Config) GetInstallConfigPath() string {
	// 获取 exe 所在的目录
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	exeDir := filepath.Dir(exePath)
	return filepath.Join(exeDir, ConfigFileName)
}

// NewConfig 创建一个新的 Config 实例
func NewConfig(jdmHome string) *Config {
	jdkHome := filepath.Join(jdmHome, "versions")
	if jdkHome == jdmHome {
		jdkHome = jdmHome + "versions"
	}
	return &Config{
		JDMHome: jdmHome,
		JDKHome: jdkHome,
		Mirror:  DefaultMirror,
		Proxy:   DefaultProxy,
		Default: "",
		Aliases: make(map[string]string),
	}
}

// Load 加载配置
// 加载顺序：
// 1. 首先尝试加载用户目录的配置 (~/.jdm/config.json)
// 2. 如果用户配置不存在，尝试加载安装目录的配置 (exe同目录/config.json)
// 3. 如果都不存在，使用默认配置
func (c *Config) Load() error {
	configPath := c.GetConfigPath()

	// 尝试加载用户配置
	data, err := os.ReadFile(configPath)
	if err == nil {
		// 用户配置存在，解析并返回
		if err := c.unmarshalConfig(data); err != nil {
			return &ConfigError{Source: SourceUser, Path: configPath, Err: err}
		}
		return c.ensureDirectories()
	}

	// 用户配置不存在，尝试加载安装目录配置
	var loadErr error
	installConfigPath := c.GetInstallConfigPath()
	if installConfigPath != "" {
		data, err := os.ReadFile(installConfigPath)
		if err == nil {
			// 安装配置存在，解析并合并到当前配置
			var installConfig Config
			if err := installConfig.unmarshalConfig(data); err != nil {
				// 安装配置损坏时必须报错，否则会静默退回默认配置
				loadErr = &ConfigError{Source: SourceInstall, Path: installConfigPath, Err: err}
			} else {
				// 使用安装配置中的值（如果存在）
				if installConfig.Mirror != "" && installConfig.Mirror != DefaultMirror {
					c.Mirror = installConfig.Mirror
				}
				if installConfig.Proxy != "" {
					c.Proxy = installConfig.Proxy
				}
				if installConfig.JDMHome != "" {
					c.JDMHome = installConfig.JDMHome
				}
				if installConfig.JDKHome != "" {
					c.JDKHome = installConfig.JDKHome
				}
			}
		}
	}

	// 确保目录存在
	if err := c.ensureDirectories(); err != nil {
		return err
	}
	return loadErr
}

// unmarshalConfig 解析配置内容，并把旧版配置项名称规范化为当前名称
// 同时存在旧名称与当前名称时以当前名称为准
func (c *Config) unmarshalConfig(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	// 旧名称统一交给 CanonicalKey 映射，避免别名规则出现两份
	var legacyRaw json.RawMessage
	hasProxy := false
	for key, raw := range fields {
		if canonical := CanonicalKey(key); canonical != key {
			legacyRaw = raw
			delete(fields, key)
			continue
		}
		if strings.EqualFold(key, "proxy") {
			hasProxy = true
		}
	}
	if legacyRaw != nil && !hasProxy {
		fields["proxy"] = legacyRaw
	}

	normalized, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	return json.Unmarshal(normalized, c)
}

// Save 保存配置到用户目录
func (c *Config) Save() error {
	if err := c.ensureDirectories(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.GetConfigPath(), data, 0644)
}

// SaveToInstallDir 保存配置到安装目录（用于初始化安装目录的配置模板）
func (c *Config) SaveToInstallDir() error {
	installConfigPath := c.GetInstallConfigPath()
	if installConfigPath == "" {
		return nil
	}

	// 只保存必要的配置项到安装目录
	installConfig := map[string]string{
		"jdm_home": c.JDMHome,
		"jdk_home": c.JDKHome,
		"mirror":   c.Mirror,
		"proxy":    c.Proxy,
	}

	data, err := json.MarshalIndent(installConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(installConfigPath, data, 0644)
}

func (c *Config) ensureDirectories() error {
	dirs := []string{c.JDMHome, c.JDKHome}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

// Get 获取配置项的值
func (c *Config) Get(key string) string {
	switch CanonicalKey(key) {
	case "jdm_home":
		return c.JDMHome
	case "jdk_home":
		return c.JDKHome
	case "mirror":
		return c.Mirror
	case "proxy":
		return c.Proxy
	case "default":
		return c.Default
	case "aliases":
		return formatAliases(c.Aliases)
	default:
		return ""
	}
}

// Set 设置配置项的值
func (c *Config) Set(key, value string) error {
	switch CanonicalKey(key) {
	case "jdm_home":
		c.JDMHome = value
	case "jdk_home":
		c.JDKHome = value
	case "mirror":
		c.Mirror = value
	case "proxy":
		c.Proxy = value
	case "default":
		c.Default = value
	case "aliases":
		return errors.New(`aliases is a map; use "jdm alias set <name> <version>" or "jdm alias del <name>"`)
	default:
		return fmt.Errorf("unknown config key %q (supported: %s)", key, strings.Join(supportedKeys, ", "))
	}
	return c.Save()
}

// formatAliases 把别名映射序列化为 JSON，供 config get/list 展示
func formatAliases(aliases map[string]string) string {
	data, err := json.Marshal(aliases)
	if err != nil || string(data) == "null" {
		return "{}"
	}
	return string(data)
}

// SetAlias 设置版本别名
func (c *Config) SetAlias(name, version string) {
	if c.Aliases == nil {
		c.Aliases = make(map[string]string)
	}
	c.Aliases[name] = version
}

// RemoveAlias 移除版本别名
func (c *Config) RemoveAlias(name string) {
	if c.Aliases != nil {
		delete(c.Aliases, name)
	}
}

// ResolveVersion 解析版本号（支持别名）
func (c *Config) ResolveVersion(version string) string {
	if c.Aliases != nil {
		if v, ok := c.Aliases[version]; ok {
			return v
		}
	}
	return version
}

// GetAll 获取所有配置项（用于显示）
func (c *Config) GetAll() map[string]string {
	return map[string]string{
		"jdm_home": c.JDMHome,
		"jdk_home": c.JDKHome,
		"mirror":   c.Mirror,
		"proxy":    c.Proxy,
		"default":  c.Default,
		"aliases":  formatAliases(c.Aliases),
	}
}
