// rust_tools.go registers modern Rust-based CLI replacements for coreutils and
// developer productivity tools. Most are distributed as pre-compiled binaries
// using Rust target triples (e.g., x86_64-unknown-linux-musl).
package registry

import "github.com/omargallob/devops-starter/pkg/tooldef"

func registerRustTools(r *Registry) {
	// bat - only x86_64 available for macOS in v0.24.0
	r.register(&tooldef.Tool{
		Name:            "bat",
		Version:         "0.24.0",
		Description:     "Cat clone with syntax highlighting",
		Group:           tooldef.GroupRustTools,
		Format:          tooldef.FormatTarGz,
		BinaryName:      "bat",
		StripComponents: 1,
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/sharkdp/bat/releases/download/v0.24.0/bat-v0.24.0-x86_64-unknown-linux-musl.tar.gz",
			"linux/arm64":  "https://github.com/sharkdp/bat/releases/download/v0.24.0/bat-v0.24.0-aarch64-unknown-linux-musl.tar.gz",
			"darwin/amd64": "https://github.com/sharkdp/bat/releases/download/v0.24.0/bat-v0.24.0-x86_64-apple-darwin.tar.gz",
		},
	})

	// eza
	r.register(&tooldef.Tool{
		Name:        "eza",
		Version:     "0.20.14",
		Description: "Modern ls replacement",
		Group:       tooldef.GroupRustTools,
		Format:      tooldef.FormatBinary,
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/eza-community/eza/releases/download/v0.20.14/eza_x86_64-unknown-linux-musl",
			"linux/arm64":  "https://github.com/eza-community/eza/releases/download/v0.20.14/eza_aarch64-unknown-linux-musl",
			"darwin/amd64": "https://github.com/eza-community/eza/releases/download/v0.20.14/eza_x86_64-apple-darwin",
			"darwin/arm64": "https://github.com/eza-community/eza/releases/download/v0.20.14/eza_aarch64-apple-darwin",
		},
	})

	// fd - nested archive: fd-v{version}-{triple}/fd
	r.register(&tooldef.Tool{
		Name:            "fd",
		Version:         "10.2.0",
		Description:     "Fast find alternative",
		Group:           tooldef.GroupRustTools,
		Format:          tooldef.FormatTarGz,
		BinaryName:      "fd",
		StripComponents: 1,
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/sharkdp/fd/releases/download/v10.2.0/fd-v10.2.0-x86_64-unknown-linux-musl.tar.gz",
			"linux/arm64":  "https://github.com/sharkdp/fd/releases/download/v10.2.0/fd-v10.2.0-aarch64-unknown-linux-musl.tar.gz",
			"darwin/amd64": "https://github.com/sharkdp/fd/releases/download/v10.2.0/fd-v10.2.0-x86_64-apple-darwin.tar.gz",
			"darwin/arm64": "https://github.com/sharkdp/fd/releases/download/v10.2.0/fd-v10.2.0-aarch64-apple-darwin.tar.gz",
		},
	})

	// ripgrep - tag uses no v prefix; nested: ripgrep-{version}-{triple}/rg
	r.register(&tooldef.Tool{
		Name:            "ripgrep",
		Version:         "14.1.1",
		Description:     "Fast grep alternative",
		Group:           tooldef.GroupRustTools,
		Format:          tooldef.FormatTarGz,
		BinaryName:      "rg",
		InstallName:     "rg",
		StripComponents: 1,
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/BurntSushi/ripgrep/releases/download/14.1.1/ripgrep-14.1.1-x86_64-unknown-linux-musl.tar.gz",
			"linux/arm64":  "https://github.com/BurntSushi/ripgrep/releases/download/14.1.1/ripgrep-14.1.1-aarch64-unknown-linux-musl.tar.gz",
			"darwin/amd64": "https://github.com/BurntSushi/ripgrep/releases/download/14.1.1/ripgrep-14.1.1-x86_64-apple-darwin.tar.gz",
			"darwin/arm64": "https://github.com/BurntSushi/ripgrep/releases/download/14.1.1/ripgrep-14.1.1-aarch64-apple-darwin.tar.gz",
		},
	})

	// delta - tag uses no v prefix; nested: delta-{version}-{triple}/delta
	r.register(&tooldef.Tool{
		Name:            "delta",
		Version:         "0.18.2",
		Description:     "Git diff viewer",
		Group:           tooldef.GroupRustTools,
		Format:          tooldef.FormatTarGz,
		BinaryName:      "delta",
		StripComponents: 1,
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/dandavison/delta/releases/download/0.18.2/delta-0.18.2-x86_64-unknown-linux-musl.tar.gz",
			"linux/arm64":  "https://github.com/dandavison/delta/releases/download/0.18.2/delta-0.18.2-aarch64-unknown-linux-musl.tar.gz",
			"darwin/amd64": "https://github.com/dandavison/delta/releases/download/0.18.2/delta-0.18.2-x86_64-apple-darwin.tar.gz",
			"darwin/arm64": "https://github.com/dandavison/delta/releases/download/0.18.2/delta-0.18.2-aarch64-apple-darwin.tar.gz",
		},
	})

	// zoxide - filename uses no v prefix: zoxide-{version}-{triple}.tar.gz; flat archive
	r.register(&tooldef.Tool{
		Name:        "zoxide",
		Version:     "0.9.6",
		Description: "Smarter cd command",
		Group:       tooldef.GroupRustTools,
		Format:      tooldef.FormatTarGz,
		BinaryName:  "zoxide",
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/ajeetdsouza/zoxide/releases/download/v0.9.6/zoxide-0.9.6-x86_64-unknown-linux-musl.tar.gz",
			"linux/arm64":  "https://github.com/ajeetdsouza/zoxide/releases/download/v0.9.6/zoxide-0.9.6-aarch64-unknown-linux-musl.tar.gz",
			"darwin/amd64": "https://github.com/ajeetdsouza/zoxide/releases/download/v0.9.6/zoxide-0.9.6-x86_64-apple-darwin.tar.gz",
			"darwin/arm64": "https://github.com/ajeetdsouza/zoxide/releases/download/v0.9.6/zoxide-0.9.6-aarch64-apple-darwin.tar.gz",
		},
	})

	// starship - filename has no version: starship-{triple}.tar.gz; flat archive
	r.register(&tooldef.Tool{
		Name:        "starship",
		Version:     "1.21.1",
		Description: "Cross-shell prompt",
		Group:       tooldef.GroupRustTools,
		Format:      tooldef.FormatTarGz,
		BinaryName:  "starship",
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/starship/starship/releases/download/v1.21.1/starship-x86_64-unknown-linux-musl.tar.gz",
			"linux/arm64":  "https://github.com/starship/starship/releases/download/v1.21.1/starship-aarch64-unknown-linux-musl.tar.gz",
			"darwin/amd64": "https://github.com/starship/starship/releases/download/v1.21.1/starship-x86_64-apple-darwin.tar.gz",
			"darwin/arm64": "https://github.com/starship/starship/releases/download/v1.21.1/starship-aarch64-apple-darwin.tar.gz",
		},
	})

	// tokei - v13 alpha has no pre-built binaries; use v12.1.2 (x86_64 only for macOS); flat archive
	r.register(&tooldef.Tool{
		Name:        "tokei",
		Version:     "12.1.2",
		Description: "Code statistics",
		Group:       tooldef.GroupRustTools,
		Format:      tooldef.FormatTarGz,
		BinaryName:  "tokei",
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/XAMPPRocky/tokei/releases/download/v12.1.2/tokei-x86_64-unknown-linux-musl.tar.gz",
			"linux/arm64":  "https://github.com/XAMPPRocky/tokei/releases/download/v12.1.2/tokei-aarch64-unknown-linux-gnu.tar.gz",
			"darwin/amd64": "https://github.com/XAMPPRocky/tokei/releases/download/v12.1.2/tokei-x86_64-apple-darwin.tar.gz",
		},
	})

	// hyperfine - nested: hyperfine-v{version}-{triple}/hyperfine
	r.register(&tooldef.Tool{
		Name:            "hyperfine",
		Version:         "1.19.0",
		Description:     "Command-line benchmarking",
		Group:           tooldef.GroupRustTools,
		Format:          tooldef.FormatTarGz,
		BinaryName:      "hyperfine",
		StripComponents: 1,
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/sharkdp/hyperfine/releases/download/v1.19.0/hyperfine-v1.19.0-x86_64-unknown-linux-musl.tar.gz",
			"linux/arm64":  "https://github.com/sharkdp/hyperfine/releases/download/v1.19.0/hyperfine-v1.19.0-aarch64-unknown-linux-musl.tar.gz",
			"darwin/amd64": "https://github.com/sharkdp/hyperfine/releases/download/v1.19.0/hyperfine-v1.19.0-x86_64-apple-darwin.tar.gz",
			"darwin/arm64": "https://github.com/sharkdp/hyperfine/releases/download/v1.19.0/hyperfine-v1.19.0-aarch64-apple-darwin.tar.gz",
		},
	})

	// procs - uses non-standard platform names (aarch64-mac, x86_64-mac); flat zip
	r.register(&tooldef.Tool{
		Name:        "procs",
		Version:     "0.14.8",
		Description: "Modern ps replacement",
		Group:       tooldef.GroupRustTools,
		Format:      tooldef.FormatZip,
		BinaryName:  "procs",
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/dalance/procs/releases/download/v0.14.8/procs-v0.14.8-x86_64-linux.zip",
			"darwin/amd64": "https://github.com/dalance/procs/releases/download/v0.14.8/procs-v0.14.8-x86_64-mac.zip",
			"darwin/arm64": "https://github.com/dalance/procs/releases/download/v0.14.8/procs-v0.14.8-aarch64-mac.zip",
		},
	})

	// bottom - tag has no v prefix; naming: bottom_{triple}.tar.gz; flat archive
	r.register(&tooldef.Tool{
		Name:        "bottom",
		Version:     "0.10.2",
		Description: "System monitor",
		Group:       tooldef.GroupRustTools,
		Format:      tooldef.FormatTarGz,
		BinaryName:  "btm",
		InstallName: "btm",
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/ClementTsang/bottom/releases/download/0.10.2/bottom_x86_64-unknown-linux-musl.tar.gz",
			"linux/arm64":  "https://github.com/ClementTsang/bottom/releases/download/0.10.2/bottom_aarch64-unknown-linux-musl.tar.gz",
			"darwin/amd64": "https://github.com/ClementTsang/bottom/releases/download/0.10.2/bottom_x86_64-apple-darwin.tar.gz",
			"darwin/arm64": "https://github.com/ClementTsang/bottom/releases/download/0.10.2/bottom_aarch64-apple-darwin.tar.gz",
		},
	})

	// gitui - org moved to gitui-org; non-standard naming; flat archive
	r.register(&tooldef.Tool{
		Name:        "gitui",
		Version:     "0.26.3",
		Description: "Terminal UI for git",
		Group:       tooldef.GroupRustTools,
		Format:      tooldef.FormatTarGz,
		BinaryName:  "gitui",
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/gitui-org/gitui/releases/download/v0.26.3/gitui-linux-x86_64.tar.gz",
			"linux/arm64":  "https://github.com/gitui-org/gitui/releases/download/v0.26.3/gitui-linux-aarch64.tar.gz",
			"darwin/amd64": "https://github.com/gitui-org/gitui/releases/download/v0.26.3/gitui-mac-x86.tar.gz",
			"darwin/arm64": "https://github.com/gitui-org/gitui/releases/download/v0.26.3/gitui-mac.tar.gz",
		},
	})

	// dust - only x86_64 macOS available; nested: dust-v{version}-{triple}/dust
	r.register(&tooldef.Tool{
		Name:            "dust",
		Version:         "1.1.1",
		Description:     "Disk usage analyzer",
		Group:           tooldef.GroupRustTools,
		Format:          tooldef.FormatTarGz,
		BinaryName:      "dust",
		StripComponents: 1,
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/bootandy/dust/releases/download/v1.1.1/dust-v1.1.1-x86_64-unknown-linux-musl.tar.gz",
			"linux/arm64":  "https://github.com/bootandy/dust/releases/download/v1.1.1/dust-v1.1.1-aarch64-unknown-linux-musl.tar.gz",
			"darwin/amd64": "https://github.com/bootandy/dust/releases/download/v1.1.1/dust-v1.1.1-x86_64-apple-darwin.tar.gz",
		},
	})

	// bandwhich - flat archive
	r.register(&tooldef.Tool{
		Name:        "bandwhich",
		Version:     "0.23.1",
		Description: "Network bandwidth monitor",
		Group:       tooldef.GroupRustTools,
		Format:      tooldef.FormatTarGz,
		BinaryName:  "bandwhich",
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/imsnif/bandwhich/releases/download/v0.23.1/bandwhich-v0.23.1-x86_64-unknown-linux-musl.tar.gz",
			"linux/arm64":  "https://github.com/imsnif/bandwhich/releases/download/v0.23.1/bandwhich-v0.23.1-aarch64-unknown-linux-musl.tar.gz",
			"darwin/amd64": "https://github.com/imsnif/bandwhich/releases/download/v0.23.1/bandwhich-v0.23.1-x86_64-apple-darwin.tar.gz",
			"darwin/arm64": "https://github.com/imsnif/bandwhich/releases/download/v0.23.1/bandwhich-v0.23.1-aarch64-apple-darwin.tar.gz",
		},
	})

	// sd - nested: sd-v{version}-{triple}/sd
	r.register(&tooldef.Tool{
		Name:            "sd",
		Version:         "1.0.0",
		Description:     "Intuitive find & replace",
		Group:           tooldef.GroupRustTools,
		Format:          tooldef.FormatTarGz,
		BinaryName:      "sd",
		StripComponents: 1,
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/chmln/sd/releases/download/v1.0.0/sd-v1.0.0-x86_64-unknown-linux-musl.tar.gz",
			"linux/arm64":  "https://github.com/chmln/sd/releases/download/v1.0.0/sd-v1.0.0-aarch64-unknown-linux-musl.tar.gz",
			"darwin/amd64": "https://github.com/chmln/sd/releases/download/v1.0.0/sd-v1.0.0-x86_64-apple-darwin.tar.gz",
			"darwin/arm64": "https://github.com/chmln/sd/releases/download/v1.0.0/sd-v1.0.0-aarch64-apple-darwin.tar.gz",
		},
	})

	// tealdeer
	r.register(&tooldef.Tool{
		Name:        "tealdeer",
		Version:     "1.7.1",
		Description: "Fast tldr client",
		Group:       tooldef.GroupRustTools,
		Format:      tooldef.FormatBinary,
		InstallName: "tldr",
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/tealdeer-rs/tealdeer/releases/download/v1.7.1/tealdeer-v1.7.1-x86_64-unknown-linux-musl",
			"linux/arm64":  "https://github.com/tealdeer-rs/tealdeer/releases/download/v1.7.1/tealdeer-v1.7.1-aarch64-unknown-linux-musl",
			"darwin/amd64": "https://github.com/tealdeer-rs/tealdeer/releases/download/v1.7.1/tealdeer-v1.7.1-x86_64-apple-darwin",
			"darwin/arm64": "https://github.com/tealdeer-rs/tealdeer/releases/download/v1.7.1/tealdeer-v1.7.1-aarch64-apple-darwin",
		},
	})

	// xh - nested: xh-v{version}-{triple}/xh
	r.register(&tooldef.Tool{
		Name:            "xh",
		Version:         "0.23.1",
		Description:     "Friendly HTTP client",
		Group:           tooldef.GroupRustTools,
		Format:          tooldef.FormatTarGz,
		BinaryName:      "xh",
		StripComponents: 1,
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/ducaale/xh/releases/download/v0.23.1/xh-v0.23.1-x86_64-unknown-linux-musl.tar.gz",
			"linux/arm64":  "https://github.com/ducaale/xh/releases/download/v0.23.1/xh-v0.23.1-aarch64-unknown-linux-musl.tar.gz",
			"darwin/amd64": "https://github.com/ducaale/xh/releases/download/v0.23.1/xh-v0.23.1-x86_64-apple-darwin.tar.gz",
			"darwin/arm64": "https://github.com/ducaale/xh/releases/download/v0.23.1/xh-v0.23.1-aarch64-apple-darwin.tar.gz",
		},
	})

	// yazi - filename has no version: yazi-{triple}.zip; nested: yazi-{triple}/yazi
	r.register(&tooldef.Tool{
		Name:            "yazi",
		Version:         "25.5.28",
		Description:     "Terminal file manager",
		Group:           tooldef.GroupRustTools,
		Format:          tooldef.FormatZip,
		BinaryName:      "yazi",
		StripComponents: 1,
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/sxyazi/yazi/releases/download/v25.5.28/yazi-x86_64-unknown-linux-musl.zip",
			"linux/arm64":  "https://github.com/sxyazi/yazi/releases/download/v25.5.28/yazi-aarch64-unknown-linux-musl.zip",
			"darwin/amd64": "https://github.com/sxyazi/yazi/releases/download/v25.5.28/yazi-x86_64-apple-darwin.zip",
			"darwin/arm64": "https://github.com/sxyazi/yazi/releases/download/v25.5.28/yazi-aarch64-apple-darwin.zip",
		},
	})

	// atuin - filename has no version: atuin-{triple}.tar.gz; nested: atuin-{triple}/atuin
	r.register(&tooldef.Tool{
		Name:            "atuin",
		Version:         "18.4.0",
		Description:     "Shell history manager",
		Group:           tooldef.GroupRustTools,
		Format:          tooldef.FormatTarGz,
		BinaryName:      "atuin",
		StripComponents: 1,
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/atuinsh/atuin/releases/download/v18.4.0/atuin-x86_64-unknown-linux-musl.tar.gz",
			"linux/arm64":  "https://github.com/atuinsh/atuin/releases/download/v18.4.0/atuin-aarch64-unknown-linux-musl.tar.gz",
			"darwin/amd64": "https://github.com/atuinsh/atuin/releases/download/v18.4.0/atuin-x86_64-apple-darwin.tar.gz",
			"darwin/arm64": "https://github.com/atuinsh/atuin/releases/download/v18.4.0/atuin-aarch64-apple-darwin.tar.gz",
		},
	})

	// zellij - filename has no version: zellij-{triple}.tar.gz; flat archive
	r.register(&tooldef.Tool{
		Name:        "zellij",
		Version:     "0.41.2",
		Description: "Terminal workspace",
		Group:       tooldef.GroupRustTools,
		Format:      tooldef.FormatTarGz,
		BinaryName:  "zellij",
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/zellij-org/zellij/releases/download/v0.41.2/zellij-x86_64-unknown-linux-musl.tar.gz",
			"linux/arm64":  "https://github.com/zellij-org/zellij/releases/download/v0.41.2/zellij-aarch64-unknown-linux-musl.tar.gz",
			"darwin/amd64": "https://github.com/zellij-org/zellij/releases/download/v0.41.2/zellij-x86_64-apple-darwin.tar.gz",
			"darwin/arm64": "https://github.com/zellij-org/zellij/releases/download/v0.41.2/zellij-aarch64-apple-darwin.tar.gz",
		},
	})

	// just - tag has no v prefix; filename: just-{version}-{triple}.tar.gz; flat archive
	r.register(&tooldef.Tool{
		Name:        "just",
		Version:     "1.38.0",
		Description: "Command runner",
		Group:       tooldef.GroupRustTools,
		Format:      tooldef.FormatTarGz,
		BinaryName:  "just",
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/casey/just/releases/download/1.38.0/just-1.38.0-x86_64-unknown-linux-musl.tar.gz",
			"linux/arm64":  "https://github.com/casey/just/releases/download/1.38.0/just-1.38.0-aarch64-unknown-linux-musl.tar.gz",
			"darwin/amd64": "https://github.com/casey/just/releases/download/1.38.0/just-1.38.0-x86_64-apple-darwin.tar.gz",
			"darwin/arm64": "https://github.com/casey/just/releases/download/1.38.0/just-1.38.0-aarch64-apple-darwin.tar.gz",
		},
	})

	// watchexec - filename has no v prefix: watchexec-{version}-{triple}.tar.xz; nested dir
	r.register(&tooldef.Tool{
		Name:            "watchexec",
		Version:         "2.2.1",
		Description:     "File watcher and command executor",
		Group:           tooldef.GroupRustTools,
		Format:          tooldef.FormatTarXz,
		BinaryName:      "watchexec",
		StripComponents: 1,
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/watchexec/watchexec/releases/download/v2.2.1/watchexec-2.2.1-x86_64-unknown-linux-musl.tar.xz",
			"linux/arm64":  "https://github.com/watchexec/watchexec/releases/download/v2.2.1/watchexec-2.2.1-aarch64-unknown-linux-musl.tar.xz",
			"darwin/amd64": "https://github.com/watchexec/watchexec/releases/download/v2.2.1/watchexec-2.2.1-x86_64-apple-darwin.tar.xz",
			"darwin/arm64": "https://github.com/watchexec/watchexec/releases/download/v2.2.1/watchexec-2.2.1-aarch64-apple-darwin.tar.xz",
		},
	})
}
