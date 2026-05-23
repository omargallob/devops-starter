// languages.go registers programming language runtime managers.
package registry

import "github.com/omargallob/devops-starter/pkg/tooldef"

// registerLanguages adds language runtime tools to the registry.
// Currently includes mise (polyglot version manager, formerly rtx).
func registerLanguages(r *Registry) {
	r.register(&tooldef.Tool{
		Name:        "mise",
		Version:     "2025.1.6",
		Description: "Polyglot runtime manager (formerly rtx)",
		Group:       tooldef.GroupLanguages,
		Format:      tooldef.FormatTarGz,
		BinaryName:  "mise",
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/jdx/mise/releases/download/v2025.1.6/mise-v2025.1.6-linux-x64.tar.gz",
			"linux/arm64":  "https://github.com/jdx/mise/releases/download/v2025.1.6/mise-v2025.1.6-linux-arm64.tar.gz",
			"darwin/amd64": "https://github.com/jdx/mise/releases/download/v2025.1.6/mise-v2025.1.6-macos-x64.tar.gz",
			"darwin/arm64": "https://github.com/jdx/mise/releases/download/v2025.1.6/mise-v2025.1.6-macos-arm64.tar.gz",
		},
	})
}
