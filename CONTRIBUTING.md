# Contributing to devops-starter

Thank you for considering a contribution. This document explains how to set up
a development environment, the conventions we follow, and the process for
submitting changes.

---

## Table of Contents

- [Prerequisites](#prerequisites)
- [Development Setup](#development-setup)
- [Running Tests](#running-tests)
- [Code Style and Linting](#code-style-and-linting)
- [Commit Messages](#commit-messages)
- [Pull Request Process](#pull-request-process)
- [Adding a New Tool](#adding-a-new-tool)
- [Project Layout](#project-layout)

## Prerequisites

| Tool | Version | Required? |
|------|---------|-----------|
| Go | 1.26+ | Yes |
| mise | latest | Recommended (manages Node, Python for hooks) |
| Bazel (via Bazelisk) | 9+ | Optional (for hermetic builds) |
| golangci-lint | latest | For `make lint` |
| shellcheck | latest | For `make lint` |
| pre-commit | latest | For commit hooks |

## Development Setup

```sh
# Clone and enter the repo
git clone https://github.com/omargallob/devops-starter.git
cd devops-starter

# Install dev tools and pre-commit hooks
make setup

# Verify everything works
make check
```

`make setup` installs runtime versions via mise, installs pre-commit, and
activates commit-msg and pre-commit hooks.

## Running Tests

```sh
make test            # race-enabled tests
make test-coverage   # tests + coverage report
make bazel-test      # hermetic tests via Bazel
```

## Code Style and Linting

- Go code: `gofmt` + `golangci-lint` (config in `.golangci.yml`)
- Shell scripts: `shellcheck` + `shfmt`
- YAML files: `yamllint`
- GitHub Actions: `actionlint`

Run everything at once:

```sh
make check    # fmt + vet + lint + test
```

Or use pre-commit to run hooks on staged files:

```sh
make lint-local   # runs all pre-commit hooks on all files
```

## Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/) enforced by
commitlint (config: `.commitlintrc.yaml`).

Format:

```
<type>(<scope>): <short description>

[optional body]

[optional footer]
```

Allowed types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`,
`build`, `ci`, `chore`, `revert`.

Examples:

```
feat(registry): add podman to containers group
fix(installer): handle redirect URLs for GitHub releases
docs: update tool catalog in README
chore(deps): bump cobra to v1.11.0
```

## Pull Request Process

1. **Branch** from `main` with a descriptive name (e.g., `feat/add-podman`,
   `fix/arm64-checksum`).
2. **Make your changes** with tests where applicable.
3. **Run `make check`** to verify locally.
4. **Push** and open a PR against `main`.
5. **PR title** must follow the conventional commit format (enforced by CI via
   `.github/workflows/pr-title.yml`).
6. CI runs tests, linting, and builds on both Go and Bazel. All checks must
   pass.
7. A maintainer will review and merge via squash-merge.

## Adding a New Tool

To add a tool to the registry:

1. Identify which group it belongs to (see `internal/registry/`).
2. Open the appropriate file (e.g., `internal/registry/utilities.go`).
3. Add a new `r.register(&tooldef.Tool{...})` entry:

```go
r.register(&tooldef.Tool{
    Name:        "mytool",
    Version:     "1.2.3",
    Description: "Short description of mytool",
    Group:       tooldef.GroupUtilities,
    Format:      tooldef.FormatTarGz,  // or FormatBinary
    URLTemplate: "https://github.com/org/mytool/releases/download/v{{.Version}}/mytool-{{.OS}}-{{.Arch}}.tar.gz",
    BinaryName:  "mytool",  // path within archive (omit if FormatBinary)
})
```

4. If the tool has per-platform URLs that don't follow a template, use the
   `URLs` map instead of `URLTemplate`:

```go
URLs: map[string]string{
    "linux/amd64":  "https://example.com/mytool-linux-x86_64.tar.gz",
    "darwin/arm64": "https://example.com/mytool-macos-aarch64.tar.gz",
},
```

5. Run `make test` to ensure registry loads correctly.
6. Run `make gazelle` if you added new files.
7. Update the Tool Catalog table in `README.md`.

### Tool Fields Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `Name` | string | Yes | Canonical tool name (used as install filename) |
| `Version` | string | Yes | Pinned version |
| `Description` | string | Yes | Short human-readable description |
| `Group` | Group | Yes | Functional category (see below) |
| `Format` | ArchiveFormat | Yes | `FormatTarGz`, `FormatTarXz`, `FormatZip`, or `FormatBinary` |
| `URLTemplate` | string | * | Go template URL (mutually exclusive with `URLs`) |
| `URLs` | map[string]string | * | Per-platform URL overrides (keys: `"linux/amd64"`, etc.) |
| `BinaryName` | string | No | Path to binary inside archive (defaults to `Name`) |
| `InstallName` | string | No | Filename in install dir (defaults to `Name`) |
| `StripComponents` | int | No | Leading path components to strip on extraction |
| `Checksums` | map[string]string | No | SHA256 hex digests per platform |
| `ManagedBy` | string | No | Delegate to external manager (e.g., `"mise"`) |
| `PostInstall` | string | No | Shell command to run after install |
| `Dependencies` | []string | No | Tools that must be installed first |

\* One of `URLTemplate` or `URLs` is required (unless `ManagedBy` is set).

### Template Variables

Available variables in `URLTemplate`:

| Variable | Value |
|----------|-------|
| `{{.Name}}` | Tool name |
| `{{.Version}}` | Tool version |
| `{{.OS}}` | `linux` or `darwin` |
| `{{.Arch}}` | `amd64` or `arm64` |
| `{{.Format}}` | Archive extension (e.g., `tar.gz`) |
| `{{.BinaryName}}` | Binary name |

### Available Groups

| Constant | Use for |
|----------|---------|
| `GroupLanguages` | Language runtimes (Go, Node, Python) |
| `GroupContainers` | Container tools (Docker, Podman, Buildah) |
| `GroupKubernetes` | Kubernetes ecosystem (kubectl, helm, k9s) |
| `GroupInfra` | Infrastructure-as-code (Terraform, Pulumi) |
| `GroupCloud` | Cloud CLIs (aws, gcloud, az) |
| `GroupAnsible` | Ansible and related tools |
| `GroupRustTools` | Rust-based dev tools (ripgrep, fd, bat) |
| `GroupUtilities` | General utilities, linters, formatters |

### Multi-Binary Tools

When a single archive contains multiple useful binaries, register each
binary as a separate tool pointing to the same URL:

```go
// lcov and genhtml come from the same tarball
r.register(&tooldef.Tool{
    Name:            "lcov",
    Version:         "2.4",
    Description:     "Linux code coverage reporting tool",
    Group:           tooldef.GroupUtilities,
    Format:          tooldef.FormatTarGz,
    URLTemplate:     "https://github.com/linux-test-project/lcov/releases/download/v{{.Version}}/lcov-{{.Version}}.tar.gz",
    BinaryName:      "bin/lcov",
    StripComponents: 1,
})

r.register(&tooldef.Tool{
    Name:            "genhtml",
    Version:         "2.4",
    Description:     "Generate HTML coverage reports from lcov data",
    Group:           tooldef.GroupUtilities,
    Format:          tooldef.FormatTarGz,
    URLTemplate:     "https://github.com/linux-test-project/lcov/releases/download/v{{.Version}}/lcov-{{.Version}}.tar.gz",
    BinaryName:      "bin/genhtml",
    StripComponents: 1,
})
```

Key points:
- `StripComponents: 1` removes the version-prefixed directory (e.g., `lcov-2.4/`)
- `BinaryName: "bin/lcov"` specifies the path *after* stripping to the desired binary
- Each tool is installed independently (the tarball is downloaded per tool)

### Understanding `StripComponents`

Many release tarballs wrap files in a version-prefixed directory:

```
mytool-1.2.3/
  bin/
    mytool
    mytool-helper
  README.md
```

Setting `StripComponents: 1` removes the top-level `mytool-1.2.3/` prefix, so
the installer sees `bin/mytool` as the binary path. Combine with
`BinaryName: "bin/mytool"` to extract the correct file.

### Managed Tools

For tools managed by an external tool manager (e.g., mise), set `ManagedBy`
instead of providing a URL:

```go
r.register(&tooldef.Tool{
    Name:        "go",
    Version:     "1.24.3",
    Description: "Go programming language",
    Group:       tooldef.GroupLanguages,
    ManagedBy:   "mise",
})
```

These tools are installed via `mise install` rather than direct download.

## Project Layout

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed design documentation.

Quick reference:

```
cmd/devops-starter/    CLI entry point
internal/cli/          Command definitions (one file per command)
internal/config/       YAML config loading
internal/installer/    Download + install orchestration
internal/registry/     Tool definitions by group
internal/platform/     OS/arch detection
internal/dotfiles/     Symlink management
internal/state/        Persistent state store
internal/tui/          Bubbletea interactive TUI (full-screen, status bar)
internal/updater/      Async update checker (GitHub Releases API)
pkg/tooldef/           Public types shared across packages
```
