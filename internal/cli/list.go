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
		Long:  "Show all available tools grouped by registration source (built-in, mise, plugin) and then by category.",
		RunE:  runList,
	}
}

// listStyles holds the lipgloss styles for the list command.
type listStyles struct {
	section   lipgloss.Style
	header    lipgloss.Style
	installed lipgloss.Style
	outdated  lipgloss.Style
	detected  lipgloss.Style
	disabled  lipgloss.Style
	dim       lipgloss.Style
	plugin    lipgloss.Style
}

func newListStyles() listStyles {
	return listStyles{
		section:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")),
		header:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4")),
		installed: lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		outdated:  lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		detected:  lipgloss.NewStyle().Foreground(lipgloss.Color("6")),
		disabled:  lipgloss.NewStyle().Faint(true).Strikethrough(true),
		dim:       lipgloss.NewStyle().Faint(true),
		plugin:    lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("5")),
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

	if ts.Status == state.StatusDetected && ts.DetectedPath != "" {
		detail := ts.DetectedPath
		if ts.DetectedVersion != "" {
			detail += " v" + ts.DetectedVersion
		}
		line += fmt.Sprintf("  (system: %s)", detail)
	}

	if ts.RegistrationSource == state.RegistrationPlugin && ts.PluginFilePath != "" {
		fmt.Print(style.Render(line))
		fmt.Println(s.plugin.Render(fmt.Sprintf("  [plugin: %s]", ts.PluginFilePath)))
		return
	}

	fmt.Println(style.Render(line))
}

// filterGroupsByRegistration returns a copy of groups containing only tools
// with the given registration source. Empty groups are excluded.
func filterGroupsByRegistration(groups []state.GroupState, src state.RegistrationSource) []state.GroupState {
	var result []state.GroupState
	for _, g := range groups {
		var tools []state.ToolState
		for _, t := range g.Tools {
			if t.RegistrationSource == src {
				tools = append(tools, t)
			}
		}
		if len(tools) > 0 {
			result = append(result, state.GroupState{Name: g.Name, Tools: tools})
		}
	}
	return result
}

// printSection renders a meta-section (Built-in / Mise / Plugin) with its groups.
func printSection(title string, groups []state.GroupState, s listStyles) {
	fmt.Println()
	fmt.Println(s.section.Render(fmt.Sprintf("══ %s ══", title)))
	for _, group := range groups {
		fmt.Println(s.header.Render(fmt.Sprintf("\n  [%s]", group.Name)))
		currentSubgroup := ""
		for ti := range group.Tools {
			ts := &group.Tools[ti]
			if ts.Subgroup != "" && ts.Subgroup != currentSubgroup {
				currentSubgroup = ts.Subgroup
				fmt.Println(s.dim.Render(fmt.Sprintf("    ── %s ──", currentSubgroup)))
			}
			renderToolLine(ts, s)
		}
	}
}

// runList resolves the full state of all tools and prints them separated by
// registration source: built-in tools, mise-managed tools, and plugin tools.
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

	if builtins := filterGroupsByRegistration(groups, state.RegistrationBuiltin); len(builtins) > 0 {
		printSection("Built-in Tools", builtins, s)
	}
	if miseGroups := filterGroupsByRegistration(groups, state.RegistrationMise); len(miseGroups) > 0 {
		printSection("Mise-managed  (from .mise.toml)", miseGroups, s)
	}
	if pluginGroups := filterGroupsByRegistration(groups, state.RegistrationPlugin); len(pluginGroups) > 0 {
		printSection("Plugin Tools", pluginGroups, s)
	}

	fmt.Println()
	return nil
}
