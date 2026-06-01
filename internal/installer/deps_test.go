package installer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/omargallob/devops-starter/pkg/tooldef"
)

func TestCheckDependencies_NoDeps(t *testing.T) {
	inst := &Installer{InstallDir: t.TempDir()}
	tool := &tooldef.Tool{Name: "mytool"}

	if err := inst.checkDependencies(tool); err != nil {
		t.Errorf("expected no error for tool without dependencies, got: %v", err)
	}
}

func TestCheckDependencies_EmptySlice(t *testing.T) {
	inst := &Installer{InstallDir: t.TempDir()}
	tool := &tooldef.Tool{Name: "mytool", Dependencies: []string{}}

	if err := inst.checkDependencies(tool); err != nil {
		t.Errorf("expected no error for empty dependencies, got: %v", err)
	}
}

func TestCheckDependencies_DepInInstallDir(t *testing.T) {
	installDir := t.TempDir()
	// Place a fake binary in the install dir.
	depPath := filepath.Join(installDir, "fakecli")
	if err := os.WriteFile(depPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	inst := &Installer{InstallDir: installDir}
	tool := &tooldef.Tool{
		Name:         "mytool",
		Dependencies: []string{"fakecli"},
	}

	if err := inst.checkDependencies(tool); err != nil {
		t.Errorf("expected no error when dependency exists in install dir, got: %v", err)
	}
}

func TestCheckDependencies_DepOnPATH(t *testing.T) {
	inst := &Installer{InstallDir: t.TempDir()}
	// "sh" is always available on PATH.
	tool := &tooldef.Tool{
		Name:         "mytool",
		Dependencies: []string{"sh"},
	}

	if err := inst.checkDependencies(tool); err != nil {
		t.Errorf("expected no error when dependency is on PATH, got: %v", err)
	}
}

func TestCheckDependencies_SingleMissing(t *testing.T) {
	inst := &Installer{InstallDir: t.TempDir()}
	tool := &tooldef.Tool{
		Name:         "mytool",
		Dependencies: []string{"nonexistent-tool-xyz-12345"},
	}

	err := inst.checkDependencies(tool)
	if err == nil {
		t.Fatal("expected error for missing dependency, got nil")
	}
	if !strings.Contains(err.Error(), "nonexistent-tool-xyz-12345") {
		t.Errorf("error should mention the missing dependency, got: %v", err)
	}
	if !strings.Contains(err.Error(), "mytool") {
		t.Errorf("error should mention the tool name, got: %v", err)
	}
	if !strings.Contains(err.Error(), "requires") {
		t.Errorf("error should contain 'requires', got: %v", err)
	}
}

func TestCheckDependencies_MultipleMissing(t *testing.T) {
	inst := &Installer{InstallDir: t.TempDir()}
	tool := &tooldef.Tool{
		Name:         "mytool",
		Dependencies: []string{"missing-aaa-xyz", "missing-bbb-xyz"},
	}

	err := inst.checkDependencies(tool)
	if err == nil {
		t.Fatal("expected error for missing dependencies, got nil")
	}
	if !strings.Contains(err.Error(), "missing-aaa-xyz") {
		t.Errorf("error should mention missing-aaa-xyz, got: %v", err)
	}
	if !strings.Contains(err.Error(), "missing-bbb-xyz") {
		t.Errorf("error should mention missing-bbb-xyz, got: %v", err)
	}
}

func TestCheckDependencies_PartialMissing(t *testing.T) {
	inst := &Installer{InstallDir: t.TempDir()}
	// "sh" exists, "nonexistent-xyz" does not.
	tool := &tooldef.Tool{
		Name:         "mytool",
		Dependencies: []string{"sh", "nonexistent-tool-xyz-12345"},
	}

	err := inst.checkDependencies(tool)
	if err == nil {
		t.Fatal("expected error when one dependency is missing, got nil")
	}
	if strings.Contains(err.Error(), "sh") {
		t.Errorf("error should not mention satisfied dependency 'sh', got: %v", err)
	}
	if !strings.Contains(err.Error(), "nonexistent-tool-xyz-12345") {
		t.Errorf("error should mention the missing dependency, got: %v", err)
	}
}

func TestCheckDependencies_BinaryNameMapping(t *testing.T) {
	// "neovim" maps to "nvim" via the probe's BinName.
	// Create a fake nvim binary on a temp PATH to test the mapping.
	binDir := t.TempDir()
	nvimPath := filepath.Join(binDir, "nvim")
	if err := os.WriteFile(nvimPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	origPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir+":"+origPath)

	inst := &Installer{InstallDir: t.TempDir()}
	tool := &tooldef.Tool{
		Name:         "mytool",
		Dependencies: []string{"neovim"},
	}

	if err := inst.checkDependencies(tool); err != nil {
		t.Errorf("expected no error when neovim (nvim) is on PATH, got: %v", err)
	}
}

func TestCheckDependencies_DryRunAlsoChecks(t *testing.T) {
	inst := &Installer{
		InstallDir: t.TempDir(),
		DryRun:     true,
		Platform:   tooldef.Platform{OS: "linux", Arch: "amd64"},
	}
	tool := &tooldef.Tool{
		Name:         "mytool",
		Version:      "1.0.0",
		InstallMode:  tooldef.InstallModeEget,
		Dependencies: []string{"nonexistent-tool-xyz-12345"},
	}

	err := inst.Install(context.Background(), tool)
	if err == nil {
		t.Fatal("expected Install to fail on missing dependency even in dry-run mode")
	}
	if !strings.Contains(err.Error(), "nonexistent-tool-xyz-12345") {
		t.Errorf("error should mention missing dependency, got: %v", err)
	}
}
