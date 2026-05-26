// bootstrap.go provides automatic download of the eget binary, used as the
// underlying engine for installing tool binaries from GitHub releases and
// direct URLs.
package installer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// egetVersion is the pinned version of eget to bootstrap.
const egetVersion = "1.3.4"

// egetRepo is the GitHub repository for eget.
const egetRepo = "zyedidia/eget"

// EnsureEget checks for the eget binary in installDir or PATH.
// If missing, it downloads the correct platform binary from GitHub releases.
func EnsureEget(ctx context.Context, installDir string) (string, error) {
	// Check install dir first.
	egetPath := filepath.Join(installDir, "eget")
	if _, err := os.Stat(egetPath); err == nil {
		return egetPath, nil
	}

	// Check PATH.
	if found, err := exec.LookPath("eget"); err == nil {
		return found, nil
	}

	// Download eget.
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return "", fmt.Errorf("creating install dir: %w", err)
	}

	url, err := egetDownloadURL()
	if err != nil {
		return "", err
	}

	if err := downloadFile(ctx, url, egetPath); err != nil {
		return "", fmt.Errorf("downloading eget: %w", err)
	}

	if err := os.Chmod(egetPath, 0o755); err != nil {
		return "", fmt.Errorf("chmod eget: %w", err)
	}

	return egetPath, nil
}

// egetDownloadURL returns the download URL for the eget binary matching the
// current OS and architecture.
func egetDownloadURL() (string, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	var platform string
	switch {
	case goos == "darwin" && goarch == "amd64":
		platform = "darwin_amd64"
	case goos == "darwin" && goarch == "arm64":
		platform = "darwin_arm64"
	case goos == "linux" && goarch == "amd64":
		platform = "linux_amd64"
	case goos == "linux" && goarch == "arm64":
		platform = "linux_arm64"
	default:
		return "", fmt.Errorf("unsupported platform: %s/%s", goos, goarch)
	}

	return fmt.Sprintf(
		"https://github.com/%s/releases/download/v%s/eget-%s-%s.tar.gz",
		egetRepo, egetVersion, egetVersion, platform,
	), nil
}

// downloadFile performs a simple HTTP GET and writes the response body to dest.
// This is intentionally minimal — it's only used for the eget bootstrap.
func downloadFile(ctx context.Context, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "devops-starter/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d fetching %s", resp.StatusCode, url)
	}

	// eget releases are tar.gz archives — extract the binary.
	tmpDir, err := os.MkdirTemp("", "eget-bootstrap-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, "eget.tar.gz")
	f, err := os.Create(archivePath)
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return err
	}
	f.Close()

	// Extract the eget binary from the tarball.
	extractDir := filepath.Join(tmpDir, "extracted")
	if err := os.MkdirAll(extractDir, 0o755); err != nil {
		return err
	}

	if err := Extract(archivePath, extractDir, tooldef.FormatTarGz, 1); err != nil {
		return fmt.Errorf("extracting eget archive: %w", err)
	}

	// Copy the binary to the final destination.
	egetBinary := filepath.Join(extractDir, "eget")
	data, err := os.ReadFile(egetBinary)
	if err != nil {
		return fmt.Errorf("reading extracted eget binary: %w", err)
	}

	return os.WriteFile(dest, data, 0o755)
}
