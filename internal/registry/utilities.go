// utilities.go registers general-purpose developer utilities that don't fit
// neatly into other categories (JSON/YAML processors, fuzzy finders, security
// tools, Git helpers, linters, task runners, editors, and pip/npm-distributed
// workflow tools).
package registry

import "github.com/omargallob/devops-starter/pkg/tooldef"

// registerUtilities adds general-purpose developer tools to the registry.
// Includes jq, yq, fzf, direnv, age, sops, gh, trivy, lazygit, shellcheck,
// shfmt, task, and neovim.
func registerUtilities(r *Registry) {
	r.register(&tooldef.Tool{
		Name:        "jq",
		Version:     "1.7.1",
		Description: "JSON processor",
		Group:       tooldef.GroupUtilities,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "jqlang/jq",
	})

	r.register(&tooldef.Tool{
		Name:        "yq",
		Version:     "4.44.6",
		Description: "YAML processor",
		Group:       tooldef.GroupUtilities,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "mikefarah/yq",
		Asset:       "yq_{{.OS}}_{{.Arch}}",
	})

	r.register(&tooldef.Tool{
		Name:        "fzf",
		Version:     "0.57.0",
		Description: "Fuzzy finder",
		Group:       tooldef.GroupUtilities,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "junegunn/fzf",
	})

	r.register(&tooldef.Tool{
		Name:        "direnv",
		Version:     "2.35.0",
		Description: "Environment variable manager",
		Group:       tooldef.GroupUtilities,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "direnv/direnv",
	})

	r.register(&tooldef.Tool{
		Name:        "age",
		Version:     "1.2.0",
		Description: "File encryption tool",
		Group:       tooldef.GroupUtilities,
		InstallMode: tooldef.InstallModeEgetURL,
		Format:      tooldef.FormatTarGz,
		BinaryName:  "age",
		URLTemplate: "https://dl.filippo.io/age/v{{.Version}}?for={{.OS}}/{{.Arch}}",
	})

	r.register(&tooldef.Tool{
		Name:        "sops",
		Version:     "3.9.4",
		Description: "Secrets encryption",
		Group:       tooldef.GroupUtilities,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "getsops/sops",
		Asset:       "sops-*{{.OS}}*{{.Arch}}*",
	})

	r.register(&tooldef.Tool{
		Name:        "gh",
		Version:     "2.63.2",
		Description: "GitHub CLI",
		Group:       tooldef.GroupUtilities,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "cli/cli",
		BinaryName:  "gh",
	})

	r.register(&tooldef.Tool{
		Name:        "trivy",
		Version:     "0.70.0",
		Description: "Security scanner",
		Group:       tooldef.GroupUtilities,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "aquasecurity/trivy",
	})

	r.register(&tooldef.Tool{
		Name:        "lazygit",
		Version:     "0.44.1",
		Description: "Terminal UI for git",
		Group:       tooldef.GroupUtilities,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "jesseduffield/lazygit",
	})

	r.register(&tooldef.Tool{
		Name:        "shellcheck",
		Version:     "0.10.0",
		Description: "Shell script linter",
		Group:       tooldef.GroupUtilities,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "koalaman/shellcheck",
		BinaryName:  "shellcheck",
	})

	r.register(&tooldef.Tool{
		Name:        "shfmt",
		Version:     "3.10.0",
		Description: "Shell script formatter",
		Group:       tooldef.GroupUtilities,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "mvdan/sh",
		Asset:       "shfmt_*_{{.OS}}_{{.Arch}}",
	})

	r.register(&tooldef.Tool{
		Name:        "task",
		Version:     "3.40.1",
		Description: "Task runner",
		Group:       tooldef.GroupUtilities,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "go-task/task",
	})

	r.register(&tooldef.Tool{
		Name:        "neovim",
		Version:     "0.10.3",
		Description: "Hyperextensible text editor",
		Group:       tooldef.GroupUtilities,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "neovim/neovim",
		InstallName: "nvim",
		BinaryName:  "nvim",
	})

	r.register(&tooldef.Tool{
		Name:        "lcov",
		Version:     "2.4",
		Description: "Linux code coverage reporting tool",
		Group:       tooldef.GroupUtilities,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "linux-test-project/lcov",
		BinaryName:  "bin/lcov",
	})

	r.register(&tooldef.Tool{
		Name:        "genhtml",
		Version:     "2.4",
		Description: "Generate HTML coverage reports from lcov data",
		Group:       tooldef.GroupUtilities,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "linux-test-project/lcov",
		BinaryName:  "bin/genhtml",
	})

	r.register(&tooldef.Tool{
		Name:         "pre-commit",
		Version:      "4.0.1",
		Description:  "Git hooks framework for code quality checks",
		Group:        tooldef.GroupUtilities,
		InstallMode:  tooldef.InstallModeMise,
		MiseBackend:  "pipx:pre-commit",
		Dependencies: []string{"mise"},
	})

	r.register(&tooldef.Tool{
		Name:         "yamllint",
		Version:      "1.35.1",
		Description:  "YAML file linter",
		Group:        tooldef.GroupUtilities,
		InstallMode:  tooldef.InstallModeMise,
		MiseBackend:  "pipx:yamllint",
		Dependencies: []string{"mise"},
	})

	r.register(&tooldef.Tool{
		Name:         "commitlint",
		Version:      "19.6.1",
		Description:  "Lint commit messages against Conventional Commits rules",
		Group:        tooldef.GroupUtilities,
		InstallMode:  tooldef.InstallModeMise,
		MiseBackend:  "npm:@commitlint/cli",
		InstallName:  "commitlint",
		Dependencies: []string{"mise"},
	})

	r.register(&tooldef.Tool{
		Name:         "semantic-release",
		Version:      "24.2.3",
		Description:  "Automated changelog and version management from commit history",
		Group:        tooldef.GroupUtilities,
		InstallMode:  tooldef.InstallModeMise,
		MiseBackend:  "npm:semantic-release",
		Dependencies: []string{"mise"},
	})
}
