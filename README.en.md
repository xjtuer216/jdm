# JDM - JDK Version Manager

<p align="center">
  <img src="https://img.shields.io/badge/version-0.1.0-blue" alt="Version">
  <img src="https://img.shields.io/badge/platform-Windows-green" alt="Platform">
  <img src="https://img.shields.io/badge/language-Go-blue" alt="Language">
  <img src="https://img.shields.io/badge/license-MIT-orange" alt="License">
</p>

JDM (JDK Version Manager) is a JDK version management tool for Windows, similar to nvm for Node.js. It helps you easily install, manage, and switch between multiple JDK versions on Windows.

## Features

- **Version Installation** - One-click installation of any JDK version, supporting all Adoptium (Temurin) releases
- **Quick Switching** - Instantly switch JDK versions using symbolic links, no terminal restart needed
- **Version Listing** - View locally installed and remotely available versions with automatic LTS detection
- **Alias Management** - Set custom aliases for frequently used versions
- **Download Acceleration** - Built-in GitHub download proxy, no VPN needed for users in China
- **Progress Display** - Real-time progress and speed tracking during download and extraction
- **No Admin Required** - Automatic fallback to Directory Junction when symlinks fail, no administrator privileges needed
- **Flexible Configuration** - Custom mirror sources, installation directories, and dual-layer config
- **Graphical Installer** - Inno Setup wizard with bilingual (English/Chinese) interface

## Installation

### Method 1: Using the Installer (Recommended)

1. Download `jdm-setup-*.exe` from [Releases](https://github.com/xjtuer216/jdm/releases)
2. Double-click to run the installer
3. Choose the installation directory and JDK storage location
4. Open a new terminal after installation to start using JDM

### Method 2: Manual Installation

1. Download the latest `jdm.exe` from [Releases](https://github.com/xjtuer216/jdm/releases)
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

# Set a mirror source
jdm config set mirror https://api.adoptium.net/v3

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
| `mirror` | Adoptium API mirror | `https://api.adoptium.net/v3` |
| `download_mirror` | GitHub download proxy | `https://ghproxy.net` |
| `default` | Default JDK version | (not set) |
| `aliases` | Version aliases | `{}` |

### Custom Mirror Sources

Users in China can configure a domestic mirror for faster downloads:

```powershell
jdm config set mirror https://mirrors.aliyun.com/Adoptium
```

Or edit the configuration file directly:

```json
{
  "mirror": "https://mirrors.aliyun.com/Adoptium"
}
```

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

A: Check your network connection and ensure the mirror source is accessible. Users in China can configure a download proxy:
```powershell
jdm config set download_mirror https://ghproxy.net
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
