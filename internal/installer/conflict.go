// conflict.go detects tools already present on the system PATH that would
// conflict with managed installations, and provides resolution strategies.
package installer

import (
	"path/filepath"
	"strings"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/internal/state"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// ConflictInfo describes a tool that already exists on the system PATH
// outside the managed install directory.
type ConflictInfo struct {
	// Tool is the registry definition of the tool to be installed.
	Tool *tooldef.Tool
	// SystemPath is the absolute path to the existing system binary.
	SystemPath string
	// SystemVersion is the detected version of the system binary (may be empty if detection fails).
	SystemVersion string
	// WantedVersion is the version that would be installed from the registry.
	WantedVersion string
	// SavedAction is the conflict resolution action stored in config (may be empty).
	SavedAction string
}

// DetectConflicts checks each tool in the list against the system PATH.
// A tool is considered to have a conflict when:
//   - It is found on PATH via exec.LookPath
//   - The resolved path is NOT inside installDir (i.e., it's a system-managed binary)
//
// Tools that are already installed in installDir are not conflicts — they are
// simply upgrades/re-installs of previously managed tools.
func DetectConflicts(tools []*tooldef.Tool, installDir string, overrides map[string]config.ToolOverride) []ConflictInfo {
	absInstallDir, _ := filepath.Abs(installDir)
	var conflicts []ConflictInfo

	for _, t := range tools {
		binName := t.GetInstallName()
		systemPath := state.LookupInPath(binName)
		if systemPath == "" {
			continue
		}

		// Resolve to absolute path for comparison.
		absSystemPath, err := filepath.Abs(systemPath)
		if err != nil {
			absSystemPath = systemPath
		}

		// If the binary lives inside our install dir, it's ours — not a conflict.
		if isInsideDir(absSystemPath, absInstallDir) {
			continue
		}

		// Detect version (best-effort).
		version, _ := state.DetectVersionAtPath(t.Name, systemPath)

		info := ConflictInfo{
			Tool:          t,
			SystemPath:    systemPath,
			SystemVersion: version,
			WantedVersion: t.Version,
		}

		// Attach saved conflict action from config.
		if override, ok := overrides[t.Name]; ok && override.Conflict != "" {
			info.SavedAction = override.Conflict
		}

		conflicts = append(conflicts, info)
	}

	return conflicts
}

// UnresolvedConflicts returns only the conflicts that have no saved action in config.
func UnresolvedConflicts(conflicts []ConflictInfo) []ConflictInfo {
	var unresolved []ConflictInfo
	for _, c := range conflicts {
		if c.SavedAction == "" {
			unresolved = append(unresolved, c)
		}
	}
	return unresolved
}

// ApplyConflictActions partitions tools into those to install, those to skip,
// and those to link, based on conflict info and resolutions.
// The resolutions map keys are tool names, values are "skip", "overwrite", or "link".
func ApplyConflictActions(tools []*tooldef.Tool, conflicts []ConflictInfo, resolutions map[string]string) (install, skip []*tooldef.Tool, links map[string]string) {
	links = make(map[string]string) // toolName -> systemPath

	// Build lookup from conflicts.
	conflictMap := make(map[string]ConflictInfo, len(conflicts))
	for _, c := range conflicts {
		conflictMap[c.Tool.Name] = c
	}

	for _, t := range tools {
		conflict, hasConflict := conflictMap[t.Name]
		if !hasConflict {
			install = append(install, t)
			continue
		}

		// Determine action: explicit resolution > saved action > default (overwrite).
		action := "overwrite"
		if resolution, ok := resolutions[t.Name]; ok {
			action = resolution
		} else if conflict.SavedAction != "" {
			action = conflict.SavedAction
		}

		switch action {
		case "skip":
			skip = append(skip, t)
		case "link":
			links[t.Name] = conflict.SystemPath
			skip = append(skip, t) // don't install, will link instead
		default: // "overwrite"
			install = append(install, t)
		}
	}

	return install, skip, links
}

// isInsideDir checks whether path is inside dir (or is dir itself).
func isInsideDir(path, dir string) bool {
	// Ensure dir ends with separator for prefix matching.
	if !strings.HasSuffix(dir, string(filepath.Separator)) {
		dir += string(filepath.Separator)
	}
	return strings.HasPrefix(path, dir) || path == strings.TrimSuffix(dir, string(filepath.Separator))
}

// LookupToolInPath finds the tool's binary on the system PATH using the
// state package's probe-aware lookup.
func LookupToolInPath(tool *tooldef.Tool) string {
	return state.LookupInPath(tool.Name)
}
