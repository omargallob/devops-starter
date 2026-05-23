// Package tui implements the interactive terminal UI for devops-starter status.
// It uses Bubble Tea (charmbracelet/bubbletea) with an Elm-architecture pattern.
package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/internal/installer"
	"github.com/omargallob/devops-starter/internal/state"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// Model is the top-level Bubble Tea model for the status TUI.
type Model struct {
	groups     []groupModel
	cursor     int  // index into the flat list of visible items
	width      int
	height     int
	quitting   bool
	installing map[string]bool          // tools currently being installed
	spinners   map[string]spinner.Model // per-tool spinners
	err        error
	message    string // transient status message

	// Dependencies injected at creation
	cfg       *config.Config
	inst      *installer.Installer
	store     *state.Store
	platform  tooldef.Platform
	installDir string
}

// groupModel represents a collapsible group in the TUI.
type groupModel struct {
	Name      string
	Collapsed bool
	Tools     []toolModel
}

// toolModel represents a single tool row in the TUI.
type toolModel struct {
	state.ToolState
}

// NewModel creates a new TUI model from the resolved state.
func NewModel(
	groups []state.GroupState,
	cfg *config.Config,
	inst *installer.Installer,
	store *state.Store,
	platform tooldef.Platform,
	installDir string,
) Model {
	gm := make([]groupModel, 0, len(groups))
	for _, g := range groups {
		tools := make([]toolModel, 0, len(g.Tools))
		for _, t := range g.Tools {
			tools = append(tools, toolModel{ToolState: t})
		}
		gm = append(gm, groupModel{
			Name:  g.Name,
			Tools: tools,
		})
	}

	return Model{
		groups:     gm,
		installing: make(map[string]bool),
		spinners:   make(map[string]spinner.Model),
		cfg:        cfg,
		inst:       inst,
		store:      store,
		platform:   platform,
		installDir: installDir,
	}
}

// Init implements tea.Model. It returns no initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// visibleItem represents either a group header or a tool row for cursor navigation.
type visibleItem struct {
	isGroup    bool
	groupIdx   int
	toolIdx    int // -1 for group headers
}

// visibleItems returns the flat list of navigable items based on collapse state.
func (m Model) visibleItems() []visibleItem {
	var items []visibleItem
	for gi, g := range m.groups {
		items = append(items, visibleItem{isGroup: true, groupIdx: gi, toolIdx: -1})
		if !g.Collapsed {
			for ti := range g.Tools {
				items = append(items, visibleItem{isGroup: false, groupIdx: gi, toolIdx: ti})
			}
		}
	}
	return items
}

// selectedTools returns all tools that are currently selected (checkbox toggled).
func (m Model) selectedTools() []*tooldef.Tool {
	var tools []*tooldef.Tool
	for _, g := range m.groups {
		for _, t := range g.Tools {
			if t.Selected && t.Status != state.StatusDisabled && t.Status != state.StatusCurrent {
				tools = append(tools, t.Tool)
			}
		}
	}
	return tools
}

// currentTool returns the tool at the cursor position, or nil if on a group header.
func (m Model) currentTool() *toolModel {
	items := m.visibleItems()
	if m.cursor < 0 || m.cursor >= len(items) {
		return nil
	}
	item := items[m.cursor]
	if item.isGroup {
		return nil
	}
	return &m.groups[item.groupIdx].Tools[item.toolIdx]
}
