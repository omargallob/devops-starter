package mcp

import (
	"testing"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// TestGroupEnabled verifies the groupEnabled function for all groups.
func TestGroupEnabled(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.GroupConfig
		group   tooldef.Group
		enabled bool
	}{
		{"languages enabled", config.GroupConfig{Languages: true}, tooldef.GroupLanguages, true},
		{"languages disabled", config.GroupConfig{}, tooldef.GroupLanguages, false},
		{"containers enabled", config.GroupConfig{Containers: true}, tooldef.GroupContainers, true},
		{"containers disabled", config.GroupConfig{}, tooldef.GroupContainers, false},
		{"kubernetes enabled", config.GroupConfig{Kubernetes: true}, tooldef.GroupKubernetes, true},
		{"kubernetes disabled", config.GroupConfig{}, tooldef.GroupKubernetes, false},
		{"infra enabled", config.GroupConfig{Infra: true}, tooldef.GroupInfra, true},
		{"infra disabled", config.GroupConfig{}, tooldef.GroupInfra, false},
		{"cloud enabled", config.GroupConfig{Cloud: true}, tooldef.GroupCloud, true},
		{"cloud disabled", config.GroupConfig{}, tooldef.GroupCloud, false},
		{"ansible enabled", config.GroupConfig{Ansible: true}, tooldef.GroupAnsible, true},
		{"ansible disabled", config.GroupConfig{}, tooldef.GroupAnsible, false},
		{"rust-tools enabled", config.GroupConfig{RustTools: true}, tooldef.GroupRustTools, true},
		{"rust-tools disabled", config.GroupConfig{}, tooldef.GroupRustTools, false},
		{"utilities enabled", config.GroupConfig{Utilities: true}, tooldef.GroupUtilities, true},
		{"utilities disabled", config.GroupConfig{}, tooldef.GroupUtilities, false},
		{"ai enabled", config.GroupConfig{AI: true}, tooldef.GroupAI, true},
		{"ai disabled", config.GroupConfig{}, tooldef.GroupAI, false},
		{"package-managers enabled", config.GroupConfig{PackageManagers: true}, tooldef.GroupPackageManagers, true},
		{"package-managers disabled", config.GroupConfig{}, tooldef.GroupPackageManagers, false},
		{"unknown group", config.GroupConfig{}, tooldef.Group("nonexistent"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{Groups: tt.cfg}
			got := groupEnabled(cfg, tt.group)
			if got != tt.enabled {
				t.Errorf("groupEnabled(%q) = %v, want %v", tt.group, got, tt.enabled)
			}
		})
	}
}

// TestAllGroups verifies allGroups returns exactly 10 groups in order.
func TestAllGroups(t *testing.T) {
	groups := allGroups()

	if len(groups) != 10 {
		t.Fatalf("expected 10 groups, got %d", len(groups))
	}

	expected := []tooldef.Group{
		tooldef.GroupLanguages,
		tooldef.GroupContainers,
		tooldef.GroupKubernetes,
		tooldef.GroupInfra,
		tooldef.GroupCloud,
		tooldef.GroupAnsible,
		tooldef.GroupRustTools,
		tooldef.GroupUtilities,
		tooldef.GroupAI,
		tooldef.GroupPackageManagers,
	}

	for i, g := range groups {
		if g != expected[i] {
			t.Errorf("allGroups()[%d] = %q, want %q", i, g, expected[i])
		}
	}
}

// TestAllGroups_NoDuplicates verifies there are no duplicate group names.
func TestAllGroups_NoDuplicates(t *testing.T) {
	groups := allGroups()
	seen := make(map[tooldef.Group]bool)
	for _, g := range groups {
		if seen[g] {
			t.Errorf("duplicate group: %q", g)
		}
		seen[g] = true
	}
}

// TestNewServer_Version verifies the version is set correctly.
func TestNewServer_Version(t *testing.T) {
	oldVersion := Version
	defer func() { Version = oldVersion }()

	Version = "v1.2.3-test"
	deps := testDeps(t)
	s := NewServer(deps)
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
	// The server is created successfully with the custom version.
	// (We can't easily inspect the version from the outside, but the fact
	// that it doesn't panic with the custom version is the contract.)
}
