# Modify jdm ls-remote Command

> **Historical design note (2026-04)**: kept for rationale. Current behavior lives in `core/internal/web/adoptium.go` and `core/cmd/ls_remote.go`.
> Evidence paths below reference the removed `.sisyphus/` tooling and no longer exist.

## TL;DR

> **Quick Summary**: Refactor `jdm ls-remote` to remove hardcoded default version and add dynamic major version discovery. Two modes: no-arg shows latest per major version in table format; with-arg shows all builds for that major version.
>
> **Deliverables**:
> - `core/internal/web/adoptium.go` — Add `ListAvailableReleases()` and `GetLatestPerMajor()`
> - `core/internal/jdk/manager.go` — Modify `ListRemote()` to support empty version (dynamic discovery)
> - `core/cmd/ls_remote.go` — Update command to handle two display modes (table vs list)
>
> **Estimated Effort**: Short
> **Parallel Execution**: NO — sequential (3 dependent tasks)
> **Critical Path**: Task 1 → Task 2 → Task 3

---

## Context

### Original Request
修改 `jdm ls-remote` 指令：1) version不能写死；2) 改为 `jdm ls-remote [version]`，不传参时查询所有远程可用JDK版本（表格形式，倒序），传参时查询该大版本全部构建（倒序）。

### Interview Summary
**Key Decisions**:
- 无参展示: 表格形式（类似nvm-windows多列表格）
- 大版本范围: 动态发现（Adoptium API `/v3/info/available_releases`）
- 传参展示: 该大版本全部构建，无限制
- 排序: 倒序（最新在前）

**Research Findings**:
- nvm-windows 通过单次请求 `index.json` 获取全部版本，本地分类
- Adoptium API 按大版本请求，需先动态发现大版本列表，再逐个请求
- 现有 `adoptium.go` 已有 `ListAllVersions()` 但硬编码了大版本列表
- 现有 `compareVersions()` 可用于排序

---

## Work Objectives

### Core Objective
Remove hardcoded version default from ls-remote, add dynamic major version discovery via Adoptium API, implement two display modes (table for all, list for specific major).

### Concrete Deliverables
- `core/internal/web/adoptium.go` — 2 new functions
- `core/internal/jdk/manager.go` — 1 modified function
- `core/cmd/ls_remote.go` — updated command handler

### Definition of Done
- [ ] `jdm ls-remote` (no arg) displays latest version per major version in table, sorted descending
- [ ] `jdm ls-remote 17` displays all builds for Java 17, sorted descending, LTS marked
- [ ] `go build` succeeds with no errors
- [ ] `go vet ./...` passes

### Must Have
- Dynamic major version discovery via `/v3/info/available_releases`
- Table display for no-arg mode
- Descending sort for both modes
- LTS marker `[LTS]` on applicable versions

### Must NOT Have (Guardrails)
- No hardcoded version lists (remove `[8, 11, 17, 21, 25]`)
- No changes to other commands (ls, install, use, etc.)
- No changes to config loading or symlink logic
- No test file modifications (existing tests remain untouched)

---

## Verification Strategy

### Test Decision
- **Infrastructure exists**: YES (Go testing, testify)
- **Automated tests**: Tests-after (add tests after implementation)
- **Framework**: Go built-in `testing` + `testify`
- **Agent-Executed QA**: ALWAYS (mandatory for all tasks)

### QA Policy
Every task includes agent-executed QA scenarios.

---

## Execution Strategy

### Sequential Waves (3 tasks, dependent)

```
Wave 1 (Start — web layer):
└── Task 1: Add dynamic discovery + latest-per-major to adoptium.go

Wave 2 (After Wave 1 — business logic):
└── Task 2: Modify ListRemote() in manager.go

Wave 3 (After Wave 2 — CLI layer):
└── Task 3: Update ls_remote.go command handler + display logic
```

### Dependency Matrix
- **1**: — → 2
- **2**: 1 → 3
- **3**: 2 → —

### Agent Dispatch Summary
- **Wave 1**: Task 1 → `quick` (2 new functions, ~40 lines)
- **Wave 2**: Task 2 → `quick` (1 modified function, ~20 lines changed)
- **Wave 3**: Task 3 → `quick` (command handler update, ~30 lines changed)

---

## TODOs

- [x] 1. Add dynamic discovery functions to adoptium.go

  **What to do**:
  - Add `ListAvailableReleases() ([]int, int, int, error)` — calls `GET {mirror}/v3/info/available_releases`, returns `available_releases` list, `most_recent_lts`, `most_recent_feature_version`
  - Add `GetLatestPerMajor(majorVersion int) (*RemoteVersion, error)` — calls `FetchVersions(majorVersion)` and returns the first (already sorted newest-first) element
  - Add response type `AvailableReleases` struct with JSON tags matching API response
  - Place new functions after `ListAllVersions()` (around line 146)
  - Reuse existing `AdoptiumClient` and `http.Client`

  **Must NOT do**:
  - Do NOT modify existing `FetchVersions()`, `FetchLatestLTS()`, `ListAllVersions()` functions
  - Do NOT change any existing struct definitions
  - Do NOT add new dependencies

  **Recommended Agent Profile**:
  - **Category**: `quick` — straightforward function additions following existing patterns
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: NO (must complete before Task 2)
  - **Blocks**: Task 2
  - **Blocked By**: None

  **References**:
  - `core/internal/web/adoptium.go:62-108` — `FetchVersions()` pattern for HTTP GET + JSON parsing + filtering
  - `core/internal/web/adoptium.go:131-146` — `ListAllVersions()` pattern for iterating over multiple major versions
  - `core/internal/web/adoptium.go:39-46` — `RemoteVersion` struct (reuse this type)
  - Adoptium API docs: `https://api.adoptium.net/q/swagger/` — `/v3/info/available_releases` endpoint

  **Acceptance Criteria**:
  - [ ] `ListAvailableReleases()` returns correct major versions from API
  - [ ] `GetLatestPerMajor(17)` returns the latest JDK 17 build
  - [ ] New code compiles: `cd core && go build ./internal/web/`

  **QA Scenarios**:

  ```
  Scenario: ListAvailableReleases returns available major versions
    Tool: Bash (cd core && go run -exec "echo" ./internal/web/ 2>&1; compile check)
    Steps:
      1. Run: cd core && go vet ./internal/web/
    Expected Result: Exit code 0, no vet errors
    Evidence: .sisyphus/evidence/task-1-vet-pass.txt

  Scenario: GetLatestPerMajor returns latest build for Java 17
    Tool: Bash
    Steps:
      1. Create temp test file core/internal/web/adoptium_new_test.go
      2. Write test calling GetLatestPerMajor(17) with real API
      3. Assert returned version is non-empty and starts with expected prefix
      4. Run: cd core && go test -v ./internal/web/ -run TestGetLatestPerMajor
    Expected Result: Test passes, version string is valid
    Evidence: .sisyphus/evidence/task-1-getlatest-test.txt
  ```

  **Commit**: YES
  - Message: `feat(ls-remote): add dynamic version discovery to adoptium client`
  - Files: `core/internal/web/adoptium.go`
  - Pre-commit: `cd core && go vet ./internal/web/`

---

- [x] 2. Modify ListRemote() in manager.go to support dynamic discovery

  **What to do**:
  - Modify `ListRemote(version string) ([]web.RemoteVersion, error)` in `core/internal/jdk/manager.go` (line 68-71)
  - New logic:
    - If `version == ""`: call `vm.Client.ListAvailableReleases()` to get all major versions, then for each major version call `vm.Client.GetLatestPerMajor()`, collect all results into a single slice, sort by version descending using `web.compareVersions` (may need to export it or add a wrapper), return
    - If `version != ""`: keep existing behavior — `vm.Client.FetchVersions(majorVersion)`
  - May need to add an exported `CompareVersions(v1, v2 string) int` wrapper in adoptium.go if `compareVersions` is unexported (it is — lowercase)

  **Must NOT do**:
  - Do NOT change any other methods in VersionManager (Install, Uninstall, Use, etc.)
  - Do NOT modify config loading or symlink logic
  - Do NOT change error handling patterns

  **Recommended Agent Profile**:
  - **Category**: `quick` — single function modification
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: NO (depends on Task 1)
  - **Blocks**: Task 3
  - **Blocked By**: Task 1

  **References**:
  - `core/internal/jdk/manager.go:68-71` — Current `ListRemote()` implementation (4 lines, replace with new logic)
  - `core/internal/jdk/manager.go:273-282` — `parseMajorVersion()` helper (reuse for version != "" branch)
  - `core/internal/web/adoptium.go:187-199` — `compareVersions()` (unexported, need to export as `CompareVersions`)
  - Task 1 output — new `ListAvailableReleases()` and `GetLatestPerMajor()` functions

  **Acceptance Criteria**:
  - [ ] `ListRemote("")` returns latest version per major version, sorted descending
  - [ ] `ListRemote("17")` returns all builds for Java 17, sorted descending
  - [ ] `cd core && go build ./internal/jdk/` succeeds

  **QA Scenarios**:

  ```
  Scenario: ListRemote with empty string returns versions from all majors
    Tool: Bash
    Steps:
      1. Run: cd core && go vet ./internal/jdk/
    Expected Result: Exit code 0, no vet errors
    Evidence: .sisyphus/evidence/task-2-vet-pass.txt

  Scenario: ListRemote with specific version returns all builds
    Tool: Bash
    Steps:
      1. Create temp test file core/internal/jdk/manager_listremote_test.go
      2. Write test calling ListRemote("17") with real config
      3. Assert returned slice has multiple entries, all starting with "17."
      4. Run: cd core && go test -v ./internal/jdk/ -run TestListRemote
    Expected Result: Test passes, all versions are Java 17.x
    Evidence: .sisyphus/evidence/task-2-listremote-specific.txt
  ```

  **Commit**: YES
  - Message: `feat(ls-remote): support dynamic discovery in ListRemote`
  - Files: `core/internal/jdk/manager.go`, `core/internal/web/adoptium.go` (export CompareVersions)
  - Pre-commit: `cd core && go vet ./internal/jdk/ ./internal/web/`

---

- [x] 3. Update ls_remote.go command handler with two display modes

  **What to do**:
  - Modify `core/cmd/ls_remote.go`:
    - Remove hardcoded `version := "17"` default
    - If no args (`len(args) == 0`): call `vm.ListRemote("")`, display results in table format
    - If arg provided (`len(args) > 0`): call `vm.ListRemote(args[0])`, display as flat list with LTS markers
  - Add table display function — use simple text/tabwriter (Go standard library `text/tabwriter`, no new dependencies)
  - Table format for no-arg mode: columns showing Version, Major, LTS status, Vendor
  - Keep existing `isLTS()` function but fix it — current implementation checks last 3 chars for "LTS", but adoptium.go already has `isLTS()` using `strings.Contains(strings.ToLower(releaseName), "lts")`. Use the strings.Contains approach (more robust)
  - Update `Short` description to reflect new behavior

  **Must NOT do**:
  - Do NOT add new imports beyond standard library + existing cobra/jdk imports
  - Do NOT modify rootCmd or other commands
  - Do NOT change error handling patterns (keep existing fmt.Fprintf + os.Exit)

  **Recommended Agent Profile**:
  - **Category**: `quick` — command handler update with display logic
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: NO (depends on Task 2)
  - **Blocks**: None (final task)
  - **Blocked By**: Task 2

  **References**:
  - `core/cmd/ls_remote.go:1-51` — Entire file to modify
  - `core/cmd/ls.go` — Similar command structure for reference (list local versions)
  - `core/internal/web/adoptium.go:39-46` — `RemoteVersion` struct fields available for display
  - Go standard lib: `text/tabwriter` for table formatting

  **Acceptance Criteria**:
  - [ ] `jdm ls-remote` (no arg) displays table with columns: Version, Major, LTS, Vendor
  - [ ] `jdm ls-remote 17` displays flat list of Java 17 builds with [LTS] markers
  - [ ] `cd core && go build -o jdm.exe .` succeeds
  - [ ] `cd core && go vet ./cmd/` passes

  **QA Scenarios**:

  ```
  Scenario: jdm ls-remote (no arg) displays table
    Tool: Bash
    Steps:
      1. Build: cd core && go build -o jdm.exe .
      2. Run: cd core && .\jdm.exe ls-remote
      3. Capture output
    Expected Result: Output contains table with version columns, sorted newest first
    Evidence: .sisyphus/evidence/task-3-lsremote-all.txt

  Scenario: jdm ls-remote 17 displays all Java 17 builds
    Tool: Bash
    Steps:
      1. Run: cd core && .\jdm.exe ls-remote 17
      2. Capture output
    Expected Result: Output lists multiple Java 17.x versions, sorted newest first, with [LTS] markers
    Evidence: .sisyphus/evidence/task-3-lsremote-17.txt

  Scenario: jdm ls-remote with invalid major version handles gracefully
    Tool: Bash
    Steps:
      1. Run: cd core && .\jdm.exe ls-remote 999
      2. Capture output and exit code
    Expected Result: Shows "No versions found" message, exit code 0 (not crash)
    Evidence: .sisyphus/evidence/task-3-lsremote-invalid.txt
  ```

  **Commit**: YES
  - Message: `feat(ls-remote): add table display and remove hardcoded default`
  - Files: `core/cmd/ls_remote.go`
  - Pre-commit: `cd core && go build -o jdm.exe . && go vet ./cmd/`

---

## Final Verification Wave (MANDATORY — after ALL implementation tasks)

> Run after all 3 tasks complete.

- [x] F1. **Build Verification**
  ```bash
  cd core && go build -o jdm.exe .
  cd core && go vet ./...
  ```
  Expected: Exit code 0, no errors

- [x] F2. **Functional QA — No-arg mode**
  ```bash
  cd core && .\jdm.exe ls-remote
  ```
  Expected: Table output with versions from multiple major versions, sorted newest first

- [x] F3. **Functional QA — Specific version mode**
  ```bash
  cd core && .\jdm.exe ls-remote 17
  ```
  Expected: All Java 17 builds listed, sorted newest first, [LTS] markers present

- [x] F4. **Edge case QA — Invalid version**
  ```bash
  cd core && .\jdm.exe ls-remote 999
  ```
  Expected: Graceful "No versions found" message

---

## Commit Strategy

- **1**: `feat(ls-remote): add dynamic version discovery to adoptium client` — adoptium.go
- **2**: `feat(ls-remote): support dynamic discovery in ListRemote` — manager.go, adoptium.go (export CompareVersions)
- **3**: `feat(ls-remote): add table display and remove hardcoded default` — ls_remote.go

---

## Success Criteria

### Verification Commands
```bash
cd core && go build -o jdm.exe .     # Expected: exit 0
cd core && go vet ./...               # Expected: no issues
.\jdm.exe ls-remote                   # Expected: table output
.\jdm.exe ls-remote 17                # Expected: Java 17 builds list
```

### Final Checklist
- [ ] No hardcoded version defaults remain in ls_remote.go
- [ ] Dynamic discovery via /v3/info/available_releases works
- [ ] Table display for no-arg mode
- [ ] Flat list display for specific version mode
- [ ] Both modes sorted descending (newest first)
- [ ] LTS markers present on applicable versions
- [ ] Build succeeds, vet passes
