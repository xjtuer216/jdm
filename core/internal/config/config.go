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

	jdmHomeKey = "jdm_home"
	jdkHomeKey = "jdk_home"
	mirrorKey  = "mirror"
	proxyKey   = "proxy"
	defaultKey = "default"
	aliasesKey = "aliases"
)

// settableKeys 是可通过 config set 修改的标量配置项
var settableKeys = []string{jdmHomeKey, jdkHomeKey, mirrorKey, proxyKey, defaultKey}

// readableKeys 是 config get/list 可读取的配置项，aliases 是其中唯一的映射项
var readableKeys = append(append([]string{}, settableKeys...), aliasesKey)

// SettableKeys 返回可通过 config set 修改的配置项名称
func SettableKeys() []string {
	return append([]string(nil), settableKeys...)
}

// ReadableKeys 返回 config get/list 可读取的配置项名称
func ReadableKeys() []string {
	return append([]string(nil), readableKeys...)
}

// isLegacyProxyKey 判断配置项名称是否为 proxy 的旧名称
func isLegacyProxyKey(key string) bool {
	return strings.EqualFold(key, LegacyProxyKey)
}

// IsDeprecatedKey 判断配置项名称是否已弃用
func IsDeprecatedKey(key string) bool {
	return isLegacyProxyKey(key)
}

// CanonicalKey 返回配置项的规范名称：旧名称映射为当前名称，已知名称统一为规范拼写
// 与 encoding/json 一致，名称匹配不区分大小写
func CanonicalKey(key string) string {
	if isLegacyProxyKey(key) {
		return proxyKey
	}
	for _, known := range readableKeys {
		if strings.EqualFold(key, known) {
			return known
		}
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

	// 已知名称统一为规范拼写；旧名称仅在缺少 proxy 时生效
	normalized := make(map[string]json.RawMessage, len(fields))
	var legacyRaw json.RawMessage
	for key, raw := range fields {
		if isLegacyProxyKey(key) {
			legacyRaw = raw
			continue
		}
		normalized[CanonicalKey(key)] = raw
	}
	if legacyRaw != nil {
		if _, exists := normalized[proxyKey]; !exists {
			normalized[proxyKey] = legacyRaw
		}
	}

	normalizedData, err := json.Marshal(normalized)
	if err != nil {
		return err
	}
	return json.Unmarshal(normalizedData, c)
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

// scalarField 返回标量配置项对应的字段指针，非标量或未知配置项返回 nil
func (c *Config) scalarField(canonicalKey string) *string {
	switch canonicalKey {
	case jdmHomeKey:
		return &c.JDMHome
	case jdkHomeKey:
		return &c.JDKHome
	case mirrorKey:
		return &c.Mirror
	case proxyKey:
		return &c.Proxy
	case defaultKey:
		return &c.Default
	default:
		return nil
	}
}

// Get 获取配置项的值
func (c *Config) Get(key string) (string, error) {
	canonical := CanonicalKey(key)
	if canonical == aliasesKey {
		return formatAliases(c.Aliases)
	}
	if field := c.scalarField(canonical); field != nil {
		return *field, nil
	}
	return "", nil
}

// Set 设置配置项的值
func (c *Config) Set(key, value string) error {
	canonical := CanonicalKey(key)
	if canonical == aliasesKey {
		return errors.New(`aliases is a map; use "jdm alias set <name> <version>" or "jdm alias del <name>"`)
	}
	field := c.scalarField(canonical)
	if field == nil {
		return fmt.Errorf("unknown config key %q (settable: %s)", key, strings.Join(settableKeys, ", "))
	}
	*field = value
	return c.Save()
}

// formatAliases 把别名映射序列化为 JSON，供 config get/list 展示
func formatAliases(aliases map[string]string) (string, error) {
	data, err := json.Marshal(aliases)
	if err != nil {
		return "", err
	}
	if string(data) == "null" {
		return "{}", nil
	}
	return string(data), nil
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
func (c *Config) GetAll() (map[string]string, error) {
	all := make(map[string]string, len(readableKeys))
	for _, key := range readableKeys {
		value, err := c.Get(key)
		if err != nil {
			return nil, err
		}
		all[key] = value
	}
	return all, nil
}
