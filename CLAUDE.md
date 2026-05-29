# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`devops-starter` is a cross-platform DevOps tool installer and dotfile manager written in Go. It downloads 61+ pre-built CLI binaries directly (no package manager) and manages shell/editor configuration via symlinks.

## Common Commands

```bash
# Build
make build               # compile for current platform
make install             # build and copy to ~/.local/bin

# Test
make test                # run all tests with race detector
make test-coverage       # generate coverage report
go test ./internal/installer/... -run TestInstallTarGz  # single test

# Lint & Validate
make lint                # golangci-lint + shellcheck
make check               # fmt + vet + lint + test (full validation)

# Bazel (used in CI)
make bazel-build         # hermetic build
make bazel-test          # hermetic test

# Dev setup
make setup               # install dev tools and pre-commit hooks
```

## Architecture

### High-Level Data Flow

```
CLI (cobra)
  ├─ config.Load()      ← YAML (~/.config/devops-starter/config.yaml)
  ├─ platform.Detect()  ← OS, arch, distro
  ├─ registry.New()     ← all 61 tool definitions (Go structs)
  └─ installer.Install()
       ├─ HTTP GET (URLTemplate or per-platform URL)
       ├─ SHA256 checksum verification
       ├─ tar.gz extraction or raw binary rename
       └─ state.Store() ← record installed version as JSON
```

### Package Map

| Package | Role |
|---|---|
| `cmd/devops-starter/` | Entry point only — no business logic |
| `internal/cli/` | One file per Cobra subcommand; `root.go` registers global flags (`--config`, `--dry-run`, `--yes`) |
| `internal/registry/` | All 61 tool definitions as `tooldef.Tool` Go structs, split into group files |
| `internal/installer/` | Download → verify → extract → place; uses functional options pattern; worker pool (default 4) |
| `internal/platform/` | Detects OS / arch / distro; drives URL and binary selection |
| `internal/config/` | YAML load/save; XDG_CONFIG_HOME; group flags; per-tool version overrides |
| `internal/state/` | JSON store of managed tools + installed versions (separate from config by design) |
| `internal/dotfiles/` | Symlink manager with auto-backup to `~/.dotfiles.bak` |
| `internal/tui/` | Bubble Tea Elm-pattern UI (`Model` / `Update` / `View`); 4 screens |
| `internal/updater/` | Async GitHub Releases check on TUI startup (5 s timeout, non-blocking) |
| `pkg/tooldef/` | Public types: `Tool`, `Platform`, `Group`, `ArchiveFormat`, `InstallMode` |

### Config vs State Separation

- **Config** (`~/.config/devops-starter/config.yaml`) — declarative intent (what *should* be installed).
- **State** (JSON store) — actual installed versions. Kept separate to enable drift detection, the `adopt` command, and clean uninstall without touching user config.

### Install Modes

| Mode | How | Used for |
|---|---|---|
| `eget` | `eget <owner/repo>` with auto asset detection | ~53 tools (k9s, fzf, ripgrep, …) |
| `eget-url` | `eget <url>` with Go template URL resolution | kubectl, helm, terraform, vault, … |
| `custom` | Download → checksum → extract → post-install script | aws-cli, azure-cli, gcloud-cli |
| `mise` | `mise use <backend>@<version>` | language runtimes |

### TUI

Bubble Tea Elm pattern: `Model` (state) → `Update(tea.Msg)` (transitions) → `View()` (render). Bubble Tea model types use **value receivers** by design (not a bug; golangci-lint exempts `hugeParam` for these types).

## Contract-Oriented Development

Every public function and interface method must document its preconditions, postconditions, and error contract in its godoc comment.

### Interface contracts

The three interfaces in `internal/cli/deps.go` (`ToolInstaller`, `ToolRegistry`, `StateStore`) are the primary abstraction boundary between CLI commands and their dependencies. Each method must carry explicit pre/postcondition docs:

```go
// Install installs the given tool.
// Precondition: tool != nil, tool.Name != "", platform is supported by tool.
// Postcondition: on nil return, binary exists at installDir/<tool.GetInstallName()> with executable permissions.
// Returns ErrUnsupportedPlatform if tool.Platforms excludes the current platform.
Install(ctx context.Context, tool *tooldef.Tool) error
```

### Typed sentinel errors

Packages whose errors callers must distinguish must export sentinel errors, not raw strings. Use `var ErrXxx = errors.New(...)` at the package level; callers use `errors.Is()`.

Packages with ad-hoc string errors to convert:
- `internal/platform` — "unsupported operating system / architecture"
- `internal/installer` — "no URL template or override for tool"
- `internal/registry` — tool-not-found path in `Get()`

### Tool validation

`tooldef.Tool` is constructed in registry code and consumed by the installer. Add `func (t *Tool) Validate() error` that checks required fields per `InstallMode`:
- All modes: `Name` and `Version` non-empty
- `eget`: `Repo` non-empty
- `eget-url` / `custom`: `URLTemplate` non-empty, or at least one entry in `URLs`
- `mise` (per-tool): `MiseBackend` non-empty

Call `Validate()` in `registry.New()` so misconfigured tools fail at startup, not mid-install.

### State store invariant

After `Store.Record(name, version)` + `Store.Save()` succeed, `Store.GetVersion(name) == version` must hold. Tests must assert this invariant directly, not just assert no error.

### CLI–TUI synchronization contract

The CLI (`internal/cli/`) and TUI (`internal/tui/`) share state through `[]state.GroupState`, produced by `state.ResolveAll()`. **Any change to `state.ToolState` or `state.Status` requires updating all three of the following or behaviour will diverge between interactive and non-interactive modes:**

| Layer | File | What to update |
|---|---|---|
| State resolution | `internal/state/state.go` | `resolveTool()`, `resolveToolStatus()` |
| TUI rendering | `internal/tui/view.go` | How the new field/status is displayed |
| Plain-text output | `internal/tui/table.go` | `PrintTable()` — the `--no-tui` path |
| TUI action guards | `internal/tui/update.go` | All `if t.Status == ...` eligibility checks |

`state.Status` values drive which TUI actions are available (install, remove, verify). When adding a new `Status` constant, audit every `switch`/`if` in `internal/tui/update.go` that branches on status.

The TUI injects a concrete `*installer.Installer` (not the `ToolInstaller` interface), so TUI install/remove paths cannot use mock installers. Keep side-effectful TUI logic in `internal/tui/commands.go` (the `tea.Cmd` functions) so it stays isolated and testable.

### Testing contracts

Table-driven test rows must name the contract being exercised, not just the input:

```go
{"eget mode requires non-empty Repo", ...}   // good
{"tool1", ...}                                // avoid
```

## Adding Tools / Groups / Commands

| Task | Action |
|---|---|
| Add a tool | `r.register(...)` in the matching `internal/registry/<group>.go` |
| Add a group | New `internal/registry/<group>.go`, add constant to `pkg/tooldef/group.go`, register in `registry.go` |
| Add a CLI subcommand | Create `internal/cli/<cmd>.go` with `newXxxCmd()` + `runXxx()`, register in `root.go` |
| Add an archive format | Add to `pkg/tooldef/`, implement extraction in `internal/installer/` |

## Code Conventions

### Linting

Ten linters are active (see `.golangci.yml`): `errcheck`, `govet`, `staticcheck`, `unused`, `ineffassign`, `gocritic`, `revive`, `cyclop` (max 15), `gocognit`, `dupl` (threshold 150), `gosec`.

Intentional exemptions:
- `gosec` G104/G204/G301-G306/G703 — expected for a CLI tool that runs external binaries.
- `dupl` disabled in `internal/registry/` — tool definitions are intentionally repetitive.
- `unused-parameter` exemptions in `internal/cli/` — Cobra `cmd`/`args` convention.

### Commit Messages

Conventional Commits are enforced via commitlint + pre-commit hook.  
Valid types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`.

### Build System

- **Make** — local development (fast iteration).
- **Bazel** — CI and release (hermetic, reproducible cross-compilation for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64). Contributors do not need to use Bazel locally.

## Key Files

| File | Purpose |
|---|---|
| `configs/default.yaml` | Template written to user config on first run |
| `dotfiles/` | Managed shell/editor configs symlinked into `$HOME` |
| `.golangci.yml` | Linter configuration |
| `.commitlintrc.yaml` | Conventional Commits enforcement |
| `.pre-commit-config.yaml` | Git hooks (lint, commitlint) |
| `.mise.toml` | Dev tool versions (Go 1.26.3, Python 3.12, Node 22) |
| `ARCHITECTURE.md` | Detailed design decisions and extension points |
| `CONTRIBUTING.md` | Dev setup walkthrough |
