// eget.go implements tool installation via the eget binary.
// Supports two modes:
//   - Repo mode: eget <owner/repo> --tag <version> (for GitHub releases)
//   - URL mode: eget <direct-url> (for non-GitHub downloads)
package installer

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// installWithEget installs a tool using eget in GitHub repo mode.
// Requires tool.Repo to be set (e.g., "derailed/k9s").
func (inst *Installer) installWithEget(ctx context.Context, tool *tooldef.Tool) error {
	egetBin, err := EnsureEget(ctx, inst.InstallDir)
	if err != nil {
		return fmt.Errorf("ensuring eget available: %w", err)
	}

	args, err := buildEgetRepoArgs(tool, inst.InstallDir, inst.Platform)
	if err != nil {
		return fmt.Errorf("building eget args for %s: %w", tool.Name, err)
	}
	return runEget(ctx, egetBin, args, tool.Name)
}

// installWithEgetURL installs a tool using eget in direct URL mode.
// The URL is resolved from the tool's URLTemplate using the current platform.
func (inst *Installer) installWithEgetURL(ctx context.Context, tool *tooldef.Tool) error {
	egetBin, err := EnsureEget(ctx, inst.InstallDir)
	if err != nil {
		return fmt.Errorf("ensuring eget available: %w", err)
	}

	url, err := ResolveURL(tool, inst.Platform)
	if err != nil {
		return fmt.Errorf("resolving URL for %s: %w", tool.Name, err)
	}

	args := buildEgetURLArgs(tool, url, inst.InstallDir)
	return runEget(ctx, egetBin, args, tool.Name)
}

// buildEgetRepoArgs constructs the eget CLI arguments for repo mode.
func buildEgetRepoArgs(tool *tooldef.Tool, installDir string, platform tooldef.Platform) ([]string, error) {
	args := []string{tool.Repo, "--to", installDir}

	// Pin the version. Most GitHub repos tag releases as "v<version>", but
	// some (e.g. jqlang/jq uses "jq-<version>") need an explicit override.
	if tool.Version != "" {
		prefix := tool.TagPrefix
		if prefix == "" {
			prefix = "v"
		}
		args = append(args, "--tag", prefix+tool.Version)
	}

	// Asset filter narrows which release asset to download. The pattern may
	// contain {{.OS}}/{{.Arch}} template directives that need rendering.
	if tool.Asset != "" {
		asset, err := ResolveAsset(tool, platform)
		if err != nil {
			return nil, err
		}
		args = append(args, "--asset", asset)
	}

	// Specify the binary to extract from the archive.
	if tool.BinaryName != "" {
		args = append(args, "--file", tool.GetBinaryName())
	}

	// Rename the installed binary if InstallName differs.
	if tool.InstallName != "" && tool.InstallName != tool.Name {
		args = append(args, "--rename", tool.GetInstallName())
	}

	// Checksum verification.
	platformKey := platform.String()
	if expected, ok := tool.Checksums[platformKey]; ok {
		args = append(args, "--sha256", expected)
	}

	// Quiet mode — we handle progress/output in the TUI layer.
	args = append(args, "-q")

	return args, nil
}

// buildEgetURLArgs constructs the eget CLI arguments for direct URL mode.
func buildEgetURLArgs(tool *tooldef.Tool, url, installDir string) []string {
	args := []string{url, "--to", installDir}

	// Specify the binary to extract from the archive.
	if tool.BinaryName != "" {
		args = append(args, "--file", tool.GetBinaryName())
	}

	// Rename the installed binary if InstallName differs.
	if tool.InstallName != "" && tool.InstallName != tool.Name {
		args = append(args, "--rename", tool.GetInstallName())
	}

	// Quiet mode.
	args = append(args, "-q")

	return args
}

// runEget executes the eget binary with the given arguments.
func runEget(ctx context.Context, egetBin string, args []string, toolName string) error {
	cmd := exec.CommandContext(ctx, egetBin, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("eget install %s failed: %w\nstderr: %s", toolName, err, stderr.String())
	}
	return nil
}
