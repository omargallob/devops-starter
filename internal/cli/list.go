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

// listStyles holds the lipgloss styles for the list command.
type listStyles struct {
	header    lipgloss.Style
	installed lipgloss.Style
	outdated  lipgloss.Style
	detected  lipgloss.Style
	disabled  lipgloss.Style
	dim       lipgloss.Style
}

func newListStyles() listStyles {
	return listStyles{
		header:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4")),
		installed: lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		outdated:  lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		detected:  lipgloss.NewStyle().Foreground(lipgloss.Color("6")),
		disabled:  lipgloss.NewStyle().Faint(true).Strikethrough(true),
		dim:       lipgloss.NewStyle().Faint(true),
	}
}

// toolStatusStyle returns the icon and style for a tool's status.
func toolStatusStyle(ts *state.ToolState, s listStyles) (string, lipgloss.Style) {
	switch ts.Status {
	case state.StatusCurrent:
		return "✓ ", s.installed
	case state.StatusOutdated:
		return "↑ ", s.outdated
	case state.StatusDetected, state.StatusUnknown:
		return "? ", s.detected
	case state.StatusDisabled:
		return "- ", s.disabled
	default:
		return "  ", s.dim
	}
}

// renderToolLine builds and prints a single tool line.
func renderToolLine(ts *state.ToolState, s listStyles) {
	statusIcon, style := toolStatusStyle(ts, s)
	line := fmt.Sprintf("  %s%-20s %-10s %s", statusIcon, ts.Name, ts.DesiredVersion, ts.Description)

	if ts.Tool != nil && ts.Tool.ManagedBy != "" {
		line += fmt.Sprintf("  (managed by %s)", ts.Tool.ManagedBy)
	}

	if ts.Status == state.StatusDetected && ts.DetectedPath != "" {
		detail := ts.DetectedPath
		if ts.DetectedVersion != "" {
			detail += " v" + ts.DetectedVersion
		}
		line += fmt.Sprintf("  (system: %s)", detail)
	}

	fmt.Println(style.Render(line))
}

// runList resolves the full state of all tools (using the state store and
// PATH detection) and prints a formatted table with version and description.
func runList(cmd *cobra.Command, args []string) error {
	cfgPath := cfgFile
	if cfgPath == "" {
		cfgPath = config.Path()
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	info, err := platform.Detect()
	if err != nil {
		return fmt.Errorf("detecting platform: %w", err)
	}

	store, err := state.LoadStore(state.Path())
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	groups := state.ResolveAll(cfg, store, info.Platform)
	s := newListStyles()

	for _, group := range groups {
		fmt.Println(s.header.Render(fmt.Sprintf("\n[%s]", group.Name)))

		currentSubgroup := ""
		for ti := range group.Tools {
			ts := &group.Tools[ti]
			if ts.Subgroup != "" && ts.Subgroup != currentSubgroup {
				currentSubgroup = ts.Subgroup
				fmt.Println(s.dim.Render(fmt.Sprintf("  ── %s ──", currentSubgroup)))
			}
			renderToolLine(ts, s)
		}
	}

	fmt.Println()
	return nil
}
