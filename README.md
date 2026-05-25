# devops-starter

[![CI](https://github.com/omargallob/devops-starter/actions/workflows/ci.yml/badge.svg)](https://github.com/omargallob/devops-starter/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/omargallob/devops-starter?label=version)](https://github.com/omargallob/devops-starter/releases/latest)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Coverage](https://raw.githubusercontent.com/omargallob/devops-starter/badges/badges/coverage.svg)](https://github.com/omargallob/devops-starter/actions/workflows/ci.yml)

An opinionated, cross-platform DevOps tool installer and dotfile manager.
Downloads and installs 61 pre-built CLI binaries (no package manager required) and manages shell/editor configuration via symlinks.
One command to bootstrap a fully-configured development workstation.

---

## Table of Contents

- [Why devops-starter?](#why-devops-starter)
- [Quick Install](#quick-install)
- [Build from Source](#build-from-source)
- [Usage](#usage)
- [Configuration](#configuration)
- [Tool Catalog](#tool-catalog)
- [Dotfiles](#dotfiles)
- [Project Structure](#project-structure)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

## Why devops-starter?

Setting up a new machine (or a CI runner) with a consistent DevOps toolchain is tedious:
different package managers, outdated repos, manual binary downloads, and scattered dotfiles.

**devops-starter** solves this by:

1. Downloading tools **directly from upstream releases** -- no Homebrew, apt, or pacman needed.
2. Verifying every download with **SHA256 checksums**.
3. Running downloads **concurrently** with progress bars.
4. Managing **dotfiles** (shell, git, tmux, starship, Neovim) via symlinks with automatic backup.
5. Providing a single **YAML config** to control everything.

## Quick Install

```sh
curl -fsSL https://raw.githubusercontent.com/omargallob/devops-starter/main/scripts/install.sh | sh
```

This detects your OS and architecture, downloads the correct binary, and places it in `~/.local/bin`.
Make sure `~/.local/bin` is in your `$PATH`.

## Build from Source

### Prerequisites

- Go 1.26+
- (Optional) Bazel 9 (installed via [Bazelisk](https://github.com/bazelbuild/bazelisk)) for hermetic builds

### Using Make

```sh
make build          # compile for current platform
make test           # run tests with race detector
make lint           # golangci-lint + shellcheck
make install        # build and copy to ~/.local/bin
make help           # list all available targets
```

### Using Bazel

```sh
bazel build //cmd/devops-starter
bazel test //...
make bazel-release   # cross-compile all platforms
```

## Usage

### Guided Setup

```sh
# Interactive wizard: groups, install dir, dotfiles
devops-starter setup

# Non-interactive (uses existing config or defaults)
devops-starter setup --non-interactive
```

See [docs/setup.md](docs/setup.md) for a detailed walkthrough of each wizard screen.

### Install Tools

```sh
# Install all enabled tools for your platform
devops-starter install

# Preview without installing
devops-starter install --dry-run

# Install only a specific group
devops-starter install --only kubernetes
```

### List Available Tools

```sh
# Show all tools with installation status
devops-starter list
```

### Adopt System Tools

```sh
# Adopt a specific tool already on your system
devops-starter adopt kubectl helm

# Adopt all detected system tools
devops-starter adopt --all-detected
```

### Remove Managed Tools

```sh
# Remove managed binaries (reverts to system version if available)
devops-starter remove terraform packer
```

### Status Dashboard

```sh
# Launch interactive full-screen TUI showing tool state
devops-starter status

# Plain text output (CI-friendly)
devops-starter status --no-tui

# Detect actual installed versions
devops-starter status --verify
```

The TUI features:
- Full-screen layout that adapts to terminal size
- Category-based navigation with install/remove/verify actions
- Status bar showing the running version and update availability
- Automatic update check (non-blocking) against GitHub Releases

### Manage Dotfiles

```sh
# Create symlinks from repo dotfiles to $HOME
devops-starter dotfiles link

# Remove symlinks managed by devops-starter
devops-starter dotfiles unlink

# Show current symlink status
devops-starter dotfiles status

# Specify a custom dotfiles source directory
devops-starter dotfiles link --source /path/to/dotfiles
```

### System Health Check

```sh
# Verify PATH, git, curl, and platform detection
devops-starter doctor
```

### Configuration Management

```sh
# Create default config at ~/.config/devops-starter/config.yaml
devops-starter config init

# Show current configuration
devops-starter config show
```

### Global Flags

| Flag | Description |
|------|-------------|
| `--config <path>` | Use a custom config file |
| `--dry-run` | Preview actions without executing |
| `--yes`, `-y` | Skip confirmation prompts |

## Configuration

The config file lives at `~/.config/devops-starter/config.yaml` (respects `$XDG_CONFIG_HOME`).

```yaml
# Where binaries are installed (must be in $PATH)
install_dir: ~/.local/bin

# Enable or disable tool groups
groups:
  languages: true
  containers: true
  kubernetes: true
  infra: true
  cloud: true
  rust_tools: true
  utilities: true

# Per-tool overrides
overrides:
  terraform:
    version: "1.7.5"     # Pin a specific version
  aws-cli:
    disabled: true       # Skip this tool entirely
```

Run `devops-starter config init` to generate the default config.

## Tool Catalog

61 tools across 7 groups:

| Group | Tools |
|-------|-------|
| **languages** | mise, + all runtimes from `.mise.toml` (e.g., go, python, node) |
| **containers** | docker, docker-compose, nerdctl |
| **kubernetes** | kubectl, helm, kustomize, k9s, kubectx, kubens, stern, argocd, flux, istioctl, cilium, kind, kubeseal, velero |
| **infra** | terraform, opentofu, pulumi, packer, vault, consul |
| **cloud** | aws-cli, eksctl |
| **rust-tools** | bat, eza, fd, ripgrep, delta, zoxide, starship, tokei, hyperfine, procs, bottom, gitui, dust, bandwhich, sd, tealdeer, xh, yazi, atuin, zellij, just, watchexec |
| **utilities** | jq, yq, fzf, direnv, age, sops, gh, trivy, lazygit, shellcheck, shfmt, task, neovim |

### Platform Availability

Most tools provide pre-built binaries for all four platforms. Exceptions:

| Tool | linux/amd64 | linux/arm64 | darwin/amd64 | darwin/arm64 |
|------|:-----------:|:-----------:|:------------:|:------------:|
| bat | yes | yes | yes | - |
| dust | yes | yes | yes | - |
| tokei | yes | yes | yes | - |
| procs | - | - | yes | yes |
| neovim | yes | - | yes | yes |
| trivy | yes | yes | yes | yes |

All other tools support linux/amd64, linux/arm64, darwin/amd64, and darwin/arm64.

Want a tool added? See [CONTRIBUTING.md](CONTRIBUTING.md#adding-a-new-tool) for how to add tools to the registry.

## Dotfiles

The `dotfiles/` directory contains managed configuration files:

| Directory | Contents |
|-----------|----------|
| `zsh/` | `.zshrc`, `.zprofile` |
| `bash/` | `.bashrc`, `.bash_profile` |
| `shared/` | Common aliases, environment variables, functions (sourced by both shells) |
| `git/` | `.gitconfig` (delta pager, histogram diff, rebase-on-pull), `.gitignore_global` |
| `tmux/` | `.tmux.conf` (Ctrl+a prefix, vim-style navigation) |
| `starship/` | `starship.toml` (DevOps-focused prompt with k8s, AWS, terraform context) |
| `nvim/` | `init.lua` + lazy.nvim plugin setup |

Shell configs replace common coreutils with Rust alternatives (ls->eza, cat->bat, grep->rg, find->fd) and initialise starship, zoxide, direnv, atuin, and mise.

When linking, existing files are backed up to `~/.dotfiles.bak` before being replaced with symlinks.

## Project Structure

```
cmd/devops-starter/       Entry point (main package)
internal/
  cli/                    Cobra commands: setup, install, list, adopt, remove, status, dotfiles, doctor, config
  config/                 YAML configuration loading/saving
  dotfiles/               Symlink manager with backup/restore
  installer/              Download, checksum, archive extraction, install orchestration
  platform/               OS/arch/distro detection
  registry/               Tool definitions organised by group
  state/                  Persistent state store for managed tools
  tui/                    Bubbletea interactive UI (full-screen, status bar, update check)
  updater/                Async update checker (GitHub Releases API)
pkg/tooldef/              Public types: Tool, Platform, Group, ArchiveFormat
configs/                  Default config template
docs/                     Documentation (setup wizard walkthrough, etc.)
dotfiles/                 Managed dotfiles (shell, git, tmux, starship, nvim)
scripts/                  Bootstrap install script
oci/                      Container test image definitions (Ubuntu, Arch)
platforms/                Bazel cross-compilation platform configs
bazel/                    Custom Bazel rules (shellcheck, golangci-lint) and macros
.github/workflows/        CI, release, and PR title enforcement workflows
```

## Development

```sh
make help           # list all targets
make check          # fmt + vet + lint + test (full local validation)
make test-coverage  # tests with coverage report
make release        # cross-compile release binaries
make gazelle        # regenerate Bazel BUILD files
make setup          # install dev tools and pre-commit hooks
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full development guide and [ARCHITECTURE.md](ARCHITECTURE.md) for design decisions.

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) before submitting a PR.

## License

[MIT](LICENSE) - Copyright (c) 2026 Omar Gallo
