package cli

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/omargallob/devops-starter/internal/dotfiles"
)

var dotfilesSource string

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

func newDotfilesLinkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "link",
		Short: "Create dotfile symlinks",
		RunE:  runDotfilesLink,
	}
}

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

func newDotfilesUnlinkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unlink",
		Short: "Remove dotfile symlinks",
		RunE:  runDotfilesUnlink,
	}
}

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

func newDotfilesStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show dotfile symlink status",
		RunE:  runDotfilesStatus,
	}
}

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
