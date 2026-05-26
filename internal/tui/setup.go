// Package tui — setup.go implements the interactive setup wizard TUI.
// It guides users through: welcome/PATH check, group selection, install directory,
// dotfiles linking, and confirmation before executing the full install.
package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/omargallob/devops-starter/internal/config"
)

// setupScreen represents which setup wizard step is active.
type setupScreen int

const (
	setupScreenWelcome    setupScreen = iota // Welcome + PATH check
	setupScreenGroups                        // Group selection (checkboxes)
	setupScreenInstallDir                    // Install directory input
	setupScreenDotfiles                      // Dotfiles category toggles
	setupScreenConfirm                       // Summary & confirm
	setupScreenProgress                      // Install progress
	setupScreenDone                          // Completion
)

// dotfileCategory represents a toggleable dotfile group in the setup wizard.
type dotfileCategory struct {
	Name    string
	Enabled bool
}

// defaultDotfileCategories returns the dotfile categories available for linking.
func defaultDotfileCategories() []dotfileCategory {
	return []dotfileCategory{
		{Name: "shell (zsh/bash)", Enabled: true},
		{Name: "git", Enabled: true},
		{Name: "tmux", Enabled: true},
		{Name: "starship", Enabled: true},
		{Name: "neovim", Enabled: true},
	}
}

// SetupModel is the Bubble Tea model for the guided setup wizard.
type SetupModel struct {
	screen setupScreen

	// PATH check
	pathOK      bool
	pathMessage string

	// Group selection
	groups      []setupGroupItem
	groupCursor int

	// Install directory
	installDirInput textinput.Model
	installDir      string

	// Dotfiles
	dotfiles      []dotfileCategory
	dotfileCursor int
	linkDotfiles  bool

	// Confirmation
	confirmed bool

	// Progress
	progressItems []progressItem
	progressDone  bool
	spinners      map[string]spinner.Model

	// Terminal
	width    int
	height   int
	quitting bool
	err      error

	// Config
	cfg     *config.Config
	cfgPath string
	dryRun  bool
}

// setupGroupItem represents a group toggle in the setup wizard.
type setupGroupItem struct {
	Name    string
	Enabled bool
}

// NewSetupModel creates a new setup wizard model. If cfg is nil, DefaultConfig is used.
// cfgPath is where the config will be saved.
func NewSetupModel(cfg *config.Config, cfgPath string, dryRun bool) SetupModel {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	// Build group items from config state
	groupNames := config.AllGroupNames()
	groups := make([]setupGroupItem, len(groupNames))
	for i, name := range groupNames {
		groups[i] = setupGroupItem{
			Name:    name,
			Enabled: cfg.IsGroupEnabled(name),
		}
	}

	// Text input for install dir
	ti := textinput.New()
	ti.SetValue(cfg.InstallDir)
	ti.CharLimit = 256
	ti.Width = 60

	// Check PATH
	pathOK, pathMsg := checkInstallDirInPath(cfg.InstallDir)

	return SetupModel{
		screen:          setupScreenWelcome,
		pathOK:          pathOK,
		pathMessage:     pathMsg,
		groups:          groups,
		installDirInput: ti,
		installDir:      cfg.InstallDir,
		dotfiles:        defaultDotfileCategories(),
		linkDotfiles:    true,
		spinners:        make(map[string]spinner.Model),
		cfg:             cfg,
		cfgPath:         cfgPath,
		dryRun:          dryRun,
	}
}

// Init implements tea.Model.
func (m SetupModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Global quit
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}

		switch m.screen {
		case setupScreenWelcome:
			return m.updateWelcome(msg)
		case setupScreenGroups:
			return m.updateGroups(msg)
		case setupScreenInstallDir:
			return m.updateInstallDir(msg)
		case setupScreenDotfiles:
			return m.updateDotfiles(msg)
		case setupScreenConfirm:
			return m.updateConfirm(msg)
		case setupScreenProgress:
			return m.updateProgress(msg)
		case setupScreenDone:
			return m.updateDone(msg)
		}

	case setupInstallCompleteMsg:
		return m.handleSetupInstallComplete(msg)

	case spinner.TickMsg:
		return m.handleSetupSpinnerTick(msg)
	}

	// Forward to text input if on install dir screen
	if m.screen == setupScreenInstallDir {
		var cmd tea.Cmd
		m.installDirInput, cmd = m.installDirInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View implements tea.Model.
func (m SetupModel) View() string {
	if m.quitting {
		return ""
	}

	switch m.screen {
	case setupScreenWelcome:
		return m.viewWelcome()
	case setupScreenGroups:
		return m.viewGroups()
	case setupScreenInstallDir:
		return m.viewInstallDir()
	case setupScreenDotfiles:
		return m.viewDotfiles()
	case setupScreenConfirm:
		return m.viewConfirm()
	case setupScreenProgress:
		return m.viewProgress()
	case setupScreenDone:
		return m.viewDone()
	}
	return ""
}

// ─── Screen Updates ──────────────────────────────────────────────────────────

func (m SetupModel) updateWelcome(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.screen = setupScreenGroups
	case "q", "esc":
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m SetupModel) updateGroups(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.groupCursor < len(m.groups)-1 {
			m.groupCursor++
		}
	case "k", "up":
		if m.groupCursor > 0 {
			m.groupCursor--
		}
	case " ":
		m.groups[m.groupCursor].Enabled = !m.groups[m.groupCursor].Enabled
	case "a":
		for i := range m.groups {
			m.groups[i].Enabled = true
		}
	case "n":
		for i := range m.groups {
			m.groups[i].Enabled = false
		}
	case "enter":
		m.screen = setupScreenInstallDir
		m.installDirInput.Focus()
		return m, textinput.Blink
	case "esc":
		m.screen = setupScreenWelcome
	}
	return m, nil
}

func (m SetupModel) updateInstallDir(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.installDir = m.installDirInput.Value()
		// Expand ~ prefix
		if strings.HasPrefix(m.installDir, "~/") {
			home, _ := os.UserHomeDir()
			m.installDir = filepath.Join(home, m.installDir[2:])
		}
		m.installDirInput.Blur()
		m.pathOK, m.pathMessage = checkInstallDirInPath(m.installDir)
		m.screen = setupScreenDotfiles
	case "esc":
		m.installDirInput.Blur()
		m.screen = setupScreenGroups
	default:
		var cmd tea.Cmd
		m.installDirInput, cmd = m.installDirInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m SetupModel) updateDotfiles(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.dotfileCursor < len(m.dotfiles)-1 {
			m.dotfileCursor++
		}
	case "k", "up":
		if m.dotfileCursor > 0 {
			m.dotfileCursor--
		}
	case " ":
		m.dotfiles[m.dotfileCursor].Enabled = !m.dotfiles[m.dotfileCursor].Enabled
	case "a":
		for i := range m.dotfiles {
			m.dotfiles[i].Enabled = true
		}
	case "n":
		for i := range m.dotfiles {
			m.dotfiles[i].Enabled = false
		}
	case "s":
		// Skip dotfiles entirely
		m.linkDotfiles = false
		m.screen = setupScreenConfirm
	case "enter":
		m.linkDotfiles = true
		m.screen = setupScreenConfirm
	case "esc":
		m.screen = setupScreenInstallDir
		m.installDirInput.Focus()
		return m, textinput.Blink
	}
	return m, nil
}

func (m SetupModel) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		m.confirmed = true
		// Apply selections to config
		groupMap := make(map[string]bool)
		for _, g := range m.groups {
			groupMap[g.Name] = g.Enabled
		}
		m.cfg.MergeGroups(groupMap)
		m.cfg.InstallDir = m.installDir

		if m.dryRun {
			m.screen = setupScreenDone
			return m, nil
		}

		// Save config
		if err := config.Save(m.cfg, m.cfgPath); err != nil {
			m.err = err
			m.screen = setupScreenDone
			return m, nil
		}

		m.screen = setupScreenDone
		return m, nil

	case "n", "N", "esc":
		m.screen = setupScreenDotfiles
	}
	return m, nil
}

func (m SetupModel) updateProgress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.progressDone {
		m.screen = setupScreenDone
		return m, nil
	}
	if msg.String() == "q" {
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m SetupModel) updateDone(_ tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.quitting = true
	return m, tea.Quit
}

// ─── Views ───────────────────────────────────────────────────────────────────

func (m SetupModel) viewWelcome() string {
	var b strings.Builder
	w := clamp(m.width, 40, 80)

	b.WriteString(titleStyle.Render("devops-starter setup"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("━", w))
	b.WriteString("\n\n")

	b.WriteString("  Welcome! This wizard will guide you through:\n\n")
	b.WriteString("    1. Selecting tool groups to install\n")
	b.WriteString("    2. Choosing your install directory\n")
	b.WriteString("    3. Linking dotfiles (shell, git, editor configs)\n")
	b.WriteString("    4. Running the installation\n\n")

	// PATH status
	if m.pathOK {
		b.WriteString(currentStyle.Render(fmt.Sprintf("  ✓ PATH check: %s", m.pathMessage)))
	} else {
		b.WriteString(outdatedStyle.Render(fmt.Sprintf("  ! PATH check: %s", m.pathMessage)))
	}
	b.WriteString("\n\n")

	b.WriteString(strings.Repeat("━", w))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(" enter continue  q/esc quit"))
	b.WriteString("\n")

	return b.String()
}

func (m SetupModel) viewGroups() string {
	var b strings.Builder
	w := clamp(m.width, 40, 80)

	b.WriteString(titleStyle.Render("devops-starter setup"))
	b.WriteString(dimStyle.Render("  [1/4 groups]"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("━", w))
	b.WriteString("\n\n")

	b.WriteString("  Select tool groups to install:\n\n")

	for i, g := range m.groups {
		cursor := "  "
		if i == m.groupCursor {
			cursor = "▸ "
		}

		check := "[ ]"
		if g.Enabled {
			check = "[✓]"
		}

		line := fmt.Sprintf("  %s%s %s", cursor, check, g.Name)
		switch {
		case i == m.groupCursor:
			b.WriteString(cursorStyle.Render(line))
		case g.Enabled:
			b.WriteString(currentStyle.Render(line))
		default:
			b.WriteString(missingStyle.Render(line))
		}
		b.WriteString("\n")

		// Show hint for the languages group
		if g.Name == "languages" {
			hint := "         ↳ Optional: uses mise to manage Go, Python, Node versions."
			hint2 := "           Deselect if you manage runtimes yourself."
			b.WriteString(dimStyle.Render(hint))
			b.WriteString("\n")
			b.WriteString(dimStyle.Render(hint2))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(strings.Repeat("━", w))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(" ↑↓/jk navigate  space toggle  a all  n none  enter next  esc back"))
	b.WriteString("\n")

	return b.String()
}

func (m SetupModel) viewInstallDir() string {
	var b strings.Builder
	w := clamp(m.width, 40, 80)

	b.WriteString(titleStyle.Render("devops-starter setup"))
	b.WriteString(dimStyle.Render("  [2/4 install dir]"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("━", w))
	b.WriteString("\n\n")

	b.WriteString("  Where should binaries be installed?\n\n")
	b.WriteString("  ")
	b.WriteString(m.installDirInput.View())
	b.WriteString("\n\n")

	if !m.pathOK {
		b.WriteString(outdatedStyle.Render(fmt.Sprintf("  ! %s", m.pathMessage)))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(strings.Repeat("━", w))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(" enter confirm  esc back"))
	b.WriteString("\n")

	return b.String()
}

func (m SetupModel) viewDotfiles() string {
	var b strings.Builder
	w := clamp(m.width, 40, 80)

	b.WriteString(titleStyle.Render("devops-starter setup"))
	b.WriteString(dimStyle.Render("  [3/4 dotfiles]"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("━", w))
	b.WriteString("\n\n")

	b.WriteString("  Select dotfile categories to link:\n\n")

	for i, d := range m.dotfiles {
		cursor := "  "
		if i == m.dotfileCursor {
			cursor = "▸ "
		}

		check := "[ ]"
		if d.Enabled {
			check = "[✓]"
		}

		line := fmt.Sprintf("  %s%s %s", cursor, check, d.Name)
		switch {
		case i == m.dotfileCursor:
			b.WriteString(cursorStyle.Render(line))
		case d.Enabled:
			b.WriteString(currentStyle.Render(line))
		default:
			b.WriteString(missingStyle.Render(line))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(strings.Repeat("━", w))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(" ↑↓/jk navigate  space toggle  a all  n none  enter next  s skip  esc back"))
	b.WriteString("\n")

	return b.String()
}

func (m SetupModel) viewConfirm() string {
	var b strings.Builder
	w := clamp(m.width, 40, 80)

	b.WriteString(titleStyle.Render("devops-starter setup"))
	b.WriteString(dimStyle.Render("  [4/4 confirm]"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("━", w))
	b.WriteString("\n\n")

	b.WriteString(headerStyle.Render("  Summary:"))
	b.WriteString("\n\n")

	// Groups
	var enabled []string
	for _, g := range m.groups {
		if g.Enabled {
			enabled = append(enabled, g.Name)
		}
	}
	fmt.Fprintf(&b, "  Install dir:  %s\n", m.installDir)
	fmt.Fprintf(&b, "  Groups:       %s\n", strings.Join(enabled, ", "))

	// Dotfiles
	if m.linkDotfiles {
		var dotEnabled []string
		for _, d := range m.dotfiles {
			if d.Enabled {
				dotEnabled = append(dotEnabled, d.Name)
			}
		}
		fmt.Fprintf(&b, "  Dotfiles:     %s\n", strings.Join(dotEnabled, ", "))
	} else {
		b.WriteString("  Dotfiles:     skipped\n")
	}

	if m.dryRun {
		b.WriteString("\n")
		b.WriteString(outdatedStyle.Render("  [dry-run] No changes will be made"))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	fmt.Fprintf(&b, "  Config will be saved to: %s\n", m.cfgPath)
	b.WriteString("\n")

	b.WriteString(strings.Repeat("━", w))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(" y/enter confirm  n/esc go back"))
	b.WriteString("\n")

	return b.String()
}

func (m SetupModel) viewProgress() string {
	var b strings.Builder
	w := clamp(m.width, 40, 80)

	b.WriteString(titleStyle.Render("devops-starter setup"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("━", w))
	b.WriteString("\n\n")

	if m.progressDone {
		b.WriteString(headerStyle.Render("  Setup complete!"))
	} else {
		b.WriteString(headerStyle.Render("  Installing..."))
	}
	b.WriteString("\n\n")

	for _, item := range m.progressItems {
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

		line := fmt.Sprintf("    %s %s", icon, item.Name)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if m.progressDone {
		b.WriteString(helpStyle.Render("  Press any key to finish..."))
		b.WriteString("\n")
	}

	b.WriteString(strings.Repeat("━", w))
	b.WriteString("\n")

	return b.String()
}

func (m SetupModel) viewDone() string {
	var b strings.Builder
	w := clamp(m.width, 40, 80)

	b.WriteString(titleStyle.Render("devops-starter setup"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("━", w))
	b.WriteString("\n\n")

	switch {
	case m.err != nil:
		b.WriteString(errStyle.Render(fmt.Sprintf("  Error: %v", m.err)))
		b.WriteString("\n\n")
	case m.dryRun:
		b.WriteString(messageStyle.Render("  [dry-run] Setup complete (no changes made)"))
		b.WriteString("\n\n")
		b.WriteString("  To apply, run without --dry-run:\n")
		b.WriteString("    devops-starter setup\n\n")
	default:
		b.WriteString(messageStyle.Render("  ✓ Configuration saved!"))
		b.WriteString("\n\n")
		b.WriteString("  Next steps:\n")
		b.WriteString("    devops-starter install    # Install selected tools\n")
		if m.linkDotfiles {
			b.WriteString("    devops-starter dotfiles link  # Link dotfile configs\n")
		}
		if !m.pathOK {
			b.WriteString("\n")
			b.WriteString(outdatedStyle.Render(fmt.Sprintf("  ! Remember to add %s to your PATH", m.installDir)))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(strings.Repeat("━", w))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(" Press any key to exit"))
	b.WriteString("\n")

	return b.String()
}

// ─── Messages ────────────────────────────────────────────────────────────────

// setupInstallCompleteMsg is sent when an install step in setup finishes.
type setupInstallCompleteMsg struct {
	Name string
	Err  error
}

func (m SetupModel) handleSetupInstallComplete(msg setupInstallCompleteMsg) (tea.Model, tea.Cmd) {
	delete(m.spinners, msg.Name)

	for i := range m.progressItems {
		if m.progressItems[i].Name == msg.Name {
			if msg.Err != nil {
				m.progressItems[i].Status = progressFailed
				m.progressItems[i].Error = msg.Err
			} else {
				m.progressItems[i].Status = progressDone
			}
			break
		}
	}

	// Check if all done
	allDone := true
	for _, item := range m.progressItems {
		if item.Status == progressInstalling || item.Status == progressWaiting {
			allDone = false
			break
		}
	}
	m.progressDone = allDone

	return m, nil
}

func (m SetupModel) handleSetupSpinnerTick(msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	for name, s := range m.spinners {
		updated, cmd := s.Update(msg)
		m.spinners[name] = updated
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// checkInstallDirInPath checks if the given directory is present in $PATH.
func checkInstallDirInPath(dir string) (found bool, absPath string) {
	pathEnv := os.Getenv("PATH")
	paths := filepath.SplitList(pathEnv)

	// Normalize for comparison
	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}

	for _, p := range paths {
		absP, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		if absP == absDir {
			return true, fmt.Sprintf("%s is in your PATH", dir)
		}
	}

	return false, fmt.Sprintf("%s is NOT in your PATH", dir)
}

// SelectedGroups returns the group selections as a map.
func (m SetupModel) SelectedGroups() map[string]bool {
	result := make(map[string]bool, len(m.groups))
	for _, g := range m.groups {
		result[g.Name] = g.Enabled
	}
	return result
}

// SelectedDotfiles returns the enabled dotfile category names.
func (m SetupModel) SelectedDotfiles() []string {
	var result []string
	for _, d := range m.dotfiles {
		if d.Enabled {
			result = append(result, d.Name)
		}
	}
	return result
}

// IsConfirmed returns whether the user confirmed the setup.
func (m SetupModel) IsConfirmed() bool {
	return m.confirmed
}

// Config returns the config as modified by the setup wizard.
func (m SetupModel) Config() *config.Config {
	return m.cfg
}
