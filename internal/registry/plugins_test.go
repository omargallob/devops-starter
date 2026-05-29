package registry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/omargallob/devops-starter/pkg/tooldef"
)

func TestValidatePluginTool_valid(t *testing.T) {
	tests := []struct {
		name string
		tool *tooldef.Tool
	}{
		{
			name: "eget",
			tool: &tooldef.Tool{
				Name:        "mytool",
				Version:     "1.0.0",
				Description: "A test tool",
				Group:       tooldef.GroupUtilities,
				InstallMode: tooldef.InstallModeEget,
				Repo:        "myorg/mytool",
			},
		},
		{
			name: "eget-url with urls",
			tool: &tooldef.Tool{
				Name:        "mytool",
				Version:     "1.0.0",
				Group:       tooldef.GroupUtilities,
				InstallMode: tooldef.InstallModeEgetURL,
				URLs:        map[string]string{"linux/amd64": "https://example.com/mytool"},
			},
		},
		{
			name: "mise",
			tool: &tooldef.Tool{
				Name:        "mytool",
				Version:     "1.0.0",
				Group:       tooldef.GroupLanguages,
				InstallMode: tooldef.InstallModeMise,
			},
		},
		{
			name: "gh-extension",
			tool: &tooldef.Tool{
				Name:        "mytool",
				Version:     "1.0.0",
				Group:       tooldef.GroupUtilities,
				InstallMode: tooldef.InstallModeGhExtension,
				Repo:        "myorg/mytool",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidatePluginTool(tc.tool); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidatePluginTool_disallowsCustom(t *testing.T) {
	tool := &tooldef.Tool{
		Name:        "mytool",
		Version:     "1.0.0",
		Group:       tooldef.GroupUtilities,
		InstallMode: tooldef.InstallModeCustom,
		URLs:        map[string]string{"linux/amd64": "https://example.com/mytool"},
	}
	if err := ValidatePluginTool(tool); err == nil {
		t.Fatal("expected error for install_mode=custom")
	}
}

func TestValidatePluginTool_disallowsPostInstall(t *testing.T) {
	tool := &tooldef.Tool{
		Name:        "mytool",
		Version:     "1.0.0",
		Group:       tooldef.GroupUtilities,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "myorg/mytool",
		PostInstall: "echo hello",
	}
	if err := ValidatePluginTool(tool); err == nil {
		t.Fatal("expected error for post_install")
	}
}

func TestValidatePluginTool_requiredFields(t *testing.T) {
	base := &tooldef.Tool{
		Name:        "mytool",
		Version:     "1.0.0",
		Group:       tooldef.GroupUtilities,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "myorg/mytool",
	}

	cases := []struct {
		name   string
		mutate func(*tooldef.Tool)
	}{
		{"empty name", func(t *tooldef.Tool) { t.Name = "" }},
		{"empty version", func(t *tooldef.Tool) { t.Version = "" }},
		{"empty group", func(t *tooldef.Tool) { t.Group = "" }},
		{"unknown group", func(t *tooldef.Tool) { t.Group = "unknown" }},
		{"eget missing repo", func(t *tooldef.Tool) { t.Repo = "" }},
		{"gh-extension missing repo", func(t *tooldef.Tool) {
			t.InstallMode = tooldef.InstallModeGhExtension
			t.Repo = ""
		}},
		{"eget-url missing urls", func(t *tooldef.Tool) {
			t.InstallMode = tooldef.InstallModeEgetURL
			t.Repo = ""
			t.URLs = nil
			t.URLTemplate = ""
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tool := *base
			tc.mutate(&tool)
			if err := ValidatePluginTool(&tool); err == nil {
				t.Fatalf("expected validation error for %q", tc.name)
			}
		})
	}
}

func TestLoadPluginFile_valid(t *testing.T) {
	content := `
tools:
  - name: myplugin-tool
    version: "1.2.3"
    description: "A plugin tool"
    group: utilities
    install_mode: eget
    repo: myorg/myplugin-tool
`
	path := pluginWriteTempFile(t, "plugin.yaml", content)

	entries, errs := LoadPluginFile(path)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Tool.Name != "myplugin-tool" {
		t.Errorf("unexpected tool name: %s", entries[0].Tool.Name)
	}
	if entries[0].FilePath != path {
		t.Errorf("unexpected file path: %s", entries[0].FilePath)
	}
}

func TestLoadPluginFile_emptyFile(t *testing.T) {
	path := pluginWriteTempFile(t, "empty.yaml", "tools: []\n")
	entries, errs := LoadPluginFile(path)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestLoadPluginFile_invalidYAML(t *testing.T) {
	path := pluginWriteTempFile(t, "bad.yaml", "{ not valid yaml ][")
	_, errs := LoadPluginFile(path)
	if len(errs) == 0 {
		t.Fatal("expected parse error")
	}
}

func TestLoadPluginFile_missingFile(t *testing.T) {
	_, errs := LoadPluginFile("/nonexistent/path/plugin.yaml")
	if len(errs) == 0 {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadPluginFile_partialValidation(t *testing.T) {
	content := `
tools:
  - name: good-tool
    version: "1.0.0"
    description: "ok"
    group: utilities
    install_mode: eget
    repo: myorg/good-tool
  - name: bad-tool
    version: "1.0.0"
    group: utilities
    install_mode: custom
    urls:
      linux/amd64: "https://example.com/bad"
`
	path := pluginWriteTempFile(t, "plugin.yaml", content)
	entries, errs := LoadPluginFile(path)
	if len(entries) != 1 {
		t.Errorf("expected 1 valid entry, got %d", len(entries))
	}
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestRegisterPlugins_collision(t *testing.T) {
	reg := New()
	builtinCount := len(reg.All())

	// kubectl already exists as a built-in; plugin entry must be skipped.
	entries := []PluginEntry{
		{
			Tool: &tooldef.Tool{
				Name:        "kubectl",
				Version:     "99.0.0",
				Description: "plugin kubectl",
				Group:       tooldef.GroupKubernetes,
				InstallMode: tooldef.InstallModeEget,
				Repo:        "myorg/kubectl",
			},
			FilePath: "/fake/plugin.yaml",
		},
	}

	reg.RegisterPlugins(entries)

	if len(reg.All()) != builtinCount {
		t.Errorf("expected %d tools after collision skip, got %d", builtinCount, len(reg.All()))
	}
	if len(reg.PluginEntries()) != 0 {
		t.Errorf("expected no plugin entries after collision, got %d", len(reg.PluginEntries()))
	}
	tool, _ := reg.Get("kubectl")
	if tool.Version == "99.0.0" {
		t.Error("built-in kubectl was overwritten by plugin")
	}
}

func TestRegisterPlugins_newTool(t *testing.T) {
	reg := New()

	entries := []PluginEntry{
		{
			Tool: &tooldef.Tool{
				Name:        "my-unique-plugin-tool",
				Version:     "1.0.0",
				Description: "A unique plugin tool",
				Group:       tooldef.GroupUtilities,
				InstallMode: tooldef.InstallModeEget,
				Repo:        "myorg/my-unique-plugin-tool",
			},
			FilePath: "/fake/plugin.yaml",
		},
	}

	reg.RegisterPlugins(entries)

	tool, ok := reg.Get("my-unique-plugin-tool")
	if !ok {
		t.Fatal("expected plugin tool to be registered")
	}
	if tool.Version != "1.0.0" {
		t.Errorf("unexpected version: %s", tool.Version)
	}

	pe := reg.PluginEntries()
	if len(pe) != 1 {
		t.Fatalf("expected 1 plugin entry, got %d", len(pe))
	}
	if pe[0].FilePath != "/fake/plugin.yaml" {
		t.Errorf("unexpected file path: %s", pe[0].FilePath)
	}
}

func TestLoadPluginDirs_precedence(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	pluginWriteTempFileInDir(t, dir1, "plugin.yaml", `
tools:
  - name: precedence-test-tool
    version: "1.0.0"
    description: "v1"
    group: utilities
    install_mode: eget
    repo: myorg/precedence-test-tool
`)
	pluginWriteTempFileInDir(t, dir2, "plugin.yaml", `
tools:
  - name: precedence-test-tool
    version: "2.0.0"
    description: "v2"
    group: utilities
    install_mode: eget
    repo: myorg/precedence-test-tool
`)

	entries, errs := LoadPluginDirs([]string{dir1, dir2})
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 deduplicated entry, got %d", len(entries))
	}
	if entries[0].Tool.Version != "2.0.0" {
		t.Errorf("expected version 2.0.0 (dir2 takes precedence), got %s", entries[0].Tool.Version)
	}
}

func TestLoadPluginDirs_emptyDir(t *testing.T) {
	entries, errs := LoadPluginDirs([]string{t.TempDir()})
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestLoadPluginDirs_nonexistentDir(t *testing.T) {
	// A directory that doesn't exist should produce zero entries and zero errors
	// (treated as an empty plugin directory, not a fatal error).
	entries, errs := LoadPluginDirs([]string{"/nonexistent/plugin/dir"})
	if len(errs) != 0 {
		t.Fatalf("unexpected errors for missing dir: %v", errs)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func pluginWriteTempFile(t *testing.T, name, content string) string {
	t.Helper()
	return pluginWriteTempFileInDir(t, t.TempDir(), name, content)
}

func pluginWriteTempFileInDir(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	return path
}
