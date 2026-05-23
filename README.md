# devops-starter

Opinionated cross-platform DevOps tool installer and dotfile manager. Installs 60+ tools as pre-built binaries with no package manager dependency.

## Supported Platforms

| OS | Architecture | Distro |
|----|-------------|--------|
| Linux | amd64, arm64 | Ubuntu, Arch |
| macOS | amd64 (Intel), arm64 (Apple Silicon) | - |

## Quick Install

```sh
curl -fsSL https://raw.githubusercontent.com/omargallob/devops-starter/main/scripts/install.sh | sh
```

Or build from source:

```sh
git clone https://github.com/omargallob/devops-starter.git
cd devops-starter
make install
```

## Usage

```sh
# Install all tools
devops-starter install

# Install only a specific group
devops-starter install --only kubernetes
devops-starter install --only rust-tools

# Preview what would be installed
devops-starter install --dry-run

# List all available tools
devops-starter list

# Manage dotfiles
devops-starter dotfiles link
devops-starter dotfiles status
devops-starter dotfiles unlink

# Check system health
devops-starter doctor

# Initialize/view configuration
devops-starter config init
devops-starter config show
```

## Tool Groups

| Group | Tools |
|-------|-------|
| **languages** | mise (Go, Node, Ruby, Python, Rust) |
| **containers** | docker, docker-compose, nerdctl |
| **kubernetes** | kubectl, helm, kustomize, k9s, kubectx, kubens, stern, argocd, flux, istioctl, cilium, kind, kubeseal, velero |
| **infra** | terraform, opentofu, pulumi, packer, vault, consul |
| **cloud** | aws-cli, eksctl |
| **rust-tools** | bat, eza, fd, ripgrep, delta, zoxide, starship, tokei, hyperfine, procs, bottom, gitui, dust, bandwhich, sd, tealdeer, xh, yazi, atuin, zellij, just, watchexec |
| **utilities** | jq, yq, fzf, direnv, age, sops, gh, trivy, lazygit, shellcheck, shfmt, task, neovim |

## Configuration

Configuration lives at `~/.config/devops-starter/config.yaml`:

```yaml
install_dir: ~/.local/bin

groups:
  languages: true
  containers: true
  kubernetes: true
  infra: true
  cloud: false        # disable on homelab
  ansible: true
  rust_tools: true
  utilities: true

overrides:
  terraform:
    version: "1.7.5"  # pin specific version
  aws-cli:
    disabled: true    # skip this tool
```

## Dotfiles

Managed configurations for:

- **Shell**: ZSH + Bash with shared aliases, env, and functions
- **Git**: delta pager, histogram diff, rebase-on-pull
- **Tmux**: Ctrl+a prefix, vim navigation, 256color
- **Starship**: DevOps-focused prompt (k8s, aws, terraform, git)
- **Neovim**: lazy.nvim with treesitter, telescope, LSP, gitsigns

Aliases include Rust-based replacements: `ls`→eza, `cat`→bat, `grep`→ripgrep, `find`→fd, `cd`→zoxide.

## Building with Bazel

This project uses Bazel 9 with Bzlmod for hermetic, reproducible builds:

```sh
# Build
bazelisk build //cmd/devops-starter

# Test
bazelisk test //...

# Cross-compile
bazelisk build //cmd/devops-starter --config=linux_arm64

# Regenerate BUILD files
bazelisk run //:gazelle

# Build container test images
bazelisk build //oci:ubuntu_test_image_tarball
```

### Prerequisites

- [Bazelisk](https://github.com/bazelbuild/bazelisk) (manages Bazel version)
- Go 1.25+ (for `go test` outside Bazel)

## Development

```sh
make test          # Run all Go tests
make lint          # golangci-lint + shellcheck
make build         # Build binary for current platform
make release       # Cross-compile all 4 platforms
make gazelle       # Regenerate BUILD.bazel files
make fmt           # Format Go + shell
```

## Architecture

```
cmd/devops-starter/     CLI entry point
internal/
  cli/                  Cobra command implementations
  config/               YAML configuration management
  dotfiles/             Symlink manager with backup
  installer/            Download, verify, extract, install
  platform/             OS/arch/distro detection + system deps
  registry/             Tool definitions (60+ tools)
pkg/tooldef/            Public tool definition types
configs/                Default configuration
dotfiles/               Dotfile content (zsh, bash, git, tmux, starship, nvim)
oci/                    Container images for integration testing
platforms/              Bazel cross-compilation targets
```

## License

See [LICENSE](LICENSE) for details.
