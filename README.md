# devops-starter

[![CI](https://github.com/omargallob/devops-starter/actions/workflows/ci.yml/badge.svg)](https://github.com/omargallob/devops-starter/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Coverage](badges/coverage.svg)](https://github.com/omargallob/devops-starter/actions/workflows/ci.yml)

An opinionated, cross-platform DevOps tool installer and dotfile manager. Downloads and installs 60+ pre-built CLI binaries (no package manager required) and manages shell/editor configuration via symlinks. One command to bootstrap a fully-configured development workstation.

## Features

- **60+ tools** across 7 categories: languages, containers, Kubernetes, infrastructure, cloud, Rust-based CLI tools, and utilities
- **Cross-platform**: Linux (Ubuntu, Arch) and macOS (Intel, Apple Silicon)
- **No package manager dependency**: downloads pre-built binaries directly from upstream releases
- **SHA256 checksum verification** for download integrity
- **Concurrent downloads** (4 by default) with progress bars
- **Dotfile management**: symlinks shell, git, tmux, starship, and Neovim configs with automatic backup of existing files
- **Configuration-driven**: YAML config to enable/disable groups and pin tool versions
- **Hermetic builds** with Bazel 9 for reproducibility and cross-compilation
- **CI/CD** via GitHub Actions (test, lint, build matrix, tag-triggered releases)

## Quick Install

```sh
curl -fsSL https://raw.githubusercontent.com/omargallob/devops-starter/main/scripts/install.sh | sh
```

This downloads the correct binary for your platform to `~/.local/bin`.

## Build from Source

### Prerequisites

- Go 1.26+
- (Optional) Bazel 9 / Bazelisk for hermetic builds

### Using Make

```sh
# Build the binary
make build

# Run tests
make test

# Lint (requires golangci-lint + shellcheck)
make lint

# Install to ~/.local/bin
make install
```

### Using Bazel

```sh
# Build
bazelisk build //cmd/devops-starter

# Test
bazelisk test //...

# Cross-compile all platforms
make bazel-release
```

## Usage

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

### Configuration

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
  ansible: true
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

## Tool Groups

| Group | Tools |
|-------|-------|
| **languages** | mise |
| **containers** | docker, docker-compose, nerdctl |
| **kubernetes** | kubectl, helm, kustomize, k9s, kubectx, kubens, stern, argocd, flux, istioctl, cilium, kind, kubeseal, velero |
| **infra** | terraform, opentofu, pulumi, packer, vault, consul |
| **cloud** | aws-cli, eksctl |
| **rust-tools** | bat, eza, fd, ripgrep, delta, zoxide, starship, tokei, hyperfine, procs, bottom, gitui, dust, bandwhich, sd, tealdeer, xh, yazi, atuin, zellij, just, watchexec |
| **utilities** | jq, yq, fzf, direnv, age, sops, gh, trivy, lazygit, shellcheck, shfmt, task, neovim |

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
cmd/devops-starter/       Entry point
internal/
  cli/                    Cobra command definitions (install, list, dotfiles, doctor, config)
  config/                 YAML configuration loading/saving
  dotfiles/               Symlink manager with backup/restore
  installer/              Download, checksum, archive extraction, install orchestration
  platform/               OS/arch/distro detection
  registry/               Tool definitions organised by group
pkg/tooldef/              Public types: Tool, Platform, Group, ArchiveFormat
configs/                  Default config template
dotfiles/                 Managed dotfiles (shell, git, tmux, starship, nvim)
scripts/                  Bootstrap install script
oci/                      Container test image definitions (Ubuntu, Arch)
platforms/                Bazel cross-compilation platform configs
bazel/                    Custom Bazel rules (shellcheck, golangci-lint) and macros
.github/workflows/        CI and release workflows
```

## Development

```sh
# Format code
make fmt

# Run vet + linters
make lint

# Run tests with race detector
make test

# Cross-compile release binaries (no Bazel)
make release

# Regenerate Bazel BUILD files
make gazelle
```

## License

[MIT](LICENSE) - Copyright (c) 2026 Omar Gallo
