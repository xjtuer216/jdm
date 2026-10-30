# brainstorm: Reorganize JDM project structure to match nvm-windows

> **Historical design note (2026-04)**: the package rename described here was implemented. Current layout is documented in `core/AGENTS.md`.

## Goal

参照nvm-windows项目结构，调整JDM项目架构，在保持现有功能不变的前提下，提升项目的可扩展性。

## What I already know

### Current JDM structure:
```
core/
├── main.go              # Entry point
├── cmd/                 # Cobra commands
│   ├── root.go
│   ├── ls.go, ls_remote.go, install.go, ...
├── internal/
│   ├── api/             # Adoptium API client
│   ├── config/          # Configuration management
│   ├── log/             # Logging
│   ├── utils/           # Symlink, extract utilities
│   └── version/         # Version manager
├── config.json          # Default configuration
└── jdm.exe              # Built binary
```

### nvm-windows structure:
```
src/
├── nvm.go               # Main entry + commands
├── arch/                # Architecture detection
├── author/              # Author/bridge
├── encoding/            # Encoding utilities
├── file/                # File operations (unzip, etc.)
├── node/                # Node version management
├── semver/              # Semantic versioning
├── upgrade/             # Upgrade/check
├── utility/             # Logging and utilities
├── web/                 # Web/API client
├── go.mod
└── manifest.json
```

### Functional mapping:

| nvm-windows | JDM (current) | Notes |
|-------------|---------------|-------|
| `arch/` | - | Architecture detection - needed? |
| `file/` | `internal/utils/` | File operations - partially covered |
| `node/` | `internal/version/` | Version management - similar concept |
| `semver/` | - | Semantic versioning - may need |
| `utility/` | `internal/log/` | Logging/utilities |
| `web/` | `internal/api/` | API client - equivalent |

## Assumptions (temporary)

- The target is a structural reorganization, not rewriting logic
- Need to maintain backward compatibility with existing commands
- Keep using Cobra for CLI framework
- Keep using logrus for logging

## Decision (ADR-lite)

**Context**: 需要在保持 Go 最佳实践和借鉴 nvm-windows 模块化之间找到平衡

**Decision**: 保持 internal/ 分层结构，但优化内部模块划分

**Consequences**:
- 保留 internal/ 约定，符合 Go 惯用法
- 重命名/拆分模块使其职责更清晰
- 需要更新所有导入路径

## Requirements

- [ ] Maintain all existing commands (ls, ls-remote, install, uninstall, use, default, local, alias, config, current)
- [ ] Maintain all existing functionality
- [ ] Improve package organization for extensibility
- [ ] Keep Go best practices
- [ ] Rename `version/` to `jdk/` for clarity
- [ ] Rename `api/` to `web/` for consistency with nvm-windows
- [ ] Move file operations from `utils/` to new `file/` module
- [ ] Add `semver/` package for version parsing
- [ ] Add `arch/` package for architecture detection
- [ ] Update all import paths

## Acceptance Criteria (evolving)

- [ ] Project compiles successfully
- [ ] All existing commands work
- [ ] Package structure is clear and extensible

## Out of Scope

- Adding new commands
- Changing core logic
- Database or persistent storage changes

## Technical Notes

- Current JDM uses `internal/` to indicate private packages (Go 1.4+ convention)
- nvm-windows uses a flatter structure but imports as `nvm/xxx`
- Both approaches are valid; the question is what fits JDM better