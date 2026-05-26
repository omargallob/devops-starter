// link.go provides symlink-based conflict resolution: linking a system binary
// into the managed install directory, and verifying symlink health on subsequent runs.
package installer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// LinkResult describes the outcome of a symlink verification check.
type LinkResult struct {
	ToolName   string
	LinkPath   string
	TargetPath string
	Broken     bool
}

// Link creates a symlink in installDir pointing to the system binary at systemPath.
// If a file already exists at the link destination, it is removed first.
func (inst *Installer) Link(tool *tooldef.Tool, systemPath string) error {
	if inst.DryRun {
		fmt.Printf("[dry-run] Would symlink %s -> %s\n", filepath.Join(inst.InstallDir, tool.GetInstallName()), systemPath)
		return nil
	}

	if err := inst.EnsureDir(); err != nil {
		return fmt.Errorf("ensuring install dir: %w", err)
	}

	linkPath := filepath.Join(inst.InstallDir, tool.GetInstallName())

	// Remove existing file/symlink if present.
	if _, err := os.Lstat(linkPath); err == nil {
		if err := os.Remove(linkPath); err != nil {
			return fmt.Errorf("removing existing file at %s: %w", linkPath, err)
		}
	}

	if err := os.Symlink(systemPath, linkPath); err != nil {
		return fmt.Errorf("creating symlink %s -> %s: %w", linkPath, systemPath, err)
	}

	return nil
}

// VerifyLinks checks all tools with conflict=link to ensure their symlinks
// still point to valid targets. Returns a list of results indicating which
// links are healthy and which are broken.
func VerifyLinks(installDir string, tools []*tooldef.Tool, overrides map[string]ToolConflictGetter) []LinkResult {
	var results []LinkResult

	for _, t := range tools {
		override, ok := overrides[t.Name]
		if !ok || override.GetConflict() != "link" {
			continue
		}

		linkPath := filepath.Join(installDir, t.GetInstallName())

		// Check if it's a symlink.
		fi, err := os.Lstat(linkPath)
		if err != nil {
			// Link doesn't exist at all.
			results = append(results, LinkResult{
				ToolName:   t.Name,
				LinkPath:   linkPath,
				TargetPath: "",
				Broken:     true,
			})
			continue
		}

		if fi.Mode()&os.ModeSymlink == 0 {
			// Not a symlink (maybe a regular binary from a previous overwrite).
			continue
		}

		target, err := os.Readlink(linkPath)
		if err != nil {
			results = append(results, LinkResult{
				ToolName:   t.Name,
				LinkPath:   linkPath,
				TargetPath: "",
				Broken:     true,
			})
			continue
		}

		// Verify target exists.
		if _, err := os.Stat(target); err != nil {
			results = append(results, LinkResult{
				ToolName:   t.Name,
				LinkPath:   linkPath,
				TargetPath: target,
				Broken:     true,
			})
		} else {
			results = append(results, LinkResult{
				ToolName:   t.Name,
				LinkPath:   linkPath,
				TargetPath: target,
				Broken:     false,
			})
		}
	}

	return results
}

// ToolConflictGetter is a minimal interface for reading conflict config.
// This avoids importing the config package directly.
type ToolConflictGetter interface {
	GetConflict() string
}
