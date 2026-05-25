package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/omargallob/devops-starter/internal/config"
)

func newTestSetupModel() SetupModel {
	cfg := config.DefaultConfig()
	return NewSetupModel(cfg, "/tmp/test-config.yaml", false)
}

func TestSetupModel_Init(t *testing.T) {
	m := newTestSetupModel()

	if m.screen != setupScreenWelcome {
		t.Errorf("expected initial screen to be welcome, got %d", m.screen)
	}

	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestSetupModel_WelcomeView(t *testing.T) {
	m := newTestSetupModel()
	m.width = 80

	view := m.View()
	if !strings.Contains(view, "devops-starter setup") {
		t.Error("welcome view should contain title")
	}
	if !strings.Contains(view, "Welcome") {
		t.Error("welcome view should contain welcome message")
	}
	if !strings.Contains(view, "PATH check") {
		t.Error("welcome view should show PATH check status")
	}
}

func TestSetupModel_WelcomeToGroups(t *testing.T) {
	m := newTestSetupModel()

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(SetupModel)

	if m.screen != setupScreenGroups {
		t.Errorf("expected groups screen after enter, got %d", m.screen)
	}
}

func TestSetupModel_WelcomeQuit(t *testing.T) {
	m := newTestSetupModel()

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = result.(SetupModel)

	if !m.quitting {
		t.Error("should be quitting after 'q'")
	}
	if cmd == nil {
		t.Error("quit should return a command")
	}
}

func TestSetupModel_GroupSelection_Toggle(t *testing.T) {
	m := newTestSetupModel()
	m.screen = setupScreenGroups

	// All groups start enabled (from DefaultConfig). Toggle first one off.
	if !m.groups[0].Enabled {
		t.Fatal("first group should start enabled")
	}

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = result.(SetupModel)

	if m.groups[0].Enabled {
		t.Error("first group should be disabled after space")
	}

	// Toggle back on
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = result.(SetupModel)

	if !m.groups[0].Enabled {
		t.Error("first group should be re-enabled after second space")
	}
}

func TestSetupModel_GroupSelection_AllOn(t *testing.T) {
	m := newTestSetupModel()
	m.screen = setupScreenGroups

	// Disable a few
	m.groups[0].Enabled = false
	m.groups[2].Enabled = false

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = result.(SetupModel)

	for i, g := range m.groups {
		if !g.Enabled {
			t.Errorf("group %d (%s) should be enabled after 'a'", i, g.Name)
		}
	}
}

func TestSetupModel_GroupSelection_NoneOff(t *testing.T) {
	m := newTestSetupModel()
	m.screen = setupScreenGroups

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = result.(SetupModel)

	for i, g := range m.groups {
		if g.Enabled {
			t.Errorf("group %d (%s) should be disabled after 'n'", i, g.Name)
		}
	}
}

func TestSetupModel_Navigation(t *testing.T) {
	m := newTestSetupModel()
	m.screen = setupScreenGroups

	// Move cursor down
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = result.(SetupModel)
	if m.groupCursor != 1 {
		t.Errorf("expected cursor=1, got %d", m.groupCursor)
	}

	// Move cursor up
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = result.(SetupModel)
	if m.groupCursor != 0 {
		t.Errorf("expected cursor=0, got %d", m.groupCursor)
	}

	// Don't go below 0
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = result.(SetupModel)
	if m.groupCursor != 0 {
		t.Errorf("expected cursor=0 (clamped), got %d", m.groupCursor)
	}
}

func TestSetupModel_GroupsToInstallDir(t *testing.T) {
	m := newTestSetupModel()
	m.screen = setupScreenGroups

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(SetupModel)

	if m.screen != setupScreenInstallDir {
		t.Errorf("expected install dir screen, got %d", m.screen)
	}
}

func TestSetupModel_InstallDirView(t *testing.T) {
	m := newTestSetupModel()
	m.screen = setupScreenInstallDir
	m.width = 80

	view := m.View()
	if !strings.Contains(view, "install") {
		t.Error("install dir view should mention install")
	}
	if !strings.Contains(view, "binaries") {
		t.Error("install dir view should mention binaries")
	}
}

func TestSetupModel_InstallDirToDotfiles(t *testing.T) {
	m := newTestSetupModel()
	m.screen = setupScreenInstallDir

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(SetupModel)

	if m.screen != setupScreenDotfiles {
		t.Errorf("expected dotfiles screen, got %d", m.screen)
	}
}

func TestSetupModel_InstallDirBack(t *testing.T) {
	m := newTestSetupModel()
	m.screen = setupScreenInstallDir

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = result.(SetupModel)

	if m.screen != setupScreenGroups {
		t.Errorf("expected groups screen on esc, got %d", m.screen)
	}
}

func TestSetupModel_DotfilesToggle(t *testing.T) {
	m := newTestSetupModel()
	m.screen = setupScreenDotfiles

	// First dotfile category starts enabled
	if !m.dotfiles[0].Enabled {
		t.Fatal("first dotfile should start enabled")
	}

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = result.(SetupModel)

	if m.dotfiles[0].Enabled {
		t.Error("first dotfile should be disabled after space")
	}
}

func TestSetupModel_DotfilesSkip(t *testing.T) {
	m := newTestSetupModel()
	m.screen = setupScreenDotfiles

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = result.(SetupModel)

	if m.screen != setupScreenConfirm {
		t.Errorf("expected confirm screen after skip, got %d", m.screen)
	}
	if m.linkDotfiles {
		t.Error("linkDotfiles should be false after skip")
	}
}

func TestSetupModel_DotfilesToConfirm(t *testing.T) {
	m := newTestSetupModel()
	m.screen = setupScreenDotfiles

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(SetupModel)

	if m.screen != setupScreenConfirm {
		t.Errorf("expected confirm screen, got %d", m.screen)
	}
	if !m.linkDotfiles {
		t.Error("linkDotfiles should be true after enter")
	}
}

func TestSetupModel_ConfirmView(t *testing.T) {
	m := newTestSetupModel()
	m.screen = setupScreenConfirm
	m.width = 80

	view := m.View()
	if !strings.Contains(view, "Summary") {
		t.Error("confirm view should show summary")
	}
	if !strings.Contains(view, "Install dir") {
		t.Error("confirm view should show install dir")
	}
	if !strings.Contains(view, "Groups") {
		t.Error("confirm view should show groups")
	}
}

func TestSetupModel_ConfirmAccept_DryRun(t *testing.T) {
	m := newTestSetupModel()
	m.screen = setupScreenConfirm
	m.dryRun = true

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = result.(SetupModel)

	if m.screen != setupScreenDone {
		t.Errorf("expected done screen after confirm in dry-run, got %d", m.screen)
	}
	if !m.confirmed {
		t.Error("should be confirmed")
	}
}

func TestSetupModel_ConfirmReject(t *testing.T) {
	m := newTestSetupModel()
	m.screen = setupScreenConfirm

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = result.(SetupModel)

	if m.screen != setupScreenDotfiles {
		t.Errorf("expected dotfiles screen after reject, got %d", m.screen)
	}
}

func TestSetupModel_DoneQuit(t *testing.T) {
	m := newTestSetupModel()
	m.screen = setupScreenDone

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(SetupModel)

	if !m.quitting {
		t.Error("should be quitting from done screen")
	}
	if cmd == nil {
		t.Error("quit should return a command")
	}
}

func TestSetupModel_CtrlCQuits(t *testing.T) {
	m := newTestSetupModel()
	m.screen = setupScreenGroups

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = result.(SetupModel)

	if !m.quitting {
		t.Error("ctrl+c should quit from any screen")
	}
	if cmd == nil {
		t.Error("quit should return a command")
	}
}

func TestSetupModel_SelectedGroups(t *testing.T) {
	m := newTestSetupModel()
	m.groups[0].Enabled = false // languages
	m.groups[3].Enabled = false // infra

	selected := m.SelectedGroups()
	if selected["languages"] != false {
		t.Error("languages should be disabled")
	}
	if selected["infra"] != false {
		t.Error("infra should be disabled")
	}
	if selected["kubernetes"] != true {
		t.Error("kubernetes should be enabled")
	}
}

func TestSetupModel_SelectedDotfiles(t *testing.T) {
	m := newTestSetupModel()
	m.dotfiles[0].Enabled = false // shell
	m.dotfiles[2].Enabled = false // tmux

	result := m.SelectedDotfiles()
	for _, name := range result {
		if strings.Contains(name, "shell") {
			t.Error("shell should not be in selected dotfiles")
		}
		if strings.Contains(name, "tmux") {
			t.Error("tmux should not be in selected dotfiles")
		}
	}
	if len(result) != 3 {
		t.Errorf("expected 3 selected dotfiles, got %d", len(result))
	}
}

func TestSetupModel_WindowSize(t *testing.T) {
	m := newTestSetupModel()

	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = result.(SetupModel)

	if m.width != 120 {
		t.Errorf("width = %d, want 120", m.width)
	}
	if m.height != 40 {
		t.Errorf("height = %d, want 40", m.height)
	}
}

func TestSetupModel_DoneView_DryRun(t *testing.T) {
	m := newTestSetupModel()
	m.screen = setupScreenDone
	m.dryRun = true
	m.confirmed = true
	m.width = 80

	view := m.View()
	if !strings.Contains(view, "dry-run") {
		t.Error("done view in dry-run should mention dry-run")
	}
}

func TestSetupModel_DoneView_Success(t *testing.T) {
	m := newTestSetupModel()
	m.screen = setupScreenDone
	m.confirmed = true
	m.width = 80

	view := m.View()
	if !strings.Contains(view, "saved") || !strings.Contains(view, "Configuration") {
		t.Error("done view should show success message")
	}
	if !strings.Contains(view, "devops-starter install") {
		t.Error("done view should show next steps")
	}
}

func TestSetupModel_GroupsView(t *testing.T) {
	m := newTestSetupModel()
	m.screen = setupScreenGroups
	m.width = 80

	view := m.View()
	if !strings.Contains(view, "languages") {
		t.Error("groups view should show languages")
	}
	if !strings.Contains(view, "kubernetes") {
		t.Error("groups view should show kubernetes")
	}
	if !strings.Contains(view, "[✓]") {
		t.Error("groups view should show checkmarks for enabled groups")
	}
}

func TestSetupModel_GroupsView_LanguagesHint(t *testing.T) {
	m := newTestSetupModel()
	m.screen = setupScreenGroups
	m.width = 80

	view := m.View()
	if !strings.Contains(view, "Optional") {
		t.Error("groups view should show 'Optional' hint for languages group")
	}
	if !strings.Contains(view, "mise") {
		t.Error("groups view should mention mise in languages hint")
	}
	if !strings.Contains(view, "Deselect if you manage runtimes yourself") {
		t.Error("groups view should tell users they can deselect languages")
	}
}

func TestSetupModel_DotfilesNavigation(t *testing.T) {
	m := newTestSetupModel()
	m.screen = setupScreenDotfiles

	// Move down
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = result.(SetupModel)
	if m.dotfileCursor != 1 {
		t.Errorf("expected dotfileCursor=1, got %d", m.dotfileCursor)
	}

	// Move up
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = result.(SetupModel)
	if m.dotfileCursor != 0 {
		t.Errorf("expected dotfileCursor=0, got %d", m.dotfileCursor)
	}
}

func TestCheckInstallDirInPath(t *testing.T) {
	// Test with a directory that's likely in PATH
	ok, msg := checkInstallDirInPath("/usr/bin")
	if !ok {
		t.Errorf("/usr/bin should be in PATH, got: %s", msg)
	}

	// Test with a directory not in PATH
	ok, msg = checkInstallDirInPath("/nonexistent/path/bin")
	if ok {
		t.Errorf("/nonexistent/path/bin should NOT be in PATH, got: %s", msg)
	}
	if !strings.Contains(msg, "NOT") {
		t.Errorf("message should contain NOT, got: %s", msg)
	}
}
