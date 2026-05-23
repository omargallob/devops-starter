package registry

import "github.com/omargallob/devops-starter/pkg/tooldef"

func registerContainers(r *Registry) {
	r.register(&tooldef.Tool{
		Name:        "docker",
		Version:     "27.5.1",
		Description: "Container runtime CLI",
		Group:       tooldef.GroupContainers,
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
		Version:     "2.32.4",
		Description: "Docker Compose plugin",
		Group:       tooldef.GroupContainers,
		Format:      tooldef.FormatBinary,
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/docker/compose/releases/download/v2.32.4/docker-compose-linux-x86_64",
			"linux/arm64":  "https://github.com/docker/compose/releases/download/v2.32.4/docker-compose-linux-aarch64",
			"darwin/amd64": "https://github.com/docker/compose/releases/download/v2.32.4/docker-compose-darwin-x86_64",
			"darwin/arm64": "https://github.com/docker/compose/releases/download/v2.32.4/docker-compose-darwin-aarch64",
		},
	})

	r.register(&tooldef.Tool{
		Name:        "nerdctl",
		Version:     "2.0.3",
		Description: "containerd CLI",
		Group:       tooldef.GroupContainers,
		Format:      tooldef.FormatTarGz,
		URLTemplate: "https://github.com/containerd/nerdctl/releases/download/v{{.Version}}/nerdctl-{{.Version}}-{{.OS}}-{{.Arch}}.tar.gz",
		Platforms: []tooldef.Platform{
			{OS: "linux", Arch: "amd64"},
			{OS: "linux", Arch: "arm64"},
		},
	})
}
