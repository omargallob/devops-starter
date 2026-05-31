// table.go provides plain-text (non-interactive) output for devops-starter status --no-tui.
package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/omargallob/devops-starter/internal/state"
)

// statusCounters tracks the number of tools in each status for the summary line.
type statusCounters struct {
	current, outdated, missing, disabled, unknown, detected, linked, unavailable int
}

// increment adds one to the counter matching the given status.
func (c *statusCounters) increment(s state.Status) {
	switch s {
	case state.StatusCurrent:
		c.current++
	case state.StatusOutdated:
		c.outdated++
	case state.StatusMissing:
		c.missing++
	case state.StatusDisabled:
		c.disabled++
	case state.StatusUnknown:
		c.unknown++
	case state.StatusDetected:
		c.detected++
	case state.StatusLinked:
		c.linked++
	case state.StatusUnavailable:
		c.unavailable++
	}
}

// summaryParts returns a formatted summary string from the counters.
func (c *statusCounters) summaryParts() string {
	var parts []string
	if c.current > 0 {
		parts = append(parts, fmt.Sprintf("%d current", c.current))
	}
	if c.linked > 0 {
		parts = append(parts, fmt.Sprintf("%d linked", c.linked))
	}
	if c.detected > 0 {
		parts = append(parts, fmt.Sprintf("%d detected", c.detected))
	}
	if c.outdated > 0 {
		parts = append(parts, fmt.Sprintf("%d outdated", c.outdated))
	}
	if c.missing > 0 {
		parts = append(parts, fmt.Sprintf("%d missing", c.missing))
	}
	if c.disabled > 0 {
		parts = append(parts, fmt.Sprintf("%d disabled", c.disabled))
	}
	if c.unknown > 0 {
		parts = append(parts, fmt.Sprintf("%d unknown", c.unknown))
	}
	if c.unavailable > 0 {
		parts = append(parts, fmt.Sprintf("%d unavailable", c.unavailable))
	}
	return strings.Join(parts, ", ")
}

// formatToolRow formats a single tool's columns for table output.
func formatToolRow(t *state.ToolState) (installed, desired, origin string) {
	installed = formatInstalledVersion(t)
	desired = formatDesiredVersion(t)
	origin = formatOrigin(t)
	return
}

// formatInstalledVersion returns the display string for the installed version column.
func formatInstalledVersion(t *state.ToolState) string {
	switch t.Status {
	case state.StatusDetected:
		return "(system)"
	case state.StatusLinked:
		if t.DetectedVersion != "" {
			return t.DetectedVersion
		}
		return "(linked)"
	default:
		if t.InstalledVersion != "" {
			return t.InstalledVersion
		}
		return "-"
	}
}

// formatDesiredVersion returns the display string for the desired version column.
func formatDesiredVersion(t *state.ToolState) string {
	switch t.Status {
	case state.StatusDisabled:
		return "-"
	case state.StatusUnavailable:
		return "n/a"
	default:
		return t.DesiredVersion
	}
}

// formatOrigin returns the display string for the origin column.
func formatOrigin(t *state.ToolState) string {
	switch {
	case t.Status == state.StatusLinked && t.DetectedPath != "":
		return fmt.Sprintf("linked (%s)", t.DetectedPath)
	case t.Source == state.SourceSystem && t.DetectedPath != "":
		if t.ConflictPolicy != "" {
			return fmt.Sprintf("system (%s) [%s]", t.DetectedPath, t.ConflictPolicy)
		}
		return fmt.Sprintf("system (%s)", t.DetectedPath)
	case t.RegistrationSource == state.RegistrationPlugin:
		return "plugin"
	case t.RegistrationSource == state.RegistrationMise:
		return "mise"
	case t.RegistrationSource == state.RegistrationBuiltin:
		return "builtin"
	default:
		return "-"
	}
}

const tableWidth = 90

// printTableSection writes one registration-source section of the status table.
func printTableSection(w io.Writer, title string, groups []state.GroupState, counters *statusCounters) {
	if len(groups) == 0 {
		return
	}
	fmt.Fprintf(w, "\n── %s %s\n", title, strings.Repeat("─", tableWidth-4-len(title)))
	fmt.Fprintf(w, "%-14s %-18s %-12s %-12s %-10s %s\n",
		"GROUP", "TOOL", "INSTALLED", "DESIRED", "ORIGIN", "STATUS")
	fmt.Fprintf(w, "%s\n", strings.Repeat("─", tableWidth))

	for gi := range groups {
		g := &groups[gi]
		currentSubgroup := ""
		for ti := range g.Tools {
			t := &g.Tools[ti]
			if t.Subgroup != "" && t.Subgroup != currentSubgroup {
				currentSubgroup = t.Subgroup
				fmt.Fprintf(w, "%-14s ── %s ──\n", g.Name, currentSubgroup)
			}
			installed, desired, origin := formatToolRow(t)
			fmt.Fprintf(w, "%-14s %-18s %-12s %-12s %-10s %s\n",
				g.Name, t.Name, installed, desired, origin, t.Status.String())
			counters.increment(t.Status)
		}
	}
}

// filterTableGroups returns groups containing only tools with the given registration source.
// An unset (zero-value) RegistrationSource is treated as RegistrationBuiltin.
func filterTableGroups(groups []state.GroupState, src state.RegistrationSource) []state.GroupState {
	var result []state.GroupState
	for _, g := range groups {
		var tools []state.ToolState
		for _, t := range g.Tools {
			ts := t.RegistrationSource
			if ts == "" {
				ts = state.RegistrationBuiltin
			}
			if ts == src {
				tools = append(tools, t)
			}
		}
		if len(tools) > 0 {
			result = append(result, state.GroupState{Name: g.Name, Tools: tools})
		}
	}
	return result
}

// PrintTable writes a non-interactive plain-text status table to the given writer,
// separated into three sections by registration source.
// Suitable for CI, scripting, or piping to other tools.
func PrintTable(w io.Writer, groups []state.GroupState) {
	var counters statusCounters

	printTableSection(w, "Built-in Tools", filterTableGroups(groups, state.RegistrationBuiltin), &counters)
	printTableSection(w, "Mise-managed (from .mise.toml)", filterTableGroups(groups, state.RegistrationMise), &counters)
	printTableSection(w, "Plugin Tools", filterTableGroups(groups, state.RegistrationPlugin), &counters)

	fmt.Fprintf(w, "%s\n", strings.Repeat("─", tableWidth))
	fmt.Fprintf(w, "Summary: %s\n", counters.summaryParts())
}
