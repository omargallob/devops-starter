// infra.go registers infrastructure-as-code and secrets management tools.
package registry

import "github.com/omargallob/devops-starter/pkg/tooldef"

// registerInfra adds infrastructure tools to the registry.
// Includes terraform, opentofu, pulumi, packer, vault, and consul.
func registerInfra(r *Registry) {
	r.register(&tooldef.Tool{
		Name:        "terraform",
		Version:     "1.15.5",
		Description: "Infrastructure as Code",
		Group:       tooldef.GroupInfra,
		InstallMode: tooldef.InstallModeEgetURL,
		Format:      tooldef.FormatZip,
		URLTemplate: "https://releases.hashicorp.com/terraform/{{.Version}}/terraform_{{.Version}}_{{.OS}}_{{.Arch}}.zip",
	})

	r.register(&tooldef.Tool{
		Name:        "opentofu",
		Version:     "1.12.1",
		Description: "Open-source Terraform alternative",
		Group:       tooldef.GroupInfra,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "opentofu/opentofu",
		InstallName: "tofu",
		Asset:       "tofu_*_{{.OS}}_{{.Arch}}*",
	})

	r.register(&tooldef.Tool{
		Name:        "pulumi",
		Version:     "3.144.1",
		Description: "Infrastructure as Code SDK",
		Group:       tooldef.GroupInfra,
		InstallMode: tooldef.InstallModeCustom,
		Format:      tooldef.FormatTarGz,
		BinaryName:  "pulumi/pulumi",
		URLs: map[string]string{
			"linux/amd64":  "https://get.pulumi.com/releases/sdk/pulumi-v3.144.1-linux-x64.tar.gz",
			"linux/arm64":  "https://get.pulumi.com/releases/sdk/pulumi-v3.144.1-linux-arm64.tar.gz",
			"darwin/amd64": "https://get.pulumi.com/releases/sdk/pulumi-v3.144.1-darwin-x64.tar.gz",
			"darwin/arm64": "https://get.pulumi.com/releases/sdk/pulumi-v3.144.1-darwin-arm64.tar.gz",
		},
	})

	r.register(&tooldef.Tool{
		Name:        "packer",
		Version:     "1.15.3",
		Description: "Machine image builder",
		Group:       tooldef.GroupInfra,
		InstallMode: tooldef.InstallModeEgetURL,
		Format:      tooldef.FormatZip,
		URLTemplate: "https://releases.hashicorp.com/packer/{{.Version}}/packer_{{.Version}}_{{.OS}}_{{.Arch}}.zip",
	})

	r.register(&tooldef.Tool{
		Name:        "vault",
		Version:     "2.0.1",
		Description: "Secrets management",
		Group:       tooldef.GroupInfra,
		InstallMode: tooldef.InstallModeEgetURL,
		Format:      tooldef.FormatZip,
		URLTemplate: "https://releases.hashicorp.com/vault/{{.Version}}/vault_{{.Version}}_{{.OS}}_{{.Arch}}.zip",
	})

	r.register(&tooldef.Tool{
		Name:        "consul",
		Version:     "2.0.0",
		Description: "Service mesh and KV store",
		Group:       tooldef.GroupInfra,
		InstallMode: tooldef.InstallModeEgetURL,
		Format:      tooldef.FormatZip,
		URLTemplate: "https://releases.hashicorp.com/consul/{{.Version}}/consul_{{.Version}}_{{.OS}}_{{.Arch}}.zip",
	})
}
