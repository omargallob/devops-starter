package tui

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/internal/state"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

func newTestModel() Model {
	groups := []state.GroupState{
		{
			Name: "kubernetes",
			Tools: []state.ToolState{
				{Name: "kubectl", Group: "kubernetes", Description: "CLI for K8s", DesiredVersion: "1.31.4", InstalledVersion: "1.31.4", Status: state.StatusCurrent, Tool: &tooldef.Tool{Name: "kubectl", Version: "1.31.4"}},
				{Name: "helm", Group: "kubernetes", Description: "Package manager", DesiredVersion: "3.16.4", InstalledVersion: "", Status: state.StatusMissing, Tool: &tooldef.Tool{Name: "helm", Version: "3.16.4"}},
				{Name: "k9s", Group: "kubernetes", Description: "TUI for K8s", DesiredVersion: "0.32.7", InstalledVersion: "0.31.0", Status: state.StatusOutdated, Tool: &tooldef.Tool{Name: "k9s", Version: "0.32.7"}},
			},
		},
		{
			Name: "utilities",
			Tools: []state.ToolState{
				{Name: "fzf", Group: "utilities", Description: "Fuzzy finder", DesiredVersion: "0.57.0", InstalledVersion: "", Status: state.StatusMissing, Tool: &tooldef.Tool{Name: "fzf", Version: "0.57.0"}},
				{Name: "jq", Group: "utilities", Description: "JSON processor", DesiredVersion: "1.7.1", InstalledVersion: "1.7.1", Status: state.StatusCurrent, Tool: &tooldef.Tool{Name: "jq", Version: "1.7.1"}},
			},
		},
	}

	cfg := config.DefaultConfig()
	store := &state.Store{Tools: map[string]state.InstalledTool{}}
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}

	return NewModel(groups, cfg, nil, store, plat, "/usr/local/bin")
}

func TestNewModel(t *testing.T) {
	m := newTestModel()

	if len(m.groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(m.groups))
	}
	if m.groups[0].Name != "kubernetes" {
		t.Errorf("first group: got %s, want kubernetes", m.groups[0].Name)
	}
	if len(m.groups[0].Tools) != 3 {
		t.Errorf("expected 3 tools in kubernetes, got %d", len(m.groups[0].Tools))
	}
	if m.screen != screenGroups {
		t.Errorf("initial screen: got %d, want screenGroups", m.screen)
	}
}

func TestModel_Init(t *testing.T) {
	m := newTestModel()
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestModel_ViewGroups(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 24

	view := m.View()
	if !strings.Contains(view, "devops-starter status") {
		t.Error("view should contain title")
	}
	if !strings.Contains(view, "kubernetes") {
		t.Error("view should contain group name")
	}
	if !strings.Contains(view, "utilities") {
		t.Error("view should contain second group")
	}
}

func TestModel_NavigateGroups(t *testing.T) {
	m := newTestModel()
	m.width = 80

	// Move cursor down
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = result.(Model)
	if m.groupCursor != 1 {
		t.Errorf("expected groupCursor=1, got %d", m.groupCursor)
	}

	// Move cursor back up
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = result.(Model)
	if m.groupCursor != 0 {
		t.Errorf("expected groupCursor=0, got %d", m.groupCursor)
	}

	// Don't go above 0
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = result.(Model)
	if m.groupCursor != 0 {
		t.Errorf("expected groupCursor=0 (clamped), got %d", m.groupCursor)
	}
}

func TestModel_EnterGroup(t *testing.T) {
	m := newTestModel()
	m.width = 80

	// Press enter on first group
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(Model)

	if m.screen != screenTools {
		t.Errorf("expected screenTools, got %d", m.screen)
	}
	if m.selectedGroup != 0 {
		t.Errorf("expected selectedGroup=0, got %d", m.selectedGroup)
	}
}

func TestModel_ViewTools(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.screen = screenTools
	m.selectedGroup = 0

	view := m.View()
	if !strings.Contains(view, "kubernetes") {
		t.Error("tool view should show group name")
	}
	if !strings.Contains(view, "kubectl") {
		t.Error("tool view should contain kubectl")
	}
	if !strings.Contains(view, "helm") {
		t.Error("tool view should contain helm")
	}
}

func TestModel_NavigateTools(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.screen = screenTools
	m.selectedGroup = 0

	// Move down
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = result.(Model)
	if m.toolCursor != 1 {
		t.Errorf("expected toolCursor=1, got %d", m.toolCursor)
	}

	// Move down again
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = result.(Model)
	if m.toolCursor != 2 {
		t.Errorf("expected toolCursor=2, got %d", m.toolCursor)
	}

	// Don't go past end
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = result.(Model)
	if m.toolCursor != 2 {
		t.Errorf("expected toolCursor=2 (clamped), got %d", m.toolCursor)
	}
}

func TestModel_ToggleSelection(t *testing.T) {
	m := newTestModel()
	m.screen = screenTools
	m.selectedGroup = 0
	m.toolCursor = 1 // helm (missing)

	// Space to toggle
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = result.(Model)

	if !m.groups[0].Tools[1].Selected {
		t.Error("helm should be selected")
	}

	// Toggle back
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = result.(Model)

	if m.groups[0].Tools[1].Selected {
		t.Error("helm should be deselected")
	}
}

func TestModel_CannotSelectCurrentTool(t *testing.T) {
	m := newTestModel()
	m.screen = screenTools
	m.selectedGroup = 0
	m.toolCursor = 0 // kubectl (current)

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = result.(Model)

	if m.groups[0].Tools[0].Selected {
		t.Error("current tool should not be selectable")
	}
}

func TestModel_SelectAll(t *testing.T) {
	m := newTestModel()
	m.screen = screenTools
	m.selectedGroup = 0

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = result.(Model)

	// kubectl is current, should NOT be selected; helm (missing) and k9s (outdated) should be
	if m.groups[0].Tools[0].Selected {
		t.Error("kubectl (current) should not be selected by 'a'")
	}
	if !m.groups[0].Tools[1].Selected {
		t.Error("helm (missing) should be selected by 'a'")
	}
	if !m.groups[0].Tools[2].Selected {
		t.Error("k9s (outdated) should be selected by 'a'")
	}
}

func TestModel_DeselectAll(t *testing.T) {
	m := newTestModel()
	m.screen = screenTools
	m.selectedGroup = 0
	m.groups[0].Tools[1].Selected = true
	m.groups[0].Tools[2].Selected = true

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = result.(Model)

	for _, tool := range m.groups[0].Tools {
		if tool.Selected {
			t.Errorf("tool %s should be deselected after 'n'", tool.Name)
		}
	}
}

func TestModel_BackToGroups(t *testing.T) {
	m := newTestModel()
	m.screen = screenTools
	m.selectedGroup = 0

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = result.(Model)

	if m.screen != screenGroups {
		t.Errorf("expected screenGroups after esc, got %d", m.screen)
	}
}

func TestModel_Quit(t *testing.T) {
	m := newTestModel()

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = result.(Model)

	if !m.quitting {
		t.Error("should be quitting")
	}
	if cmd == nil {
		t.Error("quit should return a command")
	}
}

func TestModel_ViewQuitting(t *testing.T) {
	m := newTestModel()
	m.quitting = true

	view := m.View()
	if view != "" {
		t.Errorf("quitting view should be empty, got %q", view)
	}
}

func TestModel_ViewProgress(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.screen = screenProgress
	m.selectedGroup = 0
	m.progressTools = []progressItem{
		{Name: "helm", Status: progressInstalling},
		{Name: "k9s", Status: progressDone},
	}

	view := m.View()
	if !strings.Contains(view, "Installing") {
		t.Error("progress view should show Installing header")
	}
	if !strings.Contains(view, "helm") {
		t.Error("progress view should show helm")
	}
	if !strings.Contains(view, "k9s") {
		t.Error("progress view should show k9s")
	}
}

func TestModel_ViewProgressDone(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.screen = screenProgress
	m.selectedGroup = 0
	m.progressDone = true
	m.progressTools = []progressItem{
		{Name: "helm", Status: progressDone},
		{Name: "k9s", Status: progressFailed, Error: nil},
	}

	view := m.View()
	if !strings.Contains(view, "Installed") {
		t.Error("done progress view should show Installed header")
	}
	if !strings.Contains(view, "1 installed") {
		t.Error("should show install count")
	}
	if !strings.Contains(view, "1 failed") {
		t.Error("should show failure count")
	}
}

func TestModel_ProgressDismiss(t *testing.T) {
	m := newTestModel()
	m.screen = screenProgress
	m.progressDone = true
	m.returnToScreen = screenGroups
	m.progressTools = []progressItem{{Name: "helm", Status: progressDone}}

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(Model)

	if m.screen != screenGroups {
		t.Errorf("expected return to groups after dismissing progress, got %d", m.screen)
	}
}

func TestModel_WindowSizeMsg(t *testing.T) {
	m := newTestModel()

	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = result.(Model)

	if m.width != 120 {
		t.Errorf("width = %d, want 120", m.width)
	}
	if m.height != 40 {
		t.Errorf("height = %d, want 40", m.height)
	}
}

func TestModel_InstallableToolsInGroup(t *testing.T) {
	m := newTestModel()

	// Without selection filter
	tools := m.installableToolsInGroup(0, false)
	if len(tools) != 2 { // helm (missing) + k9s (outdated)
		t.Errorf("expected 2 installable tools, got %d", len(tools))
	}

	// With selection filter, none selected
	tools = m.installableToolsInGroup(0, true)
	if len(tools) != 0 {
		t.Errorf("expected 0 selected installable tools, got %d", len(tools))
	}

	// Select helm
	m.groups[0].Tools[1].Selected = true
	tools = m.installableToolsInGroup(0, true)
	if len(tools) != 1 {
		t.Errorf("expected 1 selected installable tool, got %d", len(tools))
	}
	if tools[0].Name != "helm" {
		t.Errorf("expected helm, got %s", tools[0].Name)
	}
}

func TestModel_InstallableToolsInGroup_Invalid(t *testing.T) {
	m := newTestModel()

	tools := m.installableToolsInGroup(-1, false)
	if tools != nil {
		t.Error("expected nil for invalid group index")
	}

	tools = m.installableToolsInGroup(99, false)
	if tools != nil {
		t.Error("expected nil for out-of-bounds group index")
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		val, lo, hi, want int
	}{
		{5, 0, 10, 5},
		{-1, 0, 10, 0},
		{15, 0, 10, 10},
		{0, 0, 10, 0},
		{10, 0, 10, 10},
	}
	for _, tc := range tests {
		got := clamp(tc.val, tc.lo, tc.hi)
		if got != tc.want {
			t.Errorf("clamp(%d, %d, %d) = %d, want %d", tc.val, tc.lo, tc.hi, got, tc.want)
		}
	}
}

func TestModel_HandleInstallComplete(t *testing.T) {
	m := newTestModel()
	m.screen = screenProgress
	m.selectedGroup = 0
	m.installing = map[string]bool{"helm": true}
	m.progressTools = []progressItem{
		{Name: "helm", Status: progressInstalling},
	}

	msg := installCompleteMsg{Name: "helm", Err: nil}
	result, _ := m.Update(msg)
	m = result.(Model)

	// Progress item should be marked done
	if m.progressTools[0].Status != progressDone {
		t.Errorf("expected progressDone, got %d", m.progressTools[0].Status)
	}
	if !m.progressDone {
		t.Error("all installs done, progressDone should be true")
	}

	// Tool state should be updated to current
	for _, tool := range m.groups[0].Tools {
		if tool.Name == "helm" {
			if tool.Status != state.StatusCurrent {
				t.Errorf("helm should be StatusCurrent after install, got %s", tool.Status.String())
			}
		}
	}
}

func TestModel_HandleInstallComplete_WithError(t *testing.T) {
	m := newTestModel()
	m.screen = screenProgress
	m.selectedGroup = 0
	m.installing = map[string]bool{"helm": true}
	m.progressTools = []progressItem{
		{Name: "helm", Status: progressInstalling},
	}

	msg := installCompleteMsg{Name: "helm", Err: fmt.Errorf("download failed")}
	result, _ := m.Update(msg)
	m = result.(Model)

	if m.progressTools[0].Status != progressFailed {
		t.Errorf("expected progressFailed, got %d", m.progressTools[0].Status)
	}
}

func TestModel_HandleVerifyComplete(t *testing.T) {
	m := newTestModel()
	m.screen = screenTools
	m.selectedGroup = 0

	msg := verifyCompleteMsg{Name: "helm", Version: "3.16.4", Err: nil}
	result, _ := m.Update(msg)
	m = result.(Model)

	for _, tool := range m.groups[0].Tools {
		if tool.Name == "helm" {
			if tool.InstalledVersion != "3.16.4" {
				t.Errorf("helm installed version should be 3.16.4, got %s", tool.InstalledVersion)
			}
			if tool.Status != state.StatusCurrent {
				t.Errorf("helm should be current after verify, got %s", tool.Status.String())
			}
		}
	}
}

func TestModel_HandleVerifyComplete_Error(t *testing.T) {
	m := newTestModel()
	m.screen = screenTools

	msg := verifyCompleteMsg{Name: "helm", Err: fmt.Errorf("binary not found")}
	result, _ := m.Update(msg)
	m = result.(Model)

	if !strings.Contains(m.message, "binary not found") {
		t.Errorf("expected error in message, got %q", m.message)
	}
}

func TestPrintTable(t *testing.T) {
	groups := []state.GroupState{
		{
			Name: "kubernetes",
			Tools: []state.ToolState{
				{Name: "kubectl", DesiredVersion: "1.31.4", InstalledVersion: "1.31.4", Status: state.StatusCurrent},
				{Name: "helm", DesiredVersion: "3.16.4", InstalledVersion: "", Status: state.StatusMissing},
				{Name: "k9s", DesiredVersion: "0.32.7", InstalledVersion: "0.31.0", Status: state.StatusOutdated},
			},
		},
		{
			Name: "utilities",
			Tools: []state.ToolState{
				{Name: "jq", DesiredVersion: "1.7.1", InstalledVersion: "1.7.1", Status: state.StatusCurrent},
				{Name: "fzf", DesiredVersion: "0.57.0", Status: state.StatusDisabled},
			},
		},
	}

	var buf bytes.Buffer
	PrintTable(&buf, groups)
	output := buf.String()

	// Check header
	if !strings.Contains(output, "GROUP") {
		t.Error("output should contain GROUP header")
	}
	if !strings.Contains(output, "TOOL") {
		t.Error("output should contain TOOL header")
	}

	// Check tool rows
	if !strings.Contains(output, "kubectl") {
		t.Error("output should contain kubectl")
	}
	if !strings.Contains(output, "current") {
		t.Error("output should contain 'current' status")
	}
	if !strings.Contains(output, "missing") {
		t.Error("output should contain 'missing' status")
	}
	if !strings.Contains(output, "outdated") {
		t.Error("output should contain 'outdated' status")
	}
	if !strings.Contains(output, "disabled") {
		t.Error("output should contain 'disabled' status")
	}

	// Check summary
	if !strings.Contains(output, "Summary:") {
		t.Error("output should contain summary line")
	}
	if !strings.Contains(output, "2 current") {
		t.Error("summary should show 2 current")
	}
	if !strings.Contains(output, "1 missing") {
		t.Error("summary should show 1 missing")
	}
	if !strings.Contains(output, "1 outdated") {
		t.Error("summary should show 1 outdated")
	}
	if !strings.Contains(output, "1 disabled") {
		t.Error("summary should show 1 disabled")
	}
}

func TestModel_ToggleDisable(t *testing.T) {
	m := newTestModel()
	m.screen = screenTools
	m.selectedGroup = 0
	m.toolCursor = 1 // helm (missing)

	// Press 'd' to disable
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = result.(Model)

	if m.groups[0].Tools[1].Status != state.StatusDisabled {
		t.Error("helm should be disabled after pressing 'd'")
	}

	// Press 'd' again to re-enable
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = result.(Model)

	if m.groups[0].Tools[1].Status != state.StatusMissing {
		t.Errorf("helm should be missing after toggling disable back, got %s", m.groups[0].Tools[1].Status.String())
	}
}
