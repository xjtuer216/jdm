package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	// DefaultMirror 是默认的 Adoptium 镜像地址
	DefaultMirror = "https://api.adoptium.net/v3"
	// ConfigFileName 是配置文件名
	ConfigFileName = "config.json"
)

type Config struct {
	JDMHome string            `json:"jdm_home"`
	JDKHome string            `json:"jdk_home"`
	Mirror  string            `json:"mirror"`
	Default string            `json:"default"`
	Aliases map[string]string `json:"aliases"`
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
		if err := json.Unmarshal(data, c); err != nil {
			return err
		}
		return c.ensureDirectories()
	}

	// 用户配置不存在，尝试加载安装目录配置
	installConfigPath := c.GetInstallConfigPath()
	if installConfigPath != "" {
		data, err := os.ReadFile(installConfigPath)
		if err == nil {
			// 安装配置存在，解析并合并到当前配置
			var installConfig Config
			if err := json.Unmarshal(data, &installConfig); err == nil {
				// 使用安装配置中的值（如果存在）
				if installConfig.Mirror != "" && installConfig.Mirror != DefaultMirror {
					c.Mirror = installConfig.Mirror
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
	return c.ensureDirectories()
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
	switch key {
	case "jdm_home":
		return c.JDMHome
	case "jdk_home":
		return c.JDKHome
	case "mirror":
		return c.Mirror
	case "default":
		return c.Default
	default:
		return ""
	}
}

// Set 设置配置项的值
func (c *Config) Set(key, value string) error {
	switch key {
	case "jdm_home":
		c.JDMHome = value
	case "jdk_home":
		c.JDKHome = value
	case "mirror":
		c.Mirror = value
	case "default":
		c.Default = value
	default:
		return nil
	}
	return c.Save()
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
		"default":  c.Default,
	}
}