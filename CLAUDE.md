
# CLAUDE.md – Contract‑Driven Development for `devops-starter`

This repository builds a cross‑platform DevOps tool installer and dotfile manager in Go.  
**All contributions must follow Contract‑Driven Development (CDD).**  
Contracts are the single source of truth – code, tests, and documentation derive from them.

## 🧠 Core CDD Principles for this Repo

1. **Design the contract first** – Every new feature begins with a clear contract:
   - A **public type** in `pkg/tooldef/` (e.g. `Tool`, `Platform`, `ArchiveFormat`, `InstallMode`).
   - A **YAML schema snippet** for configuration changes.
   - A **command line interface** (subcommand, flags, usage text).
2. **Generate, don’t write** – Use code generation where possible:
   - `go generate` for tool definitions (see `internal/registry/generate.go`).
   - `go fmt` and `go mod tidy` for consistency.
3. **Test the contract** – Every addition must include:
   - **Contract tests** that validate the contract against the implementation.
   - **Regression tests** for boundary conditions (e.g. unsupported platforms, malformed config).
4. **Fail on contract violation** – Use `panic` or structured errors only when pre‑/post‑conditions break:
   - `config.Validate()` must panic if required fields are missing.
   - `platform.Detect()` must return a structured error for unrecognised OS/arch.
5. **Version contracts explicitly** – Configuration files and state stores are versioned:
   - `ConfigVersion` constant in `internal/config/`.
   - `StateVersion` constant in `internal/state/`.
   - Migrations must be provided for backwards‑incompatible changes.

## 📁 Repository Structure (CDD‑aware)

```

.
├── pkg/                     # Public types (contracts)
│   └── tooldef/             # Tool, Platform, ArchiveFormat, InstallMode
├── internal/
│   ├── config/              # YAML configuration (load/save/validate)
│   ├── registry/            # Tool definitions (generated)
│   ├── installer/           # Download → verify → extract → place
│   ├── platform/            # OS/arch/distro detection
│   ├── state/               # JSON store of installed tools
│   ├── dotfiles/            # Symlink manager with backup
│   ├── tui/                 # Bubble Tea interfaces
│   └── test/                # Contract tests
├── cmd/                     # Cobra commands (thin wrappers)
└── configs/                 # Default configuration YAML

```

## 🔧 Mandatory Tooling

- `golangci-lint` – linting with project‑specific config.
- `go test -race` – race detector for all tests.
- `go test -cover` – coverage reports (aim for ≥80%).
- `pre-commit` – runs `make check` on every commit.

## 🧪 Contract Testing Workflow

Every new tool definition **must** have a contract test in `internal/test/contract_test.go`:

```go
func TestToolContract(t *testing.T) {
    // 1. Precondition (contract's "requires")
    tool := &tooldef.Tool{
        Name:    "kubectl",
        Version: "1.28.0",
        Platforms: []tooldef.Platform{
            {OS: "linux", Arch: "amd64", URL: "https://dl.k8s.io/release/.../bin/linux/amd64/kubectl"},
        },
    }

    // 2. Action
    installer := installer.New(tool, opts...)
    err := installer.Install(context.Background())

    // 3. Postcondition (contract's "ensures")
    require.NoError(t, err)
    assert.FileExists(t, filepath.Join(installDir, "kubectl"))

    // 4. Invariant (contract's "invariant")
    versionOutput, err := exec.Command("kubectl", "version", "--client").Output()
    require.NoError(t, err)
    assert.Contains(t, string(versionOutput), "1.28.0")
}
```

For configuration changes, write a table‑driven test that validates the new field’s marshaling/unmarshaling and default values.

🚦 Code Review Checklist

Pull requests must demonstrate:

· Contract first – A public type or YAML schema is defined before implementation.
· Generated code – Any code that can be generated (tool definitions, deep copies) is generated.
· Contract tests – New features include tests that validate the contract (inputs → outputs).
· Error handling – Contract violations cause a clear error or panic, not silent failure.
· Configuration version – Bumped if configuration format changes, with migration path.
· State version – Bumped if state format changes, with migration path.
· Documentation – Updated README.md and docs/ to reflect the new contract.

🚀 Quick Reference

Command Purpose
make test Run all tests with race detector.
make test-cover Generate coverage report.
make lint Run golangci-lint and shellcheck.
make check Run fmt, vet, lint, and test.
make generate Run go generate to refresh tool definitions.
make bazel-test Hermetic test using Bazel (used in CI).

🔄 Releasing a New Contract

1. Update the version constants in internal/config/config.go and internal/state/state.go.
2. Write a migration function for old configuration/state files.
3. Add contract tests for the migration logic.
4. Update the documentation – especially the configuration reference in docs/configuration.md.
5. Create a PR with the label contract-change for focused review.

