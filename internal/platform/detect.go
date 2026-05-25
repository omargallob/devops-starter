// Package platform provides OS, architecture, and Linux distribution detection.
package platform

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// Distro represents a Linux distribution.
type Distro string

const (
	DistroUbuntu  Distro = "ubuntu"
	DistroArch    Distro = "arch"
	DistroUnknown Distro = "unknown"
	DistroNone    Distro = "" // macOS
)

// Info contains detected platform information.
type Info struct {
	OS       string           // "linux" or "darwin"
	Arch     string           // "amd64" or "arm64"
	Distro   Distro           // Linux distro or empty for macOS
	Platform tooldef.Platform // Convenience accessor
}

// Detect returns the current platform information.
func Detect() (*Info, error) {
	os := normalizeOS(runtime.GOOS)
	arch := normalizeArch(runtime.GOARCH)

	if os != "linux" && os != "darwin" {
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	if arch != "amd64" && arch != "arm64" {
		return nil, fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	info := &Info{
		OS:   os,
		Arch: arch,
		Platform: tooldef.Platform{
			OS:   os,
			Arch: arch,
		},
	}

	if os == "linux" {
		distro, err := detectDistro()
		if err != nil {
			return nil, fmt.Errorf("detecting distro: %w", err)
		}
		info.Distro = distro
	}

	return info, nil
}

// DetectFromValues creates platform info from explicit values (useful for testing).
func DetectFromValues(osName, arch string, distro Distro) *Info {
	return &Info{
		OS:     osName,
		Arch:   arch,
		Distro: distro,
		Platform: tooldef.Platform{
			OS:   osName,
			Arch: arch,
		},
	}
}

// normalizeOS maps runtime.GOOS values to the canonical strings used
// throughout devops-starter ("linux" or "darwin").
func normalizeOS(goos string) string {
	switch goos {
	case "linux":
		return "linux"
	case "darwin":
		return "darwin"
	default:
		return goos
	}
}

// normalizeArch maps runtime.GOARCH values (and common alternatives like
// "x86_64" and "aarch64") to the canonical "amd64" or "arm64" strings.
func normalizeArch(goarch string) string {
	switch goarch {
	case "amd64", "x86_64":
		return "amd64"
	case "arm64", "aarch64":
		return "arm64"
	default:
		return goarch
	}
}

// detectDistro reads /etc/os-release to determine the Linux distribution.
func detectDistro() (Distro, error) {
	return detectDistroFromFile("/etc/os-release")
}

// detectDistroFromFile parses an os-release file for the ID field.
func detectDistroFromFile(path string) (Distro, error) {
	f, err := os.Open(path)
	if err != nil {
		return DistroUnknown, nil // Not an error - just can't detect
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			id := strings.TrimPrefix(line, "ID=")
			id = strings.Trim(id, `"`)
			return parseDistroID(id), nil
		}
	}

	return DistroUnknown, nil
}

// parseDistroID normalises a distro ID string to our Distro enum.
// Debian-based distros are treated as Ubuntu for package management purposes.
// Arch-based distros (Arch, Manjaro, EndeavourOS) map to DistroArch.
func parseDistroID(id string) Distro {
	switch strings.ToLower(id) {
	case "ubuntu", "debian": // treat debian as ubuntu for package purposes
		return DistroUbuntu
	case "arch", "manjaro", "endeavouros": // arch-based
		return DistroArch
	default:
		return DistroUnknown
	}
}
