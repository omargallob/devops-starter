// ansible.go registers Ansible automation tools. All tools in this group
// are distributed exclusively via pip and therefore use InstallMode mise
// with pipx backends, which install each tool into an isolated virtualenv.
package registry

import "github.com/omargallob/devops-starter/pkg/tooldef"

// registerAnsible adds Ansible platform tools to the registry.
// Includes ansible core, ansible-lint, and molecule.
func registerAnsible(r *Registry) {
	r.register(&tooldef.Tool{
		Name:         "ansible",
		Version:      "11.2.0",
		Description:  "IT automation and configuration management platform",
		Group:        tooldef.GroupAnsible,
		InstallMode:  tooldef.InstallModeMise,
		MiseBackend:  "pipx:ansible",
		Dependencies: []string{"mise"},
	})

	r.register(&tooldef.Tool{
		Name:         "ansible-lint",
		Version:      "24.12.2",
		Description:  "Linter for Ansible playbooks and roles",
		Group:        tooldef.GroupAnsible,
		InstallMode:  tooldef.InstallModeMise,
		MiseBackend:  "pipx:ansible-lint",
		Dependencies: []string{"mise"},
	})

	r.register(&tooldef.Tool{
		Name:         "molecule",
		Version:      "24.12.0",
		Description:  "Ansible role and collection testing framework",
		Group:        tooldef.GroupAnsible,
		InstallMode:  tooldef.InstallModeMise,
		MiseBackend:  "pipx:molecule",
		Dependencies: []string{"mise"},
	})
}
