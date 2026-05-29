// Package cli implements the devops-starter command-line interface using Cobra.
// It defines the root command and all subcommands (install, list, dotfiles, doctor, config).
// Each subcommand file provides a constructor (newXxxCmd) and a run function (runXxx).
package cli

import (
	"github.com/spf13/cobra"
)

// Global flags shared across subcommands.
var (
	cfgFile string // --config: override path to config YAML
	dryRun  bool   // --dry-run: preview mode, no side effects
	autoYes bool   // --yes/-y: skip confirmation prompts
	only    string // --only: filter tools to a single group (install command)
)

// NewRootCmd constructs the top-level Cobra command with all subcommands registered.
// Persistent flags (--config, --dry-run) are inherited by every subcommand.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "devops-starter",
		Short: "Opinionated cross-platform DevOps tool installer",
		Long: `devops-starter installs and manages DevOps tools across Linux (Ubuntu, Arch)
and macOS (Intel, Apple Silicon). It downloads pre-built binaries,
verifies checksums, and manages dotfile configurations.`,
	}

	root.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/devops-starter/config.yaml)")
	root.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "preview actions without executing")
	root.PersistentFlags().BoolVarP(&autoYes, "yes", "y", false, "skip confirmation prompts")

	root.AddCommand(
		newSetupCmd(),
		newInstallCmd(),
		newListCmd(),
		newAdoptCmd(),
		newRemoveCmd(),
		newDotfilesCmd(),
		newDoctorCmd(),
		newConfigCmd(),
		newStatusCmd(),
		newPluginsCmd(),
	)

	return root
}
