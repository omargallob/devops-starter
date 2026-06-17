package installer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// ErrGlazePkgNotAvailable is returned when gpk is not installed or no matching
// package manager is available for the current platform. The caller should fall
// back to eget-based installation.
var ErrGlazePkgNotAvailable = errors.New("glazepkg (gpk) not available")

// platformManagerOrder returns glazepkg manager names in preferred order for
// the given OS and Linux distro.
func platformManagerOrder(os, distro string) []string {
	switch os {
	case "darwin":
		return []string{"brew", "macports"}
	case "linux":
		switch distro {
		case "ubuntu":
			return []string{"apt", "snap", "flatpak", "nix"}
		case "arch":
			return []string{"pacman", "aur", "flatpak", "nix"}
		default:
			return []string{"apt", "dnf", "pacman", "snap", "flatpak", "nix"}
		}
	default:
		return nil
	}
}

// managerBinary maps a glazepkg manager name to the binary checked for availability.
func managerBinary(manager string) string {
	switch manager {
	case "apt":
		return "apt-get"
	case "chocolatey":
		return "choco"
	case "macports":
		return "port"
	default:
		return manager
	}
}

// resolveBestManager iterates the priority list and returns the first manager
// that is both declared in tool.PackageNames and has its binary available on PATH.
func resolveBestManager(tool *tooldef.Tool, managers []string) (manager, pkgName string, err error) {
	for _, m := range managers {
		pkg, ok := tool.PackageNames[m]
		if !ok {
			continue
		}
		if _, lookErr := exec.LookPath(managerBinary(m)); lookErr != nil {
			continue
		}
		return m, pkg, nil
	}
	return "", "", ErrGlazePkgNotAvailable
}

// installViaGlazePkg installs a tool using glazepkg as the package manager
// abstraction layer. It selects the best available native package manager for
// the current platform and delegates to `gpk install --manager <m> --yes <pkg>`.
//
// Returns ErrGlazePkgNotAvailable if:
//   - PreferNativeManagers is false on the installer
//   - gpk is not in PATH
//   - none of the tool's declared PackageNames have an available manager
func (inst *Installer) installViaGlazePkg(ctx context.Context, tool *tooldef.Tool) error {
	if !inst.PreferNativeManagers {
		return ErrGlazePkgNotAvailable
	}

	gpkBin, err := exec.LookPath("gpk")
	if err != nil {
		return ErrGlazePkgNotAvailable
	}

	managers := platformManagerOrder(inst.Platform.OS, inst.Distro)
	manager, pkgName, err := resolveBestManager(tool, managers)
	if err != nil {
		return ErrGlazePkgNotAvailable
	}

	args := []string{"install", "--manager", manager, "--yes", pkgName}
	cmd := exec.CommandContext(ctx, gpkBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gpk install %s via %s: %w", pkgName, manager, err)
	}
	return nil
}

// glazePkgFallback runs when glazepkg is unavailable: it falls back to eget
// (if Repo is set) or eget-url (if URLTemplate/URLs is set).
func (inst *Installer) glazePkgFallback(ctx context.Context, tool *tooldef.Tool) error {
	if tool.Repo != "" {
		return inst.installViaEget(ctx, tool)
	}
	if tool.URLTemplate != "" || len(tool.URLs) > 0 {
		return inst.installViaEgetURL(ctx, tool)
	}
	return fmt.Errorf("glazepkg not available and no binary download fallback configured for %s", tool.Name)
}
