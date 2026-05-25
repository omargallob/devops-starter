// languages.go registers programming language runtime managers and their
// managed tools discovered from .mise.toml.
package registry

import (
	"github.com/omargallob/devops-starter/internal/mise"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// Subgroup constants for the languages group.
const (
	SubgroupPlatforms = "Platforms"
	SubgroupLanguages = "Languages"
)

// registerLanguages adds the mise tool manager to the registry.
// Mise-managed language runtimes are registered separately via
// RegisterMiseTools after .mise.toml discovery.
func registerLanguages(r *Registry) {
	r.register(&tooldef.Tool{
		Name:            "mise",
		Version:         "2025.1.6",
		Description:     "Polyglot runtime manager (formerly rtx)",
		Group:           tooldef.GroupLanguages,
		Subgroup:        SubgroupPlatforms,
		Format:          tooldef.FormatTarGz,
		BinaryName:      "mise",
		StripComponents: 2,
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/jdx/mise/releases/download/v2025.1.6/mise-v2025.1.6-linux-x64.tar.gz",
			"linux/arm64":  "https://github.com/jdx/mise/releases/download/v2025.1.6/mise-v2025.1.6-linux-arm64.tar.gz",
			"darwin/amd64": "https://github.com/jdx/mise/releases/download/v2025.1.6/mise-v2025.1.6-macos-x64.tar.gz",
			"darwin/arm64": "https://github.com/jdx/mise/releases/download/v2025.1.6/mise-v2025.1.6-macos-arm64.tar.gz",
		},
	})
}

// RegisterMiseTools dynamically registers all tools found in a mise.ToolVersions
// map (parsed from .mise.toml) into the languages group. Each tool is marked
// as ManagedBy "mise" with a dependency on the mise binary.
func (r *Registry) RegisterMiseTools(versions mise.ToolVersions) {
	for name, version := range versions {
		// Skip if a tool with this name is already registered (e.g., "mise" itself,
		// or a tool that exists in another group like a utility).
		if _, exists := r.tools[name]; exists {
			continue
		}

		r.register(&tooldef.Tool{
			Name:         name,
			Version:      version,
			Description:  mise.DescriptionFor(name),
			Group:        tooldef.GroupLanguages,
			Subgroup:     SubgroupLanguages,
			ManagedBy:    "mise",
			Dependencies: []string{"mise"},
		})
	}
}
