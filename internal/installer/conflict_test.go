package installer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

func TestDetectConflicts_NoConflictsWhenToolNotOnPath(t *testing.T) {
	tools := []*tooldef.Tool{
		{Name: "nonexistent-tool-xyz", Version: "1.0.0"},
	}

	conflicts := DetectConflicts(tools, "/tmp/test-install-dir", nil)
	if len(conflicts) != 0 {
		t.Errorf("expected 0 conflicts, got %d", len(conflicts))
	}
}

func TestDetectConflicts_NoConflictWhenBinaryInInstallDir(t *testing.T) {
	// Create a temp dir to act as install dir with a binary in it.
	installDir := t.TempDir()
	binPath := filepath.Join(installDir, "mytool")
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	tools := []*tooldef.Tool{
		{Name: "mytool", Version: "1.0.0"},
	}

	// Prepend our install dir to PATH so LookupInPath finds it.
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", installDir+":"+origPath)

	conflicts := DetectConflicts(tools, installDir, nil)
	if len(conflicts) != 0 {
		t.Errorf("expected 0 conflicts (binary is in install dir), got %d", len(conflicts))
	}
}

func TestDetectConflicts_DetectsSystemBinary(t *testing.T) {
	// Create a "system" dir separate from install dir.
	systemDir := t.TempDir()
	binPath := filepath.Join(systemDir, "mytool")
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\necho v1.2.3"), 0o755); err != nil {
		t.Fatal(err)
	}

	installDir := t.TempDir()

	tools := []*tooldef.Tool{
		{Name: "mytool", Version: "2.0.0"},
	}

	origPath := os.Getenv("PATH")
	t.Setenv("PATH", systemDir+":"+origPath)

	conflicts := DetectConflicts(tools, installDir, nil)
	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}

	if conflicts[0].Tool.Name != "mytool" {
		t.Errorf("expected conflict for mytool, got %s", conflicts[0].Tool.Name)
	}
	if conflicts[0].SystemPath != binPath {
		t.Errorf("expected system path %s, got %s", binPath, conflicts[0].SystemPath)
	}
	if conflicts[0].WantedVersion != "2.0.0" {
		t.Errorf("expected wanted version 2.0.0, got %s", conflicts[0].WantedVersion)
	}
}

func TestDetectConflicts_AttachesSavedAction(t *testing.T) {
	systemDir := t.TempDir()
	binPath := filepath.Join(systemDir, "mytool")
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	installDir := t.TempDir()

	tools := []*tooldef.Tool{
		{Name: "mytool", Version: "2.0.0"},
	}

	origPath := os.Getenv("PATH")
	t.Setenv("PATH", systemDir+":"+origPath)

	overrides := map[string]config.ToolOverride{
		"mytool": {Conflict: "skip"},
	}

	conflicts := DetectConflicts(tools, installDir, overrides)
	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}
	if conflicts[0].SavedAction != "skip" {
		t.Errorf("expected saved action 'skip', got %q", conflicts[0].SavedAction)
	}
}

func TestUnresolvedConflicts(t *testing.T) {
	conflicts := []ConflictInfo{
		{Tool: &tooldef.Tool{Name: "a"}, SavedAction: "skip"},
		{Tool: &tooldef.Tool{Name: "b"}, SavedAction: ""},
		{Tool: &tooldef.Tool{Name: "c"}, SavedAction: "overwrite"},
		{Tool: &tooldef.Tool{Name: "d"}, SavedAction: ""},
	}

	unresolved := UnresolvedConflicts(conflicts)
	if len(unresolved) != 2 {
		t.Fatalf("expected 2 unresolved, got %d", len(unresolved))
	}
	if unresolved[0].Tool.Name != "b" || unresolved[1].Tool.Name != "d" {
		t.Errorf("unexpected unresolved tools: %v, %v", unresolved[0].Tool.Name, unresolved[1].Tool.Name)
	}
}

func TestApplyConflictActions(t *testing.T) {
	tools := []*tooldef.Tool{
		{Name: "a", Version: "1.0"},
		{Name: "b", Version: "2.0"},
		{Name: "c", Version: "3.0"},
		{Name: "d", Version: "4.0"}, // no conflict
	}

	conflicts := []ConflictInfo{
		{Tool: tools[0], SystemPath: "/usr/bin/a", SavedAction: "skip"},
		{Tool: tools[1], SystemPath: "/usr/bin/b", SavedAction: ""},
		{Tool: tools[2], SystemPath: "/usr/bin/c", SavedAction: ""},
	}

	resolutions := map[string]string{
		"b": "overwrite",
		"c": "link",
	}

	install, skip, links := ApplyConflictActions(tools, conflicts, resolutions)

	// "a" is skipped (saved action), "b" overwrites, "c" links, "d" installs normally.
	if len(install) != 2 {
		t.Errorf("expected 2 tools to install, got %d", len(install))
	}
	if len(skip) != 2 { // "a" (skip) + "c" (link goes to skip list)
		t.Errorf("expected 2 tools skipped, got %d", len(skip))
	}
	if len(links) != 1 {
		t.Errorf("expected 1 link, got %d", len(links))
	}
	if links["c"] != "/usr/bin/c" {
		t.Errorf("expected link for c -> /usr/bin/c, got %s", links["c"])
	}
}

func TestIsInsideDir(t *testing.T) {
	tests := []struct {
		path, dir string
		want      bool
	}{
		{"/home/user/.local/bin/kubectl", "/home/user/.local/bin", true},
		{"/home/user/.local/bin/kubectl", "/home/user/.local/bin/", true},
		{"/usr/local/bin/kubectl", "/home/user/.local/bin", false},
		{"/home/user/.local/bin", "/home/user/.local/bin", true},
	}

	for _, tt := range tests {
		got := isInsideDir(tt.path, tt.dir)
		if got != tt.want {
			t.Errorf("isInsideDir(%q, %q) = %v, want %v", tt.path, tt.dir, got, tt.want)
		}
	}
}
