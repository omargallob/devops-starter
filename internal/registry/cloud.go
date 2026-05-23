package registry

import "github.com/omargallob/devops-starter/pkg/tooldef"

func registerCloud(r *Registry) {
	r.register(&tooldef.Tool{
		Name:        "aws-cli",
		Version:     "2.22.35",
		Description: "AWS Command Line Interface",
		Group:       tooldef.GroupCloud,
		Format:      tooldef.FormatZip,
		InstallName: "aws",
		PostInstall: "./aws/install --install-dir ~/.local/aws-cli --bin-dir ~/.local/bin",
		URLs: map[string]string{
			"linux/amd64": "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip",
			"linux/arm64": "https://awscli.amazonaws.com/awscli-exe-linux-aarch64.zip",
		},
		Platforms: []tooldef.Platform{
			{OS: "linux", Arch: "amd64"},
			{OS: "linux", Arch: "arm64"},
		},
	})

	r.register(&tooldef.Tool{
		Name:        "eksctl",
		Version:     "0.198.0",
		Description: "Amazon EKS CLI",
		Group:       tooldef.GroupCloud,
		Format:      tooldef.FormatTarGz,
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/eksctl-io/eksctl/releases/download/v0.198.0/eksctl_Linux_amd64.tar.gz",
			"linux/arm64":  "https://github.com/eksctl-io/eksctl/releases/download/v0.198.0/eksctl_Linux_arm64.tar.gz",
			"darwin/amd64": "https://github.com/eksctl-io/eksctl/releases/download/v0.198.0/eksctl_Darwin_amd64.tar.gz",
			"darwin/arm64": "https://github.com/eksctl-io/eksctl/releases/download/v0.198.0/eksctl_Darwin_arm64.tar.gz",
		},
	})
}
