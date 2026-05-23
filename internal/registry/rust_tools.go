package registry

import (
	"fmt"

	"github.com/omargallob/devops-starter/pkg/tooldef"
)

func rustTargetTriple(os, arch string) string {
	switch {
	case os == "linux" && arch == "amd64":
		return "x86_64-unknown-linux-musl"
	case os == "linux" && arch == "arm64":
		return "aarch64-unknown-linux-musl"
	case os == "darwin" && arch == "amd64":
		return "x86_64-apple-darwin"
	case os == "darwin" && arch == "arm64":
		return "aarch64-apple-darwin"
	}
	return ""
}

func rustURLs(org, name, version string, format tooldef.ArchiveFormat, binaryPath string) *tooldef.Tool {
	platforms := []struct {
		os, arch string
	}{
		{"linux", "amd64"},
		{"linux", "arm64"},
		{"darwin", "amd64"},
		{"darwin", "arm64"},
	}

	urls := make(map[string]string)
	for _, p := range platforms {
		triple := rustTargetTriple(p.os, p.arch)
		var ext string
		switch format {
		case tooldef.FormatTarGz:
			ext = ".tar.gz"
		case tooldef.FormatTarXz:
			ext = ".tar.xz"
		case tooldef.FormatZip:
			ext = ".zip"
		case tooldef.FormatBinary:
			ext = ""
		}
		url := fmt.Sprintf("https://github.com/%s/releases/download/v%s/%s-v%s-%s%s", org, version, name, version, triple, ext)
		urls[p.os+"/"+p.arch] = url
	}

	t := &tooldef.Tool{
		Name:    name,
		Version: version,
		Group:   tooldef.GroupRustTools,
		Format:  format,
		URLs:    urls,
	}
	if binaryPath != "" {
		t.BinaryName = binaryPath
	}
	return t
}

func registerRustTools(r *Registry) {
	// bat
	t := rustURLs("sharkdp/bat", "bat", "0.24.0", tooldef.FormatTarGz, "")
	t.Description = "Cat clone with syntax highlighting"
	// Binary is in bat-v0.24.0-{target}/bat - we set BinaryName per the pattern
	// Since BinaryName can't be dynamic per-platform easily, leave it for the installer to resolve
	t.BinaryName = "bat"
	r.register(t)

	// eza
	{
		t := &tooldef.Tool{
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
		}
		r.register(t)
	}

	// fd
	t = rustURLs("sharkdp/fd", "fd", "10.2.0", tooldef.FormatTarGz, "fd")
	t.Description = "Fast find alternative"
	r.register(t)

	// ripgrep
	t = rustURLs("BurntSushi/ripgrep", "ripgrep", "14.1.1", tooldef.FormatTarGz, "rg")
	t.Description = "Fast grep alternative"
	t.InstallName = "rg"
	r.register(t)

	// delta
	t = rustURLs("dandavison/delta", "delta", "0.18.2", tooldef.FormatTarGz, "delta")
	t.Description = "Git diff viewer"
	r.register(t)

	// zoxide
	t = rustURLs("ajeetdsouza/zoxide", "zoxide", "0.9.6", tooldef.FormatTarGz, "zoxide")
	t.Description = "Smarter cd command"
	r.register(t)

	// starship
	t = rustURLs("starship/starship", "starship", "1.21.1", tooldef.FormatTarGz, "starship")
	t.Description = "Cross-shell prompt"
	r.register(t)

	// tokei
	t = rustURLs("XAMPPRocky/tokei", "tokei", "13.0.0-alpha.5", tooldef.FormatTarGz, "tokei")
	t.Description = "Code statistics"
	r.register(t)

	// hyperfine
	t = rustURLs("sharkdp/hyperfine", "hyperfine", "1.19.0", tooldef.FormatTarGz, "hyperfine")
	t.Description = "Command-line benchmarking"
	r.register(t)

	// procs
	t = rustURLs("dalance/procs", "procs", "0.14.8", tooldef.FormatZip, "procs")
	t.Description = "Modern ps replacement"
	r.register(t)

	// bottom
	t = rustURLs("ClementTsang/bottom", "bottom", "0.10.2", tooldef.FormatTarGz, "btm")
	t.Description = "System monitor"
	t.InstallName = "btm"
	r.register(t)

	// gitui
	t = rustURLs("extrawurst/gitui", "gitui", "0.26.3", tooldef.FormatTarGz, "gitui")
	t.Description = "Terminal UI for git"
	r.register(t)

	// dust
	t = rustURLs("bootandy/dust", "dust", "1.1.1", tooldef.FormatTarGz, "dust")
	t.Description = "Disk usage analyzer"
	r.register(t)

	// bandwhich
	t = rustURLs("imsnif/bandwhich", "bandwhich", "0.23.1", tooldef.FormatTarGz, "bandwhich")
	t.Description = "Network bandwidth monitor"
	r.register(t)

	// sd
	t = rustURLs("chmln/sd", "sd", "1.0.0", tooldef.FormatTarGz, "sd")
	t.Description = "Intuitive find & replace"
	r.register(t)

	// tealdeer
	{
		t := &tooldef.Tool{
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
		}
		r.register(t)
	}

	// xh
	t = rustURLs("ducaale/xh", "xh", "0.23.1", tooldef.FormatTarGz, "xh")
	t.Description = "Friendly HTTP client"
	r.register(t)

	// yazi
	t = rustURLs("sxyazi/yazi", "yazi", "25.3.3", tooldef.FormatZip, "yazi")
	t.Description = "Terminal file manager"
	r.register(t)

	// atuin
	t = rustURLs("atuinsh/atuin", "atuin", "18.4.0", tooldef.FormatTarGz, "atuin")
	t.Description = "Shell history manager"
	r.register(t)

	// zellij
	t = rustURLs("zellij-org/zellij", "zellij", "0.41.2", tooldef.FormatTarGz, "zellij")
	t.Description = "Terminal workspace"
	r.register(t)

	// just
	t = rustURLs("casey/just", "just", "1.38.0", tooldef.FormatTarGz, "just")
	t.Description = "Command runner"
	r.register(t)

	// watchexec
	t = rustURLs("watchexec/watchexec", "watchexec", "2.2.1", tooldef.FormatTarXz, "watchexec")
	t.Description = "File watcher and command executor"
	r.register(t)
}
