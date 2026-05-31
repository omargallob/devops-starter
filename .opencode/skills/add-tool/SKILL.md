---
name: add-tool
description: >
  Add a new tool to an existing group in devops-starter. Use when asked to
  "add a tool", "register a binary", "add kubectl", or when a user wants to
  install a new CLI tool via the devops-starter catalog. Covers tool definition,
  install mode selection, and field requirements.
---

# Add a Tool to devops-starter

Adding a tool to an **existing** group requires editing a single registry file
and running the existing tests to validate. No test changes or BUILD.bazel
changes are needed -- the structural tests automatically pick up and validate
new tools.

If the group does not exist yet, use the `add-tool-group` skill first.

## Quick reference: Required fields

Every tool must have these 5 fields:

| Field | Type | Example |
|-------|------|---------|
| `Name` | `string` | `"kubectl"` |
| `Version` | `string` | `"1.31.4"` (no `v` prefix) |
| `Description` | `string` | `"Kubernetes CLI"` |
| `Group` | `tooldef.Group` | `tooldef.GroupKubernetes` |
| `InstallMode` | `tooldef.InstallMode` | `tooldef.InstallModeEget` |

## Step 1: Choose the install mode

### Decision tree

```
Is it distributed as a GitHub release?
  YES -> Does `eget <owner/repo>` work directly?
    YES -> Use InstallModeEget (simplest)
    NO  -> Does the URL follow a predictable template?
      YES -> Use InstallModeEgetURL
      NO  -> Use InstallModeCustom
  NO -> Is it installed via mise (npm/pip/language runtime)?
    YES -> Use InstallModeMise
    NO  -> Is it a GitHub CLI extension?
      YES -> Use InstallModeGhExtension
      NO  -> Use InstallModeCustom with URLs map
```

### Mode-specific required fields

| Mode | Required | Optional |
|------|----------|----------|
| `eget` | `Repo` | `Asset`, `BinaryName`, `InstallName` |
| `eget-url` | `URLTemplate` or `URLs` + `Format` | `BinaryName`, `InstallName` |
| `custom` | `URLTemplate` or `URLs` + `Format` | `PostInstall`, `BinaryName`, `StripComponents` |
| `mise` | (none extra) | `MiseBackend`, `Dependencies` |
| `gh-extension` | `Repo` | `Dependencies` |

## Step 2: Write the tool definition

Edit the appropriate group file under `internal/registry/`. Find the
`registerXxx` function and add a `r.register(...)` call.

### Pattern A: eget (GitHub release, auto-detect)

Use when `eget <owner/repo>` just works:

```go
r.register(&tooldef.Tool{
    Name:        "k9s",
    Version:     "0.32.7",
    Description: "Kubernetes TUI",
    Group:       tooldef.GroupKubernetes,
    InstallMode: tooldef.InstallModeEget,
    Repo:        "derailed/k9s",
})
```

### Pattern B: eget with Asset glob

Use when the repo has multiple binaries or ambiguous asset names:

```go
r.register(&tooldef.Tool{
    Name:        "kustomize",
    Version:     "5.5.0",
    Description: "Kubernetes configuration management",
    Group:       tooldef.GroupKubernetes,
    InstallMode: tooldef.InstallModeEget,
    Repo:        "kubernetes-sigs/kustomize",
    Asset:       "kustomize_*",
})
```

### Pattern C: eget with BinaryName/InstallName

Use when the binary name inside the archive differs from the tool name:

```go
r.register(&tooldef.Tool{
    Name:        "ripgrep",
    Version:     "14.1.1",
    Description: "Fast grep alternative",
    Group:       tooldef.GroupRustTools,
    InstallMode: tooldef.InstallModeEget,
    Repo:        "BurntSushi/ripgrep",
    Asset:       "*musl*",
    BinaryName:  "rg",       // name inside the archive
    InstallName: "rg",       // name in ~/.local/bin
})
```

### Pattern D: eget-url with URLTemplate (raw binary)

Use for direct-download binaries with a predictable URL:

```go
r.register(&tooldef.Tool{
    Name:        "kubectl",
    Version:     "1.31.4",
    Description: "Kubernetes CLI",
    Group:       tooldef.GroupKubernetes,
    InstallMode: tooldef.InstallModeEgetURL,
    Format:      tooldef.FormatBinary,
    URLTemplate: "https://dl.k8s.io/release/v{{.Version}}/bin/{{.OS}}/{{.Arch}}/kubectl",
})
```

Template variables: `{{.Name}}`, `{{.Version}}`, `{{.OS}}`, `{{.Arch}}`,
`{{.Format}}`, `{{.BinaryName}}`.

### Pattern E: eget-url with URLTemplate (archive)

Use for ZIP/tar.gz downloads with a predictable URL:

```go
r.register(&tooldef.Tool{
    Name:        "terraform",
    Version:     "1.10.4",
    Description: "Infrastructure as Code",
    Group:       tooldef.GroupInfra,
    InstallMode: tooldef.InstallModeEgetURL,
    Format:      tooldef.FormatZip,
    URLTemplate: "https://releases.hashicorp.com/terraform/{{.Version}}/terraform_{{.Version}}_{{.OS}}_{{.Arch}}.zip",
})
```

### Pattern F: eget-url with per-platform URLs map

Use when URLs differ too much per platform for a template:

```go
r.register(&tooldef.Tool{
    Name:        "firebase-cli",
    Version:     "13.29.1",
    Description: "Firebase CLI",
    Group:       tooldef.GroupCloud,
    InstallMode: tooldef.InstallModeEgetURL,
    Format:      tooldef.FormatBinary,
    InstallName: "firebase",
    URLs: map[string]string{
        "linux/amd64":  "https://firebase.tools/bin/linux/v13.29.1",
        "linux/arm64":  "https://firebase.tools/bin/linux/arm64/v13.29.1",
        "darwin/amd64": "https://firebase.tools/bin/macos/v13.29.1",
        "darwin/arm64": "https://firebase.tools/bin/macos/arm64/v13.29.1",
    },
})
```

### Pattern G: custom with PostInstall

Use for tools that need a post-install script (e.g., aws-cli, gcloud):

```go
r.register(&tooldef.Tool{
    Name:        "aws-cli",
    Version:     "2.22.35",
    Description: "AWS CLI",
    Group:       tooldef.GroupCloud,
    InstallMode: tooldef.InstallModeCustom,
    Format:      tooldef.FormatZip,
    InstallName: "aws",
    PostInstall: "./aws/install --install-dir ~/.local/aws-cli --bin-dir ~/.local/bin",
    URLs: map[string]string{
        "linux/amd64":  "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip",
        "linux/arm64":  "https://awscli.amazonaws.com/awscli-exe-linux-aarch64.zip",
        "darwin/amd64": "https://awscli.amazonaws.com/AWSCLIV2.pkg",
        "darwin/arm64": "https://awscli.amazonaws.com/AWSCLIV2.pkg",
    },
})
```

### Pattern H: mise with MiseBackend

Use for tools distributed via npm or pip and installed through mise:

```go
r.register(&tooldef.Tool{
    Name:         "claude-code",
    Version:      "1.0.3",
    Description:  "Anthropic Claude AI coding assistant",
    Group:        tooldef.GroupAI,
    InstallMode:  tooldef.InstallModeMise,
    MiseBackend:  "npm:@anthropic-ai/claude-code",
    Dependencies: []string{"mise"},
})
```

Common `MiseBackend` prefixes: `npm:`, `pipx:`.

### Pattern I: gh-extension

Use for GitHub CLI extensions:

```go
r.register(&tooldef.Tool{
    Name:         "copilot-cli",
    Version:      "1.0.5",
    Description:  "GitHub Copilot in the CLI",
    Group:        tooldef.GroupAI,
    InstallMode:  tooldef.InstallModeGhExtension,
    Repo:         "github/gh-copilot",
    Dependencies: []string{"gh"},
})
```

## Step 3: Handle platform restrictions

If the tool does **not** support all 4 platforms (linux/amd64, linux/arm64,
darwin/amd64, darwin/arm64), set the `Platforms` field:

```go
Platforms: []tooldef.Platform{
    {OS: "linux", Arch: "amd64"},
    {OS: "linux", Arch: "arm64"},
},
```

Omitting `Platforms` means the tool supports all platforms.

## Step 4: Update the function doc comment

Update the doc comment on the `registerXxx` function to include the new tool
in its list. For example:

```go
// registerKubernetes adds Kubernetes ecosystem tools to the registry.
// Includes kubectl, helm, k9s, kustomize, and my-new-tool.
```

## Step 5: Verify

Run:

```bash
go build ./...
go test -race ./internal/registry/...
bazel run //:golangci_lint
```

The existing structural tests automatically validate:

- All 5 required fields are set
- Mode-specific fields are present (e.g., `Repo` for eget)
- `Format` is valid for eget-url/custom tools
- `URLTemplate` renders without errors for all supported platforms
- `Description` is non-empty
- No duplicate tool names exist
- Group is a known constant

No test file changes are needed when adding a tool to an existing group.

## Field reference

| Field | Type | Default | Notes |
|-------|------|---------|-------|
| `Name` | `string` | (required) | Canonical name, used as registry key |
| `Version` | `string` | (required) | Pinned version, no `v` prefix |
| `Description` | `string` | (required) | Short human-readable description |
| `Group` | `Group` | (required) | Must be a `Group*` constant |
| `InstallMode` | `InstallMode` | (required) | `eget`, `eget-url`, `custom`, `mise`, `gh-extension` |
| `Repo` | `string` | | GitHub `owner/repo` for eget and gh-extension |
| `Asset` | `string` | | Eget `--asset` glob to select release asset |
| `URLTemplate` | `string` | | Go template for download URL |
| `URLs` | `map[string]string` | | Per-platform URLs keyed by `"os/arch"` |
| `BinaryName` | `string` | `Name` | Binary name inside archive |
| `InstallName` | `string` | `Name` | Name in `~/.local/bin` |
| `Format` | `ArchiveFormat` | | `tar.gz`, `tar.xz`, `zip`, `binary` |
| `StripComponents` | `int` | `0` | Like `tar --strip-components` |
| `Checksums` | `map[string]string` | | SHA256 per `"os/arch"` |
| `Platforms` | `[]Platform` | all | Restrict to specific OS/arch pairs |
| `PostInstall` | `string` | | Shell command to run after install |
| `Dependencies` | `[]string` | | Tool names that must be installed first |
| `MiseBackend` | `string` | | Mise backend specifier (e.g., `npm:pkg`) |
| `ManagedBy` | `string` | | **Deprecated** -- use `InstallMode` instead |
| `Subgroup` | `string` | | Visual sub-category within a group |
