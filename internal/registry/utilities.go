// utilities.go registers general-purpose developer utilities that don't fit
// neatly into other categories (JSON/YAML processors, fuzzy finders, security
// tools, Git helpers, linters, task runners, and editors).
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
		Format:      tooldef.FormatBinary,
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/jqlang/jq/releases/download/jq-1.7.1/jq-linux-amd64",
			"linux/arm64":  "https://github.com/jqlang/jq/releases/download/jq-1.7.1/jq-linux-arm64",
			"darwin/amd64": "https://github.com/jqlang/jq/releases/download/jq-1.7.1/jq-macos-amd64",
			"darwin/arm64": "https://github.com/jqlang/jq/releases/download/jq-1.7.1/jq-macos-arm64",
		},
	})

	r.register(&tooldef.Tool{
		Name:        "yq",
		Version:     "4.44.6",
		Description: "YAML processor",
		Group:       tooldef.GroupUtilities,
		Format:      tooldef.FormatBinary,
		URLTemplate: "https://github.com/mikefarah/yq/releases/download/v{{.Version}}/yq_{{.OS}}_{{.Arch}}",
	})

	r.register(&tooldef.Tool{
		Name:        "fzf",
		Version:     "0.57.0",
		Description: "Fuzzy finder",
		Group:       tooldef.GroupUtilities,
		Format:      tooldef.FormatTarGz,
		URLTemplate: "https://github.com/junegunn/fzf/releases/download/v{{.Version}}/fzf-{{.Version}}-{{.OS}}_{{.Arch}}.tar.gz",
	})

	r.register(&tooldef.Tool{
		Name:        "direnv",
		Version:     "2.35.0",
		Description: "Environment variable manager",
		Group:       tooldef.GroupUtilities,
		Format:      tooldef.FormatBinary,
		URLTemplate: "https://github.com/direnv/direnv/releases/download/v{{.Version}}/direnv.{{.OS}}-{{.Arch}}",
	})

	r.register(&tooldef.Tool{
		Name:        "age",
		Version:     "1.2.0",
		Description: "File encryption tool",
		Group:       tooldef.GroupUtilities,
		Format:      tooldef.FormatTarGz,
		BinaryName:  "age/age",
		URLTemplate: "https://dl.filippo.io/age/v{{.Version}}?for={{.OS}}/{{.Arch}}",
	})

	r.register(&tooldef.Tool{
		Name:        "sops",
		Version:     "3.9.4",
		Description: "Secrets encryption",
		Group:       tooldef.GroupUtilities,
		Format:      tooldef.FormatBinary,
		URLTemplate: "https://github.com/getsops/sops/releases/download/v{{.Version}}/sops-v{{.Version}}.{{.OS}}.{{.Arch}}",
	})

	r.register(&tooldef.Tool{
		Name:        "gh",
		Version:     "2.63.2",
		Description: "GitHub CLI",
		Group:       tooldef.GroupUtilities,
		Format:      tooldef.FormatTarGz,
		BinaryName:  "gh_{{.Version}}_{{.OS}}_{{.Arch}}/bin/gh",
		URLTemplate: "https://github.com/cli/cli/releases/download/v{{.Version}}/gh_{{.Version}}_{{.OS}}_{{.Arch}}.tar.gz",
	})

	r.register(&tooldef.Tool{
		Name:        "trivy",
		Version:     "0.70.0",
		Description: "Security scanner",
		Group:       tooldef.GroupUtilities,
		Format:      tooldef.FormatTarGz,
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/aquasecurity/trivy/releases/download/v0.70.0/trivy_0.70.0_Linux-64bit.tar.gz",
			"linux/arm64":  "https://github.com/aquasecurity/trivy/releases/download/v0.70.0/trivy_0.70.0_Linux-ARM64.tar.gz",
			"darwin/amd64": "https://github.com/aquasecurity/trivy/releases/download/v0.70.0/trivy_0.70.0_macOS-64bit.tar.gz",
			"darwin/arm64": "https://github.com/aquasecurity/trivy/releases/download/v0.70.0/trivy_0.70.0_macOS-ARM64.tar.gz",
		},
	})

	r.register(&tooldef.Tool{
		Name:        "lazygit",
		Version:     "0.44.1",
		Description: "Terminal UI for git",
		Group:       tooldef.GroupUtilities,
		Format:      tooldef.FormatTarGz,
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/jesseduffield/lazygit/releases/download/v0.44.1/lazygit_0.44.1_Linux_x86_64.tar.gz",
			"linux/arm64":  "https://github.com/jesseduffield/lazygit/releases/download/v0.44.1/lazygit_0.44.1_Linux_arm64.tar.gz",
			"darwin/amd64": "https://github.com/jesseduffield/lazygit/releases/download/v0.44.1/lazygit_0.44.1_Darwin_x86_64.tar.gz",
			"darwin/arm64": "https://github.com/jesseduffield/lazygit/releases/download/v0.44.1/lazygit_0.44.1_Darwin_arm64.tar.gz",
		},
	})

	r.register(&tooldef.Tool{
		Name:        "shellcheck",
		Version:     "0.10.0",
		Description: "Shell script linter",
		Group:       tooldef.GroupUtilities,
		Format:      tooldef.FormatTarXz,
		BinaryName:  "shellcheck-v0.10.0/shellcheck",
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/koalaman/shellcheck/releases/download/v0.10.0/shellcheck-v0.10.0.linux.x86_64.tar.xz",
			"linux/arm64":  "https://github.com/koalaman/shellcheck/releases/download/v0.10.0/shellcheck-v0.10.0.linux.aarch64.tar.xz",
			"darwin/amd64": "https://github.com/koalaman/shellcheck/releases/download/v0.10.0/shellcheck-v0.10.0.darwin.x86_64.tar.xz",
			"darwin/arm64": "https://github.com/koalaman/shellcheck/releases/download/v0.10.0/shellcheck-v0.10.0.darwin.aarch64.tar.xz",
		},
	})

	r.register(&tooldef.Tool{
		Name:        "shfmt",
		Version:     "3.10.0",
		Description: "Shell script formatter",
		Group:       tooldef.GroupUtilities,
		Format:      tooldef.FormatBinary,
		URLTemplate: "https://github.com/mvdan/sh/releases/download/v{{.Version}}/shfmt_v{{.Version}}_{{.OS}}_{{.Arch}}",
	})

	r.register(&tooldef.Tool{
		Name:        "task",
		Version:     "3.40.1",
		Description: "Task runner",
		Group:       tooldef.GroupUtilities,
		Format:      tooldef.FormatTarGz,
		InstallName: "task",
		URLTemplate: "https://github.com/go-task/task/releases/download/v{{.Version}}/task_{{.OS}}_{{.Arch}}.tar.gz",
	})

	r.register(&tooldef.Tool{
		Name:            "neovim",
		Version:         "0.10.3",
		Description:     "Hyperextensible text editor",
		Group:           tooldef.GroupUtilities,
		Format:          tooldef.FormatTarGz,
		InstallName:     "nvim",
		BinaryName:      "nvim",
		StripComponents: 2,
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/neovim/neovim/releases/download/v0.10.3/nvim-linux-x86_64.tar.gz",
			"darwin/amd64": "https://github.com/neovim/neovim/releases/download/v0.10.3/nvim-macos-x86_64.tar.gz",
			"darwin/arm64": "https://github.com/neovim/neovim/releases/download/v0.10.3/nvim-macos-arm64.tar.gz",
		},
	})

	r.register(&tooldef.Tool{
		Name:            "lcov",
		Version:         "2.4",
		Description:     "Linux code coverage reporting tool",
		Group:           tooldef.GroupUtilities,
		Format:          tooldef.FormatTarGz,
		URLTemplate:     "https://github.com/linux-test-project/lcov/releases/download/v{{.Version}}/lcov-{{.Version}}.tar.gz",
		BinaryName:      "bin/lcov",
		StripComponents: 1,
	})

	r.register(&tooldef.Tool{
		Name:            "genhtml",
		Version:         "2.4",
		Description:     "Generate HTML coverage reports from lcov data",
		Group:           tooldef.GroupUtilities,
		Format:          tooldef.FormatTarGz,
		URLTemplate:     "https://github.com/linux-test-project/lcov/releases/download/v{{.Version}}/lcov-{{.Version}}.tar.gz",
		BinaryName:      "bin/genhtml",
		StripComponents: 1,
	})
}
