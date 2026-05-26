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
func formatToolRow(t *state.ToolState) (installed, desired, source string) {
	installed = t.InstalledVersion
	if installed == "" {
		installed = "-"
	}
	if t.Status == state.StatusDetected {
		installed = "(system)"
	}
	if t.Status == state.StatusLinked {
		if t.DetectedVersion != "" {
			installed = t.DetectedVersion
		} else {
			installed = "(linked)"
		}
	}

	desired = t.DesiredVersion
	if t.Status == state.StatusDisabled {
		desired = "-"
	}
	if t.Status == state.StatusUnavailable {
		desired = "n/a"
	}

	source = string(t.Source)
	if source == "" {
		source = "-"
	}
	if t.Status == state.StatusLinked && t.DetectedPath != "" {
		source = fmt.Sprintf("linked (%s)", t.DetectedPath)
	} else if t.Source == state.SourceSystem && t.DetectedPath != "" {
		if t.ConflictPolicy != "" {
			source = fmt.Sprintf("system (%s) [%s]", t.DetectedPath, t.ConflictPolicy)
		} else {
			source = fmt.Sprintf("system (%s)", t.DetectedPath)
		}
	}
	return
}

// PrintTable writes a non-interactive plain-text status table to the given writer.
// Suitable for CI, scripting, or piping to other tools.
func PrintTable(w io.Writer, groups []state.GroupState) {
	// Header
	fmt.Fprintf(w, "%-14s %-18s %-12s %-12s %-10s %s\n",
		"GROUP", "TOOL", "INSTALLED", "DESIRED", "SOURCE", "STATUS")
	fmt.Fprintf(w, "%s\n", strings.Repeat("─", 86))

	var counters statusCounters

	for gi := range groups {
		g := &groups[gi]
		currentSubgroup := ""
		for ti := range g.Tools {
			t := &g.Tools[ti]
			if t.Subgroup != "" && t.Subgroup != currentSubgroup {
				currentSubgroup = t.Subgroup
				fmt.Fprintf(w, "%-14s ── %s ──\n", g.Name, currentSubgroup)
			}

			installed, desired, source := formatToolRow(t)
			fmt.Fprintf(w, "%-14s %-18s %-12s %-12s %-10s %s\n",
				g.Name, t.Name, installed, desired, source, t.Status.String())
			counters.increment(t.Status)
		}
	}

	// Summary
	fmt.Fprintf(w, "%s\n", strings.Repeat("─", 86))
	fmt.Fprintf(w, "Summary: %s\n", counters.summaryParts())
}
