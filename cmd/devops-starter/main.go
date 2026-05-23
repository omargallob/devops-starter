// Command devops-starter is an opinionated cross-platform DevOps tool installer
// and dotfile manager. It downloads pre-built binaries for 60+ DevOps CLI tools
// (no package manager required), verifies checksums, and installs them to
// ~/.local/bin. It also manages shell, git, tmux, and editor configuration
// via symlinks.
//
// Usage:
//
//	devops-starter install          # install all enabled tools
//	devops-starter list             # show available tools
//	devops-starter dotfiles link    # symlink dotfiles to $HOME
//	devops-starter doctor           # check system health
//	devops-starter config init      # create default config
package main

import (
	"os"

	"github.com/omargallob/devops-starter/internal/cli"
)

// main initialises the root Cobra command and executes it.
// A non-nil error from cmd.Execute causes a non-zero exit code.
func main() {
	cmd := cli.NewRootCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
