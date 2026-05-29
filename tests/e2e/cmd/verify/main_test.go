package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/omargallob/devops-starter/pkg/tooldef"
)

func TestGroupMatchesFilter(t *testing.T) {
	tests := []struct {
		group  string
		filter string
		want   bool
	}{
		{"utilities", "utilities", true},
		{"utilities", "utilities,containers", true},
		{"containers", "utilities,containers", true},
		{"ai", "utilities,containers", false},
		{"utilities", "", false},
		{"utilities", "util", false},
		{"utilities", " utilities ", true},
	}

	for _, tt := range tests {
		got := groupMatchesFilter(tt.group, tt.filter)
		if got != tt.want {
			t.Errorf("groupMatchesFilter(%q, %q) = %v, want %v", tt.group, tt.filter, got, tt.want)
		}
	}
}

func TestStatusIcon(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"pass", "✓"},
		{"fail", "✗"},
		{"warn", "!"},
		{"skip", "-"},
		{"unknown", " "},
		{"", " "},
	}

	for _, tt := range tests {
		got := statusIcon(tt.status)
		if got != tt.want {
			t.Errorf("statusIcon(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestPrintResults_ExitCode(t *testing.T) {
	tests := []struct {
		name    string
		results []result
		want    int
	}{
		{
			name:    "all pass",
			results: []result{{Name: "foo", Status: "pass"}, {Name: "bar", Status: "pass"}},
			want:    0,
		},
		{
			name:    "one fail",
			results: []result{{Name: "foo", Status: "pass"}, {Name: "bar", Status: "fail", Message: "not found"}},
			want:    1,
		},
		{
			name:    "all skipped",
			results: []result{{Name: "foo", Status: "skip"}, {Name: "bar", Status: "skip"}},
			want:    0,
		},
		{
			name:    "warns are not failures",
			results: []result{{Name: "foo", Status: "warn", Message: "version mismatch"}},
			want:    0,
		},
		{
			name:    "empty results",
			results: []result{},
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout to avoid noise in test output.
			old := os.Stdout
			os.Stdout, _ = os.Open(os.DevNull)
			defer func() { os.Stdout = old }()

			got := printResults(tt.results)
			if got != tt.want {
				t.Errorf("printResults() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestVerifyTool_BinaryNotFound(t *testing.T) {
	tool := &tooldef.Tool{
		Name:    "nonexistent-tool-xyz-12345",
		Version: "1.0.0",
	}

	r := verifyTool(tool)
	if r.Status != "fail" {
		t.Errorf("expected status=fail for missing binary, got %q", r.Status)
	}
	if r.Name != tool.Name {
		t.Errorf("expected name=%q, got %q", tool.Name, r.Name)
	}
}

func TestVerifyTool_BinaryExists(t *testing.T) {
	// Create a fake binary that outputs a version string.
	tmpDir := t.TempDir()
	fakeBin := filepath.Join(tmpDir, "fake-tool")
	//nolint:gosec // test helper
	err := os.WriteFile(fakeBin, []byte("#!/bin/sh\necho 'fake-tool version 2.5.0'\n"), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	// Prepend tmpDir to PATH so LookPath finds our fake binary.
	t.Setenv("PATH", tmpDir+":"+os.Getenv("PATH"))

	tool := &tooldef.Tool{
		Name:    "fake-tool",
		Version: "2.5.0",
	}

	r := verifyTool(tool)
	// No version probe defined for "fake-tool", so it should pass with "no version probe" message.
	if r.Status != "pass" {
		t.Errorf("expected status=pass, got %q: %s", r.Status, r.Message)
	}
}

func TestVerifyTool_InstallNameOverride(t *testing.T) {
	// Tool with InstallName different from Name.
	tmpDir := t.TempDir()
	fakeBin := filepath.Join(tmpDir, "mybin")
	err := os.WriteFile(fakeBin, []byte("#!/bin/sh\necho 'mybin 1.0.0'\n"), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", tmpDir+":"+os.Getenv("PATH"))

	tool := &tooldef.Tool{
		Name:        "my-tool",
		Version:     "1.0.0",
		InstallName: "mybin",
	}

	r := verifyTool(tool)
	if r.Status != "pass" {
		t.Errorf("expected status=pass for InstallName override, got %q: %s", r.Status, r.Message)
	}
}

func TestVerifyGhExtension_GhNotInstalled(t *testing.T) {
	// Set PATH to empty dir so gh is not found.
	tmpDir := t.TempDir()
	t.Setenv("PATH", tmpDir)

	tool := &tooldef.Tool{
		Name: "copilot-cli",
		Repo: "github/gh-copilot",
	}

	r := verifyGhExtension(tool)
	if r.Status != "fail" {
		t.Errorf("expected status=fail when gh not found, got %q", r.Status)
	}
}

func TestVerifyGhExtension_ExtensionFound(t *testing.T) {
	// Create a fake gh binary that outputs the extension in its list.
	tmpDir := t.TempDir()
	fakeGh := filepath.Join(tmpDir, "gh")
	script := "#!/bin/sh\necho 'github/gh-copilot  v1.0.5'\n"
	if err := os.WriteFile(fakeGh, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", tmpDir+":"+os.Getenv("PATH"))

	// Verify the fake gh is found.
	if _, err := exec.LookPath("gh"); err != nil {
		t.Skipf("fake gh not found on PATH: %v", err)
	}

	tool := &tooldef.Tool{
		Name: "copilot-cli",
		Repo: "github/gh-copilot",
	}

	r := verifyGhExtension(tool)
	if r.Status != "pass" {
		t.Errorf("expected status=pass, got %q: %s", r.Status, r.Message)
	}
}
