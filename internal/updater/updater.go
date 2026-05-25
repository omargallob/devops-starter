// Package updater checks for newer versions of devops-starter via the GitHub Releases API.
package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	// releaseURL is the GitHub API endpoint for the latest release.
	releaseURL = "https://api.github.com/repos/omargallob/devops-starter/releases/latest"
	// httpTimeout is the maximum time for the update check HTTP request.
	httpTimeout = 5 * time.Second
)

// Result contains the outcome of an update check.
type Result struct {
	LatestVersion   string // latest version available (without "v" prefix)
	UpdateAvailable bool   // true if latest > current
	Error           error  // non-nil if the check failed (should be treated as non-fatal)
}

// githubRelease is the minimal structure we need from the GitHub API response.
type githubRelease struct {
	TagName string `json:"tag_name"`
}

// Check queries the GitHub Releases API and compares the latest release tag
// against the current version. It is designed to be called asynchronously.
// Errors are returned inside Result (non-fatal; the app should continue normally).
func Check(currentVersion string) Result {
	return checkWithURL(releaseURL, currentVersion)
}

// checkWithURL is the internal implementation that accepts a custom URL for testing.
func checkWithURL(url, currentVersion string) Result {
	if currentVersion == "" || currentVersion == "dev" {
		return Result{} // no point checking for dev builds
	}

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return Result{Error: fmt.Errorf("update check: %w", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{Error: fmt.Errorf("update check: HTTP %d", resp.StatusCode)}
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return Result{Error: fmt.Errorf("update check: %w", err)}
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(currentVersion, "v")

	if latest == "" || latest == current {
		return Result{LatestVersion: latest}
	}

	if compareVersions(latest, current) > 0 {
		return Result{LatestVersion: latest, UpdateAvailable: true}
	}

	return Result{LatestVersion: latest}
}

// compareVersions performs a simple semver-style comparison (major.minor.patch).
// Returns >0 if a > b, 0 if equal, <0 if a < b.
func compareVersions(a, b string) int {
	aParts := splitVersion(a)
	bParts := splitVersion(b)

	for i := 0; i < 3; i++ {
		if aParts[i] > bParts[i] {
			return 1
		}
		if aParts[i] < bParts[i] {
			return -1
		}
	}
	return 0
}

// splitVersion splits "1.2.3" into [1, 2, 3]. Non-numeric parts default to 0.
func splitVersion(v string) [3]int {
	var parts [3]int
	segments := strings.SplitN(v, ".", 3)
	for i, s := range segments {
		if i >= 3 {
			break
		}
		// Parse digits only (stop at first non-digit, e.g. "3-rc1" -> 3)
		var n int
		for _, ch := range s {
			if ch >= '0' && ch <= '9' {
				n = n*10 + int(ch-'0')
			} else {
				break
			}
		}
		parts[i] = n
	}
	return parts
}
