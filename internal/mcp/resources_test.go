package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/internal/platform"
	"github.com/omargallob/devops-starter/internal/state"
)

// TestConfigResourceHandler verifies the config resource returns valid YAML content.
func TestConfigResourceHandler(t *testing.T) {
	deps := testDeps(t)
	handler := configResourceHandler(deps)

	contents, err := handler(context.Background(), mcp.ReadResourceRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(contents) != 1 {
		t.Fatalf("expected 1 resource content, got %d", len(contents))
	}

	tc, ok := contents[0].(mcp.TextResourceContents)
	if !ok {
		t.Fatalf("expected TextResourceContents, got %T", contents[0])
	}
	if tc.URI != "devops-starter://config" {
		t.Errorf("expected URI devops-starter://config, got %s", tc.URI)
	}
	if tc.MIMEType != "text/yaml" {
		t.Errorf("expected MIME text/yaml, got %s", tc.MIMEType)
	}
	if tc.Text == "" {
		t.Error("expected non-empty config text")
	}
	// Should contain install_dir since we marshal the in-memory config.
	if !containsString(tc.Text, "install_dir") {
		t.Error("expected config text to contain 'install_dir'")
	}
}

// TestStateResourceHandler_NoFile verifies that when no state file exists,
// an empty JSON state is returned.
func TestStateResourceHandler_NoFile(t *testing.T) {
	// Override XDG to a temp dir so state.Path() points to a nonexistent file.
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	handler := stateResourceHandler()

	contents, err := handler(context.Background(), mcp.ReadResourceRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(contents) != 1 {
		t.Fatalf("expected 1 resource content, got %d", len(contents))
	}

	tc, ok := contents[0].(mcp.TextResourceContents)
	if !ok {
		t.Fatalf("expected TextResourceContents, got %T", contents[0])
	}
	if tc.URI != "devops-starter://state" {
		t.Errorf("expected URI devops-starter://state, got %s", tc.URI)
	}
	if tc.MIMEType != "application/json" {
		t.Errorf("expected MIME application/json, got %s", tc.MIMEType)
	}
	if tc.Text != `{"tools":{}}` {
		t.Errorf("expected empty state JSON, got %s", tc.Text)
	}
}

// TestStateResourceHandler_WithFile verifies that an existing state file is read.
func TestStateResourceHandler_WithFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Create a state file.
	stateDir := filepath.Join(tmpDir, "devops-starter")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	stateContent := `{"tools":{"kubectl":{"version":"1.31.0","installed_at":"2025-01-01T00:00:00Z"}}}`
	if err := os.WriteFile(filepath.Join(stateDir, "state.json"), []byte(stateContent), 0o644); err != nil {
		t.Fatal(err)
	}

	handler := stateResourceHandler()

	contents, err := handler(context.Background(), mcp.ReadResourceRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tc := contents[0].(mcp.TextResourceContents)
	if tc.Text != stateContent {
		t.Errorf("expected state content %q, got %q", stateContent, tc.Text)
	}
}

// TestGroupsResourceHandler verifies the groups resource returns all groups
// with correct enabled/disabled status.
func TestGroupsResourceHandler(t *testing.T) {
	deps := testDeps(t)
	handler := groupsResourceHandler(deps)

	contents, err := handler(context.Background(), mcp.ReadResourceRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(contents) != 1 {
		t.Fatalf("expected 1 resource content, got %d", len(contents))
	}

	tc, ok := contents[0].(mcp.TextResourceContents)
	if !ok {
		t.Fatalf("expected TextResourceContents, got %T", contents[0])
	}
	if tc.URI != "devops-starter://groups" {
		t.Errorf("expected URI devops-starter://groups, got %s", tc.URI)
	}
	if tc.MIMEType != "application/json" {
		t.Errorf("expected MIME application/json, got %s", tc.MIMEType)
	}

	type groupInfo struct {
		Name    string `json:"name"`
		Enabled bool   `json:"enabled"`
	}
	var groups []groupInfo
	if err := json.Unmarshal([]byte(tc.Text), &groups); err != nil {
		t.Fatalf("expected valid JSON array, got: %s", tc.Text)
	}

	// Should have 10 groups (all defined groups).
	if len(groups) != 10 {
		t.Errorf("expected 10 groups, got %d", len(groups))
	}

	// Verify enabled/disabled matches test config.
	groupMap := make(map[string]bool)
	for _, g := range groups {
		groupMap[g.Name] = g.Enabled
	}

	// testDeps enables these groups:
	expectedEnabled := map[string]bool{
		"languages":        true,
		"containers":       true,
		"kubernetes":       true,
		"infra":            true,
		"cloud":            true,
		"ansible":          false,
		"rust-tools":       true,
		"utilities":        true,
		"ai":               true,
		"package-managers": false,
	}
	for name, expected := range expectedEnabled {
		if got, ok := groupMap[name]; !ok {
			t.Errorf("group %q not found in response", name)
		} else if got != expected {
			t.Errorf("group %q: expected enabled=%v, got %v", name, expected, got)
		}
	}
}

// TestGroupsResourceHandler_AllDisabled verifies groups reflect config changes.
func TestGroupsResourceHandler_AllDisabled(t *testing.T) {
	cfg := &config.Config{
		InstallDir: t.TempDir(),
		Groups:     config.GroupConfig{}, // all false
	}
	info, err := platform.Detect()
	if err != nil {
		t.Fatal(err)
	}
	deps := &Deps{
		Config:   cfg,
		Registry: nil, // not needed for this handler
		Store:    &state.Store{Tools: make(map[string]state.InstalledTool)},
		Platform: info,
	}

	handler := groupsResourceHandler(deps)
	contents, err := handler(context.Background(), mcp.ReadResourceRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tc := contents[0].(mcp.TextResourceContents)
	type groupInfo struct {
		Name    string `json:"name"`
		Enabled bool   `json:"enabled"`
	}
	var groups []groupInfo
	if err := json.Unmarshal([]byte(tc.Text), &groups); err != nil {
		t.Fatal(err)
	}

	for _, g := range groups {
		if g.Enabled {
			t.Errorf("expected group %q to be disabled", g.Name)
		}
	}
}

// helper
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
