package pkgmgr

import (
	"context"
	"testing"
)

func TestNew_UnsupportedManager(t *testing.T) {
	_, err := New("cargo")
	if err == nil {
		t.Fatal("expected error for unsupported manager")
	}
}

func TestWithDryRun_PipDoesNotExec(t *testing.T) {
	// Use a fake bin path so we test dry-run path without needing pip installed.
	m := &pipManager{bin: "/nonexistent/pip", dryRun: true}
	// Dry-run must not execute the binary, so it must not return an exec error.
	if err := m.Install(context.Background(), "black", "24.4"); err != nil {
		t.Errorf("dry-run Install() returned error: %v", err)
	}
}

func TestWithDryRun_NpmDoesNotExec(t *testing.T) {
	m := &npmManager{bin: "/nonexistent/npm", dryRun: true}
	if err := m.Install(context.Background(), "typescript", "5.4"); err != nil {
		t.Errorf("dry-run Install() returned error: %v", err)
	}
}

func TestPipInstall_VersionlessSpecifier(t *testing.T) {
	// Dry-run: version empty → no "==" in command (just prints).
	m := &pipManager{bin: "/nonexistent/pip", dryRun: true}
	if err := m.Install(context.Background(), "black", ""); err != nil {
		t.Errorf("dry-run versionless Install() returned error: %v", err)
	}
}

func TestNpmInstall_VersionlessSpecifier(t *testing.T) {
	m := &npmManager{bin: "/nonexistent/npm", dryRun: true}
	if err := m.Install(context.Background(), "eslint", ""); err != nil {
		t.Errorf("dry-run versionless Install() returned error: %v", err)
	}
}
