package registry

import (
	"testing"

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
		if tool.Name == "" {
			t.Fatal("tool has empty name")
		}
		if tool.Version == "" {
			t.Fatalf("tool %s has empty version", tool.Name)
		}
		if tool.Group == "" {
			t.Fatalf("tool %s has empty group", tool.Name)
		}
		if tool.URLTemplate == "" && len(tool.URLs) == 0 && tool.ManagedBy == "" {
			t.Fatalf("tool %s has neither URLTemplate nor URLs (and is not externally managed)", tool.Name)
		}
	}
}
