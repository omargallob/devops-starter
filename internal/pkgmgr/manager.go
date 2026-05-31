// Package pkgmgr installs and lists globally-scoped packages distributed via
// pip (Python) and npm (Node.js). All installs target the user/global level —
// pip uses --user and npm uses -g.
package pkgmgr

import (
	"context"
	"fmt"
	"os/exec"
)

// PackageManager installs and lists packages for a single package manager.
type PackageManager interface {
	// Install installs pkg at version globally.
	// Precondition: pkg non-empty; version may be empty (installs latest).
	// Postcondition: on nil return, pkg is available at user/global scope.
	Install(ctx context.Context, pkg, version string) error

	// List returns all globally-installed packages mapped to their versions.
	List(ctx context.Context) (map[string]string, error)

	// IsAvailable returns true if the underlying manager binary is on PATH
	// or at its expected mise-managed location.
	IsAvailable() bool
}

// Option configures a PackageManager.
type Option func(*options)

type options struct {
	dryRun bool
}

// WithDryRun makes the manager print commands without executing them.
func WithDryRun(v bool) Option {
	return func(o *options) { o.dryRun = v }
}

// ErrManagerNotFound is returned when the package manager binary is not found.
var ErrManagerNotFound = fmt.Errorf("package manager binary not found")

// New returns a PackageManager for the named manager ("pip", "pipx", "npm").
// Returns ErrManagerNotFound if the binary cannot be located.
func New(manager string, opts ...Option) (PackageManager, error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	switch manager {
	case "pip", "pipx":
		bin, err := resolveBinary(manager, pythonBinCandidates(manager))
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrManagerNotFound, manager)
		}
		return &pipManager{bin: bin, dryRun: o.dryRun}, nil
	case "npm":
		bin, err := resolveBinary("npm", npmBinCandidates())
		if err != nil {
			return nil, fmt.Errorf("%w: npm", ErrManagerNotFound)
		}
		return &npmManager{bin: bin, dryRun: o.dryRun}, nil
	default:
		return nil, fmt.Errorf("unsupported package manager: %s", manager)
	}
}

// resolveBinary finds the first accessible binary from candidates, then falls
// back to a PATH lookup for fallback.
func resolveBinary(fallback string, candidates []string) (string, error) {
	for _, c := range candidates {
		if p, err := exec.LookPath(c); err == nil {
			return p, nil
		}
	}
	return exec.LookPath(fallback)
}
