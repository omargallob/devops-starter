package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/internal/tui"
)

// nonInteractive controls whether setup uses defaults without TUI prompts.
var nonInteractive bool

// newSetupCmd creates the "setup" subcommand — an interactive guided setup wizard.
func newSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Interactive guided setup wizard",
		Long: `Interactive guided setup wizard. Walks you through PATH verification,
tool group selection, install directory, and dotfiles linking.

Safe to re-run — updates your existing config without losing per-tool
overrides (version pins, disabled tools).`,
		Example: `  # First-time setup
  devops-starter setup

  # Re-run to change groups
  devops-starter setup

  # Preview what would be configured
  devops-starter setup --dry-run

  # Non-interactive (uses existing config or defaults)
  devops-starter setup --non-interactive`,
		RunE: runSetup,
	}

	cmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "use existing config/defaults without prompts (for CI)")

	return cmd
}

// runSetup is the main logic for the setup command.
func runSetup(cmd *cobra.Command, args []string) error {
	// Resolve config path
	cfgPath := cfgFile
	if cfgPath == "" {
		cfgPath = config.Path()
	}

	// Load existing config (returns defaults if file doesn't exist)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Non-interactive mode: just save existing/default config and exit
	if nonInteractive {
		if dryRun {
			fmt.Printf("[dry-run] Would save config to %s\n", cfgPath)
			return nil
		}
		if err := config.Save(cfg, cfgPath); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Printf("Config saved to %s\n", cfgPath)
		return nil
	}

	// Launch interactive TUI wizard
	model := tui.NewSetupModel(cfg, cfgPath, dryRun)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("running setup wizard: %w", err)
	}

	// Check if user confirmed
	if m, ok := finalModel.(tui.SetupModel); ok {
		if !m.IsConfirmed() {
			fmt.Println("Setup cancelled.")
			return nil
		}
	}

	return nil
}
