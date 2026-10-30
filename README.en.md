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
  <a href="https://github.com/xjtuer216/jdm/releases">Releases</a> · 
  <a href="https://github.com/xjtuer216/jdm/blob/develop/CHANGELOG.md">CHANGELOG</a>
</p>

JDM (JDK Version Manager) is a JDK version management tool for Windows, similar to nvm for Node.js. It helps you easily install, manage, and switch between multiple JDK versions on Windows.

## Features

- **Version Installation** - One-click installation of any JDK version, supporting all Adoptium (Temurin) releases
- **Quick Switching** - Instantly switch JDK versions using symbolic links, no terminal restart needed
- **Version Listing** - View locally installed and remotely available versions with automatic LTS detection
- **Alias Management** - Set custom aliases for frequently used versions
- **Domestic Download Source** - Uses the Tsinghua TUNA download source by default, falls back to GitHub when a file is missing or fails verification
- **Progress Display** - Real-time progress and speed tracking during download and extraction
- **No Admin Required** - Automatic fallback to Directory Junction when symlinks fail, no administrator privileges needed
- **Flexible Configuration** - Custom mirror sources, installation directories, and dual-layer config
- **Graphical Installer** - Inno Setup wizard with bilingual (English/Chinese) interface

## Installation

### Method 1: Using the Installer (Recommended)

1. Download `jdm-setup-*.exe` from [Gitee Releases](https://gitee.com/xjtuer216/jdm/releases) or [GitHub Releases](https://github.com/xjtuer216/jdm/releases)
2. Double-click to run the installer
3. Choose the installation directory (JDK versions are stored in its `versions` subdirectory)
4. Open a new terminal after installation to start using JDM

### Method 2: Manual Installation

1. Download the latest `jdm.exe` from [Gitee Releases](https://gitee.com/xjtuer216/jdm/releases) or [GitHub Releases](https://github.com/xjtuer216/jdm/releases)
2. Place `jdm.exe` in your preferred directory
3. Add that directory to your system PATH environment variable

## Basic Usage

```powershell
# View help
jdm help

# View JDM version
jdm version
# or
jdm -v

# View available remote versions
jdm ls-remote          # Display all versions in table format
jdm ls-remote 17       # View all available versions for JDK 17

# Install JDK
jdm install 17         # Install the latest JDK 17
jdm install 17.0.2+2   # Install a specific build version

# List locally installed versions
jdm ls

# Switch JDK version (current terminal only)
jdm use 17

# Set default JDK version (globally effective)
jdm default 17

# View currently active version
jdm current

# Uninstall JDK
jdm uninstall 17
```

### Alias Management

```powershell
# Set an alias
jdm alias set myjdk 17.0.2+2

# List all aliases
jdm alias list

# Delete an alias
jdm alias del myjdk
```

### Configuration Management

```powershell
# View all configuration
jdm config list

# Get a configuration value
jdm config get mirror
jdm config get jdk_mirror

# Set the JDK download source (Tsinghua TUNA by default)
jdm config set jdk_mirror https://mirrors.tuna.tsinghua.edu.cn/Adoptium

# Disable the download source and use GitHub directly
jdm config set jdk_mirror none

# Initialize configuration file
jdm config init
```

## Configuration

### Configuration File Locations

| Location | Description |
|----------|-------------|
| `~/.jdm/config.json` | User configuration (higher priority) |
| `<exe-dir>/config.json` | Installation directory configuration (used as template) |

### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `jdm_home` | JDM main directory | `~/.jdm` |
| `jdk_home` | JDK installation directory | `~/.jdm/versions` |
| `mirror` | Metadata source: Adoptium API base URL | `https://api.adoptium.net` |
| `jdk_mirror` | Download source base URL; `none` disables it | `https://mirrors.tuna.tsinghua.edu.cn/Adoptium` |
| ~~`download_mirror`~~ → `proxy` | GitHub download proxy (`download_mirror` is deprecated but still accepted as an alias); `none` disables it | `https://ghproxy.net` |
| `default` | Default JDK version | (not set) |
| `aliases` | Version aliases (read-only; manage with `jdm alias`) | `{}` |

### Download Source and Mirrors

A JDK installation makes two independent kinds of network requests, controlled by different options:

| Stage | Option | Purpose |
|-------|--------|---------|
| Metadata | `mirror` | Adoptium API base URL: resolves exact versions, file names, sizes, and SHA-256 |
| Download | `jdk_mirror` | Download source that serves JDK archives under a fixed directory layout |
| Download | `proxy` | Prefixes github.com URLs only; it does not understand mirror directory layouts |

The download candidate chain, tried in order (one attempt per source):

1. `jdk_mirror`, built as `{base}/{major}/jdk/x64/windows/{file name}` (the file name is taken verbatim from the metadata source)
2. `proxy` + the original GitHub URL
3. The original GitHub URL

Every source is verified against the byte size and SHA-256 from the metadata source. HTML error pages, truncated files, and checksum mismatches all count as failures, and the next source is tried; if everything fails, the reason for each source is reported. On success, JDM prints the source that was actually used.

Users in China get the Tsinghua TUNA source by default with no configuration. You can switch or disable it:

```powershell
# Switch the download source
jdm config set jdk_mirror https://mirrors.tuna.tsinghua.edu.cn/Adoptium

# Disable it and use GitHub directly
jdm config set jdk_mirror none

# Or edit ~/.jdm/config.json directly
```

Available download sources (tested on 2026-10-30):

| Site | Base URL | Coverage and limitations |
|------|----------|--------------------------|
| Tsinghua TUNA (default) | `https://mirrors.tuna.tsinghua.edu.cn/Adoptium` | Covers 8/11/17/21/25; keeps only the latest GA per major, so patches may lag behind the Adoptium API |
| Nanjing University | `https://mirror.nju.edu.cn/adoptium` | Lowercase path; syncs irregularly |
| Lanzhou University | `https://mirror.lzu.edu.cn/Adoptium` | Syncs irregularly |
| CERNET | `https://mirrors.cernet.edu.cn/Adoptium` | Redirects (302) to a nearby site |

These sites are **not usable**: Huawei Cloud returns HTTP 200 with HTML for any path; USTC only keeps the old layout; Aliyun, Tencent Cloud, and others have no Adoptium directory. JDM detects HTML/404 responses and falls back to GitHub.

Known limitations:

- Mirrors only sync a subset of releases. When the API resolves a newer patch that the mirror has not synced yet, JDM falls back to GitHub, so the first install can be slower than expected.
- Mirrors do not implement the Adoptium API. Keep `mirror` pointed at the official API; pointing it at a mirror breaks version resolution.
- The download layout is fixed to Windows x64. Non-Windows platforms and arm64 are out of scope.
- A self-hosted mirror only needs to expose the same directory layout; both http and custom paths are supported.

## Environment Variables

| Variable | Description |
|----------|-------------|
| `JDM_HOME` | JDM main directory (overrides default configuration) |
| `JAVA_HOME` | Currently active JDK path (set automatically) |

## Shell Integration

### PowerShell

Add the following to your `$PROFILE` for automatic JAVA_HOME setup:

```powershell
# JDM Shell Integration
$env:JAVA_HOME = "$env:USERPROFILE\.jdm\current"
$env:PATH = "$env:USERPROFILE\.jdm\current\bin;$env:PATH"
```

## Command Reference

| Command | Description |
|---------|-------------|
| `jdm ls` | List locally installed versions |
| `jdm ls-remote [version]` | List remotely available versions |
| `jdm install <version>` | Install a specified JDK version |
| `jdm uninstall <version>` | Uninstall a specified JDK version |
| `jdm use <version>` | Temporarily switch version (current terminal) |
| `jdm default <version>` | Set default version (globally effective) |
| `jdm local <version>` | Project-level version locking (reserved) |
| `jdm alias list` | List all aliases |
| `jdm alias set <name> <version>` | Set an alias |
| `jdm alias del <name>` | Delete an alias |
| `jdm config list` | View all configuration |
| `jdm config get <key>` | Get a configuration value |
| `jdm config set <key> <value>` | Set a configuration value |
| `jdm config init` | Initialize configuration file |
| `jdm current` | View current version |
| `jdm version` | View JDM version |
| `jdm help` | View help |

## FAQ

### Q: Symlink creation fails on Windows

A: JDM has a built-in fallback mechanism. When symlink creation fails, it automatically uses Directory Junction instead, requiring neither Developer Mode nor administrator privileges.

### Q: JDK installation fails

A: Check your network connection and make sure both the metadata source and the download source are reachable. Users in China get the Tsinghua TUNA source by default; you can also switch or disable it explicitly:
```powershell
jdm config set jdk_mirror https://mirrors.tuna.tsinghua.edu.cn/Adoptium
jdm config set jdk_mirror none   # use GitHub directly
jdm config set proxy https://ghproxy.net
```

### Q: Version switch doesn't take effect

A: Ensure `~/.jdm/current/bin` is at the front of your PATH. If you used the installer, PATH is configured automatically.

### Q: How do I completely uninstall JDM?

A: Uninstall JDM via Windows Settings → Apps, or run `unins000.exe` in the installation directory. Uninstallation cleans up all environment variables and configuration files.

## Development Guide

### Requirements

- Go 1.21+
- Windows 10/11

### Building

```bash
# Clone the repository
git clone https://github.com/xjtuer216/jdm.git
cd jdm/core

# Development build
go build -o jdm.exe .

# Build with version info
go build -ldflags "-X github.com/xjtuer216/jdm/internal/jdk.Version=1.0.0" -o jdm.exe .
```

### Testing

```bash
cd core && go test ./...
```

### Building the Installer

```powershell
# Requires Inno Setup 6 to be installed
.\installer\build.bat 1.0.0
```

## Project Structure

```
jdm/
├── core/                    # Go source code
│   ├── main.go              # Entry point → cmd.Execute()
│   ├── go.mod               # Go module definition
│   ├── config.json          # Default configuration template
│   ├── cmd/                 # CLI subcommands
│   └── internal/            # Internal packages
│       ├── arch/            # Architecture detection (x64/arm64)
│       ├── config/          # Configuration management
│       ├── file/            # File operations (symlink/junction)
│       ├── jdk/             # JDK version management
│       ├── log/             # Logging
│       ├── progress/        # Progress bars
│       ├── semver/          # Semantic version parsing
│       └── web/             # Adoptium API client
├── installer/               # Inno Setup installer
│   ├── jdm.iss              # Installer script
│   ├── build.bat            # Build script
│   └── Languages/           # Language files
└── docs/                    # Documentation
```

## License

MIT License

## Contributing

Issues and Pull Requests are welcome!
