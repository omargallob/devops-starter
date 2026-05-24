package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/omargallob/devops-starter/internal/installer"
	"github.com/omargallob/devops-starter/internal/state"
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
