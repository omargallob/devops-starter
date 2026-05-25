package tui

import (
	"context"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/omargallob/devops-starter/internal/installer"
	"github.com/omargallob/devops-starter/internal/state"
	"github.com/omargallob/devops-starter/internal/updater"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// installCompleteMsg is sent when a tool installation finishes.
type installCompleteMsg struct {
	Name string
	Err  error
}

// verifyCompleteMsg is sent when a version verification finishes.
type verifyCompleteMsg struct {
	Name    string
	Version string
	Err     error
}

// removeCompleteMsg is sent when a tool removal finishes.
type removeCompleteMsg struct {
	Name       string
	SystemPath string // path to system binary that will take over (if any)
	Err        error
}

// installToolCmd returns a Bubble Tea command that installs a single tool
// asynchronously and sends an installCompleteMsg when done.
func installToolCmd(inst *installer.Installer, tool *tooldef.Tool) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := inst.Install(ctx, tool)
		return installCompleteMsg{
			Name: tool.Name,
			Err:  err,
		}
	}
}

// verifyToolCmd returns a Bubble Tea command that runs version detection
// for a single tool and sends a verifyCompleteMsg with the result.
func verifyToolCmd(toolName, installDir string) tea.Cmd {
	return func() tea.Msg {
		version, err := state.DetectVersion(toolName, installDir)
		return verifyCompleteMsg{
			Name:    toolName,
			Version: version,
			Err:     err,
		}
	}
}

// removeToolCmd returns a Bubble Tea command that removes a managed tool binary
// and clears it from the state store. It reports whether a system version exists.
func removeToolCmd(toolName, binName, installDir string, store *state.Store) tea.Cmd {
	return func() tea.Msg {
		binPath := filepath.Join(installDir, binName)

		// Remove the binary
		err := os.Remove(binPath)
		if err != nil && !os.IsNotExist(err) {
			return removeCompleteMsg{Name: toolName, Err: err}
		}

		// Remove from state store
		if err := store.Remove(toolName); err != nil {
			return removeCompleteMsg{Name: toolName, Err: err}
		}

		// Check if a system version will take over
		systemPath := state.LookupInPath(toolName)

		return removeCompleteMsg{
			Name:       toolName,
			SystemPath: systemPath,
			Err:        nil,
		}
	}
}

// updateCheckMsg is sent when the async update check completes.
type updateCheckMsg struct {
	LatestVersion   string
	UpdateAvailable bool
}

// checkForUpdateCmd returns a Bubble Tea command that checks for a newer release.
func checkForUpdateCmd(currentVersion string) tea.Cmd {
	return func() tea.Msg {
		result := updater.Check(currentVersion)
		return updateCheckMsg{
			LatestVersion:   result.LatestVersion,
			UpdateAvailable: result.UpdateAvailable,
		}
	}
}
