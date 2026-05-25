// table.go provides plain-text (non-interactive) output for devops-starter status --no-tui.
package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/omargallob/devops-starter/internal/state"
)

// PrintTable writes a non-interactive plain-text status table to the given writer.
// Suitable for CI, scripting, or piping to other tools.
func PrintTable(w io.Writer, groups []state.GroupState) {
	// Header
	fmt.Fprintf(w, "%-14s %-18s %-12s %-12s %s\n",
		"GROUP", "TOOL", "INSTALLED", "DESIRED", "STATUS")
	fmt.Fprintf(w, "%s\n", strings.Repeat("─", 76))

	// Counters for summary
	var current, outdated, missing, disabled, unknown, detected int

	for _, g := range groups {
		currentSubgroup := ""
		for _, t := range g.Tools {
			// Insert subgroup separator row when it changes
			if t.Subgroup != "" && t.Subgroup != currentSubgroup {
				currentSubgroup = t.Subgroup
				fmt.Fprintf(w, "%-14s ── %s ──\n", g.Name, currentSubgroup)
			}

			installed := t.InstalledVersion
			if installed == "" {
				installed = "-"
			}
			if t.Status == state.StatusDetected {
				installed = "(system)"
			}

			desired := t.DesiredVersion
			if t.Status == state.StatusDisabled {
				desired = "-"
			}

			fmt.Fprintf(w, "%-14s %-18s %-12s %-12s %s\n",
				g.Name, t.Name, installed, desired, t.Status.String())

			switch t.Status {
			case state.StatusCurrent:
				current++
			case state.StatusOutdated:
				outdated++
			case state.StatusMissing:
				missing++
			case state.StatusDisabled:
				disabled++
			case state.StatusUnknown:
				unknown++
			case state.StatusDetected:
				detected++
			}
		}
	}

	// Summary
	fmt.Fprintf(w, "%s\n", strings.Repeat("─", 76))
	parts := []string{}
	if current > 0 {
		parts = append(parts, fmt.Sprintf("%d current", current))
	}
	if detected > 0 {
		parts = append(parts, fmt.Sprintf("%d detected", detected))
	}
	if outdated > 0 {
		parts = append(parts, fmt.Sprintf("%d outdated", outdated))
	}
	if missing > 0 {
		parts = append(parts, fmt.Sprintf("%d missing", missing))
	}
	if disabled > 0 {
		parts = append(parts, fmt.Sprintf("%d disabled", disabled))
	}
	if unknown > 0 {
		parts = append(parts, fmt.Sprintf("%d unknown", unknown))
	}
	fmt.Fprintf(w, "Summary: %s\n", strings.Join(parts, ", "))
}
