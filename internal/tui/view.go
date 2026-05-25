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
	detectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("5")) // magenta/purple for system-installed
	cursorStyle   = lipgloss.NewStyle().Background(lipgloss.Color("236"))
	helpStyle     = lipgloss.NewStyle().Faint(true)
	helpKeyStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true) // cyan+bold for keybinds
	messageStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	errStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	dimStyle      = lipgloss.NewStyle().Faint(true)
	statusStyle   = lipgloss.NewStyle().Faint(true)
	updateStyle   = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("3"))
)

// helpBinding represents a single key-action pair in the help bar.
type helpBinding struct {
	key  string
	desc string
}

// renderHelp formats a list of key bindings with colored keys and dim descriptions.
func renderHelp(bindings []helpBinding) string {
	var b strings.Builder
	b.WriteString(" ")
	for i, bind := range bindings {
		if i > 0 {
			b.WriteString("  ")
		}
		b.WriteString(helpKeyStyle.Render(bind.key))
		b.WriteString(helpStyle.Render(" " + bind.desc))
	}
	return b.String()
}

// View renders the TUI. Implements tea.Model.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var content string
	var helpLine string

	switch m.screen {
	case screenGroups:
		content, helpLine = m.viewGroupsContent()
	case screenTools:
		content, helpLine = m.viewToolsContent()
	case screenProgress:
		content, helpLine = m.viewProgressContent()
	case screenConfirm:
		content, helpLine = m.viewConfirmContent()
	}

	return m.composeFullScreen(content, helpLine)
}

// composeFullScreen arranges the content into a full-screen layout with a
// status bar pinned to the bottom.
func (m Model) composeFullScreen(content, helpLine string) string {
	width := m.contentWidth()
	height := m.height
	if height <= 0 {
		height = 24 // sensible default before first WindowSizeMsg
	}

	// Build the footer (help + separator + status bar): 3 lines
	separator := strings.Repeat("━", width)
	statusBar := m.renderStatusBar(width)

	footer := separator + "\n" + helpLine + "\n" + separator + "\n" + statusBar

	// Count lines in content and footer
	contentLines := strings.Count(content, "\n")
	footerLines := strings.Count(footer, "\n") + 1 // +1 for the last line without \n

	// Pad content to fill the screen above the footer
	padLines := height - contentLines - footerLines
	if padLines < 0 {
		padLines = 0
	}
	padding := strings.Repeat("\n", padLines)

	return content + padding + footer
}

// renderStatusBar produces the single-line status bar with version and update info.
func (m Model) renderStatusBar(width int) string {
	// Format version display - make dev mode obvious
	var left string
	var leftLen int
	if m.version == "dev" || m.version == "" {
		devStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
		label := fmt.Sprintf("devops-starter v%s", devVersionLabel())
		left = devStyle.Render(label)
		leftLen = len(label)
	} else {
		label := fmt.Sprintf("devops-starter v%s", m.version)
		left = statusStyle.Render(label)
		leftLen = len(label)
	}

	var right string
	if m.updateAvailable && m.latestVersion != "" {
		right = updateStyle.Render(fmt.Sprintf("update available: v%s", m.latestVersion))
	}

	if right == "" {
		return left
	}

	// Calculate spacing to right-align the update notice
	rightLen := len(fmt.Sprintf("update available: v%s", m.latestVersion))
	gap := width - leftLen - rightLen
	if gap < 2 {
		gap = 2
	}

	return left + strings.Repeat(" ", gap) + right
}

// viewGroupsContent renders the category picker screen content and help line.
func (m Model) viewGroupsContent() (string, string) {
	var b strings.Builder

	// Title bar
	b.WriteString(titleStyle.Render("devops-starter status"))
	fmt.Fprintf(&b, "  [%s/%s]\n", m.platform.OS, m.platform.Arch)
	b.WriteString(strings.Repeat("━", m.contentWidth()))
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
			if t.Status == state.StatusCurrent || t.Status == state.StatusOutdated || t.Status == state.StatusUnknown || t.Status == state.StatusDetected {
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

	helpLine := renderHelp([]helpBinding{
		{"↑↓/jk", "navigate"},
		{"enter", "select category"},
		{"a", "install group"},
		{"q", "quit"},
	})
	return b.String(), helpLine
}

// viewToolsContent renders the tool list within the selected group.
func (m Model) viewToolsContent() (string, string) {
	var b strings.Builder

	g := m.groups[m.selectedGroup]

	// Title bar
	b.WriteString(titleStyle.Render("devops-starter status"))
	fmt.Fprintf(&b, "  [%s/%s]\n", m.platform.OS, m.platform.Arch)
	b.WriteString(strings.Repeat("━", m.contentWidth()))
	b.WriteString("\n\n")

	// Group header with back hint
	b.WriteString(headerStyle.Render(fmt.Sprintf("  [%s]", g.Name)))
	b.WriteString(dimStyle.Render("  ← esc to go back"))
	b.WriteString("\n\n")

	// Tool rows
	currentSubgroup := ""
	for i, t := range g.Tools {
		// Insert subgroup header when it changes (non-selectable divider)
		if t.Subgroup != "" && t.Subgroup != currentSubgroup {
			currentSubgroup = t.Subgroup
			b.WriteString(dimStyle.Render(fmt.Sprintf("    ── %s ──", currentSubgroup)))
			b.WriteString("\n")
		}

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

	helpLine := renderHelp([]helpBinding{
		{"↑↓/jk", "navigate"},
		{"space", "select"},
		{"i", "install"},
		{"r", "remove"},
		{"a", "all"},
		{"v", "verify"},
		{"d", "disable"},
		{"esc", "back"},
		{"q", "quit"},
	})
	return b.String(), helpLine
}

// viewProgressContent renders the install progress screen.
func (m Model) viewProgressContent() (string, string) {
	var b strings.Builder

	groupName := ""
	if m.selectedGroup >= 0 && m.selectedGroup < len(m.groups) {
		groupName = m.groups[m.selectedGroup].Name
	}

	// Title
	b.WriteString(titleStyle.Render("devops-starter status"))
	fmt.Fprintf(&b, "  [%s/%s]\n", m.platform.OS, m.platform.Arch)
	b.WriteString(strings.Repeat("━", m.contentWidth()))
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
	var helpLine string
	if m.progressDone {
		var success, failed int
		for _, item := range m.progressTools {
			switch item.Status {
			case progressDone:
				success++
			case progressFailed:
				failed++
			}
		}
		summary := messageStyle.Render(fmt.Sprintf("  ✓ %d installed", success))
		if failed > 0 {
			summary += errStyle.Render(fmt.Sprintf("  ✗ %d failed", failed))
		}
		b.WriteString(summary)
		b.WriteString("\n")
		helpLine = helpStyle.Render("  Press any key to continue...")
	} else {
		helpLine = helpStyle.Render("  Installing...")
	}

	return b.String(), helpLine
}

// viewConfirmContent renders the confirmation prompt screen.
func (m Model) viewConfirmContent() (string, string) {
	var b strings.Builder

	// Title bar
	b.WriteString(titleStyle.Render("devops-starter status"))
	fmt.Fprintf(&b, "  [%s/%s]\n", m.platform.OS, m.platform.Arch)
	b.WriteString(strings.Repeat("━", m.contentWidth()))
	b.WriteString("\n\n")

	switch m.confirmType {
	case confirmInstall:
		// Determine if this is an adopt (detected tools) or regular install
		hasDetected := false
		for _, tool := range m.confirmTools {
			for _, g := range m.groups {
				for _, t := range g.Tools {
					if t.Name == tool.Name && t.Status == state.StatusDetected {
						hasDetected = true
					}
				}
			}
		}

		action := "install"
		if hasDetected {
			action = "install/adopt"
		}

		b.WriteString(headerStyle.Render(fmt.Sprintf("  Confirm %s (%d tool(s)):", action, len(m.confirmTools))))
		b.WriteString("\n\n")

		for _, tool := range m.confirmTools {
			detail := ""
			for _, g := range m.groups {
				for _, t := range g.Tools {
					if t.Name == tool.Name && t.Status == state.StatusDetected && t.DetectedPath != "" {
						detail = fmt.Sprintf("  (adopting from system: %s", t.DetectedPath)
						if t.DetectedVersion != "" {
							detail += " v" + t.DetectedVersion
						}
						detail += ")"
					}
				}
			}
			line := fmt.Sprintf("    • %s %s%s", tool.Name, tool.Version, detail)
			b.WriteString(line)
			b.WriteString("\n")
		}

	case confirmRemove:
		b.WriteString(headerStyle.Render(fmt.Sprintf("  Confirm removal (%d tool(s)):", len(m.confirmNames))))
		b.WriteString("\n\n")

		for _, name := range m.confirmNames {
			detail := ""
			for _, g := range m.groups {
				for _, t := range g.Tools {
					if t.Name == name {
						if t.InstalledVersion != "" {
							detail += fmt.Sprintf(" (managed: v%s)", t.InstalledVersion)
						}
						// Check if system version will take over
						if t.DetectedPath != "" {
							detail += fmt.Sprintf(" → system: %s", t.DetectedPath)
						} else {
							detail += " → no system version"
						}
					}
				}
			}
			line := fmt.Sprintf("    • %s%s", name, detail)
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	// Prompt
	b.WriteString("\n")
	b.WriteString(messageStyle.Render("  Proceed? (y/n)"))
	b.WriteString("\n")

	helpLine := renderHelp([]helpBinding{
		{"y", "confirm"},
		{"n/esc", "cancel"},
	})
	return b.String(), helpLine
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
	case state.StatusDetected:
		icon = "~"
		style = detectedStyle
	case state.StatusUnavailable:
		icon = "⊘"
		style = dimStyle
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
	case state.StatusDetected:
		versionInfo = fmt.Sprintf("%-10s ~  %-10s", "(system)", t.DesiredVersion)
	case state.StatusUnavailable:
		versionInfo = fmt.Sprintf("%-10s    %-10s", "-", "n/a")
	}

	// Source label
	var sourceLabel string
	switch {
	case t.Status == state.StatusUnavailable:
		sourceLabel = "[not available on this platform — install manually or use Docker]"
	case t.Source == state.SourceMise:
		sourceLabel = "[mise]"
	case t.Source == state.SourceSystem:
		sourceLabel = fmt.Sprintf("[system: %s]", t.DetectedPath)
	case t.Source == state.SourceManaged:
		sourceLabel = "[managed]"
	}

	line := fmt.Sprintf("    %s %s %-16s %s  %s  %s", sel, icon, t.Name, versionInfo, sourceLabel, t.Description)
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

// contentWidth returns the width to use for full-screen layout elements
// (separators, status bar). Uses the actual terminal width from the last
// WindowSizeMsg, with a sensible fallback before the first resize event.
func (m Model) contentWidth() int {
	switch {
	case m.width <= 0:
		return 80 // default before first WindowSizeMsg
	case m.width < 40:
		return 40 // minimum usable width
	default:
		return m.width
	}
}
