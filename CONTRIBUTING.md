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
| Bazel / Bazelisk | 9+ | Optional (for hermetic builds) |
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
internal/tui/          Bubbletea interactive components
pkg/tooldef/           Public types shared across packages
```
