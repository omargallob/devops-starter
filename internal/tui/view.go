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
	cursorStyle   = lipgloss.NewStyle().Background(lipgloss.Color("236"))
	helpStyle     = lipgloss.NewStyle().Faint(true)
	messageStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	errStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	dimStyle      = lipgloss.NewStyle().Faint(true)
)

// View renders the TUI. Implements tea.Model.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	switch m.screen {
	case screenGroups:
		return m.viewGroups()
	case screenTools:
		return m.viewTools()
	case screenProgress:
		return m.viewProgress()
	}
	return ""
}

// viewGroups renders the category picker screen.
func (m Model) viewGroups() string {
	var b strings.Builder

	// Title bar
	b.WriteString(titleStyle.Render("devops-starter status"))
	b.WriteString(fmt.Sprintf("  [%s/%s]\n", m.platform.OS, m.platform.Arch))
	b.WriteString(strings.Repeat("━", clamp(m.width, 40, 80)))
	b.WriteString("\n\n")
	b.WriteString("  Select a category:\n\n")

	for i, g := range m.groups {
		isCursor := i == m.groupCursor

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

		// Determine group status colour
		var style lipgloss.Style
		switch {
		case installed == total && total > 0:
			style = currentStyle
		case installed > 0:
			style = outdatedStyle
		default:
			style = missingStyle
		}

		// Cursor indicator
		cursor := "  "
		if isCursor {
			cursor = "▸ "
		}

		summary := fmt.Sprintf("%d/%d installed", installed, total)
		name := fmt.Sprintf("%-14s", g.Name)
		line := fmt.Sprintf("  %s%-14s %s", cursor, name, summary)

		rendered := style.Render(line)
		if isCursor {
			rendered = cursorStyle.Render(line)
		}

		b.WriteString(rendered)
		b.WriteString("\n")
	}

	// Message
	b.WriteString("\n")
	if m.err != nil {
		b.WriteString(errStyle.Render(fmt.Sprintf("  Error: %v", m.err)))
		b.WriteString("\n")
	} else if m.message != "" {
		b.WriteString(messageStyle.Render("  " + m.message))
		b.WriteString("\n")
	}

	// Help footer
	b.WriteString("\n")
	b.WriteString(strings.Repeat("━", clamp(m.width, 40, 80)))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(" ↑↓/jk navigate  enter select category  a install group  q quit"))
	b.WriteString("\n")

	return b.String()
}

// viewTools renders the tool list within the selected group.
func (m Model) viewTools() string {
	var b strings.Builder

	g := m.groups[m.selectedGroup]

	// Title bar
	b.WriteString(titleStyle.Render("devops-starter status"))
	b.WriteString(fmt.Sprintf("  [%s/%s]\n", m.platform.OS, m.platform.Arch))
	b.WriteString(strings.Repeat("━", clamp(m.width, 40, 80)))
	b.WriteString("\n\n")

	// Group header with back hint
	b.WriteString(headerStyle.Render(fmt.Sprintf("  [%s]", g.Name)))
	b.WriteString(dimStyle.Render("  ← esc to go back"))
	b.WriteString("\n\n")

	// Tool rows
	for i, t := range g.Tools {
		isCursor := i == m.toolCursor
		line := m.renderToolRow(t)

		if isCursor {
			line = cursorStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	// Message
	b.WriteString("\n")
	if m.err != nil {
		b.WriteString(errStyle.Render(fmt.Sprintf("  Error: %v", m.err)))
		b.WriteString("\n")
	} else if m.message != "" {
		b.WriteString(messageStyle.Render("  " + m.message))
		b.WriteString("\n")
	}

	// Help footer
	b.WriteString("\n")
	b.WriteString(strings.Repeat("━", clamp(m.width, 40, 80)))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(" ↑↓/jk navigate  space select  i install  a all  v verify  d disable  esc back  q quit"))
	b.WriteString("\n")

	return b.String()
}

// viewProgress renders the install progress screen.
func (m Model) viewProgress() string {
	var b strings.Builder

	groupName := ""
	if m.selectedGroup >= 0 && m.selectedGroup < len(m.groups) {
		groupName = m.groups[m.selectedGroup].Name
	}

	// Title
	b.WriteString(titleStyle.Render("devops-starter status"))
	b.WriteString(fmt.Sprintf("  [%s/%s]\n", m.platform.OS, m.platform.Arch))
	b.WriteString(strings.Repeat("━", clamp(m.width, 40, 80)))
	b.WriteString("\n\n")

	if m.progressDone {
		b.WriteString(headerStyle.Render(fmt.Sprintf("  Installed [%s]", groupName)))
	} else {
		b.WriteString(headerStyle.Render(fmt.Sprintf("  Installing [%s]...", groupName)))
	}
	b.WriteString("\n\n")

	// Progress rows
	for _, item := range m.progressTools {
		var icon string
		var style lipgloss.Style

		switch item.Status {
		case progressWaiting:
			icon = "○"
			style = dimStyle
		case progressInstalling:
			if s, ok := m.spinners[item.Name]; ok {
				icon = s.View()
			} else {
				icon = "◐"
			}
			style = outdatedStyle
		case progressDone:
			icon = "✓"
			style = currentStyle
		case progressFailed:
			icon = "✗"
			style = disabledStyle
		}

		statusText := ""
		switch item.Status {
		case progressWaiting:
			statusText = "waiting..."
		case progressInstalling:
			statusText = "installing..."
		case progressDone:
			statusText = "installed"
		case progressFailed:
			if item.Error != nil {
				statusText = fmt.Sprintf("failed: %v", item.Error)
			} else {
				statusText = "failed"
			}
		}

		line := fmt.Sprintf("    %s %-18s %s", icon, item.Name, statusText)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	// Summary when done
	b.WriteString("\n")
	if m.progressDone {
		var success, failed int
		for _, item := range m.progressTools {
			if item.Status == progressDone {
				success++
			} else if item.Status == progressFailed {
				failed++
			}
		}
		summary := messageStyle.Render(fmt.Sprintf("  ✓ %d installed", success))
		if failed > 0 {
			summary += errStyle.Render(fmt.Sprintf("  ✗ %d failed", failed))
		}
		b.WriteString(summary)
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("  Press any key to continue..."))
		b.WriteString("\n")
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(strings.Repeat("━", clamp(m.width, 40, 80)))
	b.WriteString("\n")

	return b.String()
}

// renderToolRow renders a single tool row with status icon, versions, and description.
func (m Model) renderToolRow(t toolModel) string {
	// Status indicator
	var icon string
	var style lipgloss.Style

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

	line := fmt.Sprintf("    %s %s %-16s %s  %s", sel, icon, t.Name, versionInfo, t.Description)
	return style.Render(line)
}

// clamp restricts a value to the range [lo, hi].
func clamp(val, lo, hi int) int {
	if val < lo {
		return lo
	}
	if val > hi {
		return hi
	}
	return val
}
