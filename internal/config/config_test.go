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
	if !cfg.Groups.AI {
		t.Error("expected AI group to be enabled by default")
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
		{"ai", true},
		{"package-managers", false},
		{"package_managers", false},
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

func TestConfig_MergeGroups(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Groups.Cloud = true
	cfg.Groups.Ansible = true
	cfg.Overrides["terraform"] = ToolOverride{Version: "1.8.0"}
	cfg.Overrides["kubectl"] = ToolOverride{Disabled: true}

	// Merge: disable some groups
	groups := map[string]bool{
		"languages":  true,
		"containers": false,
		"kubernetes": true,
		"infra":      false,
		"cloud":      false,
		"ansible":    true,
		"rust-tools": true,
		"utilities":  true,
	}

	cfg.MergeGroups(groups)

	// Check groups were updated
	if cfg.Groups.Containers != false {
		t.Error("containers should be disabled after merge")
	}
	if cfg.Groups.Infra != false {
		t.Error("infra should be disabled after merge")
	}
	if cfg.Groups.Cloud != false {
		t.Error("cloud should be disabled after merge")
	}
	if cfg.Groups.Languages != true {
		t.Error("languages should still be enabled")
	}

	// Check overrides were NOT touched
	if cfg.Overrides["terraform"].Version != "1.8.0" {
		t.Errorf("terraform override should be preserved, got %q", cfg.Overrides["terraform"].Version)
	}
	if cfg.Overrides["kubectl"].Disabled != true {
		t.Error("kubectl disabled override should be preserved")
	}
}

func TestConfig_MergeGroups_RustToolsVariants(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Groups.RustTools = false

	// Merge using hyphenated form
	cfg.MergeGroups(map[string]bool{"rust-tools": true})
	if !cfg.Groups.RustTools {
		t.Error("rust-tools (hyphenated) should enable RustTools")
	}

	cfg.Groups.RustTools = false
	// Merge using underscore form
	cfg.MergeGroups(map[string]bool{"rust_tools": true})
	if !cfg.Groups.RustTools {
		t.Error("rust_tools (underscore) should enable RustTools")
	}
}

func TestConfig_SaveAndReload_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "roundtrip.yaml")

	original := DefaultConfig()
	original.Groups.Cloud = false
	original.Groups.Ansible = false
	original.InstallDir = "/opt/devops/bin"
	original.Overrides["helm"] = ToolOverride{Version: "3.15.0"}
	original.Overrides["fzf"] = ToolOverride{Disabled: true}

	if err := Save(original, path); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Verify all fields roundtrip correctly
	if loaded.InstallDir != "/opt/devops/bin" {
		t.Errorf("InstallDir: got %q, want /opt/devops/bin", loaded.InstallDir)
	}
	if loaded.Groups.Languages != true {
		t.Error("Languages should be true")
	}
	if loaded.Groups.Containers != true {
		t.Error("Containers should be true")
	}
	if loaded.Groups.Cloud != false {
		t.Error("Cloud should be false")
	}
	if loaded.Groups.Ansible != false {
		t.Error("Ansible should be false")
	}
	if loaded.Groups.RustTools != true {
		t.Error("RustTools should be true")
	}
	if loaded.Overrides["helm"].Version != "3.15.0" {
		t.Errorf("helm override: got %q, want 3.15.0", loaded.Overrides["helm"].Version)
	}
	if loaded.Overrides["fzf"].Disabled != true {
		t.Error("fzf should be disabled")
	}
}

func TestAllGroupNames(t *testing.T) {
	names := AllGroupNames()
	if len(names) != 10 {
		t.Errorf("expected 10 group names, got %d", len(names))
	}
	expected := []string{"languages", "containers", "kubernetes", "infra", "cloud", "ansible", "rust-tools", "utilities", "ai", "package-managers"}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("group %d: got %q, want %q", i, names[i], name)
		}
	}
}

func TestSetGroup(t *testing.T) {
	cfg := DefaultConfig()

	cfg.SetGroup("cloud", false)
	if cfg.Groups.Cloud != false {
		t.Error("SetGroup should disable cloud")
	}

	cfg.SetGroup("cloud", true)
	if cfg.Groups.Cloud != true {
		t.Error("SetGroup should enable cloud")
	}

	// Unknown group should not panic
	cfg.SetGroup("nonexistent", true)
}

func TestSaveAndLoad_ConflictField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := DefaultConfig()
	cfg.Overrides["terraform"] = ToolOverride{Version: "1.10.4", Conflict: "skip"}
	cfg.Overrides["kubectl"] = ToolOverride{Conflict: "overwrite"}
	cfg.Overrides["helm"] = ToolOverride{Conflict: "link"}

	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.Overrides["terraform"].Conflict != "skip" {
		t.Errorf("terraform conflict: got %q, want 'skip'", loaded.Overrides["terraform"].Conflict)
	}
	if loaded.Overrides["terraform"].Version != "1.10.4" {
		t.Errorf("terraform version: got %q, want '1.10.4'", loaded.Overrides["terraform"].Version)
	}
	if loaded.Overrides["kubectl"].Conflict != "overwrite" {
		t.Errorf("kubectl conflict: got %q, want 'overwrite'", loaded.Overrides["kubectl"].Conflict)
	}
	if loaded.Overrides["helm"].Conflict != "link" {
		t.Errorf("helm conflict: got %q, want 'link'", loaded.Overrides["helm"].Conflict)
	}
}

func TestIsValidConflictAction(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"skip", true},
		{"overwrite", true},
		{"link", true},
		{"", true},
		{"invalid", false},
		{"SKIP", false},
	}

	for _, tt := range tests {
		got := IsValidConflictAction(tt.input)
		if got != tt.want {
			t.Errorf("IsValidConflictAction(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
