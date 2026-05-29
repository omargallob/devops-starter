// Package config handles loading, saving, and managing the devops-starter
// YAML configuration file. The config lives at ~/.config/devops-starter/config.yaml
// (or $XDG_CONFIG_HOME/devops-starter/config.yaml) and controls which tool
// groups are enabled, per-tool version overrides, and the binary install directory.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the user's devops-starter configuration.
type Config struct {
	// InstallDir is where binaries are installed. Defaults to ~/.local/bin.
	InstallDir string `yaml:"install_dir"`

	// Groups controls which tool groups are enabled.
	Groups GroupConfig `yaml:"groups"`

	// Overrides allows per-tool version pinning.
	Overrides map[string]ToolOverride `yaml:"overrides,omitempty"`

	// PluginPaths lists additional directories to scan for plugin YAML files.
	// These take precedence over the standard project-local and user-global dirs.
	PluginPaths []string `yaml:"plugin_paths,omitempty"`
}

// GroupConfig toggles tool groups on/off.
type GroupConfig struct {
	Languages  bool `yaml:"languages"`
	Containers bool `yaml:"containers"`
	Kubernetes bool `yaml:"kubernetes"`
	Infra      bool `yaml:"infra"`
	Cloud      bool `yaml:"cloud"`
	Ansible    bool `yaml:"ansible"`
	RustTools  bool `yaml:"rust_tools"`
	Utilities  bool `yaml:"utilities"`
	AI         bool `yaml:"ai"`
}

// ConflictAction defines how to handle a tool already present on the system.
type ConflictAction string

const (
	// ConflictSkip keeps the existing system binary and does not install.
	ConflictSkip ConflictAction = "skip"
	// ConflictOverwrite installs the managed version, shadowing the system binary.
	ConflictOverwrite ConflictAction = "overwrite"
	// ConflictLink creates a symlink in install_dir pointing to the system binary.
	ConflictLink ConflictAction = "link"
)

// ValidConflictActions returns all valid conflict action values.
func ValidConflictActions() []ConflictAction {
	return []ConflictAction{ConflictSkip, ConflictOverwrite, ConflictLink}
}

// IsValidConflictAction checks whether a string is a valid conflict action.
func IsValidConflictAction(s string) bool {
	switch ConflictAction(s) {
	case ConflictSkip, ConflictOverwrite, ConflictLink:
		return true
	default:
		return s == "" // empty is valid (means "unset / prompt")
	}
}

// ToolOverride allows overriding version, disabling, or setting conflict
// resolution policy for a specific tool.
type ToolOverride struct {
	Version  string `yaml:"version,omitempty"`
	Disabled bool   `yaml:"disabled,omitempty"`
	Conflict string `yaml:"conflict,omitempty"` // "skip", "overwrite", or "link"
}

// DefaultConfig returns the default configuration with all groups enabled.
func DefaultConfig() *Config {
	return &Config{
		InstallDir: filepath.Join(homeDir(), ".local", "bin"),
		Groups: GroupConfig{
			Languages:  true,
			Containers: true,
			Kubernetes: true,
			Infra:      true,
			Cloud:      true,
			Ansible:    true,
			RustTools:  true,
			Utilities:  true,
			AI:         true,
		},
		Overrides: make(map[string]ToolOverride),
	}
}

// Path returns the default config file path.
func Path() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "devops-starter", "config.yaml")
	}
	return filepath.Join(homeDir(), ".config", "devops-starter", "config.yaml")
}

// Load reads configuration from the given path. If the file doesn't exist,
// it returns the default configuration.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	// Expand ~ in install_dir
	if cfg.InstallDir == "~/.local/bin" || cfg.InstallDir == "" {
		cfg.InstallDir = filepath.Join(homeDir(), ".local", "bin")
	}

	return cfg, nil
}

// Save writes the configuration to the given path, creating parent directories.
func Save(cfg *Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

// IsGroupEnabled checks if a group name string is enabled in the config.
func (c *Config) IsGroupEnabled(group string) bool {
	switch group {
	case "languages":
		return c.Groups.Languages
	case "containers":
		return c.Groups.Containers
	case "kubernetes":
		return c.Groups.Kubernetes
	case "infra":
		return c.Groups.Infra
	case "cloud":
		return c.Groups.Cloud
	case "ansible":
		return c.Groups.Ansible
	case "rust-tools", "rust_tools":
		return c.Groups.RustTools
	case "utilities":
		return c.Groups.Utilities
	case "ai":
		return c.Groups.AI
	default:
		return false
	}
}

// SetGroup enables or disables a group by name string.
func (c *Config) SetGroup(group string, enabled bool) {
	switch group {
	case "languages":
		c.Groups.Languages = enabled
	case "containers":
		c.Groups.Containers = enabled
	case "kubernetes":
		c.Groups.Kubernetes = enabled
	case "infra":
		c.Groups.Infra = enabled
	case "cloud":
		c.Groups.Cloud = enabled
	case "ansible":
		c.Groups.Ansible = enabled
	case "rust-tools", "rust_tools":
		c.Groups.RustTools = enabled
	case "utilities":
		c.Groups.Utilities = enabled
	case "ai":
		c.Groups.AI = enabled
	}
}

// MergeGroups applies group selections from a map into the config without
// modifying the Overrides map. This preserves per-tool version pins and
// disable flags when the user re-runs the setup wizard.
func (c *Config) MergeGroups(groups map[string]bool) {
	for group, enabled := range groups {
		c.SetGroup(group, enabled)
	}
}

// AllGroupNames returns the ordered list of group name strings.
func AllGroupNames() []string {
	return []string{
		"languages",
		"containers",
		"kubernetes",
		"infra",
		"cloud",
		"ansible",
		"rust-tools",
		"utilities",
		"ai",
	}
}

// homeDir returns the user's home directory, falling back to $HOME if
// os.UserHomeDir fails (e.g., in minimal container environments).
func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return os.Getenv("HOME")
	}
	return home
}
