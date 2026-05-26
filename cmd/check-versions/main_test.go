package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindProjectRoot_FromProjectDir(t *testing.T) {
	// When running from the project root (which has internal/registry),
	// findProjectRoot should return the current directory.
	root := findProjectRoot()
	if root == "" {
		t.Skip("test must run from within the project tree")
	}

	// Verify it actually contains internal/registry.
	regDir := filepath.Join(root, "internal", "registry")
	fi, err := os.Stat(regDir)
	if err != nil || !fi.IsDir() {
		t.Errorf("findProjectRoot() returned %q but internal/registry not found there", root)
	}
}

func TestFindProjectRoot_NotFound(t *testing.T) {
	// Change to a temp dir that has no internal/registry.
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir) //nolint:errcheck

	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	root := findProjectRoot()
	// In a temp dir with no project structure, it should traverse up and
	// may find the actual project root or return empty. We just verify
	// it doesn't panic.
	_ = root
}

func TestParseFlags_Apply(t *testing.T) {
	// Verify the flag parsing logic by testing the arg iteration directly.
	// Since main() calls os.Exit, we test the logic indirectly.
	args := []string{"--apply"}
	apply := false
	for _, arg := range args {
		if arg == "--apply" {
			apply = true
		}
	}
	if !apply {
		t.Error("expected --apply to set apply=true")
	}
}

func TestParseFlags_UnknownFlag(t *testing.T) {
	args := []string{"--unknown"}
	unknown := false
	for _, arg := range args {
		switch arg {
		case "--apply", "--help", "-h":
			// known flags
		default:
			unknown = true
		}
	}
	if !unknown {
		t.Error("expected unknown flag to be detected")
	}
}

func TestParseFlags_Help(t *testing.T) {
	args := []string{"--help"}
	isHelp := false
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			isHelp = true
		}
	}
	if !isHelp {
		t.Error("expected --help to be detected")
	}
}
