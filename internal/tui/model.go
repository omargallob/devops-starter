// Package tui implements the interactive terminal UI for devops-starter status.
// It uses Bubble Tea (charmbracelet/bubbletea) with an Elm-architecture pattern.
//
// The TUI has three screens:
//   - screenGroups: category picker (7 rows, one per tool group)
//   - screenTools: tool list within the selected group
//   - screenProgress: shows install progress with per-tool spinners
package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/internal/installer"
	"github.com/omargallob/devops-starter/internal/state"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// screen represents which view is currently active.
type screen int

const (
	screenGroups   screen = iota // Category picker
	screenTools                  // Tool list within selected group
	screenProgress               // Install progress view
	screenConfirm                // Confirmation prompt before action
)

// Model is the top-level Bubble Tea model for the status TUI.
type Model struct {
	groups      []groupModel
	screen      screen
	groupCursor int // cursor position on group screen (0 to len(groups)-1)
	toolCursor  int // cursor position on tool screen (within selected group)

	// selectedGroup is the index of the group currently being viewed in
	// screenTools or screenProgress.
	selectedGroup int

	width  int
	height int

	// Progress tracking
	installing     map[string]bool          // tools currently being installed
	spinners       map[string]spinner.Model // per-tool spinners during install
	progressTools  []progressItem           // ordered list for progress screen
	progressDone   bool                     // all installs in current batch completed
	returnToScreen screen                   // where to go when progress is dismissed

	// Confirmation screen state
	confirmType  confirmAction   // what action is being confirmed
	confirmTools []*tooldef.Tool // tools pending confirmation
	confirmNames []string        // tool names for remove confirmation

	quitting bool
	message  string // transient status message
	err      error

	// Version and update status
	version         string // current build version
	latestVersion   string // latest available version (populated async)
	updateAvailable bool   // true if a newer version exists

	// Dependencies injected at creation
	cfg        *config.Config
	inst       *installer.Installer
	store      *state.Store
	platform   tooldef.Platform
	installDir string
}

// groupModel represents a group in the TUI.
type groupModel struct {
	Name  string
	Tools []toolModel
}

// toolModel represents a single tool row in the TUI.
type toolModel struct {
	state.ToolState
}

// progressItem tracks an individual tool's install progress.
type progressItem struct {
	Name   string
	Status progressStatus
	Error  error
}

// progressStatus represents the state of a tool during installation.
type progressStatus int

const (
	progressWaiting    progressStatus = iota // queued
	progressInstalling                       // currently downloading/installing
	progressDone                             // completed successfully
	progressFailed                           // failed with error
)

// confirmAction represents the type of action awaiting confirmation.
type confirmAction int

const (
	confirmInstall confirmAction = iota // Install/adopt tools
	confirmRemove                       // Remove managed tools
)

// NewModel creates a new TUI model from the resolved state.
// Plugin tools are extracted from their functional groups and placed in a
// dedicated "plugins" group so the three registration sources stay visually
// separate in both the group picker and the tool list screens.
func NewModel(
	groups []state.GroupState,
	cfg *config.Config,
	inst *installer.Installer,
	store *state.Store,
	platform tooldef.Platform,
	installDir string,
	version string,
) Model {
	gm := make([]groupModel, 0, len(groups))
	var pluginTools []toolModel

	for _, g := range groups {
		var kept []toolModel
		for ti := range g.Tools {
			tm := toolModel{ToolState: g.Tools[ti]}
			if g.Tools[ti].RegistrationSource == state.RegistrationPlugin {
				pluginTools = append(pluginTools, tm)
			} else {
				kept = append(kept, tm)
			}
		}
		if len(kept) > 0 {
			gm = append(gm, groupModel{Name: g.Name, Tools: kept})
		}
	}

	// Append the plugins virtual group last so it stands apart in the picker.
	if len(pluginTools) > 0 {
		gm = append(gm, groupModel{Name: "plugins", Tools: pluginTools})
	}

	return Model{
		groups:     gm,
		screen:     screenGroups,
		installing: make(map[string]bool),
		spinners:   make(map[string]spinner.Model),
		cfg:        cfg,
		inst:       inst,
		store:      store,
		platform:   platform,
		installDir: installDir,
		version:    version,
	}
}

// Init implements tea.Model. Fires an async update check on startup.
func (m Model) Init() tea.Cmd {
	return checkForUpdateCmd(m.version)
}

// selectedGroupTools returns the tools in the currently selected group.
func (m Model) selectedGroupTools() []toolModel {
	if m.selectedGroup < 0 || m.selectedGroup >= len(m.groups) {
		return nil
	}
	return m.groups[m.selectedGroup].Tools
}

// currentToolInGroup returns a pointer to the tool at the tool cursor, or nil.
func (m *Model) currentToolInGroup() *toolModel {
	tools := m.selectedGroupTools()
	if m.toolCursor < 0 || m.toolCursor >= len(tools) {
		return nil
	}
	return &m.groups[m.selectedGroup].Tools[m.toolCursor]
}

// installableToolsInGroup returns tools that can be installed (missing or outdated)
// within the given group index. If onlySelected is true, only returns selected ones.
func (m Model) installableToolsInGroup(gi int, onlySelected bool) []*tooldef.Tool {
	if gi < 0 || gi >= len(m.groups) {
		return nil
	}
	var tools []*tooldef.Tool
	for ti := range m.groups[gi].Tools {
		t := &m.groups[gi].Tools[ti]
		if t.Status == state.StatusDisabled || t.Status == state.StatusCurrent {
			continue
		}
		if onlySelected && !t.Selected {
			continue
		}
		tools = append(tools, t.Tool)
	}
	return tools
}

// newSpinner creates a styled spinner for install progress.
func newSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	return s
}
