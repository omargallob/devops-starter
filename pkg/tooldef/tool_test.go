package tooldef

import "testing"

func TestPlatformString(t *testing.T) {
	tests := []struct {
		platform Platform
		want     string
	}{
		{Platform{OS: "linux", Arch: "amd64"}, "linux/amd64"},
		{Platform{OS: "darwin", Arch: "arm64"}, "darwin/arm64"},
	}

	for _, tt := range tests {
		got := tt.platform.String()
		if got != tt.want {
			t.Errorf("Platform.String() = %q, want %q", got, tt.want)
		}
	}
}

func TestGetBinaryName(t *testing.T) {
	tests := []struct {
		name       string
		tool       Tool
		wantBinary string
	}{
		{
			name:       "defaults to name",
			tool:       Tool{Name: "kubectl"},
			wantBinary: "kubectl",
		},
		{
			name:       "uses explicit binary name",
			tool:       Tool{Name: "bottom", BinaryName: "btm"},
			wantBinary: "btm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tool.GetBinaryName()
			if got != tt.wantBinary {
				t.Errorf("GetBinaryName() = %q, want %q", got, tt.wantBinary)
			}
		})
	}
}

func TestGetInstallName(t *testing.T) {
	tests := []struct {
		name        string
		tool        Tool
		wantInstall string
	}{
		{
			name:        "defaults to name",
			tool:        Tool{Name: "kubectl"},
			wantInstall: "kubectl",
		},
		{
			name:        "uses explicit install name",
			tool:        Tool{Name: "ripgrep", InstallName: "rg"},
			wantInstall: "rg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tool.GetInstallName()
			if got != tt.wantInstall {
				t.Errorf("GetInstallName() = %q, want %q", got, tt.wantInstall)
			}
		})
	}
}

func TestIsMiseManaged(t *testing.T) {
	tests := []struct {
		name string
		tool Tool
		want bool
	}{
		{
			name: "explicit mise mode",
			tool: Tool{Name: "aider", InstallMode: InstallModeMise, MiseBackend: "pipx:aider-chat"},
			want: true,
		},
		{
			name: "eget mode",
			tool: Tool{Name: "ollama", InstallMode: InstallModeEget},
			want: false,
		},
		{
			name: "gh-extension mode",
			tool: Tool{Name: "copilot", InstallMode: InstallModeGhExtension},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tool.IsMiseManaged()
			if got != tt.want {
				t.Errorf("IsMiseManaged() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsGhExtension(t *testing.T) {
	tests := []struct {
		name string
		tool Tool
		want bool
	}{
		{
			name: "gh-extension mode",
			tool: Tool{Name: "copilot-cli", InstallMode: InstallModeGhExtension},
			want: true,
		},
		{
			name: "eget mode",
			tool: Tool{Name: "ollama", InstallMode: InstallModeEget},
			want: false,
		},
		{
			name: "mise mode",
			tool: Tool{Name: "aider", InstallMode: InstallModeMise},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tool.IsGhExtension()
			if got != tt.want {
				t.Errorf("IsGhExtension() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSupportsPlatform(t *testing.T) {
	tests := []struct {
		name     string
		tool     Tool
		platform Platform
		want     bool
	}{
		{
			name:     "nil platforms supports all",
			tool:     Tool{Name: "kubectl"},
			platform: Platform{OS: "linux", Arch: "amd64"},
			want:     true,
		},
		{
			name: "explicit platforms - supported",
			tool: Tool{
				Name: "archlinux-tool",
				Platforms: []Platform{
					{OS: "linux", Arch: "amd64"},
				},
			},
			platform: Platform{OS: "linux", Arch: "amd64"},
			want:     true,
		},
		{
			name: "explicit platforms - not supported",
			tool: Tool{
				Name: "archlinux-tool",
				Platforms: []Platform{
					{OS: "linux", Arch: "amd64"},
				},
			},
			platform: Platform{OS: "darwin", Arch: "arm64"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tool.SupportsPlatform(tt.platform)
			if got != tt.want {
				t.Errorf("SupportsPlatform() = %v, want %v", got, tt.want)
			}
		})
	}
}
