package installer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/omargallob/devops-starter/pkg/tooldef"
)

func TestEnsureEget_AlreadyInDir(t *testing.T) {
	dir := t.TempDir()
	// Create a fake eget binary.
	egetPath := filepath.Join(dir, "eget")
	if err := os.WriteFile(egetPath, []byte("#!/bin/sh\necho eget"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := EnsureEget(context.Background(), dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != egetPath {
		t.Errorf("expected %s, got %s", egetPath, got)
	}
}

func TestEnsureEget_NotFound_DownloadFails(t *testing.T) {
	dir := t.TempDir()
	// No eget in dir and not in PATH — will try to download and fail
	// because we don't have a real network server for eget releases.
	// This at least exercises the error path.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately to force failure

	_, err := EnsureEget(ctx, dir)
	if err == nil {
		t.Fatal("expected error when download is cancelled")
	}
}

func TestEgetDownloadURL(t *testing.T) {
	url, err := egetDownloadURL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url == "" {
		t.Fatal("expected non-empty URL")
	}
	// Should contain the eget repo and version.
	if got := url; got == "" {
		t.Fatal("empty url")
	}
}

func TestBuildEgetRepoArgs(t *testing.T) {
	tool := &tooldef.Tool{
		Name:        "k9s",
		Version:     "0.32.7",
		Repo:        "derailed/k9s",
		Asset:       "*.tar.gz",
		BinaryName:  "k9s",
		InstallName: "k9s",
	}
	platform := tooldef.Platform{OS: "linux", Arch: "amd64"}

	args := buildEgetRepoArgs(tool, "/usr/local/bin", platform)

	// Check essential args are present.
	assertContains(t, args, "derailed/k9s")
	assertContains(t, args, "--to")
	assertContains(t, args, "/usr/local/bin")
	assertContains(t, args, "--tag")
	assertContains(t, args, "v0.32.7")
	assertContains(t, args, "--asset")
	assertContains(t, args, "*.tar.gz")
	assertContains(t, args, "-q")
}

func TestBuildEgetRepoArgs_Minimal(t *testing.T) {
	tool := &tooldef.Tool{
		Name:    "fzf",
		Version: "0.57.0",
		Repo:    "junegunn/fzf",
	}
	platform := tooldef.Platform{OS: "darwin", Arch: "arm64"}

	args := buildEgetRepoArgs(tool, "/tmp/bin", platform)

	assertContains(t, args, "junegunn/fzf")
	assertContains(t, args, "--tag")
	assertContains(t, args, "v0.57.0")
	// Should NOT have --asset since Asset is empty.
	assertNotContains(t, args, "--asset")
	// Should NOT have --file since BinaryName is empty.
	assertNotContains(t, args, "--file")
}

func TestBuildEgetRepoArgs_WithChecksum(t *testing.T) {
	tool := &tooldef.Tool{
		Name:    "stern",
		Version: "1.31.0",
		Repo:    "stern/stern",
		Checksums: map[string]string{
			"linux/amd64": "abc123def456",
		},
	}
	platform := tooldef.Platform{OS: "linux", Arch: "amd64"}

	args := buildEgetRepoArgs(tool, "/tmp/bin", platform)
	assertContains(t, args, "--sha256")
	assertContains(t, args, "abc123def456")
}

func TestBuildEgetURLArgs(t *testing.T) {
	tool := &tooldef.Tool{
		Name:        "kubectl",
		Version:     "1.31.4",
		InstallName: "kubectl",
	}

	args := buildEgetURLArgs(tool, "https://dl.k8s.io/release/v1.31.4/bin/linux/amd64/kubectl", "/usr/local/bin")

	assertContains(t, args, "https://dl.k8s.io/release/v1.31.4/bin/linux/amd64/kubectl")
	assertContains(t, args, "--to")
	assertContains(t, args, "/usr/local/bin")
	assertContains(t, args, "-q")
}

func TestBuildEgetRepoArgs_Rename(t *testing.T) {
	tool := &tooldef.Tool{
		Name:        "opentofu",
		Version:     "1.9.0",
		Repo:        "opentofu/opentofu",
		InstallName: "tofu",
	}
	platform := tooldef.Platform{OS: "linux", Arch: "amd64"}

	args := buildEgetRepoArgs(tool, "/tmp/bin", platform)
	assertContains(t, args, "--rename")
	assertContains(t, args, "tofu")
}

func TestRunEget_NotFound(t *testing.T) {
	err := runEget(context.Background(), "/nonexistent/eget", []string{"--help"}, "test")
	if err == nil {
		t.Fatal("expected error when eget binary doesn't exist")
	}
}

func TestDownloadFile_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "out")
	err := downloadFile(context.Background(), srv.URL+"/file.tar.gz", dest)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

// helpers

func assertContains(t *testing.T, slice []string, item string) {
	t.Helper()
	for _, s := range slice {
		if s == item {
			return
		}
	}
	t.Errorf("expected %v to contain %q", slice, item)
}

func assertNotContains(t *testing.T, slice []string, item string) {
	t.Helper()
	for _, s := range slice {
		if s == item {
			t.Errorf("expected %v to NOT contain %q", slice, item)
			return
		}
	}
}
