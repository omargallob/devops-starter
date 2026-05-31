---
name: add-tool-group
description: >
  Add a new tool group to devops-starter. Use when asked to "add a group",
  "create a new tool category", "register a new group", or when adding tools
  that need a group that does not exist yet. Covers all registration points
  across contracts, config, registry, state, TUI, tests, and Bazel.
---

# Add a Tool Group to devops-starter

Adding a new tool group requires changes to **10+ files** across 18 insertion
points. Follow every step below or the group will be partially broken.

## Naming conventions

Decide on the group name first. All three forms are derived from one name:

| Form | Where used | Example |
|------|-----------|---------|
| Go constant | `pkg/tooldef/tool.go` | `GroupMyGroup` (PascalCase) |
| String value | Group constant value, CLI `--only`, state | `"my-group"` (kebab-case) |
| YAML key | `configs/default.yaml`, `GroupConfig` struct tag | `my_group` (snake_case) |

If the name contains a hyphen, `IsGroupEnabled()` and `SetGroup()` must
handle **both** `"my-group"` and `"my_group"` variants (see step 6 and 7).

## Step-by-step workflow

### Step 1: Define the Group constant

**File:** `pkg/tooldef/tool.go`

Add a new constant after `GroupUtilities` in the `const` block:

```go
GroupUtilities  Group = "utilities"
GroupMyGroup    Group = "my-group"    // <-- add this
```

### Step 2: Create the registry file

**File:** `internal/registry/<my_group>.go` (new file)

Follow this pattern:

```go
package registry

import "github.com/omargallob/devops-starter/pkg/tooldef"

// registerMyGroup adds <description> tools to the registry.
func registerMyGroup(r *Registry) {
    r.register(&tooldef.Tool{
        Name:        "example-tool",
        Version:     "1.0.0",
        Description: "What this tool does",
        Group:       tooldef.GroupMyGroup,
        InstallMode: tooldef.InstallModeEget,
        Repo:        "owner/repo",
    })
}
```

### Step 3: Wire into registry.New()

**File:** `internal/registry/registry.go`

**(a)** Update the package doc comment (line 4) to include the new group name
in the comma-separated list.

**(b)** Add the `registerMyGroup(r)` call after `registerUtilities(r)`:

```go
registerUtilities(r)
registerMyGroup(r)   // <-- add this
```

### Step 4: Add field to GroupConfig

**File:** `internal/config/config.go`

Add a `bool` field to the `GroupConfig` struct after `Utilities`:

```go
Utilities  bool `yaml:"utilities"`
MyGroup    bool `yaml:"my_group"`   // <-- add this
```

### Step 5: Set default in DefaultConfig()

**File:** `internal/config/config.go`

Add the default value in `DefaultConfig()` after `Utilities: true`:

```go
Utilities:  true,
MyGroup:    false,   // <-- add this (true = opt-in by default)
```

### Step 6: Add case to IsGroupEnabled()

**File:** `internal/config/config.go`

Add a case before the `default:` return in `IsGroupEnabled()`:

```go
case "utilities":
    return c.Groups.Utilities
case "my-group", "my_group":       // <-- add this (both forms!)
    return c.Groups.MyGroup
default:
    return false
```

### Step 7: Add case to SetGroup()

**File:** `internal/config/config.go`

Add a case before the closing `}` in `SetGroup()`:

```go
case "utilities":
    c.Groups.Utilities = enabled
case "my-group", "my_group":       // <-- add this
    c.Groups.MyGroup = enabled
}
```

### Step 8: Add to AllGroupNames()

**File:** `internal/config/config.go`

Append to the returned slice in `AllGroupNames()`:

```go
"utilities",
"my-group",   // <-- add this
```

### Step 9: Add to default config YAML

**File:** `configs/default.yaml`

Add under the `groups:` section:

```yaml
utilities: true
my_group: false    # <-- add this (use snake_case)
```

### Step 10: Add to ResolveAll() group ordering

**File:** `internal/state/state.go`

Add to the `groups` slice in `ResolveAll()` after `tooldef.GroupUtilities`.
This controls **display order** in the TUI and `--no-tui` table. If the group
is not listed here, its tools will never appear in status output.

```go
tooldef.GroupUtilities,
tooldef.GroupMyGroup,   // <-- add this
```

### Step 11: Add to isValidGroup() (if plugins.go exists)

**File:** `internal/registry/plugins.go`

If this file exists on your branch, add the new group to the `isValidGroup()`
switch so plugin YAML files can use this group:

```go
tooldef.GroupUtilities,
tooldef.GroupMyGroup:   // <-- add this
    return true
```

If `plugins.go` does not exist, skip this step.

### Step 12: Add to TestAllToolsHaveValidGroup

**File:** `internal/registry/registry_test.go`

Add to the `validGroups` map:

```go
tooldef.GroupUtilities:  true,
tooldef.GroupMyGroup:    true,   // <-- add this
```

### Step 13: Add to TestAllGroupsHaveTools

**File:** `internal/registry/registry_test.go`

Add to the `groups` slice so the test verifies your group has at least one tool:

```go
tooldef.GroupUtilities,
tooldef.GroupMyGroup,   // <-- add this
```

### Step 14: Update TestAllGroupNames

**File:** `internal/config/config_test.go`

**(a)** Bump the count: change `len(names) != N` to `N+1`.

**(b)** Add the group name to the `expected` slice:

```go
expected := []string{..., "utilities", "my-group"}
```

### Step 15: Add to TestIsGroupEnabled

**File:** `internal/config/config_test.go`

Add test cases before `{"nonexistent", false}`:

```go
{"utilities", true},
{"my-group", false},      // <-- add (matches DefaultConfig)
{"my_group", false},      // <-- add (underscore variant)
{"nonexistent", false},
```

### Step 16: (Optional) Assert default in TestDefaultConfig

**File:** `internal/config/config_test.go`

Add a check for the new group's default value.

### Step 17: Add source file to BUILD.bazel

**File:** `internal/registry/BUILD.bazel`

Add the new `.go` file to the `srcs` list. The list is **alphabetically sorted**:

```python
srcs = [
    ...
    "my_group.go",   # <-- add in alphabetical position
    ...
],
```

### Step 18: Update CONTRIBUTING.md

**File:** `CONTRIBUTING.md`

Add a row to the "Available Groups" table:

```
| `GroupMyGroup` | Description of what this group covers |
```

## Verification

After completing all steps, run:

```bash
go build ./...
go test -race ./pkg/tooldef/... ./internal/config/... ./internal/registry/... ./internal/state/... ./internal/tui/...
bazel run //:golangci_lint
```

All must pass before the group is considered complete.

## Known gaps in existing groups

The following inconsistencies exist in the codebase and should be fixed if
encountered:

### GroupAnsible

- Has a constant in `pkg/tooldef/tool.go` and entries in `config.go`,
  `config_test.go`, `registry_test.go` (validGroups), and `CONTRIBUTING.md`
- **Missing**: no `ansible.go` registry file, no `registerAnsible(r)` call in
  `registry.go`, no entry in `BUILD.bazel` srcs, no entry in `state.go`
  `ResolveAll()` groups, no entry in `TestAllGroupsHaveTools`, no entry in
  `configs/default.yaml`

### GroupPackageManagers (when branch with plugins exists)

- **Missing** from `isValidGroup()` in `plugins.go` -- plugin YAML files
  using `group: package-managers` will be rejected
- **Missing** from `state.go` `ResolveAll()` groups -- tools won't appear in
  TUI
- **Missing** from `TestAllGroupsHaveTools`

### GroupAI (when branch with AI tools exists)

- **Missing** from `CONTRIBUTING.md` Available Groups table
