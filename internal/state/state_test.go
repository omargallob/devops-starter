package state

import (
	"testing"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

func TestStatus_String(t *testing.T) {
	tests := []struct {
		s    Status
		want string
	}{
		{StatusMissing, "missing"},
		{StatusCurrent, "current"},
		{StatusOutdated, "outdated"},
		{StatusDisabled, "disabled"},
		{StatusUnknown, "unknown"},
		{Status(99), "unknown"},
	}
	for _, tc := range tests {
		if got := tc.s.String(); got != tc.want {
			t.Errorf("Status(%d).String() = %q, want %q", tc.s, got, tc.want)
		}
	}
}

func TestResolveAll_BasicFlow(t *testing.T) {
	cfg := config.DefaultConfig()
	store := &Store{
		Tools: map[string]InstalledTool{},
	}
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}

	groups := ResolveAll(cfg, store, plat)

	// Should return at least a few groups
	if len(groups) == 0 {
		t.Fatal("expected at least one group")
	}

	// All tools should be StatusMissing since store is empty
	for _, g := range groups {
		for _, ts := range g.Tools {
			if ts.Status != StatusMissing {
				t.Errorf("tool %s: expected StatusMissing, got %s", ts.Name, ts.Status.String())
			}
		}
	}
}

func TestResolveAll_CurrentStatus(t *testing.T) {
	cfg := config.DefaultConfig()

	// We'll use the actual registry to find a tool name and version
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}

	// First run to get the expected tools
	emptyStore := &Store{Tools: map[string]InstalledTool{}}
	groups := ResolveAll(cfg, emptyStore, plat)
	if len(groups) == 0 || len(groups[0].Tools) == 0 {
		t.Skip("no tools in registry for this platform")
	}

	// Pick the first tool and mark it as current
	firstTool := groups[0].Tools[0]
	store := &Store{
		Tools: map[string]InstalledTool{
			firstTool.Name: {Version: firstTool.DesiredVersion},
		},
	}

	groups = ResolveAll(cfg, store, plat)
	found := false
	for _, g := range groups {
		for _, ts := range g.Tools {
			if ts.Name == firstTool.Name {
				found = true
				if ts.Status != StatusCurrent {
					t.Errorf("tool %s: expected StatusCurrent, got %s", ts.Name, ts.Status.String())
				}
			}
		}
	}
	if !found {
		t.Fatalf("tool %s not found in resolved groups", firstTool.Name)
	}
}

func TestResolveAll_OutdatedStatus(t *testing.T) {
	cfg := config.DefaultConfig()
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}

	emptyStore := &Store{Tools: map[string]InstalledTool{}}
	groups := ResolveAll(cfg, emptyStore, plat)
	if len(groups) == 0 || len(groups[0].Tools) == 0 {
		t.Skip("no tools")
	}

	firstTool := groups[0].Tools[0]
	store := &Store{
		Tools: map[string]InstalledTool{
			firstTool.Name: {Version: "0.0.1"},
		},
	}

	groups = ResolveAll(cfg, store, plat)
	for _, g := range groups {
		for _, ts := range g.Tools {
			if ts.Name == firstTool.Name {
				if ts.Status != StatusOutdated {
					t.Errorf("tool %s: expected StatusOutdated, got %s", ts.Name, ts.Status.String())
				}
				return
			}
		}
	}
	t.Fatalf("tool %s not found", firstTool.Name)
}

func TestResolveAll_UnknownStatus(t *testing.T) {
	cfg := config.DefaultConfig()
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}

	emptyStore := &Store{Tools: map[string]InstalledTool{}}
	groups := ResolveAll(cfg, emptyStore, plat)
	if len(groups) == 0 || len(groups[0].Tools) == 0 {
		t.Skip("no tools")
	}

	firstTool := groups[0].Tools[0]
	store := &Store{
		Tools: map[string]InstalledTool{
			firstTool.Name: {Version: "unknown"},
		},
	}

	groups = ResolveAll(cfg, store, plat)
	for _, g := range groups {
		for _, ts := range g.Tools {
			if ts.Name == firstTool.Name {
				if ts.Status != StatusUnknown {
					t.Errorf("tool %s: expected StatusUnknown, got %s", ts.Name, ts.Status.String())
				}
				return
			}
		}
	}
	t.Fatalf("tool %s not found", firstTool.Name)
}

func TestResolveAll_DisabledTool(t *testing.T) {
	cfg := config.DefaultConfig()
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}

	// First get a tool name
	emptyStore := &Store{Tools: map[string]InstalledTool{}}
	groups := ResolveAll(cfg, emptyStore, plat)
	if len(groups) == 0 || len(groups[0].Tools) == 0 {
		t.Skip("no tools")
	}

	toolName := groups[0].Tools[0].Name
	cfg.Overrides = map[string]config.ToolOverride{
		toolName: {Disabled: true},
	}

	groups = ResolveAll(cfg, emptyStore, plat)
	for _, g := range groups {
		for _, ts := range g.Tools {
			if ts.Name == toolName {
				if ts.Status != StatusDisabled {
					t.Errorf("tool %s: expected StatusDisabled, got %s", ts.Name, ts.Status.String())
				}
				return
			}
		}
	}
	t.Fatalf("tool %s not found", toolName)
}

func TestResolveAll_DisabledGroup(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Groups.Languages = false
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}

	store := &Store{Tools: map[string]InstalledTool{}}
	groups := ResolveAll(cfg, store, plat)

	// All tools in "languages" group should be disabled
	for _, g := range groups {
		if g.Name == "languages" {
			for _, ts := range g.Tools {
				if ts.Status != StatusDisabled {
					t.Errorf("tool %s in disabled group: expected StatusDisabled, got %s", ts.Name, ts.Status.String())
				}
			}
			return
		}
	}
}

func TestResolveAll_VersionOverride(t *testing.T) {
	cfg := config.DefaultConfig()
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}

	emptyStore := &Store{Tools: map[string]InstalledTool{}}
	groups := ResolveAll(cfg, emptyStore, plat)
	if len(groups) == 0 || len(groups[0].Tools) == 0 {
		t.Skip("no tools")
	}

	toolName := groups[0].Tools[0].Name
	cfg.Overrides = map[string]config.ToolOverride{
		toolName: {Version: "99.99.99"},
	}

	groups = ResolveAll(cfg, emptyStore, plat)
	for _, g := range groups {
		for _, ts := range g.Tools {
			if ts.Name == toolName {
				if ts.DesiredVersion != "99.99.99" {
					t.Errorf("tool %s: expected desired version 99.99.99, got %s", ts.Name, ts.DesiredVersion)
				}
				return
			}
		}
	}
	t.Fatalf("tool %s not found", toolName)
}

func TestDetectVersion_NoProbe(t *testing.T) {
	_, err := DetectVersion("nonexistent-tool", t.TempDir())
	if err == nil {
		t.Fatal("expected error for tool with no probe")
	}
}

func TestDetectVersion_BinaryNotFound(t *testing.T) {
	_, err := DetectVersion("kubectl", t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing binary")
	}
}

func TestVerifyAll_SkipsMissingBinaries(t *testing.T) {
	store := &Store{
		Tools: make(map[string]InstalledTool),
	}

	// VerifyAll should not panic or error with an empty dir
	VerifyAll(store, t.TempDir())

	// Nothing should be recorded since no binaries exist
	if len(store.Tools) != 0 {
		t.Errorf("expected 0 tools recorded, got %d", len(store.Tools))
	}
}

func TestVerifyOne_MissingBinary(t *testing.T) {
	store := &Store{
		Tools: make(map[string]InstalledTool),
	}

	_, err := VerifyOne(store, "kubectl", t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing binary")
	}
}
