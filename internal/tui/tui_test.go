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
				{Name: "kustomize", Group: "kubernetes", Description: "K8s config", DesiredVersion: "5.5.0", InstalledVersion: "", Status: state.StatusDetected, Tool: &tooldef.Tool{Name: "kustomize", Version: "5.5.0"}},
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

	return NewModel(groups, cfg, nil, store, plat, "/usr/local/bin", "0.1.4")
}

func TestNewModel(t *testing.T) {
	m := newTestModel()

	if len(m.groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(m.groups))
	}
	if m.groups[0].Name != "kubernetes" {
		t.Errorf("first group: got %s, want kubernetes", m.groups[0].Name)
	}
	if len(m.groups[0].Tools) != 4 {
		t.Errorf("expected 4 tools in kubernetes, got %d", len(m.groups[0].Tools))
	}
	if m.screen != screenGroups {
		t.Errorf("initial screen: got %d, want screenGroups", m.screen)
	}
}

func TestModel_Init(t *testing.T) {
	m := newTestModel()
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return an update check command")
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

	// Move down to last (index 3)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = result.(Model)
	if m.toolCursor != 3 {
		t.Errorf("expected toolCursor=3, got %d", m.toolCursor)
	}

	// Don't go past end
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = result.(Model)
	if m.toolCursor != 3 {
		t.Errorf("expected toolCursor=3 (clamped), got %d", m.toolCursor)
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

	// kubectl is current, should NOT be selected; helm (missing), k9s (outdated), kustomize (detected) should be
	if m.groups[0].Tools[0].Selected {
		t.Error("kubectl (current) should not be selected by 'a'")
	}
	if !m.groups[0].Tools[1].Selected {
		t.Error("helm (missing) should be selected by 'a'")
	}
	if !m.groups[0].Tools[2].Selected {
		t.Error("k9s (outdated) should be selected by 'a'")
	}
	if !m.groups[0].Tools[3].Selected {
		t.Error("kustomize (detected) should be selected by 'a'")
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
	if len(tools) != 3 { // helm (missing) + k9s (outdated) + kustomize (detected)
		t.Errorf("expected 3 installable tools, got %d", len(tools))
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

func TestModel_ViewToolsDetected(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.screen = screenTools
	m.selectedGroup = 0

	view := m.View()
	if !strings.Contains(view, "kustomize") {
		t.Error("tool view should contain kustomize")
	}
	if !strings.Contains(view, "(system)") {
		t.Error("detected tool should show (system) in version column")
	}
	if !strings.Contains(view, "~") {
		t.Error("detected tool should show ~ icon")
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

func TestModel_ViewToolsSourceLabels(t *testing.T) {
	groups := []state.GroupState{
		{
			Name: "languages",
			Tools: []state.ToolState{
				{Name: "go", Group: "languages", Description: "Go language", DesiredVersion: "1.26.3", InstalledVersion: "1.26.3", Status: state.StatusCurrent, Source: state.SourceMise, Tool: &tooldef.Tool{Name: "go", Version: "1.26.3", ManagedBy: "mise"}},
				{Name: "python", Group: "languages", Description: "Python", DesiredVersion: "3.12", InstalledVersion: "3.12.7", Status: state.StatusCurrent, Source: state.SourceSystem, DetectedPath: "/usr/bin/python3", Tool: &tooldef.Tool{Name: "python", Version: "3.12"}},
				{Name: "mise", Group: "languages", Description: "Tool manager", DesiredVersion: "2025.1.6", InstalledVersion: "2025.1.6", Status: state.StatusCurrent, Source: state.SourceManaged, Tool: &tooldef.Tool{Name: "mise", Version: "2025.1.6"}},
			},
		},
	}

	cfg := config.DefaultConfig()
	store := &state.Store{Tools: map[string]state.InstalledTool{}}
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}
	m := NewModel(groups, cfg, nil, store, plat, "/usr/local/bin", "0.1.4")
	m.width = 100
	m.screen = screenTools
	m.selectedGroup = 0

	view := m.View()

	if !strings.Contains(view, "[mise]") {
		t.Error("tool view should show [mise] label for mise-managed tools")
	}
	if !strings.Contains(view, "[system: /usr/bin/python3]") {
		t.Error("tool view should show [system: /path] label for system-detected tools")
	}
	if !strings.Contains(view, "[managed]") {
		t.Error("tool view should show [managed] label for devops-starter-managed tools")
	}
}

func TestPrintTable_SourceColumn(t *testing.T) {
	groups := []state.GroupState{
		{
			Name: "languages",
			Tools: []state.ToolState{
				{Name: "go", DesiredVersion: "1.26.3", InstalledVersion: "1.26.3", Status: state.StatusCurrent, Source: state.SourceMise},
				{Name: "mise", DesiredVersion: "2025.1.6", InstalledVersion: "2025.1.6", Status: state.StatusCurrent, Source: state.SourceManaged},
				{Name: "node", DesiredVersion: "22", InstalledVersion: "", Status: state.StatusDetected, Source: state.SourceSystem, DetectedPath: "/usr/local/bin/node"},
				{Name: "ruby", DesiredVersion: "3.3", InstalledVersion: "", Status: state.StatusMissing, Source: state.SourceNone},
			},
		},
	}

	var buf bytes.Buffer
	PrintTable(&buf, groups)
	output := buf.String()

	// Check SOURCE header is present
	if !strings.Contains(output, "SOURCE") {
		t.Error("output should contain SOURCE header")
	}

	// Check source values appear in output
	if !strings.Contains(output, "mise") {
		t.Error("output should contain 'mise' source for mise-managed tool")
	}
	if !strings.Contains(output, "managed") {
		t.Error("output should contain 'managed' source for devops-starter tool")
	}
	if !strings.Contains(output, "system (/usr/local/bin/node)") {
		t.Error("output should contain 'system (/path)' for system-detected tool")
	}
	// Missing tools with no source should show "-"
	if !strings.Contains(output, "-") {
		t.Error("output should contain '-' for tools with no source")
	}
}

func TestPrintTable_LinkedAndConflict(t *testing.T) {
	groups := []state.GroupState{
		{
			Name: "infra",
			Tools: []state.ToolState{
				{
					Name:            "terraform",
					DesiredVersion:  "1.9.0",
					DetectedVersion: "1.8.5",
					Status:          state.StatusLinked,
					Source:          state.SourceSystem,
					DetectedPath:    "/usr/local/bin/terraform",
					ConflictPolicy:  "link",
				},
				{
					Name:            "vault",
					DesiredVersion:  "1.17.0",
					DetectedVersion: "1.16.0",
					Status:          state.StatusDetected,
					Source:          state.SourceSystem,
					DetectedPath:    "/usr/local/bin/vault",
					ConflictPolicy:  "skip",
				},
				{
					Name:           "consul",
					DesiredVersion: "1.19.0",
					Status:         state.StatusDetected,
					Source:         state.SourceSystem,
					DetectedPath:   "/usr/local/bin/consul",
					ConflictPolicy: "overwrite",
				},
			},
		},
	}

	var buf bytes.Buffer
	PrintTable(&buf, groups)
	output := buf.String()

	// Linked tool should show "linked" source
	if !strings.Contains(output, "linked") {
		t.Error("output should contain 'linked' for a StatusLinked tool")
	}
	if !strings.Contains(output, "/usr/local/bin/terraform") {
		t.Error("output should show detected path for linked tool")
	}

	// Conflict policy should appear for system-detected tools with policy set
	if !strings.Contains(output, "[skip]") {
		t.Error("output should show [skip] conflict policy for vault")
	}
	if !strings.Contains(output, "[overwrite]") {
		t.Error("output should show [overwrite] conflict policy for consul")
	}

	// Summary should count linked tools
	if !strings.Contains(output, "1 linked") {
		t.Error("summary should show 1 linked")
	}
	if !strings.Contains(output, "2 detected") {
		t.Error("summary should show 2 detected")
	}
}

func TestModel_ViewToolsLinkedLabel(t *testing.T) {
	groups := []state.GroupState{
		{
			Name: "infra",
			Tools: []state.ToolState{
				{
					Name:            "terraform",
					Group:           "infra",
					Description:     "Infrastructure as code",
					DesiredVersion:  "1.9.0",
					DetectedVersion: "1.8.5",
					Status:          state.StatusLinked,
					Source:          state.SourceSystem,
					DetectedPath:    "/usr/local/bin/terraform",
					ConflictPolicy:  "link",
					Tool:            &tooldef.Tool{Name: "terraform", Version: "1.9.0"},
				},
			},
		},
	}

	cfg := config.DefaultConfig()
	store := &state.Store{Tools: map[string]state.InstalledTool{}}
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}
	m := NewModel(groups, cfg, nil, store, plat, "/usr/local/bin", "0.1.4")
	m.width = 120
	m.screen = screenTools
	m.selectedGroup = 0

	view := m.View()

	if !strings.Contains(view, "[linked: /usr/local/bin/terraform]") {
		t.Error("tool view should show [linked: /path] label for linked tools")
	}
}

// ─── Tests for dev_version.go ────────────────────────────────────────────────

func TestDevVersionLabel(t *testing.T) {
	label := devVersionLabel()
	// The manifest has "0.1.4", so dev label should be "0.1.5-dev"
	if !strings.HasSuffix(label, "-dev") {
		t.Errorf("devVersionLabel() = %q, want suffix '-dev'", label)
	}
	if label == "dev" {
		t.Error("devVersionLabel() should include a version number, not just 'dev'")
	}
	// Verify it produces a semver-like prefix
	parts := strings.SplitN(strings.TrimSuffix(label, "-dev"), ".", 3)
	if len(parts) != 3 {
		t.Errorf("devVersionLabel() = %q, want format X.Y.Z-dev", label)
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{9, "9"},
		{10, "10"},
		{42, "42"},
		{100, "100"},
		{999, "999"},
	}
	for _, tt := range tests {
		got := itoa(tt.input)
		if got != tt.want {
			t.Errorf("itoa(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// ─── Tests for view.go helpers ───────────────────────────────────────────────

func TestContentWidth(t *testing.T) {
	tests := []struct {
		name  string
		width int
		want  int
	}{
		{"zero defaults to 80", 0, 80},
		{"negative defaults to 80", -1, 80},
		{"below minimum clamps to 40", 20, 40},
		{"at minimum returns 40", 40, 40},
		{"normal width passes through", 120, 120},
		{"wide terminal passes through", 300, 300},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{width: tt.width}
			got := m.contentWidth()
			if got != tt.want {
				t.Errorf("contentWidth() with width=%d: got %d, want %d", tt.width, got, tt.want)
			}
		})
	}
}

func TestRenderHelp(t *testing.T) {
	bindings := []helpBinding{
		{"q", "quit"},
		{"enter", "select"},
	}
	result := renderHelp(bindings)

	// Keys should appear in output
	if !strings.Contains(result, "q") {
		t.Error("renderHelp should contain key 'q'")
	}
	if !strings.Contains(result, "enter") {
		t.Error("renderHelp should contain key 'enter'")
	}
	// Descriptions should appear
	if !strings.Contains(result, "quit") {
		t.Error("renderHelp should contain description 'quit'")
	}
	if !strings.Contains(result, "select") {
		t.Error("renderHelp should contain description 'select'")
	}
}

func TestRenderStatusBar_ReleaseVersion(t *testing.T) {
	m := Model{
		version: "1.2.3",
		width:   80,
	}
	bar := m.renderStatusBar(80)
	if !strings.Contains(bar, "v1.2.3") {
		t.Errorf("status bar should contain version, got: %q", bar)
	}
	if !strings.Contains(bar, "devops-starter") {
		t.Error("status bar should contain app name")
	}
}

func TestRenderStatusBar_DevVersion(t *testing.T) {
	m := Model{
		version: "dev",
		width:   80,
	}
	bar := m.renderStatusBar(80)
	if !strings.Contains(bar, "-dev") {
		t.Errorf("status bar in dev mode should contain '-dev', got: %q", bar)
	}
}

func TestRenderStatusBar_UpdateAvailable(t *testing.T) {
	m := Model{
		version:         "1.0.0",
		latestVersion:   "2.0.0",
		updateAvailable: true,
		width:           100,
	}
	bar := m.renderStatusBar(100)
	if !strings.Contains(bar, "update available") {
		t.Errorf("status bar should show update notice, got: %q", bar)
	}
	if !strings.Contains(bar, "v2.0.0") {
		t.Errorf("status bar should show latest version, got: %q", bar)
	}
}

func TestRenderStatusBar_NoUpdate(t *testing.T) {
	m := Model{
		version:         "1.0.0",
		updateAvailable: false,
		width:           80,
	}
	bar := m.renderStatusBar(80)
	if strings.Contains(bar, "update") {
		t.Errorf("status bar should not show update notice when not available, got: %q", bar)
	}
}

func TestComposeFullScreen_PadsContent(t *testing.T) {
	m := Model{
		version: "1.0.0",
		width:   80,
		height:  24,
	}
	content := "line1\nline2\n"
	helpLine := "help text"

	result := m.composeFullScreen(content, helpLine)

	// Should contain the content
	if !strings.Contains(result, "line1") {
		t.Error("full screen should contain content")
	}
	// Should contain help
	if !strings.Contains(result, "help text") {
		t.Error("full screen should contain help line")
	}
	// Should contain status bar
	if !strings.Contains(result, "devops-starter") {
		t.Error("full screen should contain status bar")
	}
	// Total lines should approximately match height
	lines := strings.Count(result, "\n") + 1
	if lines < m.height-1 || lines > m.height+1 {
		t.Errorf("full screen should fill terminal height (%d), got %d lines", m.height, lines)
	}
}

func TestUpdateCheckMsg_SetsFields(t *testing.T) {
	m := newTestModel()
	m.width = 80
	m.height = 24

	msg := updateCheckMsg{
		LatestVersion:   "2.0.0",
		UpdateAvailable: true,
	}
	result, _ := m.Update(msg)
	updated := result.(Model)

	if updated.latestVersion != "2.0.0" {
		t.Errorf("expected latestVersion=2.0.0, got %s", updated.latestVersion)
	}
	if !updated.updateAvailable {
		t.Error("expected updateAvailable=true")
	}
}

func TestViewGroupsContent_ContainsStatusBar(t *testing.T) {
	m := newTestModel()
	m.width = 100
	m.height = 30

	view := m.View()

	// Should have separator lines (full width)
	if !strings.Contains(view, strings.Repeat("━", 100)) {
		t.Error("view should have full-width separators matching terminal width")
	}
	// Should have the version in status bar
	if !strings.Contains(view, "devops-starter") {
		t.Error("view should contain status bar with app name")
	}
}
