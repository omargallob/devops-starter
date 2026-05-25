# Architecture

This document describes the internal design of devops-starter: how packages
relate to each other, the key data flows, and the reasoning behind major
decisions.

---

## High-Level Flow

```
CLI (cobra)
  │
  ├─ config.Load()          ← YAML config (~/.config/devops-starter/config.yaml)
  ├─ platform.Detect()      ← OS, arch, distro detection
  ├─ registry.New()         ← all 61 tool definitions
  │
  └─ installer.Install()    ← download → verify → extract → place binary
       │
       ├─ HTTP GET (URLTemplate or per-platform URLs)
       ├─ SHA256 checksum verification
       ├─ Archive extraction (tar.gz) or raw binary rename
       └─ state.Store()     ← record installed version
```

    ## Architecture Diagram

    See the interactive Excalidraw diagram here:

    - [docs/code-architecture-interactions.excalidraw](docs/code-architecture-interactions.excalidraw)

    The diagram highlights three primary concern areas:

    - Platform matrix and URL/checksum drift risk between `internal/platform`, `internal/registry`, and installer inputs.
    - Download pipeline reliability and integrity verification in `internal/installer`.
    - Desired config vs persisted state drift between `internal/config` and `internal/state`.

## Package Responsibilities

### `cmd/devops-starter/`

Minimal entry point. Constructs the root Cobra command and calls `Execute()`.
Contains no business logic.

### `internal/cli/`

One file per subcommand (`install.go`, `list.go`, `adopt.go`, `remove.go`,
`status.go`, `dotfiles.go`, `doctor.go`, `config.go`). Each exports a
`newXxxCmd()` constructor and a `runXxx()` function.

`root.go` defines global persistent flags (`--config`, `--dry-run`, `--yes`)
and registers all subcommands.

### `internal/config/`

Loads and validates the YAML config file. Handles:
- Default path resolution (XDG_CONFIG_HOME)
- Group enable/disable flags
- Per-tool version overrides and disable flags
- Config initialization (writes `configs/default.yaml` to the user path)

### `internal/registry/`

Defines all 61 tools as `tooldef.Tool` structs, organized into 7 group files.
`registry.go` provides the `Registry` type with lookup methods:
- `New()` — constructs and populates the registry
- `Get(name)` — single tool lookup
- `GetByGroup(group)` — all tools in a group
- `All()` — sorted list of all tools
- `Names()` — sorted tool name list

### `internal/installer/`

Core orchestration package. Responsibilities:
1. **Download** — HTTP GET with redirect following
2. **Checksum** — SHA256 verification against known hashes
3. **Extract** — tar.gz decompression or raw binary handling
4. **Install** — place binary in `InstallDir`, set executable permissions
5. **Concurrency** — worker pool (default 4) for parallel installs
6. **Progress** — reports download/install progress to the TUI or stdout

Configured via functional options pattern (`WithDryRun`, `WithConcurrency`,
`WithStateStore`).

### `internal/platform/`

Detects the runtime environment:
- `OS` — linux, darwin
- `Arch` — amd64, arm64
- `Distro` — ubuntu, arch (Linux only, via `/etc/os-release`)

Used by the installer to select the correct download URL and by the registry
to filter tools that don't support the current platform.

### `internal/dotfiles/`

Manages symlinks between the repo's `dotfiles/` directory and `$HOME`.
Operations:
- **Link** — creates symlinks, backs up existing files to `~/.dotfiles.bak`
- **Unlink** — removes managed symlinks
- **Status** — reports which files are linked, broken, or unmanaged

### `internal/state/`

Persistent JSON store tracking which tools are managed and their installed
versions. Used by `adopt`, `remove`, and `status` commands to differentiate
managed vs. system-installed tools.

### `internal/tui/`

Bubbletea-based interactive components for the `status` command. Provides a
full-screen dashboard view of tool state with install/remove actions.

Architecture:
- **Elm pattern** — `Model`, `Update`, `View` in separate files
- **Full-screen layout** — content fills terminal height, footer pinned to bottom
- **Status bar** — displays version (with `-dev` suffix for dev builds) and
  update availability (checked async on startup via `internal/updater`)
- **Styled help bar** — keybinds rendered in bold cyan, descriptions in dim text
- **Dynamic width** — adapts to terminal resize via `tea.WindowSizeMsg`

### `internal/updater/`

Non-blocking update checker. On TUI startup, queries the GitHub Releases API
(`/repos/omargallob/devops-starter/releases/latest`) with a 5-second timeout.
Compares the remote tag against the running version using simple semver comparison.
Results are displayed in the TUI status bar when an update is available.

### `pkg/tooldef/`

Public types shared across internal packages:
- `Tool` — name, version, description, group, format, URL template, platform map
- `Platform` — OS + Arch pair
- `Group` — enum for tool categories
- `ArchiveFormat` — `FormatBinary` or `FormatTarGz`

This is the only `pkg/` package, intended to be importable by external tools
if needed.

## Design Decisions

### Direct binary downloads (no package manager)

Package managers introduce variability: different versions across distros,
authentication requirements, and platform gaps. By downloading directly from
GitHub Releases (or official download URLs), we get:
- Consistent versions across all platforms
- No dependency on system package manager state
- Works in minimal containers and CI runners

### Dual build system (Make + Bazel)

- **Make** — developer ergonomics; fast iteration, familiar interface
- **Bazel** — hermetic CI builds, reproducible cross-compilation, deterministic outputs

Contributors can use Make exclusively. Bazel is used in CI for release artifacts.

### Functional options for Installer

The installer uses the functional options pattern (`WithDryRun`,
`WithConcurrency`, etc.) rather than a config struct. This keeps the API
extensible without breaking changes and makes tests cleaner (only set what
you need).

### State store separate from config

Config is *declarative intent* (what should be installed). State is *actual
state* (what is installed). Keeping them separate allows:
- Detecting drift (config says X, state says Y)
- The `adopt` command to work independently of config
- Clean uninstall without modifying user config

### Registry as code (not YAML/JSON)

Tool definitions are Go structs, not a data file. This enables:
- Compile-time type checking
- IDE autocomplete for tool fields
- Complex URL template logic without a DSL
- Easy testing via Go test functions

## Extension Points

| Want to... | Do this |
|---|---|
| Add a new tool | Add a `r.register(...)` call in the appropriate registry file |
| Add a new group | Create `internal/registry/<group>.go`, add the group constant to `pkg/tooldef/group.go`, register it in `registry.go` |
| Support a new platform | Add platform constants to `pkg/tooldef/`, update `internal/platform/` detection |
| Add a new archive format | Add format to `pkg/tooldef/`, implement extraction in `internal/installer/` |
| Add a new CLI command | Create `internal/cli/<cmd>.go` with `newXxxCmd()`, register in `root.go` |
