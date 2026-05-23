package cli

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/omargallob/devops-starter/internal/dotfiles"
)

// dotfilesSource holds the --source flag value for overriding the dotfiles directory.
var dotfilesSource string

// newDotfilesCmd creates the "dotfiles" parent command with link/unlink/status subcommands.
func newDotfilesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dotfiles",
		Short: "Manage dotfile symlinks",
		Long:  "Create, remove, and inspect dotfile symlinks between the repo and your home directory.",
	}

	cmd.PersistentFlags().StringVar(&dotfilesSource, "source", "", "path to dotfiles source directory")

	cmd.AddCommand(
		newDotfilesLinkCmd(),
		newDotfilesUnlinkCmd(),
		newDotfilesStatusCmd(),
	)

	return cmd
}

// resolveDotfilesSource determines the dotfiles source directory by checking:
// 1. The --source flag (highest priority)
// 2. Relative paths from the executable location
// 3. The current working directory
// Returns an error if no dotfiles/ directory can be found.
func resolveDotfilesSource() (string, error) {
	if dotfilesSource != "" {
		return dotfilesSource, nil
	}

	// Try relative to executable
	exe, err := os.Executable()
	if err == nil {
		// Assume repo structure: binary is in cmd/ or root, dotfiles/ is at repo root
		candidates := []string{
			exe + "/../dotfiles",
			exe + "/../../dotfiles",
		}
		for _, c := range candidates {
			if info, err := os.Stat(c); err == nil && info.IsDir() {
				return c, nil
			}
		}
	}

	// Try current working directory
	cwd, err := os.Getwd()
	if err == nil {
		candidate := cwd + "/dotfiles"
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("cannot find dotfiles source directory; use --source flag")
}

// newDotfilesLinkCmd creates the "dotfiles link" subcommand.
func newDotfilesLinkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "link",
		Short: "Create dotfile symlinks",
		RunE:  runDotfilesLink,
	}
}

// runDotfilesLink creates symlinks for all default mappings from the dotfiles
// source directory to $HOME. Conflicting files are backed up to ~/.dotfiles.bak.
func runDotfilesLink(cmd *cobra.Command, args []string) error {
	source, err := resolveDotfilesSource()
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	linker := dotfiles.NewLinker(source, home, dryRun)
	mappings := dotfiles.DefaultMappings()

	results := linker.Link(mappings)
	printDotfileResults(results)
	return nil
}

// newDotfilesUnlinkCmd creates the "dotfiles unlink" subcommand.
func newDotfilesUnlinkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unlink",
		Short: "Remove dotfile symlinks",
		RunE:  runDotfilesUnlink,
	}
}

// runDotfilesUnlink removes only symlinks that point to our source files.
// Non-symlinks and symlinks pointing elsewhere are left untouched.
func runDotfilesUnlink(cmd *cobra.Command, args []string) error {
	source, err := resolveDotfilesSource()
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	linker := dotfiles.NewLinker(source, home, dryRun)
	mappings := dotfiles.DefaultMappings()

	results := linker.Unlink(mappings)
	printDotfileResults(results)
	return nil
}

// newDotfilesStatusCmd creates the "dotfiles status" subcommand.
func newDotfilesStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show dotfile symlink status",
		RunE:  runDotfilesStatus,
	}
}

// runDotfilesStatus inspects each mapping and prints whether it is linked,
// conflicting, missing, or broken.
func runDotfilesStatus(cmd *cobra.Command, args []string) error {
	source, err := resolveDotfilesSource()
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	linker := dotfiles.NewLinker(source, home, false)
	mappings := dotfiles.DefaultMappings()

	results := linker.Status(mappings)
	printDotfileResults(results)
	return nil
}

// printDotfileResults renders link results with coloured indicators:
//   - green checkmark for linked
//   - red X for conflict
//   - dim circle for missing
//   - yellow bang for broken symlinks
func printDotfileResults(results []dotfiles.LinkResult) {
	linkedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	conflictStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	missingStyle := lipgloss.NewStyle().Faint(true)
	brokenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))

	for _, r := range results {
		var style lipgloss.Style
		var indicator string

		switch r.Status {
		case dotfiles.StatusLinked:
			style = linkedStyle
			indicator = "✓"
		case dotfiles.StatusConflict:
			style = conflictStyle
			indicator = "✗"
		case dotfiles.StatusMissing:
			style = missingStyle
			indicator = "○"
		case dotfiles.StatusBroken:
			style = brokenStyle
			indicator = "!"
		}

		line := fmt.Sprintf("  %s %-30s → %s  [%s]", indicator, r.Mapping.Dest, r.Mapping.Source, r.Status)
		fmt.Println(style.Render(line))
	}
}
