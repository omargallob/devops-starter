// Package updater checks for new versions of tools in the registry by querying
// GitHub Releases API and other version sources. It can optionally rewrite the
// registry Go source files to bump pinned versions.
package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// UpdateInfo describes a tool that has a newer version available.
type UpdateInfo struct {
	ToolName       string
	CurrentVersion string
	LatestVersion  string
	Repo           string // GitHub owner/repo
	Group          string // tool group (for locating registry file)
}

// CheckResult is the outcome of a full version check across all tools.
type CheckResult struct {
	Updates []UpdateInfo
	Skipped []string // tools that couldn't be checked (no repo, API failure, etc.)
	Errors  []error
}

// Options configures the version checker.
type Options struct {
	// HTTPClient allows injecting a custom HTTP client (for testing).
	HTTPClient *http.Client
	// Token is a GitHub personal access token for higher rate limits.
	Token string
	// Out is where progress messages are written.
	Out io.Writer
}

// DefaultOptions returns options with a default HTTP client and 10s timeout.
func DefaultOptions() *Options {
	return &Options{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		Out:        io.Discard,
	}
}

// githubRelease is the subset of the GitHub release API response we need.
type githubRelease struct {
	TagName string `json:"tag_name"`
}

// hashicorpCheckpoint is the response from the HashiCorp checkpoint API.
type hashicorpCheckpoint struct {
	CurrentVersion string `json:"current_version"`
}

// hashicorpTools maps tool names to their HashiCorp product names.
var hashicorpTools = map[string]string{
	"terraform": "terraform",
	"vault":     "vault",
	"consul":    "consul",
	"packer":    "packer",
}

// versionTagRegex extracts a semver version from a git tag.
var versionTagRegex = regexp.MustCompile(`v?(\d+\.\d+\.\d+)`)

// CheckAll queries for the latest version of every tool in the given list.
func CheckAll(ctx context.Context, tools []*tooldef.Tool, opts *Options) *CheckResult {
	if opts == nil {
		opts = DefaultOptions()
	}

	result := &CheckResult{}

	for _, t := range tools {
		// Skip mise-managed tools (their versions come from .mise.toml).
		if t.EffectiveInstallMode() == tooldef.InstallModeMise {
			result.Skipped = append(result.Skipped, t.Name)
			continue
		}

		latest, err := checkOne(ctx, t, opts)
		if err != nil {
			result.Skipped = append(result.Skipped, t.Name)
			result.Errors = append(result.Errors, fmt.Errorf("%s: %w", t.Name, err))
			continue
		}

		if latest == "" {
			result.Skipped = append(result.Skipped, t.Name)
			continue
		}

		if latest != t.Version {
			result.Updates = append(result.Updates, UpdateInfo{
				ToolName:       t.Name,
				CurrentVersion: t.Version,
				LatestVersion:  latest,
				Repo:           t.Repo,
				Group:          string(t.Group),
			})
		}
	}

	return result
}

// checkOne determines the latest version for a single tool.
func checkOne(ctx context.Context, t *tooldef.Tool, opts *Options) (string, error) {
	// Strategy 1: HashiCorp checkpoint API.
	if product, ok := hashicorpTools[t.Name]; ok {
		return checkHashiCorp(ctx, product, opts)
	}

	// Strategy 2: GitHub Releases API (for tools with a Repo field).
	if t.Repo != "" {
		return checkGitHubRelease(ctx, t.Repo, opts)
	}

	// Strategy 3: Try to derive repo from name for well-known patterns.
	// If no strategy matches, skip.
	return "", nil
}

// checkGitHubRelease queries the GitHub Releases API for the latest release tag.
func checkGitHubRelease(ctx context.Context, repo string, opts *Options) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	if opts.Token != "" {
		req.Header.Set("Authorization", "Bearer "+opts.Token)
	}

	resp, err := opts.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %d for %s", resp.StatusCode, repo)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("decoding release for %s: %w", repo, err)
	}

	return extractVersion(release.TagName), nil
}

// checkHashiCorp queries the HashiCorp checkpoint API for the latest version.
func checkHashiCorp(ctx context.Context, product string, opts *Options) (string, error) {
	url := fmt.Sprintf("https://checkpoint-api.hashicorp.com/v1/check/%s", product)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := opts.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HashiCorp checkpoint returned %d for %s", resp.StatusCode, product)
	}

	var checkpoint hashicorpCheckpoint
	if err := json.NewDecoder(resp.Body).Decode(&checkpoint); err != nil {
		return "", fmt.Errorf("decoding checkpoint for %s: %w", product, err)
	}

	return checkpoint.CurrentVersion, nil
}

// extractVersion strips a leading "v" and any suffix from a tag to get a clean version.
func extractVersion(tag string) string {
	tag = strings.TrimSpace(tag)
	matches := versionTagRegex.FindStringSubmatch(tag)
	if len(matches) >= 2 {
		return matches[1]
	}
	// Fallback: strip leading 'v'.
	return strings.TrimPrefix(tag, "v")
}
