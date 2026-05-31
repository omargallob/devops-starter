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

// npmManager installs Node packages at global scope via npm.
type npmManager struct {
	bin    string
	dryRun bool
}

// Install runs: npm install -g <pkg>@<version>
// When version is empty the version specifier is omitted.
func (m *npmManager) Install(ctx context.Context, pkg, version string) error {
	spec := pkg
	if version != "" {
		spec += "@" + version
	}

	args := []string{"install", "-g", spec}
	if m.dryRun {
		fmt.Printf("[dry-run] %s %s\n", m.bin, strings.Join(args, " "))
		return nil
	}

	cmd := exec.CommandContext(ctx, m.bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm install -g %s: %w", spec, err)
	}
	return nil
}

// List returns globally-installed npm packages via `npm list -g --depth=0 --json`.
func (m *npmManager) List(ctx context.Context) (map[string]string, error) {
	cmd := exec.CommandContext(ctx, m.bin, "list", "-g", "--depth=0", "--json")
	out, err := cmd.Output()
	if err != nil {
		// npm list exits non-zero when there are peer-dep warnings; try to parse anyway.
		if len(out) == 0 {
			return nil, fmt.Errorf("npm list -g: %w", err)
		}
	}

	var result struct {
		Dependencies map[string]struct {
			Version string `json:"version"`
		} `json:"dependencies"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("parsing npm list output: %w", err)
	}

	pkgs := make(map[string]string, len(result.Dependencies))
	for name, dep := range result.Dependencies {
		pkgs[strings.ToLower(name)] = dep.Version
	}
	return pkgs, nil
}

// IsAvailable reports whether the npm binary is accessible.
func (m *npmManager) IsAvailable() bool {
	_, err := exec.LookPath(m.bin)
	return err == nil
}

// npmBinCandidates returns npm binary paths to probe in priority order:
// mise-managed Node installation first, then system PATH.
func npmBinCandidates() []string {
	home, _ := os.UserHomeDir()
	miseBase := filepath.Join(home, ".local", "share", "mise", "installs", "node")

	var candidates []string
	for _, ver := range []string{"22", "20", "18"} {
		candidates = append(candidates,
			filepath.Join(miseBase, ver, "bin", "npm"),
		)
	}
	candidates = append(candidates, "npm")
	return candidates
}
