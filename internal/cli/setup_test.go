package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/omargallob/devops-starter/internal/config"
)

func TestSetupCmd_Exists(t *testing.T) {
	root := NewRootCmd()

	// Find setup in subcommands
	var found bool
	for _, cmd := range root.Commands() {
		if cmd.Use != "setup" {
			continue
		}
		found = true
		if cmd.Short == "" {
			t.Error("setup command should have a Short description")
		}
		if cmd.Long == "" {
			t.Error("setup command should have a Long description")
		}
		if cmd.Example == "" {
			t.Error("setup command should have an Example")
		}
		break
	}
	if !found {
		t.Fatal("setup command not registered on root")
	}
}

func TestSetupCmd_Flags(t *testing.T) {
	root := NewRootCmd()

	var setupCmd *cobra.Command
	for _, cmd := range root.Commands() {
		if cmd.Use == "setup" {
			setupCmd = cmd
			break
		}
	}
	if setupCmd == nil {
		t.Fatal("setup command not found")
	}

	// Check --non-interactive flag
	flag := setupCmd.Flags().Lookup("non-interactive")
	if flag == nil {
		t.Error("expected --non-interactive flag")
	}

	// Check that --dry-run is inherited from root persistent flags
	dryRunFlag := setupCmd.InheritedFlags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Error("expected inherited --dry-run flag")
	}
}

func TestSetupCmd_LoadsExistingConfig(t *testing.T) {
	// Create a temp config file with custom settings
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	cfg := config.DefaultConfig()
	cfg.Groups.Cloud = false
	cfg.Groups.Ansible = false
	cfg.Overrides["terraform"] = config.ToolOverride{Version: "1.8.0"}

	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	// Load and verify it picks up existing settings
	loaded, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded.Groups.Cloud != false {
		t.Error("expected Cloud group to be disabled from existing config")
	}
	if loaded.Groups.Ansible != false {
		t.Error("expected Ansible group to be disabled from existing config")
	}
	if loaded.Overrides["terraform"].Version != "1.8.0" {
		t.Errorf("expected terraform override 1.8.0, got %q", loaded.Overrides["terraform"].Version)
	}
}

func TestSetupCmd_CreatesDefaultConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "subdir", "config.yaml")

	// Loading a non-existent path should return defaults
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.Groups.Languages {
		t.Error("default config should have Languages enabled")
	}
	if !cfg.Groups.Kubernetes {
		t.Error("default config should have Kubernetes enabled")
	}
	if cfg.InstallDir == "" {
		t.Error("default config should have InstallDir set")
	}

	// Save should create parent directories
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatalf("Save() should create dirs: %v", err)
	}
	if _, err := os.Stat(cfgPath); err != nil {
		t.Fatalf("config file should exist after save: %v", err)
	}
}

func TestSetupCmd_NonInteractive(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	root := NewRootCmd()
	root.SetArgs([]string{"setup", "--non-interactive", "--config", cfgPath})

	if err := root.Execute(); err != nil {
		t.Fatalf("setup --non-interactive should succeed: %v", err)
	}

	// Config should have been saved
	if _, err := os.Stat(cfgPath); err != nil {
		t.Error("config file should exist after non-interactive setup")
	}

	// Verify it contains defaults
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}
	if !cfg.Groups.Languages {
		t.Error("saved config should have default groups enabled")
	}
}

func TestSetupCmd_NonInteractiveDryRun(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	root := NewRootCmd()
	root.SetArgs([]string{"setup", "--non-interactive", "--dry-run", "--config", cfgPath})

	if err := root.Execute(); err != nil {
		t.Fatalf("setup --non-interactive --dry-run should succeed: %v", err)
	}

	// Config should NOT have been written
	if _, err := os.Stat(cfgPath); !os.IsNotExist(err) {
		t.Error("config file should not exist in dry-run mode")
	}
}
