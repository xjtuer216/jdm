# CORE (Go Source)

**Module**: `github.com/xjtuer216/jdm`
**Language**: Go 1.21
**Framework**: Cobra CLI

## OVERVIEW
All Go source code for JDM CLI. Single binary output (`jdm.exe`).

## STRUCTURE
```
core/
├── main.go              # Entry → cmd.Execute()
├── go.mod               # Module root (not at repo root)
├── go.sum               # Dependencies lock
├── config.json          # Default config template
├── cmd/                 # 11 Cobra subcommands
└── internal/            # 7 private packages
```

## WHERE TO LOOK
| Task | File | Notes |
|------|------|-------|
| Add command | `cmd/<name>.go` | Register in `root.go` |
| Root command | `cmd/root.go` | Execute(), init(), config loading |
| Version manager | `internal/jdk/` | Install/Use/List/Uninstall |
| Config manager | `internal/config/` | Load/Save/Get/Set |
| API client | `internal/web/` | Adoptium API |
| File ops | `internal/file/` | Symlink, extract |
| Logging | `internal/log/` | logrus setup |
| Arch detection | `internal/arch/` | x64, arm64 |
| Semver parsing | `internal/semver/` | Version comparison |

## CONVENTIONS
- **Package naming**: Single word, lowercase (`config`, `jdk`, `web`)
- **Error handling**: Return errors, don't panic
- **Tests**: `*_test.go` alongside source, use `testify`
- **Version injection**: `-ldflags "-X github.com/xjtuer216/jdm/internal/jdk.Version=X.Y.Z"`

## ANTI-PATTERNS
- **NEVER** suppress type errors with `as any` or `//nolint` without reason
- **NEVER** modify `go.mod` manually - use `go get`
- **DO NOT** add commands without registering in `root.go`
- **Binary committed**: `jdm.exe` tracked (should be gitignored)

## COMMANDS
```bash
# All go commands require cd core first
cd core && go build -o jdm.exe .
cd core && go test ./...
cd core && go vet ./...
```

## NOTES
- **Non-standard layout**: Go module at `core/`, not repo root
- **Windows only**: Symlink operations need Developer Mode or admin
- **No linter config**: Consider adding `.golangci.yml`
