package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/internal/installer"
	"github.com/omargallob/devops-starter/internal/platform"
	"github.com/omargallob/devops-starter/internal/state"
	"github.com/omargallob/devops-starter/internal/tui"
)

var (
	statusNoTUI  bool
	statusVerify bool
)

// newStatusCmd creates the "status" subcommand which launches an interactive
// TUI showing current vs desired state of all tools, or prints a plain table
// with --no-tui.
func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Interactive tool status dashboard",
		Long: `Launch a TUI showing current vs desired state of all tools with
install actions. Use --no-tui for plain text output suitable for CI.`,
		RunE: runStatus,
	}

	cmd.Flags().BoolVar(&statusNoTUI, "no-tui", false, "print plain table (no interactive UI)")
	cmd.Flags().BoolVar(&statusVerify, "verify", false, "detect versions by running installed binaries")

	return cmd
}

// runStatus is the main logic for the status command.
func runStatus(cmd *cobra.Command, args []string) error {
	// Detect platform
	info, err := platform.Detect()
	if err != nil {
		return fmt.Errorf("detecting platform: %w", err)
	}

	// Load config
	cfgPath := cfgFile
	if cfgPath == "" {
		cfgPath = config.Path()
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Load state store
	store, err := state.LoadStore(state.Path())
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	// If --verify, detect real versions from binaries
	if statusVerify {
		state.VerifyAll(store, cfg.InstallDir)
		// Persist verified state
		if err := store.Save(); err != nil {
			return fmt.Errorf("saving state after verify: %w", err)
		}
	}

	// Resolve full state
	groups := state.ResolveAll(cfg, store, info.Platform)

	// Non-interactive mode
	if statusNoTUI {
		tui.PrintTable(os.Stdout, groups)
		return nil
	}

	// Create installer for TUI install actions
	inst := installer.New(
		cfg.InstallDir,
		info.Platform,
		installer.WithStateStore(store),
	)

	// Launch TUI
	model := tui.NewModel(groups, cfg, inst, store, info.Platform, cfg.InstallDir, Version())
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("running TUI: %w", err)
	}

	return nil
}
