// package_managers.go registers package manager tools across language ecosystems.
// Covers standalone package-manager binaries (pnpm, bun, uv, etc.) and
// mise-backend tools (yarn, poetry) that have no standalone binary release.
package registry

import "github.com/omargallob/devops-starter/pkg/tooldef"

// registerPackageManagers adds package manager tools to the registry.
// JavaScript subgroup: pnpm, bun, volta, yarn.
// Python subgroup: uv, rye, pipx, poetry.
// Rust subgroup: cargo-binstall.
func registerPackageManagers(r *Registry) {
	// JavaScript

	r.register(&tooldef.Tool{
		Name:        "pnpm",
		Version:     "9.15.4",
		Description: "Fast, disk-efficient npm replacement",
		Group:       tooldef.GroupPackageManagers,
		Subgroup:    "JavaScript",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "pnpm/pnpm",
	})

	r.register(&tooldef.Tool{
		Name:        "bun",
		Version:     "1.1.42",
		Description: "JavaScript runtime and package manager",
		Group:       tooldef.GroupPackageManagers,
		Subgroup:    "JavaScript",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "oven-sh/bun",
		BinaryName:  "bun",
	})

	r.register(&tooldef.Tool{
		Name:        "volta",
		Version:     "2.0.2",
		Description: "Node.js tool manager; pins npm/yarn/node versions per project",
		Group:       tooldef.GroupPackageManagers,
		Subgroup:    "JavaScript",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "volta-cli/volta",
	})

	r.register(&tooldef.Tool{
		Name:         "yarn",
		Version:      "1.22.22",
		Description:  "Node package manager with deterministic lock files",
		Group:        tooldef.GroupPackageManagers,
		Subgroup:     "JavaScript",
		InstallMode:  tooldef.InstallModeMise,
		MiseBackend:  "npm:yarn",
		Dependencies: []string{"mise"},
	})

	// Python

	r.register(&tooldef.Tool{
		Name:        "uv",
		Version:     "0.5.21",
		Description: "Rust-based Python package installer and resolver",
		Group:       tooldef.GroupPackageManagers,
		Subgroup:    "Python",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "astral-sh/uv",
		BinaryName:  "uv",
	})

	r.register(&tooldef.Tool{
		Name:        "rye",
		Version:     "0.44.0",
		Description: "Python project and dependency manager",
		Group:       tooldef.GroupPackageManagers,
		Subgroup:    "Python",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "astral-sh/rye",
	})

	r.register(&tooldef.Tool{
		Name:        "pipx",
		Version:     "1.7.1",
		Description: "Install Python CLI tools in isolated virtualenvs",
		Group:       tooldef.GroupPackageManagers,
		Subgroup:    "Python",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "pypa/pipx",
	})

	r.register(&tooldef.Tool{
		Name:         "poetry",
		Version:      "1.8.5",
		Description:  "Python dependency and virtualenv manager",
		Group:        tooldef.GroupPackageManagers,
		Subgroup:     "Python",
		InstallMode:  tooldef.InstallModeMise,
		MiseBackend:  "pipx:poetry",
		Dependencies: []string{"mise"},
	})

	// Rust

	r.register(&tooldef.Tool{
		Name:        "cargo-binstall",
		Version:     "1.10.10",
		Description: "Install Cargo crates from pre-built binaries without compiling",
		Group:       tooldef.GroupPackageManagers,
		Subgroup:    "Rust",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "cargo-bins/cargo-binstall",
	})
}
