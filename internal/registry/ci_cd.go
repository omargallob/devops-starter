package registry

import "github.com/omargallob/devops-starter/pkg/tooldef"

// registerCICD adds CI/CD and release automation tools to the registry.
func registerCICD(r *Registry) {
	r.register(&tooldef.Tool{
		Name:        "goreleaser",
		Version:     "2.6.1",
		Description: "Release automation for Go projects",
		Group:       tooldef.GroupCICD,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "goreleaser/goreleaser",
	})

	r.register(&tooldef.Tool{
		Name:         "release-please",
		Version:      "16",
		Description:  "Automated CHANGELOG and version management from conventional commits",
		Group:        tooldef.GroupCICD,
		InstallMode:  tooldef.InstallModeMise,
		MiseBackend:  "npm:release-please",
		Dependencies: []string{"mise"},
	})

	r.register(&tooldef.Tool{
		Name:         "semantic-release",
		Version:      "24.2.3",
		Description:  "Automated changelog and version management from commit history",
		Group:        tooldef.GroupCICD,
		InstallMode:  tooldef.InstallModeMise,
		MiseBackend:  "npm:semantic-release",
		Dependencies: []string{"mise"},
	})
}
