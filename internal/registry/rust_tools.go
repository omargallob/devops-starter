// rust_tools.go registers modern Rust-based CLI replacements for coreutils and
// developer productivity tools. Most are distributed as pre-compiled binaries
// via GitHub releases using Rust target triples.
package registry

import "github.com/omargallob/devops-starter/pkg/tooldef"

func registerRustTools(r *Registry) {
	r.register(&tooldef.Tool{
		Name:        "bat",
		Version:     "0.26.1",
		Description: "Cat clone with syntax highlighting",
		Group:       tooldef.GroupRustTools,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "sharkdp/bat",
		Asset:       "*musl*",
	})

	r.register(&tooldef.Tool{
		Name:        "eza",
		Version:     "0.23.4",
		Description: "Modern ls replacement",
		Group:       tooldef.GroupRustTools,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "eza-community/eza",
		Asset:       "*musl*",
	})

	r.register(&tooldef.Tool{
		Name:        "fd",
		Version:     "10.4.2",
		Description: "Fast find alternative",
		Group:       tooldef.GroupRustTools,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "sharkdp/fd",
		Asset:       "*musl*",
	})

	r.register(&tooldef.Tool{
		Name:        "ripgrep",
		Version:     "15.1.0",
		Description: "Fast grep alternative",
		Group:       tooldef.GroupRustTools,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "BurntSushi/ripgrep",
		Asset:       "*musl*",
		BinaryName:  "rg",
		InstallName: "rg",
	})

	r.register(&tooldef.Tool{
		Name:        "delta",
		Version:     "0.19.2",
		Description: "Git diff viewer",
		Group:       tooldef.GroupRustTools,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "dandavison/delta",
		Asset:       "*musl*",
	})

	r.register(&tooldef.Tool{
		Name:        "zoxide",
		Version:     "0.9.9",
		Description: "Smarter cd command",
		Group:       tooldef.GroupRustTools,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "ajeetdsouza/zoxide",
		Asset:       "*musl*",
	})

	r.register(&tooldef.Tool{
		Name:        "starship",
		Version:     "1.25.1",
		Description: "Cross-shell prompt",
		Group:       tooldef.GroupRustTools,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "starship/starship",
		Asset:       "*musl*",
	})

	r.register(&tooldef.Tool{
		Name:        "tokei",
		Version:     "14.0.0",
		Description: "Code statistics",
		Group:       tooldef.GroupRustTools,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "XAMPPRocky/tokei",
	})

	r.register(&tooldef.Tool{
		Name:        "hyperfine",
		Version:     "1.20.0",
		Description: "Command-line benchmarking",
		Group:       tooldef.GroupRustTools,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "sharkdp/hyperfine",
		Asset:       "*musl*",
	})

	r.register(&tooldef.Tool{
		Name:        "procs",
		Version:     "0.14.11",
		Description: "Modern ps replacement",
		Group:       tooldef.GroupRustTools,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "dalance/procs",
	})

	r.register(&tooldef.Tool{
		Name:        "bottom",
		Version:     "0.12.3",
		Description: "System monitor",
		Group:       tooldef.GroupRustTools,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "ClementTsang/bottom",
		Asset:       "*musl*",
		BinaryName:  "btm",
		InstallName: "btm",
	})

	r.register(&tooldef.Tool{
		Name:        "gitui",
		Version:     "0.28.1",
		Description: "Terminal UI for git",
		Group:       tooldef.GroupRustTools,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "gitui-org/gitui",
	})

	r.register(&tooldef.Tool{
		Name:        "dust",
		Version:     "1.2.4",
		Description: "Disk usage analyzer",
		Group:       tooldef.GroupRustTools,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "bootandy/dust",
		Asset:       "*musl*",
	})

	r.register(&tooldef.Tool{
		Name:        "bandwhich",
		Version:     "0.23.1",
		Description: "Network bandwidth monitor",
		Group:       tooldef.GroupRustTools,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "imsnif/bandwhich",
		Asset:       "*musl*",
	})

	r.register(&tooldef.Tool{
		Name:        "sd",
		Version:     "1.1.0",
		Description: "Intuitive find & replace",
		Group:       tooldef.GroupRustTools,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "chmln/sd",
		Asset:       "*musl*",
	})

	r.register(&tooldef.Tool{
		Name:        "tealdeer",
		Version:     "1.8.1",
		Description: "Fast tldr client",
		Group:       tooldef.GroupRustTools,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "tealdeer-rs/tealdeer",
		Asset:       "*musl*",
		InstallName: "tldr",
	})

	r.register(&tooldef.Tool{
		Name:        "xh",
		Version:     "0.25.3",
		Description: "Friendly HTTP client",
		Group:       tooldef.GroupRustTools,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "ducaale/xh",
		Asset:       "*musl*",
	})

	r.register(&tooldef.Tool{
		Name:        "yazi",
		Version:     "26.5.6",
		Description: "Terminal file manager",
		Group:       tooldef.GroupRustTools,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "sxyazi/yazi",
		Asset:       "*musl*",
	})

	r.register(&tooldef.Tool{
		Name:        "atuin",
		Version:     "18.16.1",
		Description: "Shell history manager",
		Group:       tooldef.GroupRustTools,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "atuinsh/atuin",
		Asset:       "*musl*",
	})

	r.register(&tooldef.Tool{
		Name:        "zellij",
		Version:     "0.44.3",
		Description: "Terminal workspace",
		Group:       tooldef.GroupRustTools,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "zellij-org/zellij",
		Asset:       "*musl*",
	})

	r.register(&tooldef.Tool{
		Name:        "just",
		Version:     "1.51.0",
		Description: "Command runner",
		Group:       tooldef.GroupRustTools,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "casey/just",
		Asset:       "*musl*",
	})

	r.register(&tooldef.Tool{
		Name:        "watchexec",
		Version:     "2.5.1",
		Description: "File watcher and command executor",
		Group:       tooldef.GroupRustTools,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "watchexec/watchexec",
		Asset:       "*musl*",
	})
}
