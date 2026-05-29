// kubernetes.go registers Kubernetes ecosystem tools: cluster management CLIs,
// package managers, GitOps controllers, service mesh, development frameworks,
// and backup utilities.
package registry

import "github.com/omargallob/devops-starter/pkg/tooldef"

// registerKubernetes adds all Kubernetes-related tools to the registry.
//
// Subgroups:
//   - Core CLIs: kubectl, helm, kustomize
//   - Cluster Management: k9s, kubectx, kubens, stern, lens (headless)
//   - GitOps: argocd, flux
//   - Networking: istioctl, cilium
//   - Development: kind, kubebuilder, operator-sdk, skaffold, tilt
//   - Security: kubeseal, kyverno
//   - Backup & Recovery: velero
func registerKubernetes(r *Registry) {
	// ── Core CLIs ─────────────────────────────────────────────────────────────

	r.register(&tooldef.Tool{
		Name:        "kubectl",
		Version:     "1.31.4",
		Description: "Kubernetes CLI",
		Group:       tooldef.GroupKubernetes,
		Subgroup:    "Core CLIs",
		InstallMode: tooldef.InstallModeEgetURL,
		Format:      tooldef.FormatBinary,
		URLTemplate: "https://dl.k8s.io/release/v{{.Version}}/bin/{{.OS}}/{{.Arch}}/kubectl",
	})

	r.register(&tooldef.Tool{
		Name:        "helm",
		Version:     "3.16.4",
		Description: "Kubernetes package manager",
		Group:       tooldef.GroupKubernetes,
		Subgroup:    "Core CLIs",
		InstallMode: tooldef.InstallModeEgetURL,
		Format:      tooldef.FormatTarGz,
		BinaryName:  "helm",
		URLTemplate: "https://get.helm.sh/helm-v{{.Version}}-{{.OS}}-{{.Arch}}.tar.gz",
	})

	r.register(&tooldef.Tool{
		Name:        "kustomize",
		Version:     "5.5.0",
		Description: "Kubernetes configuration management",
		Group:       tooldef.GroupKubernetes,
		Subgroup:    "Core CLIs",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "kubernetes-sigs/kustomize",
		Asset:       "kustomize_*_{{.OS}}_{{.Arch}}*",
	})

	// ── Cluster Management ─────────────────────────────────────────────────────

	r.register(&tooldef.Tool{
		Name:        "k9s",
		Version:     "0.32.7",
		Description: "Kubernetes TUI",
		Group:       tooldef.GroupKubernetes,
		Subgroup:    "Cluster Management",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "derailed/k9s",
	})

	r.register(&tooldef.Tool{
		Name:        "kubectx",
		Version:     "0.9.5",
		Description: "Kubernetes context switcher",
		Group:       tooldef.GroupKubernetes,
		Subgroup:    "Cluster Management",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "ahmetb/kubectx",
		Asset:       "kubectx_*",
	})

	r.register(&tooldef.Tool{
		Name:        "kubens",
		Version:     "0.9.5",
		Description: "Kubernetes namespace switcher",
		Group:       tooldef.GroupKubernetes,
		Subgroup:    "Cluster Management",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "ahmetb/kubectx",
		Asset:       "kubens_*",
	})

	r.register(&tooldef.Tool{
		Name:        "stern",
		Version:     "1.31.0",
		Description: "Multi-pod log tailing for Kubernetes",
		Group:       tooldef.GroupKubernetes,
		Subgroup:    "Cluster Management",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "stern/stern",
	})

	// ── GitOps ────────────────────────────────────────────────────────────────

	r.register(&tooldef.Tool{
		Name:        "argocd",
		Version:     "2.13.3",
		Description: "Argo CD CLI",
		Group:       tooldef.GroupKubernetes,
		Subgroup:    "GitOps",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "argoproj/argo-cd",
		Asset:       "argocd-{{.OS}}-{{.Arch}}",
	})

	r.register(&tooldef.Tool{
		Name:        "flux",
		Version:     "2.4.0",
		Description: "Flux CD CLI",
		Group:       tooldef.GroupKubernetes,
		Subgroup:    "GitOps",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "fluxcd/flux2",
	})

	// ── Networking ────────────────────────────────────────────────────────────

	r.register(&tooldef.Tool{
		Name:        "istioctl",
		Version:     "1.24.2",
		Description: "Istio service mesh CLI",
		Group:       tooldef.GroupKubernetes,
		Subgroup:    "Networking",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "istio/istio",
		Asset:       "istioctl-*-{{.OS}}-{{.Arch}}*",
	})

	r.register(&tooldef.Tool{
		Name:        "cilium",
		Version:     "0.16.22",
		Description: "Cilium CLI",
		Group:       tooldef.GroupKubernetes,
		Subgroup:    "Networking",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "cilium/cilium-cli",
		Asset:       "cilium-{{.OS}}-{{.Arch}}*",
	})

	// ── Development ──────────────────────────────────────────────────────────

	r.register(&tooldef.Tool{
		Name:        "kind",
		Version:     "0.25.0",
		Description: "Kubernetes in Docker",
		Group:       tooldef.GroupKubernetes,
		Subgroup:    "Development",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "kubernetes-sigs/kind",
		Asset:       "kind-{{.OS}}-{{.Arch}}",
	})

	r.register(&tooldef.Tool{
		Name:        "kubebuilder",
		Version:     "4.4.0",
		Description: "SDK for building Kubernetes APIs",
		Group:       tooldef.GroupKubernetes,
		Subgroup:    "Development",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "kubernetes-sigs/kubebuilder",
		Asset:       "kubebuilder_{{.OS}}_{{.Arch}}",
	})

	r.register(&tooldef.Tool{
		Name:        "operator-sdk",
		Version:     "1.38.0",
		Description: "Operator framework SDK CLI",
		Group:       tooldef.GroupKubernetes,
		Subgroup:    "Development",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "operator-framework/operator-sdk",
		Asset:       "operator-sdk_{{.OS}}_{{.Arch}}",
	})

	r.register(&tooldef.Tool{
		Name:        "skaffold",
		Version:     "2.14.0",
		Description: "Continuous development for Kubernetes apps",
		Group:       tooldef.GroupKubernetes,
		Subgroup:    "Development",
		InstallMode: tooldef.InstallModeEgetURL,
		Format:      tooldef.FormatBinary,
		URLTemplate: "https://storage.googleapis.com/skaffold/releases/v{{.Version}}/skaffold-{{.OS}}-{{.Arch}}",
	})

	r.register(&tooldef.Tool{
		Name:        "tilt",
		Version:     "0.33.21",
		Description: "Multi-service dev environment for K8s",
		Group:       tooldef.GroupKubernetes,
		Subgroup:    "Development",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "tilt-dev/tilt",
	})

	// ── Security ─────────────────────────────────────────────────────────────

	r.register(&tooldef.Tool{
		Name:        "kubeseal",
		Version:     "0.27.3",
		Description: "Sealed Secrets CLI",
		Group:       tooldef.GroupKubernetes,
		Subgroup:    "Security",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "bitnami-labs/sealed-secrets",
		Asset:       "kubeseal-*-{{.OS}}-{{.Arch}}*",
	})

	r.register(&tooldef.Tool{
		Name:        "kyverno",
		Version:     "1.13.4",
		Description: "Kubernetes policy engine CLI",
		Group:       tooldef.GroupKubernetes,
		Subgroup:    "Security",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "kyverno/kyverno",
		Asset:       "kyverno-cli_*_{{.OS}}_{{.Arch}}*",
		InstallName: "kyverno",
	})

	// ── Backup & Recovery ────────────────────────────────────────────────────

	r.register(&tooldef.Tool{
		Name:        "velero",
		Version:     "1.15.1",
		Description: "Kubernetes backup and restore",
		Group:       tooldef.GroupKubernetes,
		Subgroup:    "Backup & Recovery",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "vmware-tanzu/velero",
		BinaryName:  "velero",
	})
}
