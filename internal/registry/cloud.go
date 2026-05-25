// cloud.go registers cloud provider CLI tools.
package registry

import "github.com/omargallob/devops-starter/pkg/tooldef"

// registerCloud adds cloud provider tools to the registry.
// Includes aws-cli, azure-cli, gcloud-cli, firebase-cli, and eksctl.
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
			"linux/amd64":  "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip",
			"linux/arm64":  "https://awscli.amazonaws.com/awscli-exe-linux-aarch64.zip",
			"darwin/amd64": "https://awscli.amazonaws.com/AWSCLIV2.pkg",
			"darwin/arm64": "https://awscli.amazonaws.com/AWSCLIV2.pkg",
		},
	})

	r.register(&tooldef.Tool{
		Name:        "azure-cli",
		Version:     "2.67.0",
		Description: "Microsoft Azure Command Line Interface",
		Group:       tooldef.GroupCloud,
		Format:      tooldef.FormatZip,
		InstallName: "az",
		PostInstall: "python3 -m pip install --quiet --prefix ~/.local azure-cli==${VERSION}",
		URLs: map[string]string{
			"linux/amd64":  "https://azcliprod.blob.core.windows.net/releases/azure-cli-2.67.0-linux-x86_64.zip",
			"linux/arm64":  "https://azcliprod.blob.core.windows.net/releases/azure-cli-2.67.0-linux-aarch64.zip",
			"darwin/amd64": "https://azcliprod.blob.core.windows.net/releases/azure-cli-2.67.0-macos-x86_64.zip",
			"darwin/arm64": "https://azcliprod.blob.core.windows.net/releases/azure-cli-2.67.0-macos-arm64.zip",
		},
	})

	r.register(&tooldef.Tool{
		Name:        "gcloud-cli",
		Version:     "503.0.0",
		Description: "Google Cloud SDK CLI",
		Group:       tooldef.GroupCloud,
		Format:      tooldef.FormatTarGz,
		InstallName: "gcloud",
		PostInstall: "./google-cloud-sdk/install.sh --quiet --usage-reporting false --path-update false --command-completion false && ln -sf $(pwd)/google-cloud-sdk/bin/gcloud ~/.local/bin/gcloud && ln -sf $(pwd)/google-cloud-sdk/bin/gsutil ~/.local/bin/gsutil",
		URLs: map[string]string{
			"linux/amd64":  "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-503.0.0-linux-x86_64.tar.gz",
			"linux/arm64":  "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-503.0.0-linux-arm.tar.gz",
			"darwin/amd64": "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-503.0.0-darwin-x86_64.tar.gz",
			"darwin/arm64": "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-503.0.0-darwin-arm.tar.gz",
		},
	})

	r.register(&tooldef.Tool{
		Name:        "firebase-cli",
		Version:     "13.29.1",
		Description: "Firebase Command Line Interface",
		Group:       tooldef.GroupCloud,
		Format:      tooldef.FormatBinary,
		InstallName: "firebase",
		URLs: map[string]string{
			"linux/amd64":  "https://firebase.tools/bin/linux/v13.29.1",
			"linux/arm64":  "https://firebase.tools/bin/linux/arm64/v13.29.1",
			"darwin/amd64": "https://firebase.tools/bin/macos/v13.29.1",
			"darwin/arm64": "https://firebase.tools/bin/macos/arm64/v13.29.1",
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
