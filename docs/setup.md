# Setup Command

The `devops-starter setup` command is an interactive guided wizard that walks you
through the complete initial configuration of devops-starter. It is safe to
re-run at any time to modify your settings without losing per-tool overrides.

## Quick Start

```sh
devops-starter setup
```

## Usage

```
devops-starter setup [flags]

Flags:
      --non-interactive   use existing config/defaults without prompts (for CI)
      --dry-run           preview actions without executing
      --config string     config file (default: ~/.config/devops-starter/config.yaml)
```

## Wizard Screens

The setup wizard progresses through the following screens. You can navigate
forward with `enter` and backward with `esc` at any point.

### 1. Welcome & PATH Check

Displays an overview of what the wizard will configure and checks whether your
install directory is present in `$PATH`.

If the install directory is not in PATH, a warning is shown. The wizard still
proceeds -- you can add it to PATH afterward.

**Controls:** `enter` to continue, `q`/`esc` to quit.

### 2. Tool Group Selection

Presents all available tool groups as a checkbox list:

- languages
- containers
- kubernetes
- infra
- cloud
- ansible
- rust-tools
- utilities

Groups are pre-selected based on your existing config (or all enabled for
first-time users).

**Controls:**
- `j`/`k` or arrow keys to navigate
- `space` to toggle a group
- `a` to select all
- `n` to deselect all
- `enter` to proceed
- `esc` to go back

### 3. Install Directory

A text input field for the binary install directory. Defaults to `~/.local/bin`.

The `~` prefix is expanded automatically. After confirming, the PATH check is
re-evaluated against the new value.

**Controls:** type to edit, `enter` to confirm, `esc` to go back.

### 4. Dotfiles

Select which dotfile categories to symlink:

- shell (zsh/bash)
- git
- tmux
- starship
- neovim

**Controls:**
- `j`/`k` to navigate
- `space` to toggle
- `a` all, `n` none
- `enter` to proceed with linking
- `s` to skip dotfiles entirely
- `esc` to go back

### 5. Confirmation

Displays a summary of all selections:
- Install directory
- Enabled groups
- Dotfile categories (or "skipped")

Also shows where the config file will be saved.

**Controls:** `y`/`enter` to confirm, `n`/`esc` to go back.

### 6. Done

After confirmation, the config is saved. The done screen shows:
- Success message
- Next steps (`devops-starter install`, `devops-starter dotfiles link`)
- PATH warning if applicable

## Config Impact

The wizard writes to `~/.config/devops-starter/config.yaml` (or
`$XDG_CONFIG_HOME/devops-starter/config.yaml`). The resulting file looks like:

```yaml
install_dir: /home/user/.local/bin
groups:
  languages: true
  containers: true
  kubernetes: true
  infra: false
  cloud: false
  ansible: true
  rust_tools: true
  utilities: true
overrides:
  terraform:
    version: "1.8.0"
  kubectl:
    disabled: true
```

### What Gets Modified

| Field | Modified by setup? |
|-------|-------------------|
| `install_dir` | Yes |
| `groups.*` | Yes |
| `overrides` | Never -- preserved across re-runs |

## Re-running Setup

Running `devops-starter setup` again:

1. Loads your existing config
2. Pre-selects groups based on current settings
3. Pre-fills the install directory
4. Lets you change any selections
5. Saves without touching `overrides`

This means version pins and per-tool disables you've set manually are always
preserved.

## Non-interactive / CI Usage

For scripted environments, use `--non-interactive` to skip the TUI and
save the existing config (or defaults if no config exists):

```sh
# Use defaults
devops-starter setup --non-interactive

# Preview only
devops-starter setup --non-interactive --dry-run

# With custom config path
devops-starter setup --non-interactive --config /path/to/config.yaml
```

In non-interactive mode, the command:
- Loads existing config if present (preserving all settings)
- Falls back to `DefaultConfig()` if no config file exists
- Saves the config (unless `--dry-run`)

## Troubleshooting

### PATH not configured

If you see `! PATH check: ~/.local/bin is NOT in your PATH`, add the following
to your shell profile:

```sh
# ~/.zshrc or ~/.bashrc
export PATH="$HOME/.local/bin:$PATH"
```

Then restart your shell or run `source ~/.zshrc`.

### Permission denied

If saving the config fails with a permission error, ensure you have write access
to the config directory:

```sh
mkdir -p ~/.config/devops-starter
chmod 755 ~/.config/devops-starter
```

### Overrides were lost

This should not happen with the setup wizard. If it does, check that you're not
running another process that overwrites the config file concurrently. The
`MergeGroups` function explicitly preserves the `overrides` map.
