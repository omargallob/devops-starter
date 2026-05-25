package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/omargallob/devops-starter/internal/config"
)

func TestStatusCmd_Exists(t *testing.T) {
	root := NewRootCmd()
	var found bool
	for _, cmd := range root.Commands() {
		if cmd.Use == "status" {
			found = true
			if cmd.Short == "" {
				t.Error("status command should have a Short description")
			}
			break
		}
	}
	if !found {
		t.Fatal("status command not registered on root")
	}
}

func TestStatusCmd_NoTUI(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	cfg := config.DefaultConfig()
	cfg.InstallDir = filepath.Join(dir, "bin")
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Set XDG_CONFIG_HOME so state.StatePath() points to our temp dir
	xdgDir := filepath.Join(dir, "xdg")
	if err := os.MkdirAll(filepath.Join(xdgDir, "devops-starter"), 0o755); err != nil {
		t.Fatalf("failed to create xdg dir: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", xdgDir)

	// Capture stdout (command prints via tui.PrintTable to os.Stdout)
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	root := NewRootCmd()
	root.SetArgs([]string{"status", "--no-tui", "--config", cfgPath})
	err := root.Execute()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("status --no-tui should succeed: %v", err)
	}

	buf := make([]byte, 64*1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Should produce a table with tool names
	if output == "" {
		t.Error("status --no-tui should produce output")
	}
}

func TestStatusCmd_NoTUI_DisabledGroups(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	// Disable all groups
	cfg := config.DefaultConfig()
	cfg.InstallDir = filepath.Join(dir, "bin")
	cfg.Groups.Languages = false
	cfg.Groups.Containers = false
	cfg.Groups.Kubernetes = false
	cfg.Groups.Infra = false
	cfg.Groups.Cloud = false
	cfg.Groups.Ansible = false
	cfg.Groups.RustTools = false
	cfg.Groups.Utilities = false
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	xdgDir := filepath.Join(dir, "xdg")
	if err := os.MkdirAll(filepath.Join(xdgDir, "devops-starter"), 0o755); err != nil {
		t.Fatalf("failed to create xdg dir: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", xdgDir)

	// Capture stdout
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	root := NewRootCmd()
	root.SetArgs([]string{"status", "--no-tui", "--config", cfgPath})
	err := root.Execute()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("status --no-tui with all groups disabled should succeed: %v", err)
	}
}

func TestStatusCmd_Flags(t *testing.T) {
	root := NewRootCmd()
	var statusCmd *cobra.Command
	for _, cmd := range root.Commands() {
		if cmd.Use == "status" {
			statusCmd = cmd
			break
		}
	}
	if statusCmd == nil {
		t.Fatal("status command not found")
	}

	// Check --no-tui flag
	flag := statusCmd.Flags().Lookup("no-tui")
	if flag == nil {
		t.Error("expected --no-tui flag")
	}

	// Check --verify flag
	verifyFlag := statusCmd.Flags().Lookup("verify")
	if verifyFlag == nil {
		t.Error("expected --verify flag")
	}
}
