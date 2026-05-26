package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/internal/registry"
	"github.com/omargallob/devops-starter/internal/state"
)

// newRemoveCmd creates the "remove" subcommand which removes managed tool
// binaries and reverts to system-installed versions.
func newRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [tool...]",
		Short: "Remove managed tools and revert to system versions",
		Long: `Remove managed tool binaries from the install directory and clear them
from the state store. If a system-installed version exists on PATH, it will
automatically become the active version again.`,
		Args: cobra.MinimumNArgs(1),
		RunE: runRemove,
	}

	return cmd
}

// runRemove is the entry point that constructs real dependencies and delegates
// to doRemove for the actual logic.
func runRemove(cmd *cobra.Command, args []string) error {
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

	reg := registry.New()

	deps := removeDeps{
		cfg:        cfg,
		registry:   reg,
		store:      store,
		out:        os.Stdout,
		dryRun:     dryRun,
		autoYes:    autoYes,
		installDir: cfg.InstallDir,
	}

	return doRemove(deps, args)
}

// removeTarget describes a single tool to be removed.
type removeTarget struct {
	name       string
	binPath    string
	version    string
	systemPath string
}

// doRemove contains the testable remove logic, separated from Cobra wiring.
func doRemove(deps removeDeps, args []string) error {
	var targets []removeTarget
	var warnings []string

	for _, name := range args {
		tool, ok := deps.registry.Get(name)
		if !ok {
			warnings = append(warnings, fmt.Sprintf("unknown tool: %s", name))
			continue
		}

		binPath := filepath.Join(deps.installDir, tool.GetInstallName())
		if _, err := os.Stat(binPath); os.IsNotExist(err) {
			// Check if it's in the state store but binary is gone
			if deps.store.GetVersion(name) != "" {
				targets = append(targets, removeTarget{
					name:    name,
					binPath: "",
					version: deps.store.GetVersion(name),
				})
			} else {
				warnings = append(warnings, fmt.Sprintf("%s: not managed (no binary at %s)", name, binPath))
			}
			continue
		}

		// Check for system fallback
		systemPath := ""
		if path := state.LookupInPath(name); path != "" {
			absInstall, _ := filepath.Abs(deps.installDir)
			absFound, _ := filepath.Abs(filepath.Dir(path))
			if absFound != absInstall {
				systemPath = path
			}
		}

		targets = append(targets, removeTarget{
			name:       name,
			binPath:    binPath,
			version:    deps.store.GetVersion(name),
			systemPath: systemPath,
		})
	}

	// Print warnings
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	for _, w := range warnings {
		fmt.Fprintln(deps.out, warnStyle.Render("  ⚠ "+w))
	}

	if len(targets) == 0 {
		fmt.Fprintln(deps.out, "No tools to remove.")
		return nil
	}

	// Show what will be removed
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	fmt.Fprintln(deps.out, infoStyle.Render(fmt.Sprintf("\nRemoving %d managed tool(s):\n", len(targets))))

	for _, t := range targets {
		line := fmt.Sprintf("  • %s", t.name)
		if t.version != "" {
			line += fmt.Sprintf(" (managed: v%s)", t.version)
		}
		if t.systemPath != "" {
			line += fmt.Sprintf(" → will revert to system: %s", t.systemPath)
		} else {
			line += " → no system version available"
		}
		fmt.Fprintln(deps.out, line)
	}
	fmt.Fprintln(deps.out)

	if deps.dryRun {
		fmt.Fprintln(deps.out, "[dry-run] No changes made.")
		return nil
	}

	if !deps.autoYes {
		if !confirmAction("Proceed with removal?") {
			fmt.Fprintln(deps.out, "Cancelled.")
			return nil
		}
	}

	// Perform removal
	return executeRemoval(deps.out, deps.store, targets)
}

// executeRemoval performs the actual file deletion and state updates.
func executeRemoval(out io.Writer, store StateStore, targets []removeTarget) error {
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	var removed, failed int
	for _, t := range targets {
		// Remove binary if it exists
		if t.binPath != "" {
			if err := os.Remove(t.binPath); err != nil && !os.IsNotExist(err) {
				fmt.Fprintln(out, errorStyle.Render(fmt.Sprintf("  ✗ %s: failed to remove binary: %v", t.name, err)))
				failed++
				continue
			}
		}

		// Remove from state store
		if err := store.Remove(t.name); err != nil {
			fmt.Fprintln(out, errorStyle.Render(fmt.Sprintf("  ✗ %s: failed to update state: %v", t.name, err)))
			failed++
			continue
		}

		msg := fmt.Sprintf("  ✓ %s removed", t.name)
		if t.systemPath != "" {
			msg += fmt.Sprintf(" (now using: %s)", t.systemPath)
		}
		fmt.Fprintln(out, successStyle.Render(msg))
		removed++
	}

	fmt.Fprintln(out)
	if removed > 0 {
		fmt.Fprintln(out, successStyle.Render(fmt.Sprintf("✓ %d tool(s) removed", removed)))
	}
	if failed > 0 {
		return fmt.Errorf("%d removal(s) failed", failed)
	}

	return nil
}
