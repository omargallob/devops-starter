package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/omargallob/devops-starter/internal/state"
)

// Update handles messages and key presses. Implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case installCompleteMsg:
		return m.handleInstallComplete(msg)

	case verifyCompleteMsg:
		return m.handleVerifyComplete(msg)

	case spinner.TickMsg:
		return m.handleSpinnerTick(msg)
	}

	return m, nil
}

// handleKey processes keyboard input.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "j", "down":
		items := m.visibleItems()
		if m.cursor < len(items)-1 {
			m.cursor++
		}

	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}

	case " ":
		// Toggle selection on current tool
		items := m.visibleItems()
		if m.cursor >= 0 && m.cursor < len(items) {
			item := items[m.cursor]
			if !item.isGroup {
				t := &m.groups[item.groupIdx].Tools[item.toolIdx]
				if t.Status != state.StatusDisabled && t.Status != state.StatusCurrent {
					t.Selected = !t.Selected
				}
			}
		}

	case "tab", "enter":
		// Toggle collapse on group headers
		items := m.visibleItems()
		if m.cursor >= 0 && m.cursor < len(items) {
			item := items[m.cursor]
			if item.isGroup {
				m.groups[item.groupIdx].Collapsed = !m.groups[item.groupIdx].Collapsed
			}
		}

	case "a":
		// Select all installable (missing + outdated)
		for gi := range m.groups {
			for ti := range m.groups[gi].Tools {
				t := &m.groups[gi].Tools[ti]
				if t.Status == state.StatusMissing || t.Status == state.StatusOutdated {
					t.Selected = true
				}
			}
		}

	case "n":
		// Deselect all
		for gi := range m.groups {
			for ti := range m.groups[gi].Tools {
				m.groups[gi].Tools[ti].Selected = false
			}
		}

	case "i":
		// Install selected tools (or current tool if none selected)
		return m.startInstall()

	case "d":
		// Toggle disable for current tool
		items := m.visibleItems()
		if m.cursor >= 0 && m.cursor < len(items) {
			item := items[m.cursor]
			if !item.isGroup {
				t := &m.groups[item.groupIdx].Tools[item.toolIdx]
				if t.Status == state.StatusDisabled {
					t.Status = state.StatusMissing
				} else {
					t.Status = state.StatusDisabled
					t.Selected = false
				}
			}
		}

	case "v":
		// Verify version of current tool (or selected)
		return m.startVerify()
	}

	return m, nil
}

// startInstall kicks off installation for selected tools.
func (m Model) startInstall() (tea.Model, tea.Cmd) {
	selected := m.selectedTools()

	// If nothing selected, try current tool
	if len(selected) == 0 {
		t := m.currentTool()
		if t != nil && t.Status != state.StatusDisabled && t.Status != state.StatusCurrent {
			selected = append(selected, t.Tool)
		}
	}

	if len(selected) == 0 {
		m.message = "Nothing to install"
		return m, nil
	}

	var cmds []tea.Cmd
	for _, tool := range selected {
		m.installing[tool.Name] = true
		// Create spinner for this tool
		s := spinner.New()
		s.Spinner = spinner.Dot
		s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
		m.spinners[tool.Name] = s
		cmds = append(cmds, installToolCmd(m.inst, tool))
		cmds = append(cmds, s.Tick)
	}

	m.message = ""
	return m, tea.Batch(cmds...)
}

// startVerify kicks off version verification for the current tool or all selected.
func (m Model) startVerify() (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Check selected tools first
	hasSelected := false
	for _, g := range m.groups {
		for _, t := range g.Tools {
			if t.Selected {
				hasSelected = true
				cmds = append(cmds, verifyToolCmd(t.Name, m.installDir))
			}
		}
	}

	// If nothing selected, verify current tool
	if !hasSelected {
		t := m.currentTool()
		if t != nil && t.Status != state.StatusDisabled {
			cmds = append(cmds, verifyToolCmd(t.Name, m.installDir))
		}
	}

	if len(cmds) == 0 {
		m.message = "Nothing to verify"
		return m, nil
	}

	m.message = "Verifying..."
	return m, tea.Batch(cmds...)
}

// handleInstallComplete processes the result of an installation.
func (m Model) handleInstallComplete(msg installCompleteMsg) (tea.Model, tea.Cmd) {
	delete(m.installing, msg.Name)
	delete(m.spinners, msg.Name)

	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}

	// Update tool state to current
	for gi := range m.groups {
		for ti := range m.groups[gi].Tools {
			t := &m.groups[gi].Tools[ti]
			if t.Name == msg.Name {
				t.InstalledVersion = t.DesiredVersion
				t.Status = state.StatusCurrent
				t.Selected = false
			}
		}
	}

	m.message = msg.Name + " installed successfully"
	m.err = nil
	return m, nil
}

// handleVerifyComplete processes the result of a version verification.
func (m Model) handleVerifyComplete(msg verifyCompleteMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		// Tool not installed or can't be probed - leave status as-is
		m.message = msg.Name + ": " + msg.Err.Error()
		return m, nil
	}

	// Update the tool's installed version and recalculate status
	for gi := range m.groups {
		for ti := range m.groups[gi].Tools {
			t := &m.groups[gi].Tools[ti]
			if t.Name == msg.Name {
				t.InstalledVersion = msg.Version
				if msg.Version == t.DesiredVersion {
					t.Status = state.StatusCurrent
				} else if msg.Version == "" {
					t.Status = state.StatusMissing
				} else {
					t.Status = state.StatusOutdated
				}
			}
		}
	}

	// Also update the store
	if m.store != nil {
		_ = m.store.Record(msg.Name, msg.Version)
	}

	m.message = msg.Name + ": detected v" + msg.Version
	return m, nil
}

// handleSpinnerTick forwards spinner tick messages to the appropriate spinner.
func (m Model) handleSpinnerTick(msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	for name, s := range m.spinners {
		var cmd tea.Cmd
		s, cmd = s.Update(msg)
		m.spinners[name] = s
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}
