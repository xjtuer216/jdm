# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

---

## [0.1.0] - 2026-04-18

### Added

- **核心功能**
  - `jdm ls` - 列出本地已安装的 JDK 版本
  - `jdm ls-remote` - 列出 Adoptium API 远程可用版本，支持表格和列表两种显示模式
  - `jdm install <version>` - 安装指定 JDK 版本，支持短版本解析（如 `17` → `17.0.2+2`）
  - `jdm uninstall <version>` - 卸载指定 JDK 版本，防止卸载当前活跃版本
  - `jdm use <version>` - 临时切换 JDK 版本（当前终端生效）
  - `jdm default <version>` - 设置全局默认 JDK 版本
  - `jdm current` - 查看当前使用的 JDK 版本
  - `jdm version` - 查看 JDM 自身版本

- **别名管理**
  - `jdm alias set <name> <version>` - 设置版本别名
  - `jdm alias list` - 查看所有别名
  - `jdm alias del <name>` - 删除别名

- **配置管理**
  - `jdm config list` - 查看所有配置项
  - `jdm config get <key>` - 获取指定配置值
  - `jdm config set <key> <value>` - 设置配置值
  - `jdm config init` - 初始化用户配置文件
  - 双层配置加载（用户配置 `~/.jdm/config.json` > 安装目录配置）

- **下载与网络**
  - Adoptium API 集成，自动获取可用 JDK 版本
  - `download_mirror` 配置项，支持 GitHub Releases 下载代理（默认 `ghproxy.net`）
  - 自定义 API 镜像源（`mirror` 配置项）
  - 独立 HTTP 客户端，10 分钟超时，避免下载大文件超时

- **进度显示**
  - 下载进度条，实时显示已下载大小、总大小、速度和耗时
  - 解压进度条，实时显示已解压文件数和总文件数

- **符号链接管理**
  - 基于符号链接的 JDK 版本切换（`~/.jdm/current` → 目标 JDK）
  - 符号链接创建失败时自动回退到目录联接（`mklink /J`），无需管理员权限或开发者模式

- **架构检测**
  - 自动检测系统架构（x64 / arm64）
  - 自动匹配对应架构的 JDK 下载包

- **语义化版本解析**
  - 支持 `major.minor.patch+build` 格式解析
  - 支持部分匹配（如 `17` 匹配 `17.0.2+2`）
  - 版本比较和排序

- **安装程序**
  - Inno Setup 6 图形化安装向导
  - 中英文双语界面，自动检测系统语言
  - 用户模式安装（无需管理员权限），安装到 `%LOCALAPPDATA%\JDM`
  - 自定义 JDK 存储位置选择页面
  - 自动配置 PATH 和 JDM_HOME 环境变量
  - 开始菜单快捷方式
  - 完整卸载支持，自动清理环境变量和文件
  - `build.bat` 构建脚本，支持版本号参数和 Inno Setup 自动检测

- **日志系统**
  - 基于 logrus 的日志框架
  - `--log <path>` 参数启用文件日志输出

- **项目文档**
  - 中文 README.md
  - 英文 README.en.md
  - 发布操作手册（RELEASE_OPERATIONS.md）

### Changed

- 项目结构重组，Go 模块移至 `core/` 目录
- 模块路径从 `github.com/whimsy/jdm` 迁移至 `github.com/xjtuer216/jdm`
- `ls-remote` 命令重构，支持动态版本发现和 LTS 标识

### Fixed

- `jdm use [version]` 命令执行报错问题
- 安装程序版本号参数不生效（`/dMyAppVersion` 被 `#define` 覆盖）
- 非管理员模式下环境变量写入失败（HKLM → HKCU 回退）
- 卸载时 PATH 清理不完整
- 卸载时 `WizardSilent` 函数调用导致崩溃

### Technical

- **技术栈**: Go 1.21 + Cobra CLI + logrus
- **依赖**: `github.com/spf13/cobra v1.8.1`, `github.com/sirupsen/logrus v1.9.3`
- **构建**: `go build -ldflags "-X github.com/xjtuer216/jdm/internal/jdk.Version=X.Y.Z"`
- **ZIP 解压防逃逸**（Zip Slip 保护）

[Unreleased]: https://github.com/xjtuer216/jdm/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/xjtuer216/jdm/releases/tag/v0.1.0
