package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

func TestDoRemove_UnknownTool(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{Name: "kubectl", Version: "1.29.0", Group: tooldef.GroupKubernetes},
	)
	store := newFakeStore()

	cfg := config.DefaultConfig()
	installDir := t.TempDir()
	cfg.InstallDir = installDir

	var buf bytes.Buffer
	deps := removeDeps{
		cfg:        cfg,
		registry:   reg,
		store:      store,
		out:        &buf,
		dryRun:     false,
		autoYes:    true,
		installDir: installDir,
	}

	err := doRemove(deps, []string{"nonexistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "unknown tool") {
		t.Errorf("expected 'unknown tool' warning, got:\n%s", output)
	}
	if !strings.Contains(output, "No tools to remove") {
		t.Errorf("expected 'No tools to remove' message, got:\n%s", output)
	}
}

func TestDoRemove_ToolNotManaged(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{Name: "kubectl", Version: "1.29.0", Group: tooldef.GroupKubernetes},
	)
	store := newFakeStore()
	// kubectl is not in the state store and has no binary

	installDir := t.TempDir()

	var buf bytes.Buffer
	deps := removeDeps{
		cfg:        config.DefaultConfig(),
		registry:   reg,
		store:      store,
		out:        &buf,
		dryRun:     false,
		autoYes:    true,
		installDir: installDir,
	}

	err := doRemove(deps, []string{"kubectl"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "not managed") {
		t.Errorf("expected 'not managed' warning, got:\n%s", output)
	}
}

func TestDoRemove_DryRun(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{Name: "kubectl", Version: "1.29.0", Group: tooldef.GroupKubernetes},
	)
	store := newFakeStore()
	store.tools["kubectl"] = "1.29.0"

	// Create a fake binary
	installDir := t.TempDir()
	binPath := filepath.Join(installDir, "kubectl")
	if err := os.WriteFile(binPath, []byte("fake"), 0o755); err != nil {
		t.Fatalf("failed to create fake binary: %v", err)
	}

	var buf bytes.Buffer
	deps := removeDeps{
		cfg:        config.DefaultConfig(),
		registry:   reg,
		store:      store,
		out:        &buf,
		dryRun:     true,
		autoYes:    true,
		installDir: installDir,
	}

	err := doRemove(deps, []string{"kubectl"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[dry-run]") {
		t.Errorf("expected '[dry-run]' message, got:\n%s", output)
	}

	// Binary should still exist
	if _, err := os.Stat(binPath); err != nil {
		t.Error("dry-run should not remove binary")
	}

	// State should still have the tool
	if store.GetVersion("kubectl") != "1.29.0" {
		t.Error("dry-run should not modify state")
	}
}

func TestDoRemove_RemovesBinaryAndState(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{Name: "kubectl", Version: "1.29.0", Group: tooldef.GroupKubernetes},
	)
	store := newFakeStore()
	store.tools["kubectl"] = "1.29.0"

	// Create a fake binary
	installDir := t.TempDir()
	binPath := filepath.Join(installDir, "kubectl")
	if err := os.WriteFile(binPath, []byte("fake"), 0o755); err != nil {
		t.Fatalf("failed to create fake binary: %v", err)
	}

	var buf bytes.Buffer
	deps := removeDeps{
		cfg:        config.DefaultConfig(),
		registry:   reg,
		store:      store,
		out:        &buf,
		dryRun:     false,
		autoYes:    true,
		installDir: installDir,
	}

	err := doRemove(deps, []string{"kubectl"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Binary should be removed
	if _, err := os.Stat(binPath); !os.IsNotExist(err) {
		t.Error("expected binary to be removed")
	}

	// State should no longer have the tool
	if store.GetVersion("kubectl") != "" {
		t.Error("expected tool removed from state")
	}

	output := buf.String()
	if !strings.Contains(output, "kubectl removed") {
		t.Errorf("expected removal confirmation, got:\n%s", output)
	}
}

func TestDoRemove_StateOnlyCleanup(t *testing.T) {
	// Tool is in state store but binary is already gone
	reg := newFakeRegistry(
		&tooldef.Tool{Name: "kubectl", Version: "1.29.0", Group: tooldef.GroupKubernetes},
	)
	store := newFakeStore()
	store.tools["kubectl"] = "1.28.0"

	installDir := t.TempDir()
	// No binary created

	var buf bytes.Buffer
	deps := removeDeps{
		cfg:        config.DefaultConfig(),
		registry:   reg,
		store:      store,
		out:        &buf,
		dryRun:     false,
		autoYes:    true,
		installDir: installDir,
	}

	err := doRemove(deps, []string{"kubectl"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// State should be cleaned up
	if store.GetVersion("kubectl") != "" {
		t.Error("expected tool removed from state")
	}
}

func TestDoRemove_MultipleTools(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{Name: "kubectl", Version: "1.29.0", Group: tooldef.GroupKubernetes},
		&tooldef.Tool{Name: "helm", Version: "3.14.0", Group: tooldef.GroupKubernetes},
	)
	store := newFakeStore()
	store.tools["kubectl"] = "1.29.0"
	store.tools["helm"] = "3.14.0"

	installDir := t.TempDir()
	for _, name := range []string{"kubectl", "helm"} {
		if err := os.WriteFile(filepath.Join(installDir, name), []byte("fake"), 0o755); err != nil {
			t.Fatalf("failed to create fake binary: %v", err)
		}
	}

	var buf bytes.Buffer
	deps := removeDeps{
		cfg:        config.DefaultConfig(),
		registry:   reg,
		store:      store,
		out:        &buf,
		dryRun:     false,
		autoYes:    true,
		installDir: installDir,
	}

	err := doRemove(deps, []string{"kubectl", "helm"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both should be removed
	if store.GetVersion("kubectl") != "" {
		t.Error("expected kubectl removed from state")
	}
	if store.GetVersion("helm") != "" {
		t.Error("expected helm removed from state")
	}

	output := buf.String()
	if !strings.Contains(output, "2 tool(s) removed") {
		t.Errorf("expected '2 tool(s) removed' summary, got:\n%s", output)
	}
}
