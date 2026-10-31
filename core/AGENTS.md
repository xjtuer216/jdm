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
└── internal/            # 8 private packages
```

## WHERE TO LOOK
| Task | File | Notes |
|------|------|-------|
| Add command | `cmd/<name>.go` | Register in `root.go` |
| Root command | `cmd/root.go` | Execute(), init(), config loading |
| Version manager | `internal/jdk/` | Install/Use/List/Uninstall |
| Download source | `internal/jdk/download.go` | Candidate chain, SHA-256, temp cleanup |
| Config manager | `internal/config/config.go` | Load/Save/Get/Set; key registry |
| API client | `internal/web/` | Adoptium API + User-Agent |
| Domain terms | `../GLOSSARY.md` | Metadata source / download source / proxy |
| Review rules | `../CODING_STANDARDS.md` | Judgement calls enforced in review |
| File ops | `internal/file/` | Symlink, extract |
| Logging | `internal/log/` | logrus setup |
| Arch detection | `internal/arch/` | x64, arm64 |
| Semver parsing | `internal/semver/` | Version comparison |

## CONVENTIONS
- **Package naming**: Single word, lowercase (`config`, `jdk`, `web`)
- **Error handling**: Return errors, don't panic
- **Tests**: `*_test.go` alongside source, standard library `testing` (testify is not a dependency); HTTP tests use `httptest`
- **Config keys**: a scalar key must be registered in `NewConfig`, `settableKeys`, `scalarField`, `EnsureInstallConfig`, install-config merge, and `config.json`
- **HTTP**: every request carries `User-Agent: JDM/<version>`
- **Version injection**: `-ldflags "-X github.com/xjtuer216/jdm/internal/jdk.Version=X.Y.Z"`

## ANTI-PATTERNS
- **NEVER** suppress type errors with `as any` or `//nolint` without reason
- **NEVER** modify `go.mod` manually - use `go get`
- **DO NOT** add commands without registering in `root.go`
- **Binary committed**: `jdm.exe` tracked (should be gitignored)

## COMMANDS
```bash
# From repo root: gofmt + vet + tests
./scripts/check.sh

# All go commands require cd core first
cd core && go build -o jdm.exe .
cd core && go test ./...
cd core && go vet ./...
```

## NOTES
- **Non-standard layout**: Go module at `core/`, not repo root
- **Windows only**: Symlink operations need Developer Mode or admin
- **Guardrail**: `../scripts/check.sh` runs gofmt, vet, and tests; CI runs the same script
