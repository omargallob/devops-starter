// Package tooldef provides the core type definitions shared across devops-starter.
// These types are in a public package (pkg/) to allow external consumers to
// programmatically inspect or extend the tool catalog.
//
// Key types:
//   - Tool: complete metadata for a single installable binary
//   - Platform: OS/architecture pair (e.g., linux/amd64)
//   - Group: functional category for organising tools
//   - ArchiveFormat: download artifact type (tar.gz, zip, raw binary, etc.)
package tooldef

// ArchiveFormat represents the format of a downloaded archive.
type ArchiveFormat string

const (
	FormatTarGz  ArchiveFormat = "tar.gz"
	FormatTarXz  ArchiveFormat = "tar.xz"
	FormatZip    ArchiveFormat = "zip"
	FormatBinary ArchiveFormat = "binary" // Raw binary, no archive
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

const (
	GroupLanguages   Group = "languages"
	GroupContainers  Group = "containers"
	GroupKubernetes  Group = "kubernetes"
	GroupInfra       Group = "infra"
	GroupCloud       Group = "cloud"
	GroupAnsible     Group = "ansible"
	GroupRustTools   Group = "rust-tools"
	GroupUtilities   Group = "utilities"
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

	// URLTemplate is a Go text/template string that produces the download URL.
	// Available template variables: {{.Name}}, {{.Version}}, {{.OS}}, {{.Arch}},
	// {{.Format}}, {{.BinaryName}}
	// Platform-specific overrides in URLs map take precedence.
	URLTemplate string `yaml:"url_template"`

	// URLs provides per-platform download URL overrides.
	// Keys are "os/arch" strings (e.g., "linux/amd64").
	URLs map[string]string `yaml:"urls,omitempty"`

	// BinaryName is the name of the binary inside the archive.
	// If empty, defaults to Tool.Name.
	BinaryName string `yaml:"binary_name,omitempty"`

	// InstallName is what the binary will be named in ~/.local/bin.
	// If empty, defaults to Tool.Name.
	InstallName string `yaml:"install_name,omitempty"`

	// Format is the archive format of the download.
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
	// rather than downloaded directly. The installer will delegate to the
	// manager binary instead of resolving a download URL.
	ManagedBy string `yaml:"managed_by,omitempty"`
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
