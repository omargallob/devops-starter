package cli

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/internal/platform"
	"github.com/omargallob/devops-starter/internal/state"
)

// newListCmd creates the "list" subcommand which displays all available tools
// grouped by category, with a checkmark for already-installed tools.
func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available tools",
		Long:  "Show all available tools grouped by category with installation status.",
		RunE:  runList,
	}
}

// runList resolves the full state of all tools (using the state store and
// PATH detection) and prints a formatted table with version and description.
func runList(cmd *cobra.Command, args []string) error {
	cfgPath := cfgFile
	if cfgPath == "" {
		cfgPath = config.ConfigPath()
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Detect platform
	info, err := platform.Detect()
	if err != nil {
		return fmt.Errorf("detecting platform: %w", err)
	}

	// Load state store
	store, err := state.LoadStore(state.StatePath())
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	// Resolve full state (same logic as status command)
	groups := state.ResolveAll(cfg, store, info.Platform)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4"))
	installedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	outdatedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	detectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	disabledStyle := lipgloss.NewStyle().Faint(true).Strikethrough(true)
	dimStyle := lipgloss.NewStyle().Faint(true)

	for _, group := range groups {
		fmt.Println(headerStyle.Render(fmt.Sprintf("\n[%s]", group.Name)))

		for _, ts := range group.Tools {
			var statusIcon string
			var style lipgloss.Style

			switch ts.Status {
			case state.StatusCurrent:
				statusIcon = "✓ "
				style = installedStyle
			case state.StatusOutdated:
				statusIcon = "↑ "
				style = outdatedStyle
			case state.StatusDetected:
				statusIcon = "? "
				style = detectedStyle
			case state.StatusUnknown:
				statusIcon = "? "
				style = detectedStyle
			case state.StatusDisabled:
				statusIcon = "- "
				style = disabledStyle
			default: // StatusMissing
				statusIcon = "  "
				style = dimStyle
			}

			line := fmt.Sprintf("  %s%-20s %-10s %s", statusIcon, ts.Name, ts.DesiredVersion, ts.Description)

			// Append system binary info for detected tools
			if ts.Status == state.StatusDetected && ts.DetectedPath != "" {
				detail := ts.DetectedPath
				if ts.DetectedVersion != "" {
					detail += " v" + ts.DetectedVersion
				}
				line += fmt.Sprintf("  (system: %s)", detail)
			}

			fmt.Println(style.Render(line))
		}
	}

	fmt.Println()
	return nil
}
