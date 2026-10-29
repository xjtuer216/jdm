# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

JDM (JDK Version Manager) is a Windows CLI tool for managing multiple JDK versions, similar to nvm for Node.js. It allows installing, listing, and switching between different JDK versions on Windows.

## Tech Stack

- **Language**: Go 1.21+
- **CLI Framework**: Cobra
- **Logging**: logrus
- **HTTP**: Standard library `net/http` (via Adoptium API)
- **Storage**: JSON files

## Common Commands

```bash
# Build
cd core
go build -o jdm.exe .

# Build with version
go build -ldflags "-X github.com/whimsy/jdm/internal/version.Version=1.0.0" -o jdm.exe .

# Run tests
go test ./...

# Run a single test
go test -v ./internal/config/...
```

## Architecture

```
CLI Entry (main.go)
        │
  Command Layer (Cobra commands in core/cmd/)
        │
        ├─ Config Manager (internal/config/)
        ├─ Version Manager (internal/version/manager.go)
        ├─ Adoptium API Client (internal/api/adoptium.go)
        └─ Utils (internal/utils/: symlink, extract)
```

## Key Commands

| Command | Description |
|---------|-------------|
| `jdm ls` | List locally installed JDK versions |
| `jdm ls-remote [version]` | List remote available versions |
| `jdm install <version>` | Install JDK (e.g., `jdm install 17`) |
| `jdm uninstall <version>` | Uninstall JDK |
| `jdm use <version>` | Switch JDK in current terminal |
| `jdm default <version>` | Set global default JDK |
| `jdm local <version>` | Project-level version lock (creates .jdmrc) |
| `jdm alias list/set/del` | Manage version aliases |
| `jdm config list/get/set` | Manage configuration |
| `jdm current` | Show current JDK info |

## Configuration

- **User config**: `~/.jdm/config.json`
- **Defaults**: `core/config.json` (fallback)
- **JDM Home**: `~/.jdm` (configurable via `jdm_home`)
- **JDK versions**: `~/.jdm/versions/`
- **Current version**: `~/.jdm/current` (symlink)

## Important Implementation Details

### Version Switching (Shims Pattern)
On Windows, environment variable changes don't affect already-open terminals. JDM uses a symlink approach:
- `~/.jdm/current` symlink points to the active JDK version
- User adds `~/.jdm/current/bin` to PATH (must be at the front)
- `jdm use` updates the symlink, making switch instant without terminal restart

### API Integration
Fetches JDK metadata from Adoptium (Temurin) API via configurable mirror. Default: `https://api.adoptium.net/v3`

### Shell Integration
Users should add this to PowerShell `$PROFILE`:
```powershell
$env:JAVA_HOME = "$env:USERPROFILE\.jdm\current"
$env:PATH = "$env:USERPROFILE\.jdm\current\bin;$env:PATH"
```

## Directory Structure

```
core/
├── main.go              # Entry point
├── cmd/                 # Cobra commands
│   ├── root.go
│   ├── ls.go
│   ├── ls_remote.go
│   ├── install.go
│   ├── uninstall.go
│   ├── use.go
│   ├── default.go
│   ├── local.go
│   ├── alias.go
│   ├── config.go
│   └── current.go
├── internal/
│   ├── arch/            # Architecture detection (x64, arm64)
│   ├── config/          # Config management
│   ├── file/            # File operations (extract, symlink)
│   ├── jdk/             # JDK version manager
│   ├── log/             # Logging
│   ├── semver/          # Semantic version parsing
│   ├── utils/           # Utility helpers
│   └── web/             # Adoptium API client
└── config.json          # Default configuration
```

## Agent skills

### Issue tracker

Issues live as Gitee issues in `xjtuer216/jdm`, driven through the Gitee REST API v5 (no `gh`/`glab` CLI here). See `docs/agents/issue-tracker.md`.

### Triage labels

The five canonical triage roles map to the same-named Gitee labels; category roles are `bug` and `enhancement`. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context layout: root `GLOSSARY.md` plus `docs/adr/`. See `docs/agents/domain.md`.