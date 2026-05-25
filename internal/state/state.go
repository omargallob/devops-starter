// state.go defines the ToolState type and status resolution logic used by
// the TUI and plain-text status output.
package state

import (
	"sort"
	"strings"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/internal/registry"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// Status represents the install state of a tool.
type Status int

const (
	StatusMissing     Status = iota // Not installed
	StatusCurrent                   // Installed version matches desired
	StatusOutdated                  // Installed but version differs from desired
	StatusDisabled                  // Disabled in user config
	StatusUnknown                   // Binary exists but version could not be determined
	StatusDetected                  // Binary found in PATH but not managed by devops-starter
	StatusUnavailable               // Not available on the current platform
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
	case StatusDetected:
		return "detected"
	case StatusUnavailable:
		return "unavailable"
	default:
		return "unknown"
	}
}

// Source indicates how a tool is managed.
type Source string

const (
	SourceManaged Source = "managed" // installed/managed by devops-starter
	SourceMise    Source = "mise"    // managed by mise
	SourceSystem  Source = "system"  // found in PATH, not managed
	SourceNone    Source = ""        // not installed / unknown
)

// ToolState holds the computed state of a single tool for display.
type ToolState struct {
	Name             string
	Group            string
	Subgroup         string // optional visual sub-category (e.g., "Platforms", "Languages")
	Description      string
	DesiredVersion   string // from registry + config overrides
	InstalledVersion string // from state file or verify
	DetectedPath     string // path to system binary (when StatusDetected)
	DetectedVersion  string // version of system binary (when StatusDetected)
	Status           Status
	Source           Source // how the tool is managed (mise, managed, system)
	Selected         bool   // TUI selection state (not persisted)
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
				Subgroup:       t.Subgroup,
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
				ts.Status = StatusUnavailable
				gs.Tools = append(gs.Tools, ts)
				continue
			}

			// Resolve installed version from state store
			ts.InstalledVersion = store.GetVersion(t.Name)

			// Determine status and source
			switch ts.InstalledVersion {
			case "":
				// Not in state file — check if binary exists in PATH
				if path := LookupInPath(t.Name); path != "" {
					if ver, err := DetectVersionAtPath(t.Name, path); err == nil {
						// For mise-managed tools, treat PATH detection as installed
						// (they won't be in the state store since mise manages them).
						if t.ManagedBy != "" {
							ts.InstalledVersion = ver
							ts.Source = SourceMise
							if versionMatches(ver, ts.DesiredVersion) {
								ts.Status = StatusCurrent
							} else {
								ts.Status = StatusOutdated
							}
						} else {
							ts.Status = StatusDetected
							ts.Source = SourceSystem
							ts.DetectedPath = path
							ts.DetectedVersion = ver
						}
					} else if t.ManagedBy != "" {
						// Binary exists but version probe failed — still mark as detected
						ts.Status = StatusUnknown
						ts.Source = SourceMise
						ts.DetectedPath = path
					} else {
						ts.Status = StatusDetected
						ts.Source = SourceSystem
						ts.DetectedPath = path
					}
				} else {
					ts.Status = StatusMissing
				}
			case ts.DesiredVersion:
				ts.Status = StatusCurrent
				ts.Source = SourceManaged
			case "unknown":
				ts.Status = StatusUnknown
				ts.Source = SourceManaged
			default:
				ts.Status = StatusOutdated
				ts.Source = SourceManaged
			}

			gs.Tools = append(gs.Tools, ts)
		}

		if len(gs.Tools) > 0 {
			// Sort tools by subgroup (Platforms before Languages),
			// then alphabetically within each subgroup.
			sort.SliceStable(gs.Tools, func(i, j int) bool {
				si := subgroupOrder(gs.Tools[i].Subgroup)
				sj := subgroupOrder(gs.Tools[j].Subgroup)
				if si != sj {
					return si < sj
				}
				return gs.Tools[i].Name < gs.Tools[j].Name
			})
			result = append(result, gs)
		}
	}

	return result
}

// versionMatches checks if an installed version satisfies the desired version
// specification. Mise allows partial versions (e.g., "3.12" means any 3.12.x,
// "22" means any 22.x.y), so we use prefix matching on dot-separated segments.
func versionMatches(installed, desired string) bool {
	if installed == desired {
		return true
	}
	// Prefix match: "3.12" should match "3.12.7"
	// Split into segments and compare prefix.
	dParts := strings.Split(desired, ".")
	iParts := strings.Split(installed, ".")
	if len(dParts) > len(iParts) {
		return false
	}
	for i, dp := range dParts {
		if dp != iParts[i] {
			return false
		}
	}
	return true
}

// subgroupOrder returns a sort key for subgroups. "Platforms" sorts first (0),
// "Languages" second (1), empty/unknown last (2).
func subgroupOrder(subgroup string) int {
	switch subgroup {
	case "Platforms":
		return 0
	case "Languages":
		return 1
	default:
		return 2
	}
}
