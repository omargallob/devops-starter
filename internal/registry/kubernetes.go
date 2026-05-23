// kubernetes.go registers Kubernetes ecosystem tools: cluster management CLIs,
// package managers, GitOps controllers, service mesh, and backup utilities.
package registry

import "github.com/omargallob/devops-starter/pkg/tooldef"

// registerKubernetes adds all Kubernetes-related tools to the registry.
// Tools in this group: kubectl, helm, kustomize, k9s, kubectx, kubens,
// stern, argocd, flux, istioctl, cilium, kind, kubeseal, velero.
func registerKubernetes(r *Registry) {
	r.register(&tooldef.Tool{
		Name:        "kubectl",
		Version:     "1.31.4",
		Description: "Kubernetes CLI",
		Group:       tooldef.GroupKubernetes,
		Format:      tooldef.FormatBinary,
		URLTemplate: "https://dl.k8s.io/release/v{{.Version}}/bin/{{.OS}}/{{.Arch}}/kubectl",
	})

	r.register(&tooldef.Tool{
		Name:        "helm",
		Version:     "3.16.4",
		Description: "Kubernetes package manager",
		Group:       tooldef.GroupKubernetes,
		Format:      tooldef.FormatTarGz,
		BinaryName:  "{{.OS}}-{{.Arch}}/helm",
		URLTemplate: "https://get.helm.sh/helm-v{{.Version}}-{{.OS}}-{{.Arch}}.tar.gz",
	})

	r.register(&tooldef.Tool{
		Name:        "kustomize",
		Version:     "5.5.0",
		Description: "Kubernetes configuration management",
		Group:       tooldef.GroupKubernetes,
		Format:      tooldef.FormatTarGz,
		URLTemplate: "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv{{.Version}}/kustomize_v{{.Version}}_{{.OS}}_{{.Arch}}.tar.gz",
	})

	r.register(&tooldef.Tool{
		Name:        "k9s",
		Version:     "0.32.7",
		Description: "Kubernetes TUI",
		Group:       tooldef.GroupKubernetes,
		Format:      tooldef.FormatTarGz,
		URLs: map[string]string{
			"linux/amd64":  "https://github.com/derailed/k9s/releases/download/v0.32.7/k9s_Linux_amd64.tar.gz",
			"linux/arm64":  "https://github.com/derailed/k9s/releases/download/v0.32.7/k9s_Linux_arm64.tar.gz",
			"darwin/amd64": "https://github.com/derailed/k9s/releases/download/v0.32.7/k9s_Darwin_amd64.tar.gz",
			"darwin/arm64": "https://github.com/derailed/k9s/releases/download/v0.32.7/k9s_Darwin_arm64.tar.gz",
		},
	})

	r.register(&tooldef.Tool{
		Name:        "kubectx",
		Version:     "0.9.5",
		Description: "Kubernetes context switcher",
		Group:       tooldef.GroupKubernetes,
		Format:      tooldef.FormatTarGz,
		URLTemplate: "https://github.com/ahmetb/kubectx/releases/download/v{{.Version}}/kubectx_v{{.Version}}_{{.OS}}_{{.Arch}}.tar.gz",
	})

	r.register(&tooldef.Tool{
		Name:        "kubens",
		Version:     "0.9.5",
		Description: "Kubernetes namespace switcher",
		Group:       tooldef.GroupKubernetes,
		Format:      tooldef.FormatTarGz,
		URLTemplate: "https://github.com/ahmetb/kubectx/releases/download/v{{.Version}}/kubens_v{{.Version}}_{{.OS}}_{{.Arch}}.tar.gz",
	})

	r.register(&tooldef.Tool{
		Name:        "stern",
		Version:     "1.31.0",
		Description: "Multi-pod log tailing for Kubernetes",
		Group:       tooldef.GroupKubernetes,
		Format:      tooldef.FormatTarGz,
		URLTemplate: "https://github.com/stern/stern/releases/download/v{{.Version}}/stern_{{.Version}}_{{.OS}}_{{.Arch}}.tar.gz",
	})

	r.register(&tooldef.Tool{
		Name:        "argocd",
		Version:     "2.13.3",
		Description: "Argo CD CLI",
		Group:       tooldef.GroupKubernetes,
		Format:      tooldef.FormatBinary,
		URLTemplate: "https://github.com/argoproj/argo-cd/releases/download/v{{.Version}}/argocd-{{.OS}}-{{.Arch}}",
	})

	r.register(&tooldef.Tool{
		Name:        "flux",
		Version:     "2.4.0",
		Description: "Flux CD CLI",
		Group:       tooldef.GroupKubernetes,
		Format:      tooldef.FormatTarGz,
		URLTemplate: "https://github.com/fluxcd/flux2/releases/download/v{{.Version}}/flux_{{.Version}}_{{.OS}}_{{.Arch}}.tar.gz",
	})

	r.register(&tooldef.Tool{
		Name:        "istioctl",
		Version:     "1.24.2",
		Description: "Istio service mesh CLI",
		Group:       tooldef.GroupKubernetes,
		Format:      tooldef.FormatTarGz,
		URLTemplate: "https://github.com/istio/istio/releases/download/{{.Version}}/istioctl-{{.Version}}-{{.OS}}-{{.Arch}}.tar.gz",
	})

	r.register(&tooldef.Tool{
		Name:        "cilium",
		Version:     "0.16.22",
		Description: "Cilium CLI",
		Group:       tooldef.GroupKubernetes,
		Format:      tooldef.FormatTarGz,
		URLTemplate: "https://github.com/cilium/cilium-cli/releases/download/v{{.Version}}/cilium-{{.OS}}-{{.Arch}}.tar.gz",
	})

	r.register(&tooldef.Tool{
		Name:        "kind",
		Version:     "0.25.0",
		Description: "Kubernetes in Docker",
		Group:       tooldef.GroupKubernetes,
		Format:      tooldef.FormatBinary,
		URLTemplate: "https://github.com/kubernetes-sigs/kind/releases/download/v{{.Version}}/kind-{{.OS}}-{{.Arch}}",
	})

	r.register(&tooldef.Tool{
		Name:        "kubeseal",
		Version:     "0.27.3",
		Description: "Sealed Secrets CLI",
		Group:       tooldef.GroupKubernetes,
		Format:      tooldef.FormatTarGz,
		URLTemplate: "https://github.com/bitnami-labs/sealed-secrets/releases/download/v{{.Version}}/kubeseal-{{.Version}}-{{.OS}}-{{.Arch}}.tar.gz",
	})

	r.register(&tooldef.Tool{
		Name:        "velero",
		Version:     "1.15.1",
		Description: "Kubernetes backup and restore",
		Group:       tooldef.GroupKubernetes,
		Format:      tooldef.FormatTarGz,
		BinaryName:  "velero-v{{.Version}}-{{.OS}}-{{.Arch}}/velero",
		URLTemplate: "https://github.com/vmware-tanzu/velero/releases/download/v{{.Version}}/velero-v{{.Version}}-{{.OS}}-{{.Arch}}.tar.gz",
	})
}
