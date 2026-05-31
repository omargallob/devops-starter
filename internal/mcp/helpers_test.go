package mcp

import (
	"os"
	"path/filepath"
	"testing"
)

// TestHomeDirectory verifies that homeDirectory returns a non-empty path.
func TestHomeDirectory(t *testing.T) {
	home, err := homeDirectory()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if home == "" {
		t.Error("expected non-empty home directory")
	}
	// Should be an absolute path.
	if !filepath.IsAbs(home) {
		t.Errorf("expected absolute path, got %s", home)
	}
}

// TestDotfilesSourceDir verifies it returns an absolute path ending with "dotfiles".
func TestDotfilesSourceDir(t *testing.T) {
	dir := dotfilesSourceDir()
	if dir == "" {
		t.Fatal("expected non-empty path")
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("expected absolute path, got %s", dir)
	}
	if filepath.Base(dir) != "dotfiles" {
		t.Errorf("expected path to end with 'dotfiles', got %s", dir)
	}
}

// TestDotfilesSourceDir_ContainsWorkingDir verifies the path is rooted in cwd.
func TestDotfilesSourceDir_ContainsWorkingDir(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Skip("cannot determine working directory")
	}

	dir := dotfilesSourceDir()
	expected := filepath.Join(wd, "dotfiles")
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}
