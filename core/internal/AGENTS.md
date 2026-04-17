# CORE/INTERNAL (Private Packages)

**Module**: `github.com/xjtuer216/jdm/internal`
**Packages**: 7

## OVERVIEW
Internal implementation packages - not importable outside module.

## STRUCTURE
```
internal/
├── arch/          # Architecture detection (x64, arm64)
├── config/        # Config management (Load/Save/Get/Set)
├── file/          # File operations (symlink, extract)
├── jdk/           # JDK version manager (Install/Use/List/Uninstall)
├── log/           # Logging (logrus wrapper)
├── semver/        # Semantic version parsing
└── web/           # Adoptium API client
```

## WHERE TO LOOK
| Package | Purpose | Key Functions |
|---------|---------|---------------|
| `arch` | OS arch detection | `Detect()` |
| `config` | JSON config | `Load()`, `Save()`, `Get()`, `Set()` |
| `file` | File utilities | `CreateSymlink()`, `Extract()` |
| `jdk` | Version management | `Install()`, `Use()`, `List()`, `Uninstall()` |
| `log` | Logging setup | `Init()`, `Logger` |
| `semver` | Version parsing | `Parse()`, `Compare()` |
| `web` | HTTP client | `FetchVersions()`, `Download()` |

## CONVENTIONS
- **No external imports** from other internal packages without reason
- **Tests**: `*_test.go` alongside source (only `config/` and `web/` have tests)
- **Error returns**: Always return errors, never swallow

## ANTI-PATTERNS
- **NEVER** import `internal/` from outside `github.com/xjtuer216/jdm`
- **DO NOT** add business logic to `file/` or `log/` - keep them utility-only
- **NEVER** hardcode paths - use config values

## TEST COVERAGE
- ✅ `config/` - Load/Save tests
- ✅ `web/` - Table-driven parse tests
- ❌ `jdk/` - No tests (critical path!)
- ❌ `file/` - No tests
- ❌ `arch/` - No tests
- ❌ `semver/` - No tests
- ❌ `log/` - No tests

## NOTES
- **Version field**: `jdk.Version` string injected via ldflags
- **Config dual-layer**: User config (~/.jdm/config.json) overrides exe-dir config
- **Symlink switching**: `~/.jdm/current` → active JDK
