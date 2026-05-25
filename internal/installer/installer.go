// Package installer handles downloading tool binaries, verifying checksums,
// extracting archives, and installing to the target directory.
package installer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/omargallob/devops-starter/internal/state"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// Installer manages downloading and installing tool binaries.
type Installer struct {
	InstallDir  string
	Platform    tooldef.Platform
	DryRun      bool
	Concurrency int
	// StateStore is an optional state store for recording installed versions.
	// If nil, no state is recorded.
	StateStore *state.Store
}

// Option is a functional option for configuring an Installer.
type Option func(*Installer)

// WithDryRun enables dry-run mode (no actual installation).
func WithDryRun(dryRun bool) Option {
	return func(i *Installer) {
		i.DryRun = dryRun
	}
}

// WithConcurrency sets the maximum number of concurrent installations.
func WithConcurrency(n int) Option {
	return func(i *Installer) {
		i.Concurrency = n
	}
}

// WithStateStore sets the state store for recording installed tool versions.
func WithStateStore(store *state.Store) Option {
	return func(i *Installer) {
		i.StateStore = store
	}
}

// New creates a new Installer with the given install directory, platform, and options.
func New(installDir string, platform tooldef.Platform, opts ...Option) *Installer {
	i := &Installer{
		InstallDir:  installDir,
		Platform:    platform,
		Concurrency: 4,
	}
	for _, opt := range opts {
		opt(i)
	}
	return i
}

// EnsureDir creates the install directory if it doesn't exist.
func (inst *Installer) EnsureDir() error {
	return os.MkdirAll(inst.InstallDir, 0o755)
}

// IsInstalled checks if the tool binary exists at the expected path.
func (inst *Installer) IsInstalled(tool *tooldef.Tool) bool {
	binPath := filepath.Join(inst.InstallDir, tool.GetInstallName())
	_, err := os.Stat(binPath)
	return err == nil
}

// Install orchestrates download, verify, extract, and install for a single tool.
// Tools with ManagedBy set are delegated to their manager (e.g., mise install).
func (inst *Installer) Install(ctx context.Context, tool *tooldef.Tool) error {
	if inst.DryRun {
		if tool.ManagedBy != "" {
			fmt.Printf("[dry-run] Would install %s %s via %s\n", tool.Name, tool.Version, tool.ManagedBy)
		} else {
			fmt.Printf("[dry-run] Would install %s %s\n", tool.Name, tool.Version)
		}
		return nil
	}

	// Delegate to the manager if this tool is externally managed.
	if tool.ManagedBy != "" {
		return inst.installViaManager(ctx, tool)
	}

	if err := inst.EnsureDir(); err != nil {
		return fmt.Errorf("ensuring install dir: %w", err)
	}

	url, err := ResolveURL(tool, inst.Platform)
	if err != nil {
		return fmt.Errorf("resolving URL for %s: %w", tool.Name, err)
	}

	// Download to temp file
	tmpDir, err := os.MkdirTemp("", "installer-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, "download")
	if err := Download(ctx, url, archivePath); err != nil {
		return fmt.Errorf("downloading %s: %w", tool.Name, err)
	}

	// Verify checksum if available
	platformKey := inst.Platform.String()
	if expected, ok := tool.Checksums[platformKey]; ok {
		if err := VerifyChecksum(archivePath, expected); err != nil {
			return fmt.Errorf("checksum verification for %s: %w", tool.Name, err)
		}
	}

	// Extract
	extractDir := filepath.Join(tmpDir, "extracted")
	if err := os.MkdirAll(extractDir, 0o755); err != nil {
		return fmt.Errorf("creating extract dir: %w", err)
	}

	if err := Extract(archivePath, extractDir, tool.Format, tool.StripComponents); err != nil {
		return fmt.Errorf("extracting %s: %w", tool.Name, err)
	}

	// Move binary to install dir
	srcBinary := filepath.Join(extractDir, tool.GetBinaryName())
	dstBinary := filepath.Join(inst.InstallDir, tool.GetInstallName())

	data, err := os.ReadFile(srcBinary)
	if err != nil {
		return fmt.Errorf("reading extracted binary %s: %w", tool.Name, err)
	}

	if err := os.WriteFile(dstBinary, data, 0o755); err != nil {
		return fmt.Errorf("writing binary %s: %w", tool.Name, err)
	}

	// Record the installed version in state store (if configured)
	if inst.StateStore != nil {
		// Best-effort: don't fail the install if state recording fails
		_ = inst.StateStore.Record(tool.Name, tool.Version)
	}

	return nil
}

// InstallAll installs multiple tools concurrently, returning any errors.
// Tools managed by an external manager (e.g., mise) are batched and installed
// after all direct tools complete, using a single manager invocation.
func (inst *Installer) InstallAll(ctx context.Context, tools []*tooldef.Tool) []error {
	var (
		mu           sync.Mutex
		errs         []error
		wg           sync.WaitGroup
		sem          = make(chan struct{}, inst.Concurrency)
		managedTools []*tooldef.Tool
		directTools  []*tooldef.Tool
	)

	// Separate managed and direct tools.
	for _, t := range tools {
		if t.ManagedBy != "" {
			managedTools = append(managedTools, t)
		} else {
			directTools = append(directTools, t)
		}
	}

	// Install direct tools concurrently.
	for _, tool := range directTools {
		wg.Add(1)
		sem <- struct{}{}
		go func(t *tooldef.Tool) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := inst.Install(ctx, t); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(tool)
	}
	wg.Wait()

	// Install managed tools (batched by manager).
	if len(managedTools) > 0 {
		if mErr := inst.InstallMiseTools(ctx); mErr != nil {
			errs = append(errs, mErr)
		}
	}

	return errs
}

// installViaManager delegates installation of a single tool to its external
// manager. Currently only supports "mise" as a manager.
func (inst *Installer) installViaManager(ctx context.Context, tool *tooldef.Tool) error {
	switch tool.ManagedBy {
	case "mise":
		return inst.InstallMiseTools(ctx)
	default:
		return fmt.Errorf("unsupported manager %q for tool %s", tool.ManagedBy, tool.Name)
	}
}

// InstallMiseTools runs "mise install" to install all tools defined in
// .mise.toml. This is a single invocation that handles all mise-managed tools.
func (inst *Installer) InstallMiseTools(ctx context.Context) error {
	// Find mise binary — check install dir first, then PATH.
	miseBin := filepath.Join(inst.InstallDir, "mise")
	if _, err := os.Stat(miseBin); err != nil {
		// Fall back to PATH lookup.
		found, lookErr := exec.LookPath("mise")
		if lookErr != nil {
			return fmt.Errorf("mise binary not found in %s or PATH — install mise first", inst.InstallDir)
		}
		miseBin = found
	}

	cmd := exec.CommandContext(ctx, miseBin, "install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "MISE_YES=1") // non-interactive

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running mise install: %w", err)
	}
	return nil
}
