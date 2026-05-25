package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/omargallob/devops-starter/internal/config"
)

func TestConfigCmd_Exists(t *testing.T) {
	root := NewRootCmd()
	var found bool
	for _, cmd := range root.Commands() {
		if cmd.Use == "config" {
			found = true
			if cmd.Short == "" {
				t.Error("config command should have a Short description")
			}
			break
		}
	}
	if !found {
		t.Fatal("config command not registered on root")
	}
}

func TestConfigInitCmd_CreatesDefaultConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "sub", "config.yaml")

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	root := NewRootCmd()
	root.SetArgs([]string{"config", "init", "--config", cfgPath})
	err := root.Execute()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("config init should succeed: %v", err)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "Configuration created at") {
		t.Errorf("expected success message, got: %s", output)
	}

	// Verify file was created
	if _, err := os.Stat(cfgPath); err != nil {
		t.Fatalf("config file should exist after init: %v", err)
	}

	// Verify it's valid config
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("failed to load created config: %v", err)
	}
	if !cfg.Groups.Languages {
		t.Error("default config should have Languages enabled")
	}
}

func TestConfigInitCmd_RefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	// Create existing config
	cfg := config.DefaultConfig()
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatalf("failed to create existing config: %v", err)
	}

	root := NewRootCmd()
	root.SetArgs([]string{"config", "init", "--config", cfgPath})
	err := root.Execute()

	if err == nil {
		t.Fatal("config init should fail when config already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention 'already exists', got: %v", err)
	}
}

func TestConfigShowCmd_DisplaysConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	cfg := config.DefaultConfig()
	cfg.Overrides = map[string]config.ToolOverride{
		"terraform": {Version: "1.8.0"},
	}
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	root := NewRootCmd()
	root.SetArgs([]string{"config", "show", "--config", cfgPath})
	err := root.Execute()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("config show should succeed: %v", err)
	}

	buf := make([]byte, 64*1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Should contain YAML output
	if !strings.Contains(output, "install_dir:") {
		t.Errorf("expected YAML with install_dir, got:\n%s", output)
	}
	if !strings.Contains(output, "terraform") {
		t.Errorf("expected terraform override in output, got:\n%s", output)
	}
	if !strings.Contains(output, "1.8.0") {
		t.Errorf("expected version 1.8.0 in output, got:\n%s", output)
	}
}

func TestConfigShowCmd_DefaultsWhenNoFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "nonexistent.yaml")

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	root := NewRootCmd()
	root.SetArgs([]string{"config", "show", "--config", cfgPath})
	err := root.Execute()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("config show should succeed with defaults when no file: %v", err)
	}

	buf := make([]byte, 64*1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Should show default groups
	if !strings.Contains(output, "languages: true") {
		t.Errorf("expected default config with languages: true, got:\n%s", output)
	}
}
