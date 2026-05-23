package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.Groups.Languages {
		t.Error("expected Languages group to be enabled by default")
	}
	if !cfg.Groups.Kubernetes {
		t.Error("expected Kubernetes group to be enabled by default")
	}
	if !cfg.Groups.RustTools {
		t.Error("expected RustTools group to be enabled by default")
	}
	if cfg.InstallDir == "" {
		t.Error("expected InstallDir to be set")
	}
}

func TestIsGroupEnabled(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Groups.Cloud = false

	tests := []struct {
		group string
		want  bool
	}{
		{"languages", true},
		{"containers", true},
		{"kubernetes", true},
		{"infra", true},
		{"cloud", false},
		{"ansible", true},
		{"rust-tools", true},
		{"rust_tools", true},
		{"utilities", true},
		{"nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.group, func(t *testing.T) {
			got := cfg.IsGroupEnabled(tt.group)
			if got != tt.want {
				t.Errorf("IsGroupEnabled(%q) = %v, want %v", tt.group, got, tt.want)
			}
		})
	}
}

func TestLoadNonExistent(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if !cfg.Groups.Languages {
		t.Error("expected default config when file missing")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := DefaultConfig()
	cfg.Groups.Cloud = false
	cfg.Overrides["terraform"] = ToolOverride{Version: "1.7.5"}

	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.Groups.Cloud != false {
		t.Error("expected Cloud to be false after load")
	}
	if loaded.Overrides["terraform"].Version != "1.7.5" {
		t.Errorf("expected terraform override version 1.7.5, got %q", loaded.Overrides["terraform"].Version)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")

	if err := os.WriteFile(path, []byte("{{invalid yaml"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}
