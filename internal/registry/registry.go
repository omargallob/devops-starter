// Package registry provides the built-in catalog of all tools managed by
// devops-starter. Each tool group (languages, containers, kubernetes, infra,
// cloud, rust-tools, utilities) is defined in a separate file and registered
// at construction time via New(). The registry allows lookup by name, group,
// or retrieval of all tools sorted alphabetically.
package registry

import (
	"sort"

	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// Registry holds all known tool definitions indexed by name.
type Registry struct {
	tools map[string]*tooldef.Tool
}

// New creates a registry with all built-in tools registered.
// Each registerXxx function adds tools from a single group.
func New() *Registry {
	r := &Registry{
		tools: make(map[string]*tooldef.Tool),
	}
	registerLanguages(r)
	registerContainers(r)
	registerKubernetes(r)
	registerInfra(r)
	registerCloud(r)
	registerRustTools(r)
	registerUtilities(r)
	return r
}

// register adds a tool definition to the registry, keyed by tool name.
func (r *Registry) register(t *tooldef.Tool) {
	r.tools[t.Name] = t
}

// Get returns a tool by name.
func (r *Registry) Get(name string) (*tooldef.Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// GetByGroup returns all tools in a given group.
func (r *Registry) GetByGroup(group tooldef.Group) []*tooldef.Tool {
	var result []*tooldef.Tool
	for _, t := range r.tools {
		if t.Group == group {
			result = append(result, t)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// All returns all registered tools.
func (r *Registry) All() []*tooldef.Tool {
	result := make([]*tooldef.Tool, 0, len(r.tools))
	for _, t := range r.tools {
		result = append(result, t)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// Names returns all tool names sorted alphabetically.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
