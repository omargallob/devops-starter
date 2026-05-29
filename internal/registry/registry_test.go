package registry

import (
	"strings"
	"testing"
	"text/template"

	"github.com/omargallob/devops-starter/pkg/tooldef"
)

func TestNewReturnsNonEmpty(t *testing.T) {
	reg := New()
	if len(reg.All()) == 0 {
		t.Fatal("expected non-empty registry")
	}
}

func TestGetKnownTool(t *testing.T) {
	reg := New()
	tool, ok := reg.Get("kubectl")
	if !ok {
		t.Fatal("expected to find kubectl")
	}
	if tool.Name != "kubectl" {
		t.Fatalf("expected name kubectl, got %s", tool.Name)
	}
}

func TestGetUnknownTool(t *testing.T) {
	reg := New()
	_, ok := reg.Get("nonexistent-tool")
	if ok {
		t.Fatal("expected not to find nonexistent tool")
	}
}

func TestGetByGroup(t *testing.T) {
	reg := New()
	tools := reg.GetByGroup(tooldef.GroupKubernetes)
	if len(tools) == 0 {
		t.Fatal("expected non-empty kubernetes tools")
	}
	for _, tool := range tools {
		if tool.Group != tooldef.GroupKubernetes {
			t.Fatalf("tool %s has wrong group %s", tool.Name, tool.Group)
		}
	}
}

func TestNamesSorted(t *testing.T) {
	reg := New()
	names := reg.Names()
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Fatalf("names not sorted: %s before %s", names[i-1], names[i])
		}
	}
}

func TestAllToolsHaveRequiredFields(t *testing.T) {
	reg := New()
	for _, tool := range reg.All() {
		t.Run(tool.Name, func(t *testing.T) {
			if tool.Name == "" {
				t.Fatal("tool has empty name")
			}
			if tool.Version == "" {
				t.Fatal("empty version")
			}
			if tool.Group == "" {
				t.Fatal("empty group")
			}
			validateInstallMode(t, tool)
		})
	}
}

// validateInstallMode checks that a tool has the required fields for its
// install mode.
func validateInstallMode(t *testing.T, tool *tooldef.Tool) {
	t.Helper()
	mode := tool.EffectiveInstallMode()
	switch mode {
	case tooldef.InstallModeEget:
		if tool.Repo == "" {
			t.Fatal("InstallMode=eget but no Repo")
		}
	case tooldef.InstallModeEgetURL:
		if tool.URLTemplate == "" && len(tool.URLs) == 0 {
			t.Fatal("InstallMode=eget-url but no URLTemplate or URLs")
		}
	case tooldef.InstallModeCustom:
		if tool.URLTemplate == "" && len(tool.URLs) == 0 {
			t.Fatal("InstallMode=custom but no URLTemplate or URLs")
		}
	case tooldef.InstallModeMise:
		// No URL needed for mise-managed tools.
	default:
		t.Fatalf("unknown InstallMode %q", mode)
	}
}

func TestAllToolsHaveValidGroup(t *testing.T) {
	validGroups := map[tooldef.Group]bool{
		tooldef.GroupLanguages:  true,
		tooldef.GroupContainers: true,
		tooldef.GroupKubernetes: true,
		tooldef.GroupInfra:      true,
		tooldef.GroupCloud:      true,
		tooldef.GroupAnsible:    true,
		tooldef.GroupRustTools:  true,
		tooldef.GroupUtilities:  true,
	}

	reg := New()
	for _, tool := range reg.All() {
		if !validGroups[tool.Group] {
			t.Errorf("tool %s has invalid group %q", tool.Name, tool.Group)
		}
	}
}

func TestAllToolsHaveValidFormat(t *testing.T) {
	validFormats := map[tooldef.ArchiveFormat]bool{
		tooldef.FormatTarGz:  true,
		tooldef.FormatTarXz:  true,
		tooldef.FormatZip:    true,
		tooldef.FormatBinary: true,
	}

	reg := New()
	for _, tool := range reg.All() {
		mode := tool.EffectiveInstallMode()
		// eget-mode and mise-managed tools don't require a format —
		// eget auto-detects the archive type from the release asset.
		if mode == tooldef.InstallModeEget || mode == tooldef.InstallModeMise {
			continue
		}
		if !validFormats[tool.Format] {
			t.Errorf("tool %s has invalid format %q", tool.Name, tool.Format)
		}
	}
}

func TestAllToolURLTemplatesRender(t *testing.T) {
	platforms := []tooldef.Platform{
		{OS: "linux", Arch: "amd64"},
		{OS: "linux", Arch: "arm64"},
		{OS: "darwin", Arch: "amd64"},
		{OS: "darwin", Arch: "arm64"},
	}

	type urlTemplateData struct {
		Name       string
		Version    string
		OS         string
		Arch       string
		Format     string
		BinaryName string
	}

	reg := New()
	for _, tool := range reg.All() {
		if tool.URLTemplate == "" {
			continue
		}

		tmpl, err := template.New("url").Parse(tool.URLTemplate)
		if err != nil {
			t.Errorf("tool %s: URL template parse error: %v", tool.Name, err)
			continue
		}

		for _, plat := range platforms {
			// Skip if tool doesn't support this platform
			if !tool.SupportsPlatform(plat) {
				continue
			}

			data := urlTemplateData{
				Name:       tool.Name,
				Version:    tool.Version,
				OS:         plat.OS,
				Arch:       plat.Arch,
				Format:     string(tool.Format),
				BinaryName: tool.GetBinaryName(),
			}

			var buf strings.Builder
			if err := tmpl.Execute(&buf, data); err != nil {
				t.Errorf("tool %s on %s: URL template execution error: %v", tool.Name, plat, err)
				continue
			}

			url := buf.String()
			if url == "" {
				t.Errorf("tool %s on %s: rendered URL is empty", tool.Name, plat)
			}
			if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
				t.Errorf("tool %s on %s: URL %q does not start with http(s)://", tool.Name, plat, url)
			}
		}
	}
}

func TestAllToolsHaveDescription(t *testing.T) {
	reg := New()
	for _, tool := range reg.All() {
		if tool.Description == "" {
			t.Errorf("tool %s has empty description", tool.Name)
		}
	}
}

func TestNoToolNameDuplicates(t *testing.T) {
	reg := New()
	seen := make(map[string]bool)
	for _, tool := range reg.All() {
		if seen[tool.Name] {
			t.Errorf("duplicate tool name: %s", tool.Name)
		}
		seen[tool.Name] = true
	}
}

func TestAllGroupsHaveTools(t *testing.T) {
	groups := []tooldef.Group{
		tooldef.GroupLanguages,
		tooldef.GroupContainers,
		tooldef.GroupKubernetes,
		tooldef.GroupInfra,
		tooldef.GroupCloud,
		tooldef.GroupRustTools,
		tooldef.GroupUtilities,
	}

	reg := New()
	for _, group := range groups {
		tools := reg.GetByGroup(group)
		if len(tools) == 0 {
			t.Errorf("group %q has no tools registered", group)
		}
	}
}
