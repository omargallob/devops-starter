package registry

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// PluginEntry pairs a validated tool with the plugin file it was loaded from.
type PluginEntry struct {
	Tool     *tooldef.Tool
	FilePath string
}

// pluginFile is the on-disk structure of a plugin YAML file.
type pluginFile struct {
	Tools []*tooldef.Tool `yaml:"tools"`
}

// ValidatePluginTool checks a tool definition from a plugin file.
// install_mode 'custom' and 'post_install' are disallowed as they permit
// arbitrary shell execution; those belong in the Phase 2 Go SDK.
func ValidatePluginTool(t *tooldef.Tool) error {
	if t.Name == "" {
		return fmt.Errorf("tool has empty name")
	}
	if t.Version == "" {
		return fmt.Errorf("tool %q: missing version", t.Name)
	}
	if t.Group == "" {
		return fmt.Errorf("tool %q: missing group", t.Name)
	}
	if !isValidGroup(t.Group) {
		return fmt.Errorf("tool %q: unknown group %q", t.Name, t.Group)
	}
	if t.InstallMode == tooldef.InstallModeCustom {
		return fmt.Errorf("tool %q: install_mode 'custom' is not allowed in plugin files; use the Go SDK for custom install logic", t.Name)
	}
	if t.PostInstall != "" {
		return fmt.Errorf("tool %q: 'post_install' is not allowed in plugin files", t.Name)
	}

	mode := t.EffectiveInstallMode()
	switch mode {
	case tooldef.InstallModeEget:
		if t.Repo == "" {
			return fmt.Errorf("tool %q: install_mode 'eget' requires 'repo'", t.Name)
		}
	case tooldef.InstallModeEgetURL:
		if t.URLTemplate == "" && len(t.URLs) == 0 {
			return fmt.Errorf("tool %q: install_mode 'eget-url' requires 'url_template' or 'urls'", t.Name)
		}
	case tooldef.InstallModeMise:
		// No URL required.
	case tooldef.InstallModeGhExtension:
		if t.Repo == "" {
			return fmt.Errorf("tool %q: install_mode 'gh-extension' requires 'repo'", t.Name)
		}
	default:
		return fmt.Errorf("tool %q: unknown install_mode %q", t.Name, t.InstallMode)
	}

	return nil
}

// isValidGroup reports whether g is a known tool group.
func isValidGroup(g tooldef.Group) bool {
	switch g {
	case tooldef.GroupLanguages,
		tooldef.GroupContainers,
		tooldef.GroupKubernetes,
		tooldef.GroupInfra,
		tooldef.GroupCloud,
		tooldef.GroupAnsible,
		tooldef.GroupRustTools,
		tooldef.GroupUtilities,
		tooldef.GroupAI:
		return true
	}
	return false
}

// LoadPluginFile reads and validates a single plugin YAML file.
// It returns one PluginEntry per valid tool and one error per invalid tool.
func LoadPluginFile(path string) ([]PluginEntry, []error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, []error{fmt.Errorf("reading %s: %w", path, err)}
	}

	var pf pluginFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		return nil, []error{fmt.Errorf("parsing %s: %w", path, err)}
	}

	var entries []PluginEntry
	var errs []error
	for _, t := range pf.Tools {
		if err := ValidatePluginTool(t); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", filepath.Base(path), err))
			continue
		}
		entries = append(entries, PluginEntry{Tool: t, FilePath: path})
	}
	return entries, errs
}

// LoadPluginDirs scans each directory for *.yaml files and loads them in order.
// Later directories take precedence: if the same tool name appears in multiple
// files, the last occurrence wins. Callers should pass lower-precedence dirs first.
func LoadPluginDirs(dirs []string) ([]PluginEntry, []error) {
	seen := make(map[string]int) // tool name → index in entries slice
	var entries []PluginEntry
	var allErrs []error

	for _, dir := range dirs {
		matches, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
		if err != nil {
			allErrs = append(allErrs, fmt.Errorf("scanning %s: %w", dir, err))
			continue
		}
		for _, path := range matches {
			loaded, errs := LoadPluginFile(path)
			allErrs = append(allErrs, errs...)
			for _, e := range loaded {
				if idx, exists := seen[e.Tool.Name]; exists {
					entries[idx] = e // later path takes precedence
				} else {
					seen[e.Tool.Name] = len(entries)
					entries = append(entries, e)
				}
			}
		}
	}
	return entries, allErrs
}

// standardPluginDirs returns the default plugin discovery directories in
// precedence order (lowest first): project-local, then user-global.
func standardPluginDirs() []string {
	var dirs []string

	if wd, err := os.Getwd(); err == nil {
		dirs = append(dirs, filepath.Join(wd, ".devops-starter", "plugins"))
	}

	configBase := os.Getenv("XDG_CONFIG_HOME")
	if configBase == "" {
		if home, err := os.UserHomeDir(); err == nil {
			configBase = filepath.Join(home, ".config")
		}
	}
	if configBase != "" {
		dirs = append(dirs, filepath.Join(configBase, "devops-starter", "plugins"))
	}

	return dirs
}

// RegisterPlugins adds plugin entries to the registry.
// Entries whose names clash with built-ins are skipped with a warning to stderr.
func (r *Registry) RegisterPlugins(entries []PluginEntry) {
	for _, e := range entries {
		if _, exists := r.tools[e.Tool.Name]; exists {
			fmt.Fprintf(os.Stderr, "warning: plugin tool %q conflicts with a built-in and will be skipped\n", e.Tool.Name)
			continue
		}
		r.tools[e.Tool.Name] = e.Tool
		r.plugins = append(r.plugins, e)
	}
}

// PluginEntries returns all successfully registered plugin tools with their source files.
func (r *Registry) PluginEntries() []PluginEntry {
	return r.plugins
}
