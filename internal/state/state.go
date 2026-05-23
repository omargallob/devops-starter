// state.go defines the ToolState type and status resolution logic used by
// the TUI and plain-text status output.
package state

import (
	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/internal/registry"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// Status represents the install state of a tool.
type Status int

const (
	StatusMissing   Status = iota // Not installed
	StatusCurrent                 // Installed version matches desired
	StatusOutdated                // Installed but version differs from desired
	StatusDisabled                // Disabled in user config
	StatusUnknown                 // Binary exists but version could not be determined
)

// String returns a human-readable status label.
func (s Status) String() string {
	switch s {
	case StatusMissing:
		return "missing"
	case StatusCurrent:
		return "current"
	case StatusOutdated:
		return "outdated"
	case StatusDisabled:
		return "disabled"
	case StatusUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// ToolState holds the computed state of a single tool for display.
type ToolState struct {
	Name             string
	Group            string
	Description      string
	DesiredVersion   string // from registry + config overrides
	InstalledVersion string // from state file or verify
	Status           Status
	Selected         bool // TUI selection state (not persisted)
	Tool             *tooldef.Tool
}

// GroupState holds a group name and its tools for display ordering.
type GroupState struct {
	Name  string
	Tools []ToolState
}

// ResolveAll builds the full state view by combining the registry, config,
// and persisted state. It returns tools grouped in display order.
func ResolveAll(cfg *config.Config, store *Store, plat tooldef.Platform) []GroupState {
	reg := registry.New()

	groups := []tooldef.Group{
		tooldef.GroupLanguages,
		tooldef.GroupContainers,
		tooldef.GroupKubernetes,
		tooldef.GroupInfra,
		tooldef.GroupCloud,
		tooldef.GroupRustTools,
		tooldef.GroupUtilities,
	}

	var result []GroupState

	for _, group := range groups {
		tools := reg.GetByGroup(group)
		if len(tools) == 0 {
			continue
		}

		gs := GroupState{Name: string(group)}

		for _, t := range tools {
			ts := ToolState{
				Name:           t.Name,
				Group:          string(t.Group),
				Description:    t.Description,
				DesiredVersion: t.Version,
				Tool:           t,
			}

			// Apply version override from config
			if override, ok := cfg.Overrides[t.Name]; ok {
				if override.Disabled {
					ts.Status = StatusDisabled
					gs.Tools = append(gs.Tools, ts)
					continue
				}
				if override.Version != "" {
					ts.DesiredVersion = override.Version
				}
			}

			// Check if group is disabled
			if !cfg.IsGroupEnabled(string(t.Group)) {
				ts.Status = StatusDisabled
				gs.Tools = append(gs.Tools, ts)
				continue
			}

			// Check platform support
			if !t.SupportsPlatform(plat) {
				continue // skip entirely for unsupported platforms
			}

			// Resolve installed version from state store
			ts.InstalledVersion = store.GetVersion(t.Name)

			// Determine status
			switch {
			case ts.InstalledVersion == "":
				ts.Status = StatusMissing
			case ts.InstalledVersion == ts.DesiredVersion:
				ts.Status = StatusCurrent
			case ts.InstalledVersion == "unknown":
				ts.Status = StatusUnknown
			default:
				ts.Status = StatusOutdated
			}

			gs.Tools = append(gs.Tools, ts)
		}

		if len(gs.Tools) > 0 {
			result = append(result, gs)
		}
	}

	return result
}
