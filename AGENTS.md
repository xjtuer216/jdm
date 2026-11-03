# JDM PROJECT KNOWLEDGE BASE

**Updated:** 2026-10-30
**Branch:** `develop` (integration); feature work on short-lived branches

## OVERVIEW
JDM (JDK Version Manager) - Windows CLI tool for managing multiple JDK versions. Go 1.21 + Cobra CLI. Symlink-based version switching.

## STRUCTURE
```
jdm/
├── core/                  # All Go source (module root: github.com/xjtuer216/jdm)
│   ├── main.go            # Entry point → cmd.Execute()
│   ├── go.mod             # Module definition (not at repo root)
│   ├── cmd/               # Cobra CLI commands
│   ├── internal/          # Private packages
│   └── config.json        # Default config template
├── scripts/               # check.sh, setup-hooks.sh
├── .githooks/             # pre-push guardrail
├── .github/workflows/     # check workflow (GitHub mirror)
├── docs/
│   ├── adr/               # Architecture decisions
│   ├── agents/            # Issue tracker and triage docs
│   └── history/           # Historical design notes
├── installer/             # Inno Setup installer
├── AGENTS.md              # This file
├── CODING_STANDARDS.md    # Reviewer-facing rules
└── GLOSSARY.md            # Domain vocabulary
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Run all checks | `./scripts/check.sh` | gofmt + vet + tests |
| Build binary | `core/` | `cd core && go build -o jdm.exe .` |
| Add CLI command | `core/cmd/` | Register in `root.go` |
| Config keys | `core/internal/config/config.go` | `settableKeys`, `scalarField`, install merge |
| JDK operations | `core/internal/jdk/` | Install/Use/List/Uninstall |
| Download source | `core/internal/jdk/download.go` | Candidate chain + SHA-256 |
| Metadata client | `core/internal/web/` | Adoptium API + User-Agent |
| Domain terms | `GLOSSARY.md` | Metadata source / download source / proxy |
| Decisions | `docs/adr/` | ADRs |
| Issue tracker | `docs/agents/issue-tracker.md` | Gitee API auth and quirks |
| Historical designs | `docs/history/` | Installer, ls-remote, package layout |
| Installer | `installer/jdm.iss` | Inno Setup script |

## CONVENTIONS
- **Commit format**: `type(scope): description` (feat/fix/docs/refactor/test/chore)
- **Branch prefixes**: `feat/`, `fix/`, `chore/`, `docs/`, `test/`, `refactor/`, `release/`; `develop`/`master` are integration branches. Avoid `feature/*`: a remote `feature` ref makes that namespace unpushable.
- **Go module**: `github.com/xjtuer216/jdm` - run go commands from `core/`
- **Logging**: logrus, configured in `cmd/root.go`
- **Config**: JSON, dual-layer (user `~/.jdm/config.json` > exe-dir `config.json`)
- **Version injection**: `-ldflags "-X github.com/xjtuer216/jdm/internal/jdk.Version=X.Y.Z"`
- **HTTP**: every request carries `User-Agent: JDM/<version>`
- **Tests**: standard library `testing`, `httptest` for HTTP; no testify

## WORKFLOW
- Commit and push only when the user explicitly asks or the invoked skill instructs; otherwise prepare the changes and report.
- Fetch tickets from Gitee with `docs/agents/issue-tracker.md`; no hook injects spec context.
- For a new feature without an existing ticket or spec, clarify or spec it first. A provided ticket/spec is the starting point for implementation.
- Review enforces `CODING_STANDARDS.md`; mechanical checks are `scripts/check.sh` and the pre-push hook.

## ANTI-PATTERNS (THIS PROJECT)
- **No secrets**: never hardcode tokens, passwords, or API keys.
- **No manual `go.mod` edits**: use `go get`.
- **No unregistered commands**: a new Cobra command must be added in `cmd/root.go`.
- **No swallowed errors**: wrap with `%w`; clean up temp files on failure.
- **No unrelated deletions**: report dead code instead of removing it during a focused change.
- **Binary committed**: `core/jdm.exe` is tracked (should be gitignored).

## UNIQUE STYLES
- **Symlink switching**: `~/.jdm/current` symlink → active JDK, no env var changes needed.
- **Download source vs metadata source**: `jdk_mirror` serves files; `mirror` is the Adoptium API. `proxy` only prefixes github.com URLs.

## COMMANDS
```bash
# All checks (from repo root)
./scripts/check.sh

# Enable the tracked pre-push guardrail
./scripts/setup-hooks.sh

# Build
cd core && go build -o jdm.exe .

# Build with version
cd core && go build -ldflags "-X github.com/xjtuer216/jdm/internal/jdk.Version=1.0.0" -o jdm.exe .

# Tests
cd core && go test ./...
cd core && go test -v ./internal/config/...
```

## NOTES
- **Go module at `core/`**: non-standard layout; all go commands require `cd core`.
- **Guardrails**: `scripts/check.sh` runs locally; `.github/workflows/check.yml` runs on the GitHub mirror. No Gitee CI.
- **Windows only**: symlink-based switching requires Developer Mode or admin.
- **Shell integration**: users add `~/.jdm/current/bin` to the front of PATH.
