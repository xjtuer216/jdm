# JDM - JDK Version Manager

<p align="center">
  <img src="https://img.shields.io/badge/version-0.1.0-blue" alt="Version">
  <img src="https://img.shields.io/badge/platform-Windows-green" alt="Platform">
  <img src="https://img.shields.io/badge/language-Go-blue" alt="Language">
</p>

JDM (JDK Version Manager) 是一个 Windows 平台下的 JDK 版本管理工具，类似 nvm for Node.js。它可以帮助你在 Windows 上轻松安装、管理和切换多个 JDK 版本。

## 特性

- 🚀 **版本安装** - 一键安装所需的 JDK 版本
- 🔄 **快速切换** - 瞬时切换 JDK 版本，无需重启终端
- 📋 **版本列表** - 查看本地已安装和远程可用版本
- ⭐ **别名管理** - 为常用版本设置别名
- ⚙️ **配置灵活** - 支持自定义镜像源和安装目录
- 🔧 **预留接口** - 支持项目级版本锁定 (.jdmrc)

## 支持的 JDK 版本

- JDK 8, 11, 17, 21, 25+
- 支持所有 Adoptium (Temurin) 发布版本
- 自动识别 LTS 版本

## 快速开始

### 安装

1. 从 [Releases](https://github.com/whimsy/jdm/releases) 下载最新版本的 `jdm.exe`
2. 将 `jdm.exe` 放到你选择的目录
3. (可选) 将该目录添加到系统 PATH 环境变量

### 基本使用

```powershell
# 查看帮助
jdm help

# 查看版本
jdm version
# 或
jdm -v

# 查看远程可用版本
jdm ls-remote 17

# 安装 JDK
jdm install 17

# 列出本地已安装版本
jdm ls

# 切换 JDK 版本（当前终端）
jdm use 17

# 设置默认 JDK 版本
jdm default 17

# 查看当前使用的版本
jdm current

# 卸载 JDK
jdm uninstall 17
```

### 别名管理

```powershell
# 设置别名
jdm alias set myjdk 17.0.2

# 查看别名列表
jdm alias list

# 删除别名
jdm alias del myjdk
```

### 配置管理

```powershell
# 查看所有配置
jdm config list

# 查看配置项
jdm config get mirror

# 设置镜像源
jdm config set mirror https://api.adoptium.net/v3

# 初始化配置文件
jdm config init
```

## 配置说明

### 配置文件位置

| 位置 | 说明 |
|------|------|
| `~/.jdm/config.json` | 用户配置（优先级高） |
| `<exe-dir>/config.json` | 安装目录配置（作为模板） |

### 配置项说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `jdm_home` | JDM 主目录 | `~/.jdm` |
| `jdk_home` | JDK 安装目录 | `~/.jdm/versions` |
| `mirror` | 下载镜像源 | `https://api.adoptium.net/v3` |
| `default` | 默认 JDK 版本 | (未设置) |
| `aliases` | 版本别名 | `{}` |

### 自定义镜像源

国内用户可以配置国内镜像加速下载：

```powershell
jdm config set mirror https://mirrors.aliyun.com/ Adoptium
```

或者直接编辑配置文件：

```json
{
  "mirror": "https://mirrors.aliyun.com/Adoptium"
}
```

## 环境变量

| 变量 | 说明 |
|------|------|
| `JDM_HOME` | JDM 主目录（覆盖默认配置） |
| `JAVA_HOME` | 当前激活的 JDK 路径 |

## Shell 集成

### PowerShell

将以下代码添加到 `$PROFILE` 实现自动设置 JAVA_HOME：

```powershell
# JDM Shell 集成
$env:JAVA_HOME = "$env:USERPROFILE\.jdm\current"
$env:PATH = "$env:USERPROFILE\.jdm\current\bin;$env:PATH"
```

## 开发指南

### 环境要求

- Go 1.21+
- Windows 10/11

### 构建

```bash
# 克隆项目
git clone https://github.com/whimsy/jdm.git
cd jdm/jdm

# 开发构建
go build -o jdm.exe .

# 带版本号构建
go build -ldflags "-X github.com/whimsy/jdm/internal/jdk.Version=1.0.0" -o jdm.exe .
```

### 测试

```bash
go test ./...
```

## 指令集

| 命令 | 说明 |
|------|------|
| `jdm ls` | 列出本地已安装版本 |
| `jdm ls-remote [version]` | 列出远程可用版本 |
| `jdm install <version>` | 安装指定版本 JDK |
| `jdm uninstall <version>` | 卸载指定版本 JDK |
| `jdm use <version>` | 临时切换版本 |
| `jdm default <version>` | 设置默认版本 |
| `jdm local <version>` | 项目级版本锁定（预留） |
| `jdm alias list` | 查看别名列表 |
| `jdm alias set <name> <version>` | 设置别名 |
| `jdm alias del <name>` | 删除别名 |
| `jdm config list` | 查看所有配置 |
| `jdm config get <key>` | 获取配置值 |
| `jdm config set <key> <value>` | 设置配置值 |
| `jdm current` | 查看当前版本 |
| `jdm version` | 查看 jdm 版本 |
| `jdm help` | 查看帮助 |

## 常见问题

### Q: Windows 上创建符号链接失败

A: 需要开启 Windows 开发者模式，或以管理员身份运行。

### Q: 安装 JDK 失败

A: 检查网络连接，确保可以访问镜像源。尝试更换镜像：
```powershell
jdm config set mirror https://mirrors.aliyun.com/Adoptium
```

### Q: 切换版本后不生效

A: 确保 `~/.jdm/current/bin` 在 PATH 的最前端。

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！