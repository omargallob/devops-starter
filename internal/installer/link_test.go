package installer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/omargallob/devops-starter/pkg/tooldef"
)

func TestLink_CreatesSymlink(t *testing.T) {
	installDir := t.TempDir()
	systemDir := t.TempDir()

	// Create a "system binary".
	systemBin := filepath.Join(systemDir, "mytool")
	if err := os.WriteFile(systemBin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	inst := New(installDir, tooldef.Platform{OS: "linux", Arch: "amd64"})
	tool := &tooldef.Tool{Name: "mytool", Version: "1.0.0"}

	if err := inst.Link(tool, systemBin); err != nil {
		t.Fatalf("Link failed: %v", err)
	}

	linkPath := filepath.Join(installDir, "mytool")
	fi, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("expected symlink at %s, got error: %v", linkPath, err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected symlink, got mode %v", fi.Mode())
	}

	target, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("Readlink failed: %v", err)
	}
	if target != systemBin {
		t.Errorf("expected target %s, got %s", systemBin, target)
	}
}

func TestLink_OverwritesExistingFile(t *testing.T) {
	installDir := t.TempDir()
	systemDir := t.TempDir()

	systemBin := filepath.Join(systemDir, "mytool")
	if err := os.WriteFile(systemBin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Pre-create a regular file at the link destination.
	existing := filepath.Join(installDir, "mytool")
	if err := os.WriteFile(existing, []byte("old binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	inst := New(installDir, tooldef.Platform{OS: "linux", Arch: "amd64"})
	tool := &tooldef.Tool{Name: "mytool", Version: "1.0.0"}

	if err := inst.Link(tool, systemBin); err != nil {
		t.Fatalf("Link failed: %v", err)
	}

	// Verify it's now a symlink.
	fi, err := os.Lstat(existing)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected symlink after overwrite, got mode %v", fi.Mode())
	}
}

func TestLink_DryRunDoesNotCreate(t *testing.T) {
	installDir := t.TempDir()
	systemBin := "/usr/bin/fake"

	inst := New(installDir, tooldef.Platform{OS: "linux", Arch: "amd64"}, WithDryRun(true))
	tool := &tooldef.Tool{Name: "mytool", Version: "1.0.0"}

	if err := inst.Link(tool, systemBin); err != nil {
		t.Fatalf("Link (dry-run) failed: %v", err)
	}

	linkPath := filepath.Join(installDir, "mytool")
	if _, err := os.Lstat(linkPath); !os.IsNotExist(err) {
		t.Errorf("expected no file in dry-run, but found one")
	}
}

// testConflictGetter implements ToolConflictGetter for testing.
type testConflictGetter struct {
	conflict string
}

func (g testConflictGetter) GetConflict() string { return g.conflict }

func TestVerifyLinks_DetectsBrokenSymlink(t *testing.T) {
	installDir := t.TempDir()

	// Create a symlink to a non-existent target.
	linkPath := filepath.Join(installDir, "mytool")
	os.Symlink("/nonexistent/path/mytool", linkPath)

	tools := []*tooldef.Tool{
		{Name: "mytool", Version: "1.0.0"},
	}
	overrides := map[string]ToolConflictGetter{
		"mytool": testConflictGetter{conflict: "link"},
	}

	results := VerifyLinks(installDir, tools, overrides)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Broken {
		t.Error("expected broken link")
	}
	if results[0].ToolName != "mytool" {
		t.Errorf("expected tool 'mytool', got %s", results[0].ToolName)
	}
}

func TestVerifyLinks_HealthySymlink(t *testing.T) {
	installDir := t.TempDir()
	systemDir := t.TempDir()

	systemBin := filepath.Join(systemDir, "mytool")
	if err := os.WriteFile(systemBin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	linkPath := filepath.Join(installDir, "mytool")
	os.Symlink(systemBin, linkPath)

	tools := []*tooldef.Tool{
		{Name: "mytool", Version: "1.0.0"},
	}
	overrides := map[string]ToolConflictGetter{
		"mytool": testConflictGetter{conflict: "link"},
	}

	results := VerifyLinks(installDir, tools, overrides)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Broken {
		t.Error("expected healthy link, got broken")
	}
}

func TestVerifyLinks_SkipsNonLinkTools(t *testing.T) {
	installDir := t.TempDir()

	tools := []*tooldef.Tool{
		{Name: "mytool", Version: "1.0.0"},
	}
	overrides := map[string]ToolConflictGetter{
		"mytool": testConflictGetter{conflict: "overwrite"},
	}

	results := VerifyLinks(installDir, tools, overrides)
	if len(results) != 0 {
		t.Errorf("expected 0 results for non-link tool, got %d", len(results))
	}
}
