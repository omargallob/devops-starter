package pkgmgr

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// pipManager installs Python packages at user scope via pip or pipx.
type pipManager struct {
	bin    string
	dryRun bool
}

// Install runs: pip install --user <pkg>==<version>
// When version is empty the version specifier is omitted.
func (m *pipManager) Install(ctx context.Context, pkg, version string) error {
	spec := pkg
	if version != "" {
		spec += "==" + version
	}

	args := []string{"install", "--user", spec}
	if m.dryRun {
		fmt.Printf("[dry-run] %s %s\n", m.bin, strings.Join(args, " "))
		return nil
	}

	cmd := exec.CommandContext(ctx, m.bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pip install %s: %w", spec, err)
	}
	return nil
}

// List returns user-scope installed packages from `pip list --user --format=json`.
func (m *pipManager) List(ctx context.Context) (map[string]string, error) {
	cmd := exec.CommandContext(ctx, m.bin, "list", "--user", "--format=json")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("pip list: %w", err)
	}

	var entries []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(out, &entries); err != nil {
		return nil, fmt.Errorf("parsing pip list output: %w", err)
	}

	result := make(map[string]string, len(entries))
	for _, e := range entries {
		result[strings.ToLower(e.Name)] = e.Version
	}
	return result, nil
}

// IsAvailable reports whether the pip binary is accessible.
func (m *pipManager) IsAvailable() bool {
	_, err := exec.LookPath(m.bin)
	return err == nil
}

// pythonBinCandidates returns a list of pip/pipx binary paths to probe in
// priority order: mise-managed installation first, then system PATH names.
func pythonBinCandidates(manager string) []string {
	home, _ := os.UserHomeDir()
	miseBase := filepath.Join(home, ".local", "share", "mise", "installs", "python")

	var candidates []string
	// Walk common mise Python version directories for the manager binary.
	for _, ver := range []string{"3.13", "3.12", "3.11", "3.10"} {
		candidates = append(candidates,
			filepath.Join(miseBase, ver, "bin", manager),
		)
	}
	// Fallback PATH names.
	candidates = append(candidates, manager+"3", manager)
	return candidates
}
