# Coding Standards

Reviewer-facing judgement calls. Mechanical checks live in `scripts/check.sh`,
`.githooks/pre-push`, and `.github/workflows/check.yml`.

## Tests

- Each test states an external behavior. Do not assert private branches or
  internal call order.
- A failure-path test must reach the failing stage. For cleanup, assert after
  the resource was created: a test that fails before `os.CreateTemp` proves
  nothing about temp-file cleanup.
- Assert behavior, not source text, duplicated defaults, or tautologies.
- Tests are offline and deterministic. Mock HTTP with `httptest`; do not
  depend on real mirrors.

## Domain language

- Follow `GLOSSARY.md`: `mirror` is the metadata source, `jdk_mirror` is the
  download source, `proxy` is a URL-prefix proxy, and the mirror layout is the
  fixed directory path inside a download source.
- Do not call a download source or a proxy just "mirror" in code or user output.

## Configuration

- A new scalar key touches all of: the `NewConfig` default, `settableKeys`,
  `scalarField`, `EnsureInstallConfig`, the install-config merge,
  `core/config.json`, and the README tables.
- Store `none` verbatim; treat it as disabled only when reading the effective
  value. `jdk_mirror` and `proxy` share this convention.

## Network and downloads

- Every HTTP request carries `User-Agent: JDM/<version>`.
- Each download attempt checks HTTP status, content type, byte size, and
  SHA-256. On failure, delete the temp file and try the next source.
- A checksum mismatch is a failure even when the byte size matches.
- Aggregate errors name every source and its reason.

## Errors and cleanup

- Wrap errors with `%w`; do not swallow them, including file close on a success
  path.
- Leave no temp files or partial installs behind on failure.
