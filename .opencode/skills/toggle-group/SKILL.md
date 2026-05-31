---
name: toggle-group
description: >
  Activate or deactivate tool groups in devops-starter. Use when asked to
  "enable a group", "disable a group", "turn off kubernetes", "activate AI
  tools", "deactivate containers", or any request to toggle which tool
  categories are installed. Covers config file editing, CLI commands, and
  the effects on install and status.
---

# Toggle Tool Groups in devops-starter

Tool groups are enabled or disabled via the user's config file. When a group
is disabled, its tools are skipped during `install` and shown as `disabled`
in `status`.

## Config file location

The config lives at one of:

| Priority | Path |
|----------|------|
| 1 (highest) | Path passed via `--config` flag |
| 2 | `$XDG_CONFIG_HOME/devops-starter/config.yaml` |
| 3 (default) | `~/.config/devops-starter/config.yaml` |

If the file does not exist, all groups default to **enabled** (except groups
explicitly defaulted to `false` in `DefaultConfig()`).

## Method 1: Edit the config YAML (recommended for agents)

The simplest and most reliable approach. Edit (or create) the config file
directly.

### Step 1: Locate or create the config file

```bash
# Check if it exists
ls ~/.config/devops-starter/config.yaml

# If not, initialize it
devops-starter config init
```

### Step 2: Edit the groups section

The `groups` key maps group names (snake_case) to booleans:

```yaml
groups:
  languages: true
  containers: true
  kubernetes: true
  infra: true
  cloud: true
  ansible: true
  rust_tools: true
  utilities: true
```

To disable a group, set it to `false`:

```yaml
groups:
  kubernetes: false    # tools in this group will be skipped
```

To re-enable, set it back to `true`.

### Step 3: Verify

```bash
# Check effective config
devops-starter config show

# Check which tools are disabled
devops-starter list
```

### Available group names (YAML keys)

| YAML key | Group constant | Contains |
|----------|---------------|----------|
| `languages` | `GroupLanguages` | Go, Node, Python, Java, etc. |
| `containers` | `GroupContainers` | Docker, Podman, Buildah, nerdctl |
| `kubernetes` | `GroupKubernetes` | kubectl, helm, k9s, kustomize |
| `infra` | `GroupInfra` | Terraform, Pulumi, Packer |
| `cloud` | `GroupCloud` | aws-cli, gcloud, az, firebase |
| `ansible` | `GroupAnsible` | Ansible, ansible-lint, molecule |
| `rust_tools` | `GroupRustTools` | ripgrep, fd, bat, bottom, etc. |
| `utilities` | `GroupUtilities` | fzf, jq, yq, shellcheck, etc. |

Note: Use **snake_case** in YAML (`rust_tools`), not kebab-case.

## Method 2: Interactive TUI wizard

```bash
devops-starter setup
```

This launches a full-screen wizard with checkbox group selection:

- `j`/`k` or arrows to navigate
- `space` to toggle a group
- `a` to enable all
- `n` to disable all
- `enter` to proceed

On confirm, the wizard calls `MergeGroups()` which preserves any per-tool
overrides (version pins, per-tool disables) while updating group toggles.

## Method 3: Disable individual tools (not entire groups)

To disable a specific tool without disabling its whole group, use
`overrides` in the config:

```yaml
overrides:
  terraform:
    disabled: true
  helm:
    disabled: true
```

This keeps the rest of the group active.

## How disabling affects behavior

### `devops-starter install`

Disabled groups are filtered out in `filterToolsForInstall()`
(`internal/cli/install.go`). The filter chain is:

1. `--only` flag (if set, skip tools not in that group)
2. `cfg.IsGroupEnabled()` -- **disabled groups are skipped here**
3. Platform support check
4. Per-tool `overrides.disabled` check

Tools in disabled groups are silently excluded from installation.

### `devops-starter status` / `devops-starter list`

Disabled tools are still shown but marked with `StatusDisabled`:
- In the TUI: faint/strikethrough styling
- In `list` output: `- ` prefix with faint style
- In `--no-tui` table: desired version shows `-`

### `devops-starter setup`

The setup wizard reflects current group state in its checkboxes.
Previously disabled groups show as unchecked.

## Programmatic toggle (Go API)

For code changes that need to toggle groups:

```go
// Load existing config (returns defaults if file missing)
cfg, err := config.Load(config.Path())

// Toggle a single group
cfg.SetGroup("kubernetes", false)

// Or batch toggle
cfg.MergeGroups(map[string]bool{
    "kubernetes": false,
    "cloud":      true,
})

// Persist to disk (creates parent dirs if needed)
err = config.Save(cfg, config.Path())
```

Key functions in `internal/config/config.go`:

| Function | Purpose |
|----------|---------|
| `IsGroupEnabled(group)` | Check if a group is active |
| `SetGroup(group, bool)` | Toggle a single group |
| `MergeGroups(map)` | Batch toggle, preserves overrides |
| `AllGroupNames()` | Ordered list of all group name strings |
| `Load(path)` | Load config (defaults if missing) |
| `Save(cfg, path)` | Write config to disk |

Note: `SetGroup` and `IsGroupEnabled` accept both `"rust-tools"` (kebab)
and `"rust_tools"` (snake) forms for groups with hyphens.

## Troubleshooting

**Group disabled but tool still installed?**
Disabling a group does not uninstall tools. It only prevents future installs
and marks them as disabled in status. Previously installed binaries remain
in the install directory.

**Config file not taking effect?**
Check that the file path matches what the CLI is reading:
```bash
devops-starter config show
```
The `--config` flag overrides the default path.

**Group name not recognized?**
Use the YAML key form (snake_case): `rust_tools`, not `rust-tools`.
Check `AllGroupNames()` in `internal/config/config.go` for the canonical list.
