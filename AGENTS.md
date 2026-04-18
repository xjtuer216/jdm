# JDM PROJECT KNOWLEDGE BASE

**Generated:** 2026-04-17
**Commit:** 3c06bf1 - chore: update .gitignore to exclude docs directory
**Branch:** develop

## OVERVIEW
JDM (JDK Version Manager) - Windows CLI tool for managing multiple JDK versions. Go 1.21 + Cobra CLI. Symlink-based version switching.

## STRUCTURE
```
jdm/
├── core/              # All Go source code (module root: github.com/whimsy/jdm)
│   ├── main.go        # Entry point → cmd.Execute()
│   ├── go.mod         # Module definition (non-standard: not at repo root)
│   ├── cmd/           # Cobra CLI commands (11 subcommands)
│   ├── internal/      # 7 private packages
│   └── config.json    # Default config template
├── .trellis/          # Trellis workflow system (Python scripts, specs, tasks)
├── .claude/           # Claude Code hooks and agents
├── .cursor/           # Cursor IDE commands
├── docs/              # Documentation (plans, specs)
└── AGENTS.md          # This file
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Build binary | `core/` | `cd core && go build -o jdm.exe .` |
| Add CLI command | `core/cmd/` | Register in root.go |
| Modify config logic | `core/internal/config/` | Load/Save/Get/Set |
| JDK operations | `core/internal/jdk/` | Install/Use/List/Uninstall |
| API client | `core/internal/web/` | Adoptium API integration |
| Trellis workflow | `.trellis/workflow.md` | Process rules, commit format |
| Claude hooks | `.claude/hooks/` | Auto-triggered Python scripts |

## CONVENTIONS
- **Commit format**: `type(scope): description` (feat/fix/docs/refactor/test/chore)
- **Go module**: `github.com/whimsy/jdm` - must `cd core` before go commands
- **Logging**: logrus, configured in cmd/root.go
- **Config**: JSON files, dual-layer (user ~/.jdm/config.json > exe-dir/config.json)
- **Version injection**: `-ldflags "-X github.com/whimsy/jdm/internal/jdk.Version=X.Y.Z"`

## ANTI-PATTERNS (THIS PROJECT)
- **NEVER read spec/requirement files directly** - Let Hook inject context
- **NEVER execute git commit** - AI prepares, human validates
- **NEVER skip Trellis workflow** - Read `.trellis/spec/` before coding
- **DO NOT ask multiple questions** - One at a time
- **DO NOT skip directly to implementation** - Use brainstorm first for new features
- **Binary committed**: `core/jdm.exe` is tracked (should be gitignored)

## UNIQUE STYLES
- **Symlink switching**: `~/.jdm/current` symlink → active JDK, no env var changes needed
- **Trellis journal**: Max 2000 lines, auto-rotates
- **Cross-layer features**: Require `/trellis:check-cross-layer` command
- **Session recording**: Mandatory after task completion

## COMMANDS
```bash
# Build (from repo root)
cd core && go build -o jdm.exe .

# Build with version
cd core && go build -ldflags "-X github.com/whimsy/jdm/internal/jdk.Version=1.0.0" -o jdm.exe .

# Run tests
cd core && go test ./...

# Run single test
cd core && go test -v ./internal/config/...
```

## NOTES
- **Go module at core/**: Non-standard layout, all go commands require `cd core`
- **No CI/CD**: No GitHub Actions, Makefile, or release automation
- **Windows only**: Symlink-based switching requires Developer Mode or admin
- **Shell integration**: Users must add `~/.jdm/current/bin` to PATH front
- **.gitignore excludes**: .claude, .cursor, .trellis, docs/, AGENTS.md, CLAUDE.md
