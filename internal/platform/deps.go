// Package platform provides system dependency installation for supported distros.
package platform

import (
	"fmt"
	"os/exec"
	"strings"
)

// SystemDeps are the build dependencies needed for compiling language runtimes
// via mise (Python, Ruby need native extensions).
var systemDepsUbuntu = []string{
	"build-essential",
	"libssl-dev",
	"libffi-dev",
	"zlib1g-dev",
	"libreadline-dev",
	"libyaml-dev",
	"libncurses-dev",
	"curl",
	"git",
	"unzip",
	"xz-utils",
}

var systemDepsArch = []string{
	"base-devel",
	"openssl",
	"libffi",
	"zlib",
	"readline",
	"libyaml",
	"ncurses",
	"curl",
	"git",
	"unzip",
	"xz",
}

// InstallSystemDeps installs the system build dependencies for the detected distro.
func InstallSystemDeps(info *Info, dryRun bool) error {
	switch info.Distro {
	case DistroUbuntu:
		return installApt(systemDepsUbuntu, dryRun)
	case DistroArch:
		return installPacman(systemDepsArch, dryRun)
	case DistroNone:
		// macOS - no system deps needed (Xcode CLT assumed)
		return nil
	default:
		return fmt.Errorf("unsupported distro for system deps: %s", info.Distro)
	}
}

func installApt(packages []string, dryRun bool) error {
	args := append([]string{"apt-get", "install", "-y"}, packages...)
	return runSudo(args, dryRun)
}

func installPacman(packages []string, dryRun bool) error {
	args := append([]string{"pacman", "-Syu", "--noconfirm", "--needed"}, packages...)
	return runSudo(args, dryRun)
}

func runSudo(args []string, dryRun bool) error {
	cmdArgs := append([]string{"sudo"}, args...)
	if dryRun {
		fmt.Printf("[dry-run] %s\n", strings.Join(cmdArgs, " "))
		return nil
	}

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}
