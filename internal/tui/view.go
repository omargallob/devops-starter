package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/omargallob/devops-starter/internal/state"
)

// Styles used throughout the TUI.
var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4"))
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	currentStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	outdatedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	missingStyle  = lipgloss.NewStyle().Faint(true)
	disabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	unknownStyle  = lipgloss.NewStyle().Faint(true)
	selectedStyle = lipgloss.NewStyle().Bold(true)
	cursorStyle   = lipgloss.NewStyle().Background(lipgloss.Color("236"))
	helpStyle     = lipgloss.NewStyle().Faint(true)
	messageStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	errStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
)

// View renders the TUI. Implements tea.Model.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Title bar
	b.WriteString(titleStyle.Render("devops-starter status"))
	b.WriteString(fmt.Sprintf("  [%s/%s]\n", m.platform.OS, m.platform.Arch))
	b.WriteString(strings.Repeat("━", min(80, m.width)))
	b.WriteString("\n\n")

	// Render groups and tools
	items := m.visibleItems()
	for i, item := range items {
		isCursor := i == m.cursor

		var line string
		if item.isGroup {
			line = m.renderGroupHeader(item.groupIdx)
		} else {
			line = m.renderToolRow(item.groupIdx, item.toolIdx)
		}

		if isCursor {
			line = cursorStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	// Message / error
	b.WriteString("\n")
	if m.err != nil {
		b.WriteString(errStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n")
	} else if m.message != "" {
		b.WriteString(messageStyle.Render(m.message))
		b.WriteString("\n")
	}

	// Help footer
	b.WriteString("\n")
	b.WriteString(strings.Repeat("━", min(80, m.width)))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(" ↑↓/jk navigate  space select  i install  a all  n none  d disable  v verify  q quit"))
	b.WriteString("\n")

	return b.String()
}

// renderGroupHeader renders a group header line with collapse indicator and summary.
func (m Model) renderGroupHeader(gi int) string {
	g := m.groups[gi]

	arrow := "▼"
	if g.Collapsed {
		arrow = "▶"
	}

	// Count installed vs total
	var installed, total int
	for _, t := range g.Tools {
		if t.Status != state.StatusDisabled {
			total++
		}
		if t.Status == state.StatusCurrent || t.Status == state.StatusOutdated || t.Status == state.StatusUnknown {
			installed++
		}
	}

	header := fmt.Sprintf("%s [%s]", arrow, g.Name)
	summary := fmt.Sprintf("%d/%d installed", installed, total)

	// Pad to align summary on the right
	padding := 60 - len(header) - len(summary)
	if padding < 2 {
		padding = 2
	}

	return headerStyle.Render(fmt.Sprintf("%s%s%s", header, strings.Repeat(" ", padding), summary))
}

// renderToolRow renders a single tool row with status icon, versions, and description.
func (m Model) renderToolRow(gi, ti int) string {
	t := m.groups[gi].Tools[ti]

	// Status indicator
	var icon string
	var style lipgloss.Style

	// Check if currently installing (show spinner)
	if m.installing[t.Name] {
		s, ok := m.spinners[t.Name]
		if ok {
			icon = s.View()
		} else {
			icon = "◐"
		}
		style = outdatedStyle
	} else {
		switch t.Status {
		case state.StatusCurrent:
			icon = "✓"
			style = currentStyle
		case state.StatusOutdated:
			icon = "↑"
			style = outdatedStyle
		case state.StatusMissing:
			icon = "○"
			style = missingStyle
		case state.StatusDisabled:
			icon = "✗"
			style = disabledStyle
		case state.StatusUnknown:
			icon = "?"
			style = unknownStyle
		}
	}

	// Selection checkbox
	sel := " "
	if t.Selected {
		sel = "●"
	}

	// Version columns
	var versionInfo string
	switch t.Status {
	case state.StatusCurrent:
		versionInfo = fmt.Sprintf("%-10s ═  %-10s", t.InstalledVersion, t.DesiredVersion)
	case state.StatusOutdated:
		versionInfo = fmt.Sprintf("%-10s →  %-10s", t.InstalledVersion, t.DesiredVersion)
	case state.StatusMissing:
		versionInfo = fmt.Sprintf("%-10s →  %-10s", "-", t.DesiredVersion)
	case state.StatusDisabled:
		versionInfo = fmt.Sprintf("%-10s    %-10s", "-", "disabled")
	case state.StatusUnknown:
		versionInfo = fmt.Sprintf("%-10s ?  %-10s", "???", t.DesiredVersion)
	}

	if m.installing[t.Name] {
		versionInfo = fmt.Sprintf("%-10s    %-10s", "...", t.DesiredVersion)
	}

	line := fmt.Sprintf("  %s %s %-16s %s  %s", sel, icon, t.Name, versionInfo, t.Description)

	return style.Render(line)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
