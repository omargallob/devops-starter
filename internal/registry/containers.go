// containers.go registers container runtime and orchestration tools.
package registry

import "github.com/omargallob/devops-starter/pkg/tooldef"

// registerContainers adds container-related tools to the registry.
// Includes docker CLI, docker-compose, and nerdctl.
func registerContainers(r *Registry) {
	r.register(&tooldef.Tool{
		Name:        "docker",
		Version:     "27.5.1",
		Description: "Container runtime CLI",
		Group:       tooldef.GroupContainers,
		InstallMode: tooldef.InstallModeCustom,
		Format:      tooldef.FormatTarGz,
		BinaryName:  "docker/docker",
		URLs: map[string]string{
			"linux/amd64":  "https://download.docker.com/linux/static/stable/x86_64/docker-27.5.1.tgz",
			"linux/arm64":  "https://download.docker.com/linux/static/stable/aarch64/docker-27.5.1.tgz",
			"darwin/amd64": "https://download.docker.com/mac/static/stable/x86_64/docker-27.5.1.tgz",
			"darwin/arm64": "https://download.docker.com/mac/static/stable/aarch64/docker-27.5.1.tgz",
		},
	})

	r.register(&tooldef.Tool{
		Name:        "docker-compose",
		Version:     "5.1.4",
		Description: "Docker Compose plugin",
		Group:       tooldef.GroupContainers,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "docker/compose",
	})

	r.register(&tooldef.Tool{
		Name:        "nerdctl",
		Version:     "2.3.1",
		Description: "containerd CLI",
		Group:       tooldef.GroupContainers,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "containerd/nerdctl",
		Asset:       "nerdctl-*-{{.OS}}-{{.Arch}}*",
		Platforms: []tooldef.Platform{
			{OS: "linux", Arch: "amd64"},
			{OS: "linux", Arch: "arm64"},
		},
	})
}
