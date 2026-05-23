package cli

import (
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	dryRun  bool
	only    string
)

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

	root.AddCommand(
		newInstallCmd(),
		newListCmd(),
		newDotfilesCmd(),
		newDoctorCmd(),
		newConfigCmd(),
	)

	return root
}
