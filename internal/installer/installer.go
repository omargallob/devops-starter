// Package installer handles downloading tool binaries, verifying checksums,
// extracting archives, and installing to the target directory.
package installer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// Installer manages downloading and installing tool binaries.
type Installer struct {
	InstallDir  string
	Platform    tooldef.Platform
	DryRun      bool
	Concurrency int
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
func (inst *Installer) Install(ctx context.Context, tool *tooldef.Tool) error {
	if inst.DryRun {
		fmt.Printf("[dry-run] Would install %s %s\n", tool.Name, tool.Version)
		return nil
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

	return nil
}

// InstallAll installs multiple tools concurrently, returning any errors.
func (inst *Installer) InstallAll(ctx context.Context, tools []*tooldef.Tool) []error {
	var (
		mu     sync.Mutex
		errs   []error
		wg     sync.WaitGroup
		sem    = make(chan struct{}, inst.Concurrency)
	)

	for _, tool := range tools {
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
	return errs
}
