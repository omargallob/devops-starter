package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/omargallob/devops-starter/internal/config"
)

func TestListCmd_Exists(t *testing.T) {
	root := NewRootCmd()
	var found bool
	for _, cmd := range root.Commands() {
		if cmd.Use == "list" {
			found = true
			if cmd.Short == "" {
				t.Error("list command should have a Short description")
			}
			break
		}
	}
	if !found {
		t.Fatal("list command not registered on root")
	}
}

func TestListCmd_RunsWithConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	cfg := config.DefaultConfig()
	cfg.InstallDir = filepath.Join(dir, "bin")
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Set XDG_CONFIG_HOME so state.Path() points to our temp dir
	xdgDir := filepath.Join(dir, "xdg")
	if err := os.MkdirAll(filepath.Join(xdgDir, "devops-starter"), 0o755); err != nil {
		t.Fatalf("failed to create xdg dir: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", xdgDir)

	stdout, _, err := executeCommand(t, "list", "--config", cfgPath)
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	// Should produce some output with group headers
	if stdout == "" {
		// The command prints to os.Stdout via fmt.Println, not cmd.OutOrStdout().
		// This is expected behavior — stdout will be empty in our buffer but
		// the command still succeeds. The key assertion is no error.
		_ = stdout
	}
}

func TestListCmd_NoToolsForDisabledGroups(t *testing.T) {
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

	// Set XDG_CONFIG_HOME so state.Path() points to our temp dir
	xdgDir := filepath.Join(dir, "xdg")
	if err := os.MkdirAll(filepath.Join(xdgDir, "devops-starter"), 0o755); err != nil {
		t.Fatalf("failed to create xdg dir: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", xdgDir)

	_, _, err := executeCommand(t, "list", "--config", cfgPath)
	if err != nil {
		t.Fatalf("list command should succeed even with all groups disabled: %v", err)
	}
}

func TestListCmd_OutputContainsToolNames(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	// Enable only utilities group (smaller set)
	cfg := config.DefaultConfig()
	cfg.InstallDir = filepath.Join(dir, "bin")
	cfg.Groups.Languages = false
	cfg.Groups.Containers = false
	cfg.Groups.Kubernetes = false
	cfg.Groups.Infra = false
	cfg.Groups.Cloud = false
	cfg.Groups.Ansible = false
	cfg.Groups.RustTools = false
	cfg.Groups.Utilities = true
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Set XDG_CONFIG_HOME so state.Path() points to our temp dir
	xdgDir := filepath.Join(dir, "xdg")
	if err := os.MkdirAll(filepath.Join(xdgDir, "devops-starter"), 0o755); err != nil {
		t.Fatalf("failed to create xdg dir: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", xdgDir)

	// Capture real stdout since the command uses fmt.Println
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	root := NewRootCmd()
	root.SetArgs([]string{"list", "--config", cfgPath})
	err := root.Execute()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	buf := make([]byte, 64*1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// The utilities group should produce output with the group header
	if !strings.Contains(output, "utilities") {
		t.Errorf("expected output to contain 'utilities' group header, got:\n%s", output)
	}
}
