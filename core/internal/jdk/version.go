package jdk

import (
	"fmt"
	"runtime/debug"
	"time"
)

// Version 版本号，通过 ldflags 注入
// 使用方式: go build -ldflags "-X github.com/xjtuer216/jdm/internal/jdk.Version=1.0.0"
var Version = "dev"

// BuildTime 构建时间，通过 ldflags 注入
// 使用方式: go build -ldflags "-X github.com/xjtuer216/jdm/internal/jdk.BuildTime=2024-01-01"
var BuildTime = "unknown"

// GetVersion 获取当前版本号
func GetVersion() string {
	return Version
}

// GetBuildTime 获取构建时间
func GetBuildTime() string {
	if BuildTime == "unknown" {
		return time.Now().Format("2006-01-02 15:04:05")
	}
	return BuildTime
}

// PrintVersion 打印版本信息
func PrintVersion() {
	fmt.Printf("jdm v%s\n", Version)
}

// PrintFullVersion 打印完整版本信息
func PrintFullVersion() {
	fmt.Printf("jdm version %s\n", Version)
	fmt.Printf("Build time: %s\n", GetBuildTime())
}

// VersionInfo 版本信息结构
type VersionInfo struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
}

// GetVersionInfo 获取完整的版本信息
func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version:   Version,
		BuildTime: GetBuildTime(),
	}
}

// ReadBuildInfo 读取嵌入的构建信息
func ReadBuildInfo() (string, string) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return Version, BuildTime
	}

	v := Version
	bt := BuildTime

	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			// 可以使用 git commit hash
		case "vcs.time":
			// 可以使用 git commit 时间
		}
	}

	return v, bt
}
