// Package test contains contract tests that validate devops-starter's
// public types and installer behaviour against their declared contracts.
package test_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/omargallob/devops-starter/internal/installer"
	"github.com/omargallob/devops-starter/internal/registry"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// TestGlazePkgModeRequiresPackageNames validates the contract:
// a Tool with InstallMode=glazepkg must declare at least one PackageNames entry.
func TestGlazePkgModeRequiresPackageNames(t *testing.T) {
	tool := &tooldef.Tool{
		Name:        "missing-pkg-names",
		Version:     "1.0.0",
		InstallMode: tooldef.InstallModeGlazePkg,
		// PackageNames intentionally omitted
	}
	if err := tool.Validate(); err == nil {
		t.Error("expected Validate() to return error for glazepkg tool with no PackageNames")
	}
}

// TestGlazePkgModeWithPackageNamesIsValid validates that a properly declared
// glazepkg tool passes contract validation.
func TestGlazePkgModeWithPackageNamesIsValid(t *testing.T) {
	tool := &tooldef.Tool{
		Name:        "kubectl",
		Version:     "1.31.4",
		InstallMode: tooldef.InstallModeGlazePkg,
		PackageNames: map[string]string{
			"brew": "kubernetes-cli",
			"apt":  "kubectl",
		},
	}
	if err := tool.Validate(); err != nil {
		t.Errorf("expected valid glazepkg tool to pass Validate(), got: %v", err)
	}
}

// TestPackageNamesMarshal verifies the YAML round-trip for the PackageNames field.
func TestPackageNamesMarshal(t *testing.T) {
	tests := []struct {
		name         string
		packageNames map[string]string
	}{
		{
			name:         "single manager",
			packageNames: map[string]string{"brew": "kubernetes-cli"},
		},
		{
			name: "multiple managers",
			packageNames: map[string]string{
				"brew":   "kubernetes-cli",
				"apt":    "kubectl",
				"winget": "Kubernetes.kubectl",
			},
		},
		{
			name:         "nil (omitted)",
			packageNames: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &tooldef.Tool{
				Name:         "test-tool",
				Version:      "1.0.0",
				InstallMode:  tooldef.InstallModeGlazePkg,
				PackageNames: tt.packageNames,
			}

			data, err := yaml.Marshal(original)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			var restored tooldef.Tool
			if err := yaml.Unmarshal(data, &restored); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			if len(restored.PackageNames) != len(tt.packageNames) {
				t.Errorf("PackageNames length: got %d, want %d", len(restored.PackageNames), len(tt.packageNames))
			}
			for k, v := range tt.packageNames {
				if restored.PackageNames[k] != v {
					t.Errorf("PackageNames[%q] = %q, want %q", k, restored.PackageNames[k], v)
				}
			}
		})
	}
}

// TestInstallerFallback_GlazePkgUnavailable verifies that when gpk is not in
// PATH, the installer transparently falls back to eget for tools that have
// a Repo declared.
func TestInstallerFallback_GlazePkgUnavailable(t *testing.T) {
	// Ensure gpk is NOT on PATH for this test by using a controlled PATH.
	orig := os.Getenv("PATH")
	t.Cleanup(func() { os.Setenv("PATH", orig) })

	tmp := t.TempDir()
	os.Setenv("PATH", tmp) // only tmp dir — gpk not present

	// A glazepkg tool that also has a Repo for fallback.
	tool := &tooldef.Tool{
		Name:        "jq",
		Version:     "1.7.1",
		InstallMode: tooldef.InstallModeGlazePkg,
		PackageNames: map[string]string{
			"brew": "jq",
			"apt":  "jq",
		},
		// eget fallback
		Repo: "jqlang/jq",
	}

	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}
	inst := installer.New(tmp, plat, installer.WithPreferNativeManagers(true))

	// We only test that the installer routes to eget (not that it succeeds
	// in downloading — network is not guaranteed in tests). The install will
	// fail because eget is also absent, but the error must NOT be
	// ErrGlazePkgNotAvailable (it should be an eget/exec error instead).
	ctx := context.Background()
	err := inst.Install(ctx, tool)
	if err == nil {
		t.Fatal("expected install error (no eget binary), got nil")
	}
	if strings.Contains(err.Error(), "glazepkg") {
		t.Errorf("expected glazepkg fallback to eget, but got glazepkg error: %v", err)
	}
}

// TestPreferNativeManagers_False verifies that disabling PreferNativeManagers
// causes glazepkg tools to skip gpk entirely and go straight to eget.
func TestPreferNativeManagers_False(t *testing.T) {
	tmp := t.TempDir()

	tool := &tooldef.Tool{
		Name:        "fzf",
		Version:     "0.57.0",
		InstallMode: tooldef.InstallModeGlazePkg,
		PackageNames: map[string]string{
			"brew": "fzf",
		},
		Repo: "junegunn/fzf",
	}

	plat := tooldef.Platform{OS: "darwin", Arch: "arm64"}
	inst := installer.New(tmp, plat, installer.WithPreferNativeManagers(false))

	ctx := context.Background()
	err := inst.Install(ctx, tool)
	// Error expected (no eget binary in tmp), but it must not mention glazepkg.
	if err != nil && strings.Contains(err.Error(), "glazepkg") {
		t.Errorf("PreferNativeManagers=false should bypass gpk entirely, got: %v", err)
	}
}

// TestRegisteredGlazePkgToolsAreValid checks that every glazepkg-mode tool in
// the built-in registry passes contract validation (has PackageNames declared).
func TestRegisteredGlazePkgToolsAreValid(t *testing.T) {
	reg := registry.New()
	for _, tool := range reg.All() {
		if tool.EffectiveInstallMode() != tooldef.InstallModeGlazePkg {
			continue
		}
		if err := tool.Validate(); err != nil {
			t.Errorf("registry tool %q fails contract: %v", tool.Name, err)
		}
	}
}

// TestInstall_DryRun_GlazePkg verifies that dry-run does not attempt any
// real installation for glazepkg-mode tools.
func TestInstall_DryRun_GlazePkg(t *testing.T) {
	tmp := t.TempDir()
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}
	inst := installer.New(tmp, plat, installer.WithDryRun(true))

	tool := &tooldef.Tool{
		Name:        "kubectl",
		Version:     "1.31.4",
		InstallMode: tooldef.InstallModeGlazePkg,
		PackageNames: map[string]string{
			"apt": "kubectl",
		},
		URLTemplate: "https://dl.k8s.io/release/v{{.Version}}/bin/{{.OS}}/{{.Arch}}/kubectl",
		Format:      tooldef.FormatBinary,
	}

	if err := inst.Install(context.Background(), tool); err != nil {
		t.Fatalf("dry-run should not error: %v", err)
	}

	// Nothing should be installed in the tmp dir.
	entries, _ := os.ReadDir(tmp)
	for _, e := range entries {
		if e.Name() == filepath.Base(tool.Name) {
			t.Error("dry-run should not have created any files")
		}
	}
}
