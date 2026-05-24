package cli

import (
	"fmt"
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

// runRemove deletes managed binaries and removes tools from the state store.
func runRemove(cmd *cobra.Command, args []string) error {
	// Load config
	cfgPath := cfgFile
	if cfgPath == "" {
		cfgPath = config.ConfigPath()
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Load state store
	store, err := state.LoadStore(state.StatePath())
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	// Resolve tools from registry
	reg := registry.New()

	type removeTarget struct {
		name       string
		binPath    string
		version    string
		systemPath string
	}

	var targets []removeTarget
	var warnings []string

	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))

	for _, name := range args {
		tool, ok := reg.Get(name)
		if !ok {
			warnings = append(warnings, fmt.Sprintf("unknown tool: %s", name))
			continue
		}

		binPath := filepath.Join(cfg.InstallDir, tool.GetInstallName())
		if _, err := os.Stat(binPath); os.IsNotExist(err) {
			// Check if it's in the state store but binary is gone
			if store.GetVersion(name) != "" {
				// Still in state, clean it up
				targets = append(targets, removeTarget{
					name:    name,
					binPath: "",
					version: store.GetVersion(name),
				})
			} else {
				warnings = append(warnings, fmt.Sprintf("%s: not managed (no binary at %s)", name, binPath))
			}
			continue
		}

		// Check for system fallback
		systemPath := ""
		// Temporarily check if removing our binary would expose a system one
		// We look in PATH excluding our install dir
		if path := state.LookupInPath(name); path != "" {
			// Only report as system fallback if it's a different path
			absInstall, _ := filepath.Abs(cfg.InstallDir)
			absFound, _ := filepath.Abs(filepath.Dir(path))
			if absFound != absInstall {
				systemPath = path
			}
		}

		targets = append(targets, removeTarget{
			name:       name,
			binPath:    binPath,
			version:    store.GetVersion(name),
			systemPath: systemPath,
		})
	}

	// Print warnings
	for _, w := range warnings {
		fmt.Println(warnStyle.Render("  ⚠ " + w))
	}

	if len(targets) == 0 {
		fmt.Println("No tools to remove.")
		return nil
	}

	// Show what will be removed
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	fmt.Println(infoStyle.Render(fmt.Sprintf("\nRemoving %d managed tool(s):\n", len(targets))))

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
		fmt.Println(line)
	}
	fmt.Println()

	if dryRun {
		fmt.Println("[dry-run] No changes made.")
		return nil
	}

	if !confirmAction("Proceed with removal?") {
		fmt.Println("Cancelled.")
		return nil
	}

	// Perform removal
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	var removed, failed int
	for _, t := range targets {
		// Remove binary if it exists
		if t.binPath != "" {
			if err := os.Remove(t.binPath); err != nil && !os.IsNotExist(err) {
				fmt.Println(errorStyle.Render(fmt.Sprintf("  ✗ %s: failed to remove binary: %v", t.name, err)))
				failed++
				continue
			}
		}

		// Remove from state store
		if err := store.Remove(t.name); err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("  ✗ %s: failed to update state: %v", t.name, err)))
			failed++
			continue
		}

		msg := fmt.Sprintf("  ✓ %s removed", t.name)
		if t.systemPath != "" {
			msg += fmt.Sprintf(" (now using: %s)", t.systemPath)
		}
		fmt.Println(successStyle.Render(msg))
		removed++
	}

	fmt.Println()
	if removed > 0 {
		fmt.Println(successStyle.Render(fmt.Sprintf("✓ %d tool(s) removed", removed)))
	}
	if failed > 0 {
		return fmt.Errorf("%d removal(s) failed", failed)
	}

	return nil
}
