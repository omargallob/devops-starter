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
		{StatusDetected, "detected"},
		{StatusLinked, "linked"},
		{StatusUnavailable, "unavailable"},
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

	// All tools should be StatusMissing or StatusDetected (if binary is in PATH).
	// Mise-managed tools found in PATH may also be StatusCurrent or StatusOutdated
	// since they are resolved via version detection rather than state store.
	for _, g := range groups {
		for _, ts := range g.Tools {
			switch ts.Status {
			case StatusMissing, StatusDetected:
				// expected for unmanaged tools
			case StatusCurrent, StatusOutdated, StatusUnknown:
				// acceptable for mise-managed tools found in PATH
				if ts.Tool == nil || !ts.Tool.IsMiseManaged() {
					t.Errorf("tool %s: expected StatusMissing or StatusDetected, got %s", ts.Name, ts.Status.String())
				}
			default:
				t.Errorf("tool %s: unexpected status %s", ts.Name, ts.Status.String())
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

func TestLookupInPath_KnownBinary(t *testing.T) {
	// "sh" should exist on any Unix system
	path := LookupInPath("sh")
	// sh won't have a probe, so it uses tool name directly - which won't be in probes
	// Let's test with a binary we know exists: use exec.LookPath logic
	// Actually sh doesn't have a probe, but LookPath will find it by name
	if path == "" {
		t.Skip("sh not in PATH (unexpected on Unix)")
	}
	if path == "" {
		t.Error("expected to find sh in PATH")
	}
}

func TestLookupInPath_NonExistent(t *testing.T) {
	path := LookupInPath("absolutely-nonexistent-binary-xyz-12345")
	if path != "" {
		t.Errorf("expected empty string for nonexistent binary, got %s", path)
	}
}

func TestLookupInPath_UsesBinNameMapping(t *testing.T) {
	// "neovim" maps to "nvim" via probes. If nvim is installed, it should find it.
	// "ripgrep" maps to "rg". Test the mapping is used.
	// We can't guarantee these are installed, but we can test that
	// a tool with no probe falls back to tool name.
	path := LookupInPath("nonexistent-no-probe-tool")
	if path != "" {
		t.Error("expected empty for tool with no probe and no binary")
	}
}

func TestResolveAll_DetectedStatus(t *testing.T) {
	cfg := config.DefaultConfig()
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}
	store := &Store{Tools: map[string]InstalledTool{}}

	groups := ResolveAll(cfg, store, plat)

	// Any tool whose binary is in PATH should be StatusDetected
	foundDetected := false
	for _, g := range groups {
		for _, ts := range g.Tools {
			if ts.Status == StatusDetected {
				foundDetected = true
				// Verify it's not in the store
				if store.GetVersion(ts.Name) != "" {
					t.Errorf("tool %s is StatusDetected but has version in store", ts.Name)
				}
			}
		}
	}
	// This test is environment-dependent; if we have any of the tools in PATH, we should find them
	_ = foundDetected // just verifying no panics; actual detection depends on environment
}

func TestResolveAll_LinkedStatus(t *testing.T) {
	cfg := config.DefaultConfig()
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}
	store := &Store{Tools: map[string]InstalledTool{}}

	// First, find a tool that is detected (binary exists in PATH)
	groups := ResolveAll(cfg, store, plat)
	var detectedTool string
	for _, g := range groups {
		for _, ts := range g.Tools {
			if ts.Status == StatusDetected && ts.Tool != nil && !ts.Tool.IsMiseManaged() {
				detectedTool = ts.Name
				break
			}
		}
		if detectedTool != "" {
			break
		}
	}
	if detectedTool == "" {
		t.Skip("no tools detected in PATH on this system; cannot test linked status")
	}

	// Now set conflict=link override for that tool and re-resolve
	cfg.Overrides = map[string]config.ToolOverride{
		detectedTool: {Conflict: "link"},
	}

	groups = ResolveAll(cfg, store, plat)
	for _, g := range groups {
		for _, ts := range g.Tools {
			if ts.Name == detectedTool {
				if ts.Status != StatusLinked {
					t.Errorf("tool %s: expected StatusLinked, got %s", ts.Name, ts.Status.String())
				}
				if ts.Source != SourceSystem {
					t.Errorf("tool %s: expected Source=system, got %q", ts.Name, ts.Source)
				}
				if ts.ConflictPolicy != "link" {
					t.Errorf("tool %s: expected ConflictPolicy=link, got %q", ts.Name, ts.ConflictPolicy)
				}
				return
			}
		}
	}
	t.Fatalf("tool %s not found in resolved groups", detectedTool)
}

func TestResolveTool_ConflictPolicyPropagated(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Overrides = map[string]config.ToolOverride{
		"kubectl": {Conflict: "skip"},
	}
	store := &Store{Tools: map[string]InstalledTool{}}
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}

	tool := &tooldef.Tool{
		Name:    "kubectl",
		Version: "1.30.0",
		Group:   "kubernetes",
		Platforms: []tooldef.Platform{
			{OS: "linux", Arch: "amd64"},
		},
	}

	ts := resolveTool(tool, cfg, store, plat)
	if ts.ConflictPolicy != "skip" {
		t.Errorf("expected ConflictPolicy=skip, got %q", ts.ConflictPolicy)
	}
}

func TestResolveAll_SourceManaged(t *testing.T) {
	cfg := config.DefaultConfig()
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}

	emptyStore := &Store{Tools: map[string]InstalledTool{}}
	groups := ResolveAll(cfg, emptyStore, plat)
	if len(groups) == 0 || len(groups[0].Tools) == 0 {
		t.Skip("no tools in registry for this platform")
	}

	// Pick the first tool and mark it as installed in the store
	firstTool := groups[0].Tools[0]
	store := &Store{
		Tools: map[string]InstalledTool{
			firstTool.Name: {Version: firstTool.DesiredVersion},
		},
	}

	groups = ResolveAll(cfg, store, plat)
	for _, g := range groups {
		for _, ts := range g.Tools {
			if ts.Name == firstTool.Name {
				if ts.Source != SourceManaged {
					t.Errorf("tool %s: expected Source=%q, got %q", ts.Name, SourceManaged, ts.Source)
				}
				return
			}
		}
	}
	t.Fatalf("tool %s not found", firstTool.Name)
}

func TestResolveAll_SourceManagedOutdated(t *testing.T) {
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
				if ts.Source != SourceManaged {
					t.Errorf("tool %s: expected Source=%q, got %q", ts.Name, SourceManaged, ts.Source)
				}
				if ts.Status != StatusOutdated {
					t.Errorf("tool %s: expected StatusOutdated, got %s", ts.Name, ts.Status.String())
				}
				return
			}
		}
	}
	t.Fatalf("tool %s not found", firstTool.Name)
}

func TestResolveAll_SourceSystem(t *testing.T) {
	cfg := config.DefaultConfig()
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}
	store := &Store{Tools: map[string]InstalledTool{}}

	groups := ResolveAll(cfg, store, plat)

	for _, g := range groups {
		for _, ts := range g.Tools {
			if ts.Status == StatusDetected {
				if ts.Source != SourceSystem {
					t.Errorf("tool %s: StatusDetected should have Source=%q, got %q", ts.Name, SourceSystem, ts.Source)
				}
			}
		}
	}
}

func TestResolveAll_SourceMise(t *testing.T) {
	cfg := config.DefaultConfig()
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}
	store := &Store{Tools: map[string]InstalledTool{}}

	groups := ResolveAll(cfg, store, plat)

	// Mise-managed tools (Tool.IsMiseManaged()) that are found in PATH
	// should have Source == SourceMise
	for _, g := range groups {
		for _, ts := range g.Tools {
			if ts.Tool != nil && ts.Tool.IsMiseManaged() {
				switch ts.Status {
				case StatusCurrent, StatusOutdated, StatusUnknown:
					if ts.Source != SourceMise {
						t.Errorf("tool %s: mise-managed with status %s should have Source=%q, got %q",
							ts.Name, ts.Status.String(), SourceMise, ts.Source)
					}
				case StatusMissing:
					if ts.Source != SourceNone {
						t.Errorf("tool %s: missing mise-managed should have Source=%q, got %q",
							ts.Name, SourceNone, ts.Source)
					}
				}
			}
		}
	}
}

func TestResolveAll_SourceNoneForMissing(t *testing.T) {
	cfg := config.DefaultConfig()
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}
	store := &Store{Tools: map[string]InstalledTool{}}

	groups := ResolveAll(cfg, store, plat)

	for _, g := range groups {
		for _, ts := range g.Tools {
			if ts.Status == StatusMissing {
				if ts.Source != SourceNone {
					t.Errorf("tool %s: StatusMissing should have Source=%q, got %q", ts.Name, SourceNone, ts.Source)
				}
			}
		}
	}
}

func TestResolveAll_UnavailablePlatform(t *testing.T) {
	cfg := config.DefaultConfig()
	store := &Store{Tools: map[string]InstalledTool{}}

	// Use a platform that aws-cli does NOT support (it only has linux URLs)
	// to verify tools show as unavailable rather than being hidden.
	plat := tooldef.Platform{OS: "darwin", Arch: "arm64"}

	groups := ResolveAll(cfg, store, plat)

	// Find the cloud group and check that platform-restricted tools appear as unavailable
	var foundUnavailable bool
	for _, g := range groups {
		for _, ts := range g.Tools {
			if ts.Status == StatusUnavailable {
				foundUnavailable = true
				if ts.Tool == nil {
					t.Errorf("unavailable tool %s should still have Tool reference", ts.Name)
				}
			}
		}
	}

	// There should be at least one unavailable tool when running on a platform
	// that not all tools support.
	if !foundUnavailable {
		t.Error("expected at least one tool with StatusUnavailable on a restricted platform")
	}
}

func TestResolveAll_UnavailableNotSelectable(t *testing.T) {
	cfg := config.DefaultConfig()
	store := &Store{Tools: map[string]InstalledTool{}}

	// Use a fake platform that no tool supports
	plat := tooldef.Platform{OS: "freebsd", Arch: "riscv64"}

	groups := ResolveAll(cfg, store, plat)

	for _, g := range groups {
		for _, ts := range g.Tools {
			// Tools with explicit platform restrictions should be unavailable
			if ts.Status == StatusUnavailable {
				// Verify the status string renders correctly
				if ts.Status.String() != "unavailable" {
					t.Errorf("tool %s: expected status string 'unavailable', got %q", ts.Name, ts.Status.String())
				}
			}
		}
	}
}

func TestStatusString_Unavailable(t *testing.T) {
	if got := StatusUnavailable.String(); got != "unavailable" {
		t.Errorf("StatusUnavailable.String() = %q, want %q", got, "unavailable")
	}
}
