package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDetect(t *testing.T) {
	info, err := Detect()
	if err != nil {
		t.Fatalf("Detect() error: %v", err)
	}

	if info.OS != runtime.GOOS {
		t.Errorf("OS = %q, want %q", info.OS, runtime.GOOS)
	}

	expectedArch := normalizeArch(runtime.GOARCH)
	if info.Arch != expectedArch {
		t.Errorf("Arch = %q, want %q", info.Arch, expectedArch)
	}

	if info.Platform.OS != info.OS {
		t.Error("Platform.OS should match info.OS")
	}
	if info.Platform.Arch != info.Arch {
		t.Error("Platform.Arch should match info.Arch")
	}
}

func TestDetectFromValues(t *testing.T) {
	info := DetectFromValues("linux", "arm64", DistroUbuntu)

	if info.OS != "linux" {
		t.Errorf("OS = %q, want linux", info.OS)
	}
	if info.Arch != "arm64" {
		t.Errorf("Arch = %q, want arm64", info.Arch)
	}
	if info.Distro != DistroUbuntu {
		t.Errorf("Distro = %q, want ubuntu", info.Distro)
	}
}

func TestNormalizeArch(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"amd64", "amd64"},
		{"x86_64", "amd64"},
		{"arm64", "arm64"},
		{"aarch64", "arm64"},
		{"ppc64", "ppc64"},
	}

	for _, tt := range tests {
		got := normalizeArch(tt.input)
		if got != tt.want {
			t.Errorf("normalizeArch(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDetectDistroFromFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    Distro
	}{
		{
			name:    "ubuntu",
			content: "NAME=\"Ubuntu\"\nVERSION=\"24.04\"\nID=ubuntu\n",
			want:    DistroUbuntu,
		},
		{
			name:    "arch",
			content: "NAME=\"Arch Linux\"\nID=arch\n",
			want:    DistroArch,
		},
		{
			name:    "manjaro as arch",
			content: "NAME=\"Manjaro\"\nID=manjaro\n",
			want:    DistroArch,
		},
		{
			name:    "debian as ubuntu",
			content: "NAME=\"Debian GNU/Linux\"\nID=debian\n",
			want:    DistroUbuntu,
		},
		{
			name:    "quoted ID",
			content: "ID=\"ubuntu\"\n",
			want:    DistroUbuntu,
		},
		{
			name:    "unknown distro",
			content: "ID=fedora\n",
			want:    DistroUnknown,
		},
		{
			name:    "no ID field",
			content: "NAME=\"SomeOS\"\nVERSION=1.0\n",
			want:    DistroUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "os-release")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}

			got, err := detectDistroFromFile(path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("detectDistroFromFile() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectDistroMissingFile(t *testing.T) {
	got, err := detectDistroFromFile("/nonexistent/os-release")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != DistroUnknown {
		t.Errorf("expected DistroUnknown for missing file, got %q", got)
	}
}
