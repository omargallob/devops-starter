// Package registry provides the built-in catalog of all tools managed by
// devops-starter. Each tool group (languages, containers, kubernetes, infra,
// cloud, rust-tools, utilities, ai) is defined in a separate file and registered
// at construction time via New(). The registry allows lookup by name, group,
// or retrieval of all tools sorted alphabetically.
package registry

import (
	"fmt"
	"os"
	"sort"

	"github.com/omargallob/devops-starter/internal/mise"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// Registry holds all known tool definitions indexed by name.
type Registry struct {
	tools   map[string]*tooldef.Tool
	plugins []PluginEntry // registered plugin tools, in load order
}

// New creates a registry with all built-in tools registered.
// Each registerXxx function adds tools from a single group.
// Mise-managed tools are discovered from .mise.toml in the working directory.
// Optional extraPluginDirs are appended after the standard plugin discovery
// directories (project-local, then user-global); they take highest precedence.
func New(extraPluginDirs ...string) *Registry {
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
	registerAI(r)

	// Discover and register mise-managed language runtimes from .mise.toml.
	// Errors are silently ignored — if no .mise.toml is found or it's
	// unparseable, the registry simply won't include those tools.
	if wd, err := os.Getwd(); err == nil {
		if versions, err := mise.FindAndParse(wd); err == nil && versions != nil {
			r.RegisterMiseTools(versions)
		}
	}

	// Load plugins: standard dirs (project-local → user-global) then extra dirs.
	dirs := append(standardPluginDirs(), extraPluginDirs...)
	entries, errs := LoadPluginDirs(dirs)
	for _, err := range errs {
		fmt.Fprintf(os.Stderr, "warning: plugin: %v\n", err)
	}
	r.RegisterPlugins(entries)

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
