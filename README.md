# JDM - JDK Version Manager

<p align="center">
  <img src="https://img.shields.io/badge/version-0.1.1-blue" alt="Version">
  <img src="https://img.shields.io/badge/platform-Windows-green" alt="Platform">
  <img src="https://img.shields.io/badge/language-Go-blue" alt="Language">
  <img src="https://img.shields.io/badge/license-MIT-orange" alt="License">
</p>

<p align="center">
  <a href="https://gitee.com/xjtuer216/jdm">Gitee</a> · 
  <a href="https://github.com/xjtuer216/jdm">GitHub</a> · 
  <a href="https://gitee.com/xjtuer216/jdm/releases">Releases</a> · 
  <a href="https://gitee.com/xjtuer216/jdm/blob/develop/CHANGELOG.md">CHANGELOG</a>
</p>

JDM (JDK Version Manager) 是一个 Windows 平台下的 JDK 版本管理工具，类似 nvm for Node.js。它可以帮助你在 Windows 上轻松安装、管理和切换多个 JDK 版本。

## 特性

- **版本安装** - 一键安装所需的 JDK 版本，支持 Adoptium (Temurin) 所有发布版本
- **快速切换** - 瞬时切换 JDK 版本，基于符号链接技术，无需重启终端
- **版本列表** - 查看本地已安装和远程可用版本，自动识别 LTS 版本
- **别名管理** - 为常用版本设置自定义别名
- **国内下载源** - 默认使用清华 TUNA 下载源，镜像缺失或不匹配时自动回退 GitHub
- **进度显示** - 下载和解压过程实时显示进度和速度
- **免管理员** - 符号链接失败时自动回退到目录联接（Junction），无需管理员权限
- **配置灵活** - 支持自定义镜像源、安装目录和双层配置
- **图形安装** - 提供 Inno Setup 安装向导，支持中英文界面

## 安装方式

### 方式一：使用安装程序（推荐）

1. 从 [Gitee Releases](https://gitee.com/xjtuer216/jdm/releases) 或 [GitHub Releases](https://github.com/xjtuer216/jdm/releases) 下载 `jdm-setup-*.exe`
2. 双击运行安装程序
3. 选择安装目录（JDK 版本固定存放在其下的 `versions` 子目录）
4. 安装完成后打开新终端即可使用

### 方式二：手动安装

1. 从 [Gitee Releases](https://gitee.com/xjtuer216/jdm/releases) 或 [GitHub Releases](https://github.com/xjtuer216/jdm/releases) 下载最新版本的 `jdm.exe`
2. 将 `jdm.exe` 放到你选择的目录
3. 将该目录添加到系统 PATH 环境变量

## 基本使用

```powershell
# 查看帮助
jdm help

# 查看 JDM 版本
jdm version
# 或
jdm -v

# 查看远程可用版本
jdm ls-remote          # 表格形式显示所有版本
jdm ls-remote 17       # 查看 JDK 17 的所有可用版本

# 安装 JDK
jdm install 17         # 安装 JDK 17 最新版
jdm install 17.0.2+2   # 安装指定构建版本

# 列出本地已安装版本
jdm ls

# 切换 JDK 版本（当前终端生效）
jdm use 17

# 设置默认 JDK 版本（全局生效）
jdm default 17

# 查看当前使用的版本
jdm current

# 卸载 JDK
jdm uninstall 17
```

### 别名管理

```powershell
# 设置别名
jdm alias set myjdk 17.0.2+2

# 查看别名列表
jdm alias list

# 删除别名
jdm alias del myjdk
```

### 配置管理

```powershell
# 查看所有配置
jdm config list

# 获取配置项
jdm config get mirror
jdm config get jdk_mirror

# 设置 JDK 下载源（默认清华 TUNA）
jdm config set jdk_mirror https://mirrors.tuna.tsinghua.edu.cn/Adoptium

# 禁用下载源，完全回到 GitHub
jdm config set jdk_mirror none

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
| `mirror` | 元数据源：Adoptium API 基址 | `https://api.adoptium.net` |
| `jdk_mirror` | 下载源基址；`none` 表示禁用 | `https://mirrors.tuna.tsinghua.edu.cn/Adoptium` |
| ~~`download_mirror`~~ `proxy` | GitHub 下载代理（`download_mirror` 已弃用，仍作为别名接受并提示）；`none` 表示禁用 | `https://ghproxy.net` |
| `default` | 默认 JDK 版本 | (未设置) |
| `aliases` | 版本别名（只读，使用 `jdm alias` 管理） | `{}` |

### 下载源与镜像

JDK 安装涉及两段独立的网络访问，分别由不同配置项控制：

| 阶段 | 配置项 | 作用 |
|------|--------|------|
| 元数据 | `mirror` | Adoptium API 基址，解析精确版本、文件名、字节数与 SHA-256 |
| 下载 | `jdk_mirror` | 按固定目录布局提供 JDK 归档的下载源 |
| 下载 | `proxy` | 只给 github.com 链接加前缀的代理，不参与镜像目录解析 |

安装时的候选链如下，每个来源只尝试一次，失败自动换下一个：

1. `jdk_mirror`，按 `{base}/{major}/jdk/x64/windows/{文件名}` 拼出地址（文件名逐字取自元数据源）
2. `proxy` + 原始 GitHub 链接
3. 原始 GitHub 链接

所有来源都校验元数据给出的字节数与 SHA-256。镜像返回 HTML 错误页、截断文件或摘要不符都会判为失败并换源；全部失败时会列出每个来源的原因。成功时输出实际使用的来源。

国内用户默认就走清华 TUNA，无需额外配置。也可以切换或禁用：

```powershell
# 切换下载源
jdm config set jdk_mirror https://mirrors.tuna.tsinghua.edu.cn/Adoptium

# 禁用下载源，完全回到 GitHub
jdm config set jdk_mirror none

# 也可以直接编辑 ~/.jdm/config.json
```

可用下载源（实测于 2026-10-30）：

| 站点 | 基址 | 覆盖与限制 |
|------|------|-----------|
| 清华 TUNA（默认） | `https://mirrors.tuna.tsinghua.edu.cn/Adoptium` | 覆盖 8/11/17/21/25；每个大版本只保留最新 GA，补丁可能滞后于 Adoptium API |
| 南京大学 | `https://mirror.nju.edu.cn/adoptium` | 路径为小写；同步不定期 |
| 兰州大学 | `https://mirror.lzu.edu.cn/Adoptium` | 同步不定期 |
| CERNET | `https://mirrors.cernet.edu.cn/Adoptium` | 302 跳转到就近站点 |

以下站点**不可用**：华为云对任意路径返回 200 HTML；中科大只有旧目录布局；阿里云、腾讯云等没有 Adoptium 目录。JDM 会判定 HTML/404 并回退 GitHub。

已知限制：

- 镜像站只同步部分版本。API 解析出最新补丁而镜像尚未同步时，会自动回退 GitHub，因此首次安装可能比预期慢。
- 镜像站不实现 Adoptium API，`mirror` 必须保持官方地址；把 `mirror` 指向镜像站会导致版本查询失败。
- 下载源当前固定为 Windows x64 布局，非 Windows 平台与 arm64 不在支持范围。
- 自建镜像只需提供相同的目录布局，支持 http 与自定义路径。

## 环境变量

| 变量 | 说明 |
|------|------|
| `JDM_HOME` | JDM 主目录（覆盖默认配置） |
| `JAVA_HOME` | 当前激活的 JDK 路径（自动设置） |

## Shell 集成

### PowerShell

将以下代码添加到 `$PROFILE` 实现自动设置 JAVA_HOME：

```powershell
# JDM Shell 集成
$env:JAVA_HOME = "$env:USERPROFILE\.jdm\current"
$env:PATH = "$env:USERPROFILE\.jdm\current\bin;$env:PATH"
```

## 指令集

| 命令 | 说明 |
|------|------|
| `jdm ls` | 列出本地已安装版本 |
| `jdm ls-remote [version]` | 列出远程可用版本 |
| `jdm install <version>` | 安装指定版本 JDK |
| `jdm uninstall <version>` | 卸载指定版本 JDK |
| `jdm use <version>` | 临时切换版本（当前终端） |
| `jdm default <version>` | 设置默认版本（全局生效） |
| `jdm local <version>` | 项目级版本锁定（预留） |
| `jdm alias list` | 查看别名列表 |
| `jdm alias set <name> <version>` | 设置别名 |
| `jdm alias del <name>` | 删除别名 |
| `jdm config list` | 查看所有配置 |
| `jdm config get <key>` | 获取配置值 |
| `jdm config set <key> <value>` | 设置配置值 |
| `jdm config init` | 初始化配置文件 |
| `jdm current` | 查看当前版本 |
| `jdm version` | 查看 JDM 版本 |
| `jdm help` | 查看帮助 |

## 常见问题

### Q: Windows 上创建符号链接失败

A: JDM 已内置回退机制。当符号链接创建失败时，会自动使用目录联接（Directory Junction）方式，无需开启开发者模式或管理员权限。

### Q: 安装 JDK 失败

A: 检查网络连接，确保可以访问元数据源与下载源。国内用户默认使用清华 TUNA 下载源，也可显式切换或禁用：
```powershell
jdm config set jdk_mirror https://mirrors.tuna.tsinghua.edu.cn/Adoptium
jdm config set jdk_mirror none   # 完全回到 GitHub
jdm config set proxy https://ghproxy.net
```

### Q: 切换版本后不生效

A: 确保 `~/.jdm/current/bin` 在 PATH 的最前端。如果使用安装程序安装，PATH 已自动配置。

### Q: 如何完全卸载 JDM

A: 通过 Windows「设置」→「应用」找到 JDM 进行卸载，或运行安装目录下的 `unins000.exe`。卸载将清理所有环境变量和配置文件。

## 开发指南

### 环境要求

- Go 1.21+
- Windows 10/11

### 构建

```bash
# 克隆项目
git clone https://github.com/xjtuer216/jdm.git
cd jdm/core

# 开发构建
go build -o jdm.exe .

# 带版本号构建
go build -ldflags "-X github.com/xjtuer216/jdm/internal/jdk.Version=1.0.0" -o jdm.exe .
```

### 测试

```bash
# 全部检查（gofmt + go vet + go test）
./scripts/check.sh

# 启用仓库自带的 pre-push 检查
./scripts/setup-hooks.sh
```

也可以单独运行：

```bash
cd core && go test ./...
```

### 构建安装程序

```powershell
# 需要安装 Inno Setup 6
.\installer\build.bat 1.0.0
```

## 项目结构

```
jdm/
├── core/                    # Go 源代码
│   ├── main.go              # 入口 → cmd.Execute()
│   ├── go.mod               # Go 模块定义
│   ├── config.json          # 默认配置模板
│   ├── cmd/                 # CLI 子命令
│   └── internal/            # 内部包
│       ├── arch/            # 架构检测 (x64/arm64)
│       ├── config/          # 配置管理
│       ├── file/            # 文件操作（符号链接/联接）
│       ├── jdk/             # JDK 版本管理
│       ├── log/             # 日志
│       ├── progress/        # 进度条
│       ├── semver/          # 语义化版本解析
│       └── web/             # Adoptium API 客户端
├── installer/               # Inno Setup 安装程序
│   ├── jdm.iss              # 安装脚本
│   ├── build.bat            # 构建脚本
│   └── Languages/           # 多语言文件
└── docs/                    # 文档
```

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
