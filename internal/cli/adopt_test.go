package cli

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/internal/state"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

func TestAdoptCmd_Exists(t *testing.T) {
	root := NewRootCmd()
	var found bool
	for _, cmd := range root.Commands() {
		if cmd.Use == "adopt [tool...]" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("adopt command not registered on root")
	}
}

func TestDoAdopt_NoToolsToAdopt(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{Name: "kubectl", Version: "1.29.0", Group: tooldef.GroupKubernetes},
	)
	inst := newFakeInstaller()
	store := newFakeStore()

	cfg := config.DefaultConfig()
	cfg.InstallDir = "/tmp/bin"

	var buf bytes.Buffer
	deps := adoptDeps{
		cfg:       cfg,
		registry:  reg,
		installer: inst,
		store:     store,
		out:       &buf,
		dryRun:    false,
		autoYes:   true,
		adoptAll:  true,
		platform:  tooldef.Platform{OS: "darwin", Arch: "arm64"},
		detected:  map[string]state.ToolState{}, // nothing detected
	}

	err := doAdopt(deps, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "No detected tools to adopt") {
		t.Errorf("expected 'No detected tools' message, got:\n%s", buf.String())
	}
}

func TestDoAdopt_UnknownTool(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{Name: "kubectl", Version: "1.29.0", Group: tooldef.GroupKubernetes},
	)
	inst := newFakeInstaller()
	store := newFakeStore()

	cfg := config.DefaultConfig()
	cfg.InstallDir = "/tmp/bin"

	var buf bytes.Buffer
	deps := adoptDeps{
		cfg:       cfg,
		registry:  reg,
		installer: inst,
		store:     store,
		out:       &buf,
		dryRun:    false,
		autoYes:   true,
		adoptAll:  false,
		platform:  tooldef.Platform{OS: "darwin", Arch: "arm64"},
		detected:  map[string]state.ToolState{},
	}

	err := doAdopt(deps, []string{"nonexistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "unknown tool") {
		t.Errorf("expected 'unknown tool' warning, got:\n%s", output)
	}
}

func TestDoAdopt_AlreadyManaged(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{Name: "kubectl", Version: "1.29.0", Group: tooldef.GroupKubernetes},
	)
	inst := newFakeInstaller()
	store := newFakeStore()
	store.tools["kubectl"] = "1.28.0"

	cfg := config.DefaultConfig()
	cfg.InstallDir = "/tmp/bin"

	var buf bytes.Buffer
	deps := adoptDeps{
		cfg:       cfg,
		registry:  reg,
		installer: inst,
		store:     store,
		out:       &buf,
		dryRun:    false,
		autoYes:   true,
		adoptAll:  false,
		platform:  tooldef.Platform{OS: "darwin", Arch: "arm64"},
		detected:  map[string]state.ToolState{}, // not in detected — already managed
	}

	err := doAdopt(deps, []string{"kubectl"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "already managed") {
		t.Errorf("expected 'already managed' warning, got:\n%s", output)
	}
}

func TestDoAdopt_DisabledInConfig(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{Name: "kubectl", Version: "1.29.0", Group: tooldef.GroupKubernetes},
	)
	inst := newFakeInstaller()
	store := newFakeStore()

	cfg := config.DefaultConfig()
	cfg.InstallDir = "/tmp/bin"
	cfg.Overrides = map[string]config.ToolOverride{
		"kubectl": {Disabled: true},
	}

	var buf bytes.Buffer
	deps := adoptDeps{
		cfg:       cfg,
		registry:  reg,
		installer: inst,
		store:     store,
		out:       &buf,
		dryRun:    false,
		autoYes:   true,
		adoptAll:  false,
		platform:  tooldef.Platform{OS: "darwin", Arch: "arm64"},
		detected: map[string]state.ToolState{
			"kubectl": {Name: "kubectl", Status: state.StatusDetected, DetectedPath: "/usr/local/bin/kubectl"},
		},
	}

	err := doAdopt(deps, []string{"kubectl"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "disabled in config") {
		t.Errorf("expected 'disabled in config' warning, got:\n%s", output)
	}
}

func TestDoAdopt_DryRun(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{Name: "kubectl", Version: "1.29.0", Group: tooldef.GroupKubernetes},
	)
	inst := newFakeInstaller()
	store := newFakeStore()

	cfg := config.DefaultConfig()
	cfg.InstallDir = "/tmp/bin"

	var buf bytes.Buffer
	deps := adoptDeps{
		cfg:       cfg,
		registry:  reg,
		installer: inst,
		store:     store,
		out:       &buf,
		dryRun:    true,
		autoYes:   true,
		adoptAll:  false,
		platform:  tooldef.Platform{OS: "darwin", Arch: "arm64"},
		detected: map[string]state.ToolState{
			"kubectl": {Name: "kubectl", Status: state.StatusDetected, DetectedPath: "/usr/local/bin/kubectl"},
		},
	}

	err := doAdopt(deps, []string{"kubectl"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[dry-run]") {
		t.Errorf("expected '[dry-run]' message, got:\n%s", output)
	}
	if len(inst.installed) != 0 {
		t.Error("dry-run should not install anything")
	}
}

func TestDoAdopt_SuccessfulAdoption(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{Name: "kubectl", Version: "1.29.0", Group: tooldef.GroupKubernetes},
	)
	inst := newFakeInstaller()
	store := newFakeStore()

	cfg := config.DefaultConfig()
	cfg.InstallDir = "/tmp/bin"

	var buf bytes.Buffer
	deps := adoptDeps{
		cfg:       cfg,
		registry:  reg,
		installer: inst,
		store:     store,
		out:       &buf,
		dryRun:    false,
		autoYes:   true,
		adoptAll:  false,
		platform:  tooldef.Platform{OS: "darwin", Arch: "arm64"},
		detected: map[string]state.ToolState{
			"kubectl": {Name: "kubectl", Status: state.StatusDetected, DetectedPath: "/usr/local/bin/kubectl", DetectedVersion: "1.27.0"},
		},
	}

	err := doAdopt(deps, []string{"kubectl"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(inst.installed) != 1 {
		t.Fatalf("expected 1 tool installed, got %d", len(inst.installed))
	}
	if inst.installed[0].Name != "kubectl" {
		t.Errorf("expected kubectl, got %s", inst.installed[0].Name)
	}

	output := buf.String()
	if !strings.Contains(output, "1 tool(s) adopted successfully") {
		t.Errorf("expected success message, got:\n%s", output)
	}
}

func TestDoAdopt_AllDetected(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{Name: "kubectl", Version: "1.29.0", Group: tooldef.GroupKubernetes},
		&tooldef.Tool{Name: "helm", Version: "3.14.0", Group: tooldef.GroupKubernetes},
	)
	inst := newFakeInstaller()
	store := newFakeStore()

	cfg := config.DefaultConfig()
	cfg.InstallDir = "/tmp/bin"

	var buf bytes.Buffer
	deps := adoptDeps{
		cfg:       cfg,
		registry:  reg,
		installer: inst,
		store:     store,
		out:       &buf,
		dryRun:    false,
		autoYes:   true,
		adoptAll:  true,
		platform:  tooldef.Platform{OS: "darwin", Arch: "arm64"},
		detected: map[string]state.ToolState{
			"kubectl": {Name: "kubectl", Status: state.StatusDetected, DetectedPath: "/usr/local/bin/kubectl"},
			"helm":    {Name: "helm", Status: state.StatusDetected, DetectedPath: "/usr/local/bin/helm"},
		},
	}

	err := doAdopt(deps, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(inst.installed) != 2 {
		t.Fatalf("expected 2 tools installed, got %d", len(inst.installed))
	}

	output := buf.String()
	if !strings.Contains(output, "2 tool(s) adopted successfully") {
		t.Errorf("expected '2 tool(s) adopted' message, got:\n%s", output)
	}
}

func TestDoAdopt_InstallerFailure(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{Name: "kubectl", Version: "1.29.0", Group: tooldef.GroupKubernetes},
	)
	inst := newFakeInstaller()
	inst.failTools["kubectl"] = fmt.Errorf("network timeout")
	store := newFakeStore()

	cfg := config.DefaultConfig()
	cfg.InstallDir = "/tmp/bin"

	var buf bytes.Buffer
	deps := adoptDeps{
		cfg:       cfg,
		registry:  reg,
		installer: inst,
		store:     store,
		out:       &buf,
		dryRun:    false,
		autoYes:   true,
		adoptAll:  false,
		platform:  tooldef.Platform{OS: "darwin", Arch: "arm64"},
		detected: map[string]state.ToolState{
			"kubectl": {Name: "kubectl", Status: state.StatusDetected},
		},
	}

	err := doAdopt(deps, []string{"kubectl"})
	if err == nil {
		t.Fatal("expected error on installer failure")
	}
	if !strings.Contains(err.Error(), "1 adoptions failed") {
		t.Errorf("expected '1 adoptions failed', got: %v", err)
	}
}

func TestDoAdopt_VersionOverride(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{Name: "kubectl", Version: "1.29.0", Group: tooldef.GroupKubernetes},
	)
	inst := newFakeInstaller()
	store := newFakeStore()

	cfg := config.DefaultConfig()
	cfg.InstallDir = "/tmp/bin"
	cfg.Overrides = map[string]config.ToolOverride{
		"kubectl": {Version: "1.28.5"},
	}

	var buf bytes.Buffer
	deps := adoptDeps{
		cfg:       cfg,
		registry:  reg,
		installer: inst,
		store:     store,
		out:       &buf,
		dryRun:    false,
		autoYes:   true,
		adoptAll:  false,
		platform:  tooldef.Platform{OS: "darwin", Arch: "arm64"},
		detected: map[string]state.ToolState{
			"kubectl": {Name: "kubectl", Status: state.StatusDetected},
		},
	}

	err := doAdopt(deps, []string{"kubectl"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(inst.installed) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(inst.installed))
	}
	if inst.installed[0].Version != "1.28.5" {
		t.Errorf("expected version 1.28.5, got %s", inst.installed[0].Version)
	}
}

func TestDoAdopt_UnsupportedPlatform(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{
			Name:      "linux-only",
			Version:   "1.0.0",
			Group:     tooldef.GroupUtilities,
			Platforms: []tooldef.Platform{{OS: "linux", Arch: "amd64"}},
		},
	)
	inst := newFakeInstaller()
	store := newFakeStore()

	cfg := config.DefaultConfig()
	cfg.InstallDir = "/tmp/bin"

	var buf bytes.Buffer
	deps := adoptDeps{
		cfg:       cfg,
		registry:  reg,
		installer: inst,
		store:     store,
		out:       &buf,
		dryRun:    false,
		autoYes:   true,
		adoptAll:  false,
		platform:  tooldef.Platform{OS: "darwin", Arch: "arm64"},
		detected: map[string]state.ToolState{
			"linux-only": {Name: "linux-only", Status: state.StatusDetected},
		},
	}

	err := doAdopt(deps, []string{"linux-only"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "not supported on") {
		t.Errorf("expected platform warning, got:\n%s", output)
	}
}
