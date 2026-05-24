package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStatePath_DefaultsToHomeConfig(t *testing.T) {
	// Unset XDG so we exercise the home-dir branch.
	t.Setenv("XDG_CONFIG_HOME", "")

	p := StatePath()
	if !filepath.IsAbs(p) {
		t.Fatalf("expected absolute path, got %s", p)
	}
	if filepath.Base(p) != "state.json" {
		t.Fatalf("expected state.json, got %s", filepath.Base(p))
	}
}

func TestStatePath_XDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/config")

	p := StatePath()
	want := "/custom/config/devops-starter/state.json"
	if p != want {
		t.Fatalf("got %s, want %s", p, want)
	}
}

func TestLoadStore_NonExistent(t *testing.T) {
	p := filepath.Join(t.TempDir(), "does-not-exist.json")

	s, err := LoadStore(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil store")
	}
	if len(s.Tools) != 0 {
		t.Fatalf("expected empty tools map, got %d entries", len(s.Tools))
	}
}

func TestLoadStore_ValidJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "state.json")

	data := `{
		"tools": {
			"kubectl": {"version": "1.31.4", "installed_at": "2025-01-01T00:00:00Z"},
			"helm": {"version": "3.16.4", "installed_at": "2025-01-02T00:00:00Z"}
		}
	}`
	if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := LoadStore(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(s.Tools))
	}
	if s.Tools["kubectl"].Version != "1.31.4" {
		t.Errorf("kubectl version: got %s, want 1.31.4", s.Tools["kubectl"].Version)
	}
	if s.Tools["helm"].Version != "3.16.4" {
		t.Errorf("helm version: got %s, want 3.16.4", s.Tools["helm"].Version)
	}
}

func TestLoadStore_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "state.json")

	if err := os.WriteFile(p, []byte("{invalid json}"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadStore(p)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadStore_NullToolsMap(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "state.json")

	// JSON with explicit null tools
	if err := os.WriteFile(p, []byte(`{"tools": null}`), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := LoadStore(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Tools == nil {
		t.Fatal("expected non-nil tools map even when JSON has null")
	}
}

func TestStore_Save(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "sub", "state.json")

	s := &Store{
		Tools: map[string]InstalledTool{
			"fzf": {Version: "0.57.0", InstalledAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
		path: p,
	}

	if err := s.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file was written and is valid JSON
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("reading saved file: %v", err)
	}

	var loaded Store
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("unmarshaling saved file: %v", err)
	}
	if loaded.Tools["fzf"].Version != "0.57.0" {
		t.Errorf("got version %s, want 0.57.0", loaded.Tools["fzf"].Version)
	}
}

func TestStore_Record(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "state.json")

	s := &Store{
		Tools: make(map[string]InstalledTool),
		path:  p,
	}

	if err := s.Record("kubectl", "1.31.4"); err != nil {
		t.Fatalf("Record failed: %v", err)
	}

	// Verify the in-memory state
	if s.Tools["kubectl"].Version != "1.31.4" {
		t.Errorf("got version %s, want 1.31.4", s.Tools["kubectl"].Version)
	}
	if s.Tools["kubectl"].InstalledAt.IsZero() {
		t.Error("InstalledAt should not be zero")
	}

	// Verify persisted to disk
	loaded, err := LoadStore(p)
	if err != nil {
		t.Fatalf("LoadStore failed: %v", err)
	}
	if loaded.Tools["kubectl"].Version != "1.31.4" {
		t.Errorf("persisted version: got %s, want 1.31.4", loaded.Tools["kubectl"].Version)
	}
}

func TestStore_GetVersion(t *testing.T) {
	s := &Store{
		Tools: map[string]InstalledTool{
			"helm": {Version: "3.16.4"},
		},
	}

	if got := s.GetVersion("helm"); got != "3.16.4" {
		t.Errorf("got %s, want 3.16.4", got)
	}
	if got := s.GetVersion("nonexistent"); got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}
