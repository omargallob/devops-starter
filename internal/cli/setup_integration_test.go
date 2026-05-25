//go:build integration

package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/omargallob/devops-starter/internal/config"
)

func TestSetup_EndToEnd_NonInteractive(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	root := NewRootCmd()
	root.SetArgs([]string{"setup", "--non-interactive", "--config", cfgPath})

	if err := root.Execute(); err != nil {
		t.Fatalf("setup --non-interactive should succeed: %v", err)
	}

	// Verify config was written correctly
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Should have all defaults
	if !cfg.Groups.Languages {
		t.Error("Languages should be enabled by default")
	}
	if !cfg.Groups.Kubernetes {
		t.Error("Kubernetes should be enabled by default")
	}
	if !cfg.Groups.Utilities {
		t.Error("Utilities should be enabled by default")
	}
	if cfg.InstallDir == "" {
		t.Error("InstallDir should be set")
	}
}

func TestSetup_ReRun_PreservesOverrides(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	// First run: create config with overrides
	cfg := config.DefaultConfig()
	cfg.Groups.Cloud = false
	cfg.Overrides["terraform"] = config.ToolOverride{Version: "1.8.0"}
	cfg.Overrides["kubectl"] = config.ToolOverride{Disabled: true}

	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatalf("failed to save initial config: %v", err)
	}

	// Second run: non-interactive setup (simulates re-run)
	root := NewRootCmd()
	root.SetArgs([]string{"setup", "--non-interactive", "--config", cfgPath})

	if err := root.Execute(); err != nil {
		t.Fatalf("setup re-run should succeed: %v", err)
	}

	// Verify overrides are preserved
	loaded, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("failed to load config after re-run: %v", err)
	}

	if loaded.Overrides["terraform"].Version != "1.8.0" {
		t.Errorf("terraform version override lost: got %q", loaded.Overrides["terraform"].Version)
	}
	if loaded.Overrides["kubectl"].Disabled != true {
		t.Error("kubectl disabled override lost")
	}
	// Cloud should still be disabled (non-interactive preserves existing)
	if loaded.Groups.Cloud != false {
		t.Error("Cloud group should remain disabled on re-run")
	}
}

func TestSetup_PathCheck(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	// Create config with install dir not in PATH
	cfg := config.DefaultConfig()
	cfg.InstallDir = "/unlikely/path/not/in/env"
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Non-interactive should still succeed (PATH warning is advisory)
	root := NewRootCmd()
	root.SetArgs([]string{"setup", "--non-interactive", "--config", cfgPath})

	if err := root.Execute(); err != nil {
		t.Fatalf("setup should succeed even with PATH warning: %v", err)
	}
}

func TestSetup_DryRun(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	root := NewRootCmd()
	root.SetArgs([]string{"setup", "--non-interactive", "--dry-run", "--config", cfgPath})

	if err := root.Execute(); err != nil {
		t.Fatalf("setup --dry-run should succeed: %v", err)
	}

	// Config file should NOT exist
	if _, err := os.Stat(cfgPath); !os.IsNotExist(err) {
		t.Error("config file should not be created in dry-run mode")
	}
}
