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
		InstallMode: tooldef.InstallModeEgetURL,
		Format:      tooldef.FormatBinary,
		URLTemplate: "https://dl.k8s.io/release/v{{.Version}}/bin/{{.OS}}/{{.Arch}}/kubectl",
	})

	r.register(&tooldef.Tool{
		Name:        "helm",
		Version:     "3.16.4",
		Description: "Kubernetes package manager",
		Group:       tooldef.GroupKubernetes,
		InstallMode: tooldef.InstallModeEgetURL,
		Format:      tooldef.FormatTarGz,
		BinaryName:  "helm",
		URLTemplate: "https://get.helm.sh/helm-v{{.Version}}-{{.OS}}-{{.Arch}}.tar.gz",
	})

	r.register(&tooldef.Tool{
		Name:        "kustomize",
		Version:     "5.8.1",
		Description: "Kubernetes configuration management",
		Group:       tooldef.GroupKubernetes,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "kubernetes-sigs/kustomize",
		Asset:       "kustomize_*_{{.OS}}_{{.Arch}}*",
	})

	r.register(&tooldef.Tool{
		Name:        "k9s",
		Version:     "0.50.18",
		Description: "Kubernetes TUI",
		Group:       tooldef.GroupKubernetes,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "derailed/k9s",
	})

	r.register(&tooldef.Tool{
		Name:        "kubectx",
		Version:     "0.11.0",
		Description: "Kubernetes context switcher",
		Group:       tooldef.GroupKubernetes,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "ahmetb/kubectx",
		Asset:       "kubectx_*",
	})

	r.register(&tooldef.Tool{
		Name:        "kubens",
		Version:     "0.11.0",
		Description: "Kubernetes namespace switcher",
		Group:       tooldef.GroupKubernetes,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "ahmetb/kubectx",
		Asset:       "kubens_*",
	})

	r.register(&tooldef.Tool{
		Name:        "stern",
		Version:     "1.34.0",
		Description: "Multi-pod log tailing for Kubernetes",
		Group:       tooldef.GroupKubernetes,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "stern/stern",
	})

	r.register(&tooldef.Tool{
		Name:        "argocd",
		Version:     "3.4.3",
		Description: "Argo CD CLI",
		Group:       tooldef.GroupKubernetes,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "argoproj/argo-cd",
		Asset:       "argocd-{{.OS}}-{{.Arch}}",
	})

	r.register(&tooldef.Tool{
		Name:        "flux",
		Version:     "2.8.8",
		Description: "Flux CD CLI",
		Group:       tooldef.GroupKubernetes,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "fluxcd/flux2",
	})

	r.register(&tooldef.Tool{
		Name:        "istioctl",
		Version:     "1.30.0",
		Description: "Istio service mesh CLI",
		Group:       tooldef.GroupKubernetes,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "istio/istio",
		Asset:       "istioctl-*-{{.OS}}-{{.Arch}}*",
	})

	r.register(&tooldef.Tool{
		Name:        "cilium",
		Version:     "0.19.4",
		Description: "Cilium CLI",
		Group:       tooldef.GroupKubernetes,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "cilium/cilium-cli",
		Asset:       "cilium-{{.OS}}-{{.Arch}}*",
	})

	r.register(&tooldef.Tool{
		Name:        "kind",
		Version:     "0.31.0",
		Description: "Kubernetes in Docker",
		Group:       tooldef.GroupKubernetes,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "kubernetes-sigs/kind",
		Asset:       "kind-{{.OS}}-{{.Arch}}",
	})

	r.register(&tooldef.Tool{
		Name:        "kubeseal",
		Version:     "0.37.0",
		Description: "Sealed Secrets CLI",
		Group:       tooldef.GroupKubernetes,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "bitnami-labs/sealed-secrets",
		Asset:       "kubeseal-*-{{.OS}}-{{.Arch}}*",
	})

	r.register(&tooldef.Tool{
		Name:        "velero",
		Version:     "1.18.1",
		Description: "Kubernetes backup and restore",
		Group:       tooldef.GroupKubernetes,
		InstallMode: tooldef.InstallModeEget,
		Repo:        "vmware-tanzu/velero",
		BinaryName:  "velero",
	})
}
