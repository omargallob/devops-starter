// deps.go provides pre-install dependency checking for tools that declare
// required dependencies via the Dependencies field. A dependency is satisfied
// if its binary exists in the install directory or is found on the system PATH
// (using the probe's BinName mapping for tools whose binary name differs from
// the tool name, e.g., neovim → nvim).
package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/omargallob/devops-starter/internal/state"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// checkDependencies verifies that all declared dependencies for a tool are
// available, either in the install directory or on the system PATH.
// Returns an error listing all missing dependencies, or nil if all are satisfied.
func (inst *Installer) checkDependencies(tool *tooldef.Tool) error {
	if len(tool.Dependencies) == 0 {
		return nil
	}

	var missing []string
	for _, dep := range tool.Dependencies {
		if inst.isDependencyAvailable(dep) {
			continue
		}
		missing = append(missing, dep)
	}

	if len(missing) > 0 {
		return fmt.Errorf("tool %s requires %s to be installed first",
			tool.Name, strings.Join(missing, ", "))
	}
	return nil
}

// isDependencyAvailable checks whether a dependency is satisfied by looking
// in the install directory first, then falling back to the system PATH.
// Uses state.LookupInPath for proper binary name resolution (e.g., neovim → nvim).
func (inst *Installer) isDependencyAvailable(dep string) bool {
	// Check install dir first (by raw tool name).
	depPath := filepath.Join(inst.InstallDir, dep)
	if _, err := os.Stat(depPath); err == nil {
		return true
	}

	// Fall back to PATH with probe-aware binary name mapping.
	return state.LookupInPath(dep) != ""
}
