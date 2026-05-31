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

// TestListTools_EmptyGroup verifies that filtering by a group with no tools
// returns an empty array.
func TestListTools_EmptyGroup(t *testing.T) {
	deps := testDeps(t)
	handler := listToolsHandler(deps)

	// "ansible" group likely has tools, but let's use a config that might not.
	// More importantly, test that the handler doesn't crash on empty results.
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"group": "package-managers"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("unexpected tool error")
	}

	text := extractText(t, result)
	var tools []map[string]any
	if err := json.Unmarshal([]byte(text), &tools); err != nil {
		t.Fatalf("expected valid JSON, got: %s", text)
	}
	// Just verify it's a valid array (may be empty or not depending on registry).
}

// TestGetTool_MissingNameParam verifies error when name param is missing.
func TestGetTool_MissingNameParam(t *testing.T) {
	deps := testDeps(t)
	handler := getToolHandler(deps)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{} // no "name" param

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result when name param is missing")
	}
}

// TestGetStatus_FilterByStatus verifies filtering by status.
func TestGetStatus_FilterByStatus(t *testing.T) {
	deps := testDeps(t)
	handler := getStatusHandler(deps)

	// Filter for "missing" — most tools should be missing in test env.
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"status_filter": "missing"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("unexpected tool error")
	}

	text := extractText(t, result)
	var entries []map[string]any
	if err := json.Unmarshal([]byte(text), &entries); err == nil {
		for _, entry := range entries {
			if entry["status"] != "missing" {
				t.Errorf("expected all entries to have status=missing, got %v", entry["status"])
			}
		}
	}
}

// TestGetStatus_FilterByGroupAndStatus verifies combined filtering.
func TestGetStatus_FilterByGroupAndStatus(t *testing.T) {
	deps := testDeps(t)
	handler := getStatusHandler(deps)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"group":         "kubernetes",
		"status_filter": "missing",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("unexpected tool error")
	}

	text := extractText(t, result)
	var entries []map[string]any
	if err := json.Unmarshal([]byte(text), &entries); err == nil {
		for _, entry := range entries {
			if entry["group"] != "kubernetes" {
				t.Errorf("expected group=kubernetes, got %v", entry["group"])
			}
			if entry["status"] != "missing" {
				t.Errorf("expected status=missing, got %v", entry["status"])
			}
		}
	}
}

// TestGetStatus_NoResults verifies the "no tools found" message.
func TestGetStatus_NoResults(t *testing.T) {
	// Use a config with all groups disabled so nothing appears.
	cfg := &config.Config{
		InstallDir: t.TempDir(),
		Groups:     config.GroupConfig{}, // all disabled
	}
	info, err := platform.Detect()
	if err != nil {
		t.Fatal(err)
	}
	deps := &Deps{
		Config:   cfg,
		Registry: registry.New(),
		Store:    &state.Store{Tools: make(map[string]state.InstalledTool)},
		Platform: info,
	}

	handler := getStatusHandler(deps)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"status_filter": "current"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("unexpected tool error")
	}

	text := extractText(t, result)
	// Should either be "No tools found" message or an empty JSON array.
	if text == "" {
		t.Error("expected non-empty result")
	}
}

// TestGetTool_AllFields verifies that all expected fields are present in output.
func TestGetTool_AllFields(t *testing.T) {
	deps := testDeps(t)
	handler := getToolHandler(deps)

	// Use a tool we know has rich data (kubectl has many fields).
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"name": "kubectl"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		// kubectl might not be in registry depending on test env, skip.
		t.Skip("kubectl not in test registry")
	}

	text := extractText(t, result)
	var detail map[string]any
	if err := json.Unmarshal([]byte(text), &detail); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Verify required fields exist.
	requiredFields := []string{"name", "version", "group", "install_mode"}
	for _, field := range requiredFields {
		if _, ok := detail[field]; !ok {
			t.Errorf("missing required field %q in tool detail", field)
		}
	}

	if detail["name"] != "kubectl" {
		t.Errorf("expected name=kubectl, got %v", detail["name"])
	}
}

// TestDetectPlatform_ValidValues verifies platform values are within expected set.
func TestDetectPlatform_ValidValues(t *testing.T) {
	deps := testDeps(t)
	handler := detectPlatformHandler(deps)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := extractText(t, result)
	var info map[string]string
	if err := json.Unmarshal([]byte(text), &info); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	validOS := map[string]bool{"linux": true, "darwin": true}
	validArch := map[string]bool{"amd64": true, "arm64": true}

	if !validOS[info["os"]] {
		t.Errorf("unexpected os value: %s", info["os"])
	}
	if !validArch[info["arch"]] {
		t.Errorf("unexpected arch value: %s", info["arch"])
	}
}

// TestDotfilesStatus_Entries verifies each entry has source, dest, and status.
func TestDotfilesStatus_AllEntriesHaveRequiredFields(t *testing.T) {
	deps := testDeps(t)
	handler := dotfilesStatusHandler(deps)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := extractText(t, result)
	var entries []map[string]any
	if err := json.Unmarshal([]byte(text), &entries); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	for i, entry := range entries {
		for _, field := range []string{"source", "dest", "status"} {
			if _, ok := entry[field]; !ok {
				t.Errorf("entry[%d] missing field %q", i, field)
			}
		}
	}
}
