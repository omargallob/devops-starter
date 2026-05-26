// Package tooldef provides the core type definitions shared across devops-starter.
// These types are in a public package (pkg/) to allow external consumers to
// programmatically inspect or extend the tool catalog.
//
// Key types:
//   - Tool: complete metadata for a single installable binary
//   - Platform: OS/architecture pair (e.g., linux/amd64)
//   - Group: functional category for organising tools
//   - ArchiveFormat: download artifact type (tar.gz, zip, raw binary, etc.)
//   - InstallMode: how a tool is installed (eget, eget-url, mise, custom)
package tooldef

// ArchiveFormat represents the format of a downloaded archive.
type ArchiveFormat string

// ArchiveFormat identifies the packaging format of a tool download.
const (
	FormatTarGz  ArchiveFormat = "tar.gz"
	FormatTarXz  ArchiveFormat = "tar.xz"
	FormatZip    ArchiveFormat = "zip"
	FormatBinary ArchiveFormat = "binary" // Raw binary, no archive
)

// InstallMode determines how a tool is installed.
type InstallMode string

const (
	// InstallModeEget installs from a GitHub release via eget <owner/repo>.
	InstallModeEget InstallMode = "eget"

	// InstallModeEgetURL installs from a direct URL via eget <url>.
	InstallModeEgetURL InstallMode = "eget-url"

	// InstallModeCustom uses the legacy download/extract/install pipeline.
	// Reserved for tools requiring post-install scripts (e.g., aws-cli, gcloud).
	InstallModeCustom InstallMode = "custom"

	// InstallModeMise delegates installation to the mise tool manager.
	InstallModeMise InstallMode = "mise"
)

// Platform represents a target OS/architecture combination.
type Platform struct {
	OS   string // "linux" or "darwin"
	Arch string // "amd64" or "arm64"
}

func (p Platform) String() string {
	return p.OS + "/" + p.Arch
}

// Group categorizes tools by their function.
type Group string

// Group values enumerate tool categories.
const (
	GroupLanguages  Group = "languages"
	GroupContainers Group = "containers"
	GroupKubernetes Group = "kubernetes"
	GroupInfra      Group = "infra"
	GroupCloud      Group = "cloud"
	GroupAnsible    Group = "ansible"
	GroupRustTools  Group = "rust-tools"
	GroupUtilities  Group = "utilities"
)

// Tool defines a single installable tool with all metadata needed to
// download and install it across multiple platforms.
type Tool struct {
	// Name is the canonical name of the tool (e.g., "kubectl", "bat").
	Name string `yaml:"name"`

	// Version is the pinned version to install (e.g., "1.29.3").
	Version string `yaml:"version"`

	// Description is a short human-readable description.
	Description string `yaml:"description"`

	// Group is the functional category this tool belongs to.
	Group Group `yaml:"group"`

	// InstallMode determines how this tool is installed.
	// See InstallMode constants for allowed values.
	InstallMode InstallMode `yaml:"install_mode"`

	// Repo is the GitHub owner/repo for eget repo mode (e.g., "derailed/k9s").
	// Used when InstallMode is InstallModeEget.
	Repo string `yaml:"repo,omitempty"`

	// Asset is an eget --asset glob pattern to select the correct release asset
	// (e.g., "*.tar.gz", "kustomize_*"). Used when InstallMode is InstallModeEget.
	Asset string `yaml:"asset,omitempty"`

	// URLTemplate is a Go text/template string that produces the download URL.
	// Available template variables: {{.Name}}, {{.Version}}, {{.OS}}, {{.Arch}},
	// {{.Format}}, {{.BinaryName}}
	// Platform-specific overrides in URLs map take precedence.
	// Used when InstallMode is InstallModeEgetURL or InstallModeCustom.
	URLTemplate string `yaml:"url_template"`

	// URLs provides per-platform download URL overrides.
	// Keys are "os/arch" strings (e.g., "linux/amd64").
	// Used when InstallMode is InstallModeEgetURL or InstallModeCustom.
	URLs map[string]string `yaml:"urls,omitempty"`

	// BinaryName is the name of the binary inside the archive.
	// If empty, defaults to Tool.Name.
	BinaryName string `yaml:"binary_name,omitempty"`

	// InstallName is what the binary will be named in ~/.local/bin.
	// If empty, defaults to Tool.Name.
	InstallName string `yaml:"install_name,omitempty"`

	// Format is the archive format of the download.
	// Used by InstallModeCustom and InstallModeEgetURL (for non-eget extraction).
	Format ArchiveFormat `yaml:"format"`

	// StripComponents is the number of leading path components to strip
	// when extracting from an archive (like tar --strip-components).
	StripComponents int `yaml:"strip_components,omitempty"`

	// Checksums maps "os/arch" to expected SHA256 hex digest.
	Checksums map[string]string `yaml:"checksums,omitempty"`

	// Platforms lists which platforms this tool supports.
	// If nil, all platforms are assumed supported.
	Platforms []Platform `yaml:"platforms,omitempty"`

	// PostInstall is an optional shell command to run after installation.
	// Runs with the installed binary on PATH.
	PostInstall string `yaml:"post_install,omitempty"`

	// Dependencies lists other tool names that must be installed first.
	Dependencies []string `yaml:"dependencies,omitempty"`

	// ManagedBy indicates this tool is installed by another tool (e.g., "mise")
	// rather than downloaded directly.
	// Deprecated: Use InstallMode instead.
	ManagedBy string `yaml:"managed_by,omitempty"`

	// Subgroup provides an optional visual sub-category within a group.
	// Used for display purposes only (e.g., "Platforms" vs "Languages").
	Subgroup string `yaml:"subgroup,omitempty"`
}

// GetBinaryName returns the binary name, defaulting to Tool.Name.
func (t *Tool) GetBinaryName() string {
	if t.BinaryName != "" {
		return t.BinaryName
	}
	return t.Name
}

// GetInstallName returns the install name, defaulting to Tool.Name.
func (t *Tool) GetInstallName() string {
	if t.InstallName != "" {
		return t.InstallName
	}
	return t.Name
}

// SupportsPlatform checks if this tool supports the given platform.
func (t *Tool) SupportsPlatform(p Platform) bool {
	if len(t.Platforms) == 0 {
		return true
	}
	for _, supported := range t.Platforms {
		if supported.OS == p.OS && supported.Arch == p.Arch {
			return true
		}
	}
	return false
}

// EffectiveInstallMode returns the install mode, falling back to ManagedBy
// for backward compatibility during migration.
func (t *Tool) EffectiveInstallMode() InstallMode {
	if t.InstallMode != "" {
		return t.InstallMode
	}
	// Backward compat: if ManagedBy is set but InstallMode isn't, derive it.
	if t.ManagedBy != "" {
		return InstallModeMise
	}
	// Default: custom (legacy path).
	return InstallModeCustom
}
