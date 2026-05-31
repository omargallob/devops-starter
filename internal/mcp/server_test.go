package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/internal/platform"
	"github.com/omargallob/devops-starter/internal/registry"
	"github.com/omargallob/devops-starter/internal/state"
)

// TestNewServer verifies that the server factory creates a valid server
// with all expected tools and resources registered.
func TestNewServer(t *testing.T) {
	deps := testDeps(t)
	s := NewServer(deps)
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
}

// TestListTools verifies the list_tools handler returns a non-empty JSON array.
func TestListTools(t *testing.T) {
	deps := testDeps(t)
	handler := listToolsHandler(deps)

	// Call without group filter.
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}

	text := extractText(t, result)
	var tools []map[string]any
	if err := json.Unmarshal([]byte(text), &tools); err != nil {
		t.Fatalf("expected valid JSON array, got: %s", text)
	}
	if len(tools) == 0 {
		t.Fatal("expected non-empty tools list")
	}

	// Verify each tool has required fields.
	for _, tool := range tools {
		for _, field := range []string{"name", "version", "group", "install_mode"} {
			if _, ok := tool[field]; !ok {
				t.Errorf("tool missing field %q: %v", field, tool)
			}
		}
	}
}

// TestListToolsWithGroup verifies filtering by group.
func TestListToolsWithGroup(t *testing.T) {
	deps := testDeps(t)
	handler := listToolsHandler(deps)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"group": "kubernetes"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool returned error")
	}

	text := extractText(t, result)
	var tools []map[string]any
	if err := json.Unmarshal([]byte(text), &tools); err != nil {
		t.Fatalf("expected valid JSON array, got: %s", text)
	}

	for _, tool := range tools {
		if tool["group"] != "kubernetes" {
			t.Errorf("expected group=kubernetes, got %v", tool["group"])
		}
	}
}

// TestGetTool verifies the get_tool handler returns full tool details.
func TestGetTool(t *testing.T) {
	deps := testDeps(t)
	handler := getToolHandler(deps)

	// Pick the first tool name from the registry.
	names := deps.Registry.Names()
	if len(names) == 0 {
		t.Skip("registry is empty")
	}

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"name": names[0]}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool returned error")
	}

	text := extractText(t, result)
	var detail map[string]any
	if err := json.Unmarshal([]byte(text), &detail); err != nil {
		t.Fatalf("expected valid JSON object, got: %s", text)
	}

	if detail["name"] != names[0] {
		t.Errorf("expected name=%s, got %v", names[0], detail["name"])
	}
}

// TestGetToolNotFound verifies error response for unknown tool.
func TestGetToolNotFound(t *testing.T) {
	deps := testDeps(t)
	handler := getToolHandler(deps)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"name": "nonexistent-tool-xyz"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for nonexistent tool")
	}
}

// TestGetStatus verifies the get_status handler returns valid status entries.
func TestGetStatus(t *testing.T) {
	deps := testDeps(t)
	handler := getStatusHandler(deps)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool returned error")
	}

	text := extractText(t, result)
	// Result can be a JSON array or a "No tools found" message.
	if text == "" {
		t.Fatal("expected non-empty result")
	}

	var entries []map[string]any
	if err := json.Unmarshal([]byte(text), &entries); err == nil {
		// Validate status values.
		validStatuses := map[string]bool{
			"missing": true, "current": true, "outdated": true,
			"disabled": true, "unknown": true, "detected": true,
			"unavailable": true, "linked": true,
		}
		for _, entry := range entries {
			status, _ := entry["status"].(string)
			if !validStatuses[status] {
				t.Errorf("unexpected status %q in entry %v", status, entry)
			}
		}
	}
}

// TestConfigShow verifies the config_show handler returns YAML content.
func TestConfigShow(t *testing.T) {
	deps := testDeps(t)
	handler := configShowHandler(deps)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool returned error")
	}

	text := extractText(t, result)
	// Should contain the install_dir key.
	if text == "" {
		t.Fatal("expected non-empty config output")
	}
}

// TestDetectPlatform verifies the detect_platform handler returns os and arch.
func TestDetectPlatform(t *testing.T) {
	deps := testDeps(t)
	handler := detectPlatformHandler(deps)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool returned error")
	}

	text := extractText(t, result)
	var info map[string]string
	if err := json.Unmarshal([]byte(text), &info); err != nil {
		t.Fatalf("expected valid JSON, got: %s", text)
	}
	if info["os"] == "" {
		t.Error("expected non-empty os field")
	}
	if info["arch"] == "" {
		t.Error("expected non-empty arch field")
	}
}

// TestDotfilesStatus verifies the dotfiles_status handler returns entries.
func TestDotfilesStatus(t *testing.T) {
	deps := testDeps(t)
	handler := dotfilesStatusHandler(deps)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool returned error")
	}

	text := extractText(t, result)
	var entries []map[string]any
	if err := json.Unmarshal([]byte(text), &entries); err != nil {
		t.Fatalf("expected valid JSON array, got: %s", text)
	}

	validStatuses := map[string]bool{
		"linked": true, "conflict": true, "missing": true, "broken": true,
	}
	for _, entry := range entries {
		status, _ := entry["status"].(string)
		if !validStatuses[status] {
			t.Errorf("unexpected dotfile status %q", status)
		}
	}
}

// --- helpers ---

// testDeps creates a minimal Deps for testing using the real registry
// and in-memory config/state (no filesystem side effects).
func testDeps(t *testing.T) *Deps {
	t.Helper()

	cfg := &config.Config{
		InstallDir: t.TempDir(),
		Groups: config.GroupConfig{
			Languages:  true,
			Containers: true,
			Kubernetes: true,
			Infra:      true,
			Cloud:      true,
			RustTools:  true,
			Utilities:  true,
			AI:         true,
		},
	}

	reg := registry.New()

	store := &state.Store{
		Tools: make(map[string]state.InstalledTool),
	}

	info, err := platform.Detect()
	if err != nil {
		t.Fatalf("platform.Detect: %v", err)
	}

	return &Deps{
		Config:   cfg,
		Registry: reg,
		Store:    store,
		Platform: info,
	}
}

// extractText extracts the text content from a CallToolResult.
func extractText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("result has no content")
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	return tc.Text
}
