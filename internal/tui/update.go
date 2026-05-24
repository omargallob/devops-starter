package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/omargallob/devops-starter/internal/state"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// Update handles messages and key presses. Implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch m.screen {
		case screenGroups:
			return m.updateGroups(msg)
		case screenTools:
			return m.updateTools(msg)
		case screenProgress:
			return m.updateProgress(msg)
		}

	case installCompleteMsg:
		return m.handleInstallComplete(msg)

	case verifyCompleteMsg:
		return m.handleVerifyComplete(msg)

	case spinner.TickMsg:
		return m.handleSpinnerTick(msg)
	}

	return m, nil
}

// ─── Group Screen ────────────────────────────────────────────────────────────

func (m Model) updateGroups(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "j", "down":
		if m.groupCursor < len(m.groups)-1 {
			m.groupCursor++
		}

	case "k", "up":
		if m.groupCursor > 0 {
			m.groupCursor--
		}

	case "enter":
		// Drill into the selected group
		m.selectedGroup = m.groupCursor
		m.toolCursor = 0
		m.screen = screenTools
		m.message = ""
		m.err = nil

	case "a":
		// Install all missing/outdated in highlighted group
		tools := m.installableToolsInGroup(m.groupCursor, false)
		if len(tools) == 0 {
			m.message = "Nothing to install in this group"
			return m, nil
		}
		m.selectedGroup = m.groupCursor
		return m.startProgressInstall(tools, screenGroups)
	}

	return m, nil
}

// ─── Tool Screen ─────────────────────────────────────────────────────────────

func (m Model) updateTools(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	tools := m.selectedGroupTools()

	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "esc", "backspace":
		// Return to group screen
		m.screen = screenGroups
		m.message = ""
		m.err = nil

	case "j", "down":
		if m.toolCursor < len(tools)-1 {
			m.toolCursor++
		}

	case "k", "up":
		if m.toolCursor > 0 {
			m.toolCursor--
		}

	case " ":
		// Toggle selection
		t := m.currentToolInGroup()
		if t != nil && t.Status != state.StatusDisabled && t.Status != state.StatusCurrent {
			t.Selected = !t.Selected
		}

	case "a":
		// Select all installable in this group
		for i := range m.groups[m.selectedGroup].Tools {
			t := &m.groups[m.selectedGroup].Tools[i]
			if t.Status == state.StatusMissing || t.Status == state.StatusOutdated || t.Status == state.StatusDetected {
				t.Selected = true
			}
		}

	case "n":
		// Deselect all
		for i := range m.groups[m.selectedGroup].Tools {
			m.groups[m.selectedGroup].Tools[i].Selected = false
		}

	case "i", "enter":
		// Install selected (or current if none selected)
		selected := m.installableToolsInGroup(m.selectedGroup, true)
		if len(selected) == 0 {
			// Try current tool
			t := m.currentToolInGroup()
			if t != nil && t.Status != state.StatusDisabled && t.Status != state.StatusCurrent {
				selected = append(selected, t.Tool)
			}
		}
		if len(selected) == 0 {
			m.message = "Nothing to install"
			return m, nil
		}
		return m.startProgressInstall(selected, screenTools)

	case "d":
		// Toggle disable
		t := m.currentToolInGroup()
		if t != nil {
			if t.Status == state.StatusDisabled {
				t.Status = state.StatusMissing
			} else {
				t.Status = state.StatusDisabled
				t.Selected = false
			}
		}

	case "v":
		// Verify current tool
		t := m.currentToolInGroup()
		if t != nil && t.Status != state.StatusDisabled {
			m.message = "Verifying " + t.Name + "..."
			return m, verifyToolCmd(t.Name, m.installDir)
		}
	}

	return m, nil
}

// ─── Progress Screen ─────────────────────────────────────────────────────────

func (m Model) updateProgress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If progress is done, any key returns to previous screen
	if m.progressDone {
		m.screen = m.returnToScreen
		m.progressDone = false
		m.progressTools = nil
		m.message = ""
		m.err = nil
		return m, nil
	}

	// While installing, only allow quit
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// ─── Install Orchestration ───────────────────────────────────────────────────

// startProgressInstall transitions to the progress screen and kicks off
// concurrent installations for the given tools.
func (m Model) startProgressInstall(tools []*tooldef.Tool, returnTo screen) (tea.Model, tea.Cmd) {
	m.screen = screenProgress
	m.returnToScreen = returnTo
	m.progressDone = false
	m.progressTools = make([]progressItem, len(tools))

	var cmds []tea.Cmd

	for i, tool := range tools {
		m.progressTools[i] = progressItem{
			Name:   tool.Name,
			Status: progressInstalling,
		}
		m.installing[tool.Name] = true

		// Create spinner
		s := newSpinner()
		m.spinners[tool.Name] = s
		cmds = append(cmds, s.Tick)

		// Start install
		cmds = append(cmds, installToolCmd(m.inst, tool))
	}

	return m, tea.Batch(cmds...)
}

// ─── Message Handlers ────────────────────────────────────────────────────────

// handleInstallComplete processes the result of an installation.
func (m Model) handleInstallComplete(msg installCompleteMsg) (tea.Model, tea.Cmd) {
	delete(m.installing, msg.Name)
	delete(m.spinners, msg.Name)

	// Update progress item
	for i := range m.progressTools {
		if m.progressTools[i].Name == msg.Name {
			if msg.Err != nil {
				m.progressTools[i].Status = progressFailed
				m.progressTools[i].Error = msg.Err
			} else {
				m.progressTools[i].Status = progressDone
			}
			break
		}
	}

	// Update tool state in the group model
	if msg.Err == nil {
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
	}

	// Check if all installs are done
	allDone := true
	for _, item := range m.progressTools {
		if item.Status == progressInstalling || item.Status == progressWaiting {
			allDone = false
			break
		}
	}
	m.progressDone = allDone

	return m, nil
}

// handleVerifyComplete processes the result of a version verification.
func (m Model) handleVerifyComplete(msg verifyCompleteMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.message = msg.Name + ": " + msg.Err.Error()
		return m, nil
	}

	// Update the tool's installed version and recalculate status
	for gi := range m.groups {
		for ti := range m.groups[gi].Tools {
			t := &m.groups[gi].Tools[ti]
			if t.Name == msg.Name {
				t.InstalledVersion = msg.Version
				switch msg.Version {
				case t.DesiredVersion:
					t.Status = state.StatusCurrent
				case "":
					t.Status = state.StatusMissing
				default:
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

// handleSpinnerTick forwards spinner tick messages to all active spinners.
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
