package cli

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// fakeRegistry implements ToolRegistry for testing.
type fakeRegistry struct {
	tools map[string]*tooldef.Tool
}

func newFakeRegistry(tools ...*tooldef.Tool) *fakeRegistry {
	r := &fakeRegistry{tools: make(map[string]*tooldef.Tool)}
	for _, t := range tools {
		r.tools[t.Name] = t
	}
	return r
}

func (r *fakeRegistry) All() []*tooldef.Tool {
	result := make([]*tooldef.Tool, 0, len(r.tools))
	for _, t := range r.tools {
		result = append(result, t)
	}
	return result
}

func (r *fakeRegistry) Get(name string) (*tooldef.Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

func (r *fakeRegistry) GetByGroup(group tooldef.Group) []*tooldef.Tool {
	var result []*tooldef.Tool
	for _, t := range r.tools {
		if t.Group == group {
			result = append(result, t)
		}
	}
	return result
}

func (r *fakeRegistry) Names() []string {
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// fakeInstaller implements ToolInstaller for testing.
type fakeInstaller struct {
	installed []*tooldef.Tool
	linked    map[string]string // toolName -> systemPath
	failTools map[string]error  // tool names that should fail
}

func newFakeInstaller() *fakeInstaller {
	return &fakeInstaller{
		linked:    make(map[string]string),
		failTools: make(map[string]error),
	}
}

func (fi *fakeInstaller) Install(ctx context.Context, tool *tooldef.Tool) error {
	if err, ok := fi.failTools[tool.Name]; ok {
		return err
	}
	fi.installed = append(fi.installed, tool)
	return nil
}

func (fi *fakeInstaller) InstallAll(ctx context.Context, tools []*tooldef.Tool) []error {
	var errs []error
	for _, t := range tools {
		if err := fi.Install(ctx, t); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (fi *fakeInstaller) IsInstalled(tool *tooldef.Tool) bool {
	for _, t := range fi.installed {
		if t.Name == tool.Name {
			return true
		}
	}
	return false
}

func (fi *fakeInstaller) EnsureDir() error {
	return nil
}

func (fi *fakeInstaller) Link(tool *tooldef.Tool, systemPath string) error {
	fi.linked[tool.Name] = systemPath
	return nil
}

// fakeStore implements StateStore for testing.
type fakeStore struct {
	tools   map[string]string
	removed []string
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		tools: make(map[string]string),
	}
}

func (s *fakeStore) GetVersion(name string) string {
	return s.tools[name]
}

func (s *fakeStore) Record(name, version string) error {
	s.tools[name] = version
	return nil
}

func (s *fakeStore) Remove(name string) error {
	delete(s.tools, name)
	s.removed = append(s.removed, name)
	return nil
}

func (s *fakeStore) Save() error {
	return nil
}

// --- Install tests ---

func TestDoInstall_NoToolsMatchingFilter(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{Name: "kubectl", Version: "1.29.0", Group: tooldef.GroupKubernetes},
	)
	inst := newFakeInstaller()

	cfg := config.DefaultConfig()
	cfg.InstallDir = "/tmp/bin"

	var buf bytes.Buffer
	deps := installDeps{
		cfg:       cfg,
		registry:  reg,
		installer: inst,
		out:       &buf,
		dryRun:    false,
		autoYes:   true,
		only:      "nonexistent-group",
	}

	err := doInstall(deps, tooldef.Platform{OS: "darwin", Arch: "arm64"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "No tools to install") {
		t.Errorf("expected 'No tools to install' message, got:\n%s", buf.String())
	}

	if len(inst.installed) != 0 {
		t.Errorf("expected no tools installed, got %d", len(inst.installed))
	}
}

func TestDoInstall_FiltersByGroup(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{Name: "kubectl", Version: "1.29.0", Group: tooldef.GroupKubernetes},
		&tooldef.Tool{Name: "terraform", Version: "1.8.0", Group: tooldef.GroupInfra},
		&tooldef.Tool{Name: "bat", Version: "0.24.0", Group: tooldef.GroupRustTools},
	)
	inst := newFakeInstaller()

	cfg := config.DefaultConfig()
	cfg.InstallDir = "/tmp/bin"

	var buf bytes.Buffer
	deps := installDeps{
		cfg:       cfg,
		registry:  reg,
		installer: inst,
		out:       &buf,
		dryRun:    false,
		autoYes:   true,
		only:      "kubernetes",
	}

	err := doInstall(deps, tooldef.Platform{OS: "darwin", Arch: "arm64"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(inst.installed) != 1 {
		t.Fatalf("expected 1 tool installed, got %d", len(inst.installed))
	}
	if inst.installed[0].Name != "kubectl" {
		t.Errorf("expected kubectl installed, got %s", inst.installed[0].Name)
	}
}

func TestDoInstall_FiltersDisabledGroups(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{Name: "kubectl", Version: "1.29.0", Group: tooldef.GroupKubernetes},
		&tooldef.Tool{Name: "terraform", Version: "1.8.0", Group: tooldef.GroupInfra},
	)
	inst := newFakeInstaller()

	cfg := config.DefaultConfig()
	cfg.InstallDir = "/tmp/bin"
	cfg.Groups.Kubernetes = false // disable k8s

	var buf bytes.Buffer
	deps := installDeps{
		cfg:       cfg,
		registry:  reg,
		installer: inst,
		out:       &buf,
		dryRun:    false,
		autoYes:   true,
		only:      "",
	}

	err := doInstall(deps, tooldef.Platform{OS: "darwin", Arch: "arm64"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only terraform should be installed (k8s disabled)
	if len(inst.installed) != 1 {
		t.Fatalf("expected 1 tool installed, got %d", len(inst.installed))
	}
	if inst.installed[0].Name != "terraform" {
		t.Errorf("expected terraform installed, got %s", inst.installed[0].Name)
	}
}

func TestDoInstall_FiltersByPlatform(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{
			Name:    "linux-only",
			Version: "1.0.0",
			Group:   tooldef.GroupUtilities,
			Platforms: []tooldef.Platform{
				{OS: "linux", Arch: "amd64"},
			},
		},
		&tooldef.Tool{
			Name:    "universal",
			Version: "2.0.0",
			Group:   tooldef.GroupUtilities,
		},
	)
	inst := newFakeInstaller()

	cfg := config.DefaultConfig()
	cfg.InstallDir = "/tmp/bin"

	var buf bytes.Buffer
	deps := installDeps{
		cfg:       cfg,
		registry:  reg,
		installer: inst,
		out:       &buf,
		dryRun:    false,
		autoYes:   true,
		only:      "",
	}

	// Running on darwin/arm64 — linux-only tool should be skipped
	err := doInstall(deps, tooldef.Platform{OS: "darwin", Arch: "arm64"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(inst.installed) != 1 {
		t.Fatalf("expected 1 tool installed, got %d", len(inst.installed))
	}
	if inst.installed[0].Name != "universal" {
		t.Errorf("expected 'universal' installed, got %s", inst.installed[0].Name)
	}
}

func TestDoInstall_DisabledToolOverride(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{Name: "kubectl", Version: "1.29.0", Group: tooldef.GroupKubernetes},
		&tooldef.Tool{Name: "helm", Version: "3.14.0", Group: tooldef.GroupKubernetes},
	)
	inst := newFakeInstaller()

	cfg := config.DefaultConfig()
	cfg.InstallDir = "/tmp/bin"
	cfg.Overrides = map[string]config.ToolOverride{
		"helm": {Disabled: true},
	}

	var buf bytes.Buffer
	deps := installDeps{
		cfg:       cfg,
		registry:  reg,
		installer: inst,
		out:       &buf,
		dryRun:    false,
		autoYes:   true,
		only:      "",
	}

	err := doInstall(deps, tooldef.Platform{OS: "darwin", Arch: "arm64"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only kubectl should be installed (helm is disabled)
	if len(inst.installed) != 1 {
		t.Fatalf("expected 1 tool installed, got %d", len(inst.installed))
	}
	if inst.installed[0].Name != "kubectl" {
		t.Errorf("expected kubectl, got %s", inst.installed[0].Name)
	}
}

func TestDoInstall_VersionOverride(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{Name: "terraform", Version: "1.8.0", Group: tooldef.GroupInfra},
	)
	inst := newFakeInstaller()

	cfg := config.DefaultConfig()
	cfg.InstallDir = "/tmp/bin"
	cfg.Overrides = map[string]config.ToolOverride{
		"terraform": {Version: "1.7.5"},
	}

	var buf bytes.Buffer
	deps := installDeps{
		cfg:       cfg,
		registry:  reg,
		installer: inst,
		out:       &buf,
		dryRun:    false,
		autoYes:   true,
		only:      "",
	}

	err := doInstall(deps, tooldef.Platform{OS: "darwin", Arch: "arm64"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(inst.installed) != 1 {
		t.Fatalf("expected 1 tool installed, got %d", len(inst.installed))
	}
	if inst.installed[0].Version != "1.7.5" {
		t.Errorf("expected version override 1.7.5, got %s", inst.installed[0].Version)
	}
}

func TestDoInstall_DryRunSkipsInstallation(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{Name: "kubectl", Version: "1.29.0", Group: tooldef.GroupKubernetes},
	)
	inst := newFakeInstaller()

	cfg := config.DefaultConfig()
	cfg.InstallDir = "/tmp/bin"

	var buf bytes.Buffer
	deps := installDeps{
		cfg:       cfg,
		registry:  reg,
		installer: inst,
		out:       &buf,
		dryRun:    true,
		autoYes:   true,
		only:      "",
	}

	err := doInstall(deps, tooldef.Platform{OS: "darwin", Arch: "arm64"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// In dry-run, InstallAll is still called (the installer itself handles dry-run)
	// but the confirmation prompt is skipped
	if !strings.Contains(buf.String(), "kubectl") {
		t.Errorf("expected output to mention kubectl, got:\n%s", buf.String())
	}
}

func TestDoInstall_InstallerFailureReportsError(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{Name: "kubectl", Version: "1.29.0", Group: tooldef.GroupKubernetes},
		&tooldef.Tool{Name: "helm", Version: "3.14.0", Group: tooldef.GroupKubernetes},
	)
	inst := newFakeInstaller()
	inst.failTools["helm"] = fmt.Errorf("download failed: 404")

	cfg := config.DefaultConfig()
	cfg.InstallDir = "/tmp/bin"

	var buf bytes.Buffer
	deps := installDeps{
		cfg:       cfg,
		registry:  reg,
		installer: inst,
		out:       &buf,
		dryRun:    false,
		autoYes:   true,
		only:      "",
	}

	err := doInstall(deps, tooldef.Platform{OS: "darwin", Arch: "arm64"})
	if err == nil {
		t.Fatal("expected error when installer fails")
	}
	if !strings.Contains(err.Error(), "1 installations failed") {
		t.Errorf("expected '1 installations failed', got: %v", err)
	}

	// kubectl should succeed
	if len(inst.installed) != 1 {
		t.Fatalf("expected 1 successful install, got %d", len(inst.installed))
	}
	if inst.installed[0].Name != "kubectl" {
		t.Errorf("expected kubectl to succeed, got %s", inst.installed[0].Name)
	}
}

func TestDoInstall_ShowsManagedByAnnotation(t *testing.T) {
	reg := newFakeRegistry(
		&tooldef.Tool{Name: "python", Version: "3.12", Group: tooldef.GroupLanguages, ManagedBy: "mise"},
	)
	inst := newFakeInstaller()

	cfg := config.DefaultConfig()
	cfg.InstallDir = "/tmp/bin"

	var buf bytes.Buffer
	deps := installDeps{
		cfg:       cfg,
		registry:  reg,
		installer: inst,
		out:       &buf,
		dryRun:    true,
		autoYes:   true,
		only:      "",
	}

	_ = doInstall(deps, tooldef.Platform{OS: "darwin", Arch: "arm64"})

	if !strings.Contains(buf.String(), "via mise") {
		t.Errorf("expected 'via mise' annotation in output, got:\n%s", buf.String())
	}
}
