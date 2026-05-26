package installer

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/omargallob/devops-starter/internal/state"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// createTarGzFixture builds an in-memory tar.gz containing a single binary.
func createTarGzFixture(t *testing.T, binaryName string, content []byte) []byte {
	t.Helper()
	tmp := t.TempDir()
	archivePath := filepath.Join(tmp, "fixture.tar.gz")

	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)
	tw.WriteHeader(&tar.Header{
		Name: binaryName,
		Mode: 0o755,
		Size: int64(len(content)),
	})
	tw.Write(content)
	tw.Close()
	gw.Close()
	f.Close()

	data, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func sha256hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// TestInstall_EndToEnd_TarGz verifies the full install flow:
// download → checksum → extract → binary placed in install dir.
func TestInstall_EndToEnd_TarGz(t *testing.T) {
	binaryContent := []byte("#!/bin/sh\necho hello from mytool")
	archive := createTarGzFixture(t, "mytool", binaryContent)
	checksum := sha256hex(archive)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", string(rune(len(archive))))
		w.WriteHeader(http.StatusOK)
		w.Write(archive)
	}))
	defer srv.Close()

	installDir := t.TempDir()
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}
	inst := New(installDir, plat)

	tool := &tooldef.Tool{
		Name:        "mytool",
		Version:     "1.0.0",
		Format:      tooldef.FormatTarGz,
		URLTemplate: srv.URL + "/{{.Name}}-{{.Version}}.tar.gz",
		Checksums: map[string]string{
			"linux/amd64": checksum,
		},
	}

	ctx := context.Background()
	if err := inst.Install(ctx, tool); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify binary was placed correctly
	binPath := filepath.Join(installDir, "mytool")
	got, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatalf("binary not found at %s: %v", binPath, err)
	}
	if !bytes.Equal(got, binaryContent) {
		t.Errorf("binary content mismatch: got %q", got)
	}

	// Verify it's executable
	info, _ := os.Stat(binPath)
	if info.Mode()&0o111 == 0 {
		t.Error("expected executable permissions on installed binary")
	}

	// Verify IsInstalled returns true
	if !inst.IsInstalled(tool) {
		t.Error("IsInstalled should return true after install")
	}
}

// TestInstall_EndToEnd_Binary verifies install of a raw binary (no archive).
func TestInstall_EndToEnd_Binary(t *testing.T) {
	binaryContent := []byte("#!/bin/sh\necho raw binary")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(binaryContent)
	}))
	defer srv.Close()

	installDir := t.TempDir()
	plat := tooldef.Platform{OS: "darwin", Arch: "arm64"}
	inst := New(installDir, plat)

	tool := &tooldef.Tool{
		Name:        "simpletool",
		Version:     "2.0.0",
		Format:      tooldef.FormatBinary,
		URLTemplate: srv.URL + "/{{.Name}}",
	}

	ctx := context.Background()
	if err := inst.Install(ctx, tool); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	binPath := filepath.Join(installDir, "simpletool")
	got, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatalf("binary not found: %v", err)
	}
	if !bytes.Equal(got, binaryContent) {
		t.Errorf("binary content mismatch")
	}

	// Verify executable permissions
	info, _ := os.Stat(binPath)
	if info.Mode()&0o111 == 0 {
		t.Error("expected executable permissions")
	}
}

// TestInstall_ChecksumMismatch verifies that a bad checksum causes install failure.
func TestInstall_ChecksumMismatch(t *testing.T) {
	binaryContent := []byte("legit binary")
	archive := createTarGzFixture(t, "tool", binaryContent)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(archive)
	}))
	defer srv.Close()

	installDir := t.TempDir()
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}
	inst := New(installDir, plat)

	tool := &tooldef.Tool{
		Name:        "tool",
		Version:     "1.0.0",
		Format:      tooldef.FormatTarGz,
		URLTemplate: srv.URL + "/download",
		Checksums: map[string]string{
			"linux/amd64": "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
		},
	}

	ctx := context.Background()
	err := inst.Install(ctx, tool)
	if err == nil {
		t.Fatal("expected error for checksum mismatch")
	}
	// Binary should NOT be installed after checksum failure
	if inst.IsInstalled(tool) {
		t.Error("tool should not be installed after checksum mismatch")
	}
}

// TestInstall_DownloadFailure verifies install fails gracefully on HTTP error.
func TestInstall_DownloadFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	installDir := t.TempDir()
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}
	inst := New(installDir, plat)

	tool := &tooldef.Tool{
		Name:        "missing-tool",
		Version:     "1.0.0",
		Format:      tooldef.FormatBinary,
		URLTemplate: srv.URL + "/{{.Name}}",
	}

	ctx := context.Background()
	err := inst.Install(ctx, tool)
	if err == nil {
		t.Fatal("expected error when download returns 404")
	}
}

// TestInstall_WithStateStore verifies that state is recorded after install.
func TestInstall_WithStateStore(t *testing.T) {
	binaryContent := []byte("stateful binary")
	archive := createTarGzFixture(t, "statetool", binaryContent)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(archive)
	}))
	defer srv.Close()

	installDir := t.TempDir()
	stateDir := t.TempDir()
	statePath := filepath.Join(stateDir, "state.json")

	store, err := state.LoadStore(statePath)
	if err != nil {
		t.Fatalf("failed to load store: %v", err)
	}

	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}
	inst := New(installDir, plat, WithStateStore(store))

	tool := &tooldef.Tool{
		Name:        "statetool",
		Version:     "3.0.0",
		Format:      tooldef.FormatTarGz,
		URLTemplate: srv.URL + "/{{.Name}}.tar.gz",
	}

	ctx := context.Background()
	if err := inst.Install(ctx, tool); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Reload state and verify
	store2, err := state.LoadStore(statePath)
	if err != nil {
		t.Fatalf("failed to reload store: %v", err)
	}
	entry, ok := store2.Tools["statetool"]
	if !ok {
		t.Fatal("state store should have 'statetool' entry")
	}
	if entry.Version != "3.0.0" {
		t.Errorf("stored version = %q, want '3.0.0'", entry.Version)
	}
}

// TestInstallAll_EndToEnd verifies concurrent install of multiple tools.
func TestInstallAll_EndToEnd(t *testing.T) {
	tools := []struct {
		name    string
		content []byte
	}{
		{"alpha", []byte("alpha-bin")},
		{"beta", []byte("beta-bin")},
		{"gamma", []byte("gamma-bin")},
	}

	// Pre-create tar.gz archives for each tool
	archives := make(map[string][]byte)
	for _, tool := range tools {
		archives[tool.name] = createTarGzFixture(t, tool.name, tool.content)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Serve archive based on path (e.g., /alpha.tar.gz)
		for name, data := range archives {
			if r.URL.Path == "/"+name+".tar.gz" {
				w.WriteHeader(http.StatusOK)
				w.Write(data)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	installDir := t.TempDir()
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}
	inst := New(installDir, plat, WithConcurrency(2))

	var toolDefs []*tooldef.Tool
	for _, tool := range tools {
		toolDefs = append(toolDefs, &tooldef.Tool{
			Name:        tool.name,
			Version:     "1.0.0",
			Format:      tooldef.FormatTarGz,
			URLTemplate: srv.URL + "/{{.Name}}.tar.gz",
		})
	}

	ctx := context.Background()
	errs := inst.InstallAll(ctx, toolDefs)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got: %v", errs)
	}

	// Verify all binaries installed
	for _, tool := range tools {
		binPath := filepath.Join(installDir, tool.name)
		got, err := os.ReadFile(binPath)
		if err != nil {
			t.Errorf("binary %s not found: %v", tool.name, err)
			continue
		}
		if !bytes.Equal(got, tool.content) {
			t.Errorf("%s content mismatch: got %q", tool.name, got)
		}
	}
}

// TestInstall_UnsupportedManager verifies error for unknown ManagedBy value.
func TestInstall_UnsupportedManager(t *testing.T) {
	installDir := t.TempDir()
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}
	inst := New(installDir, plat)

	tool := &tooldef.Tool{
		Name:      "managed-tool",
		Version:   "1.0.0",
		ManagedBy: "unknown-manager",
	}

	ctx := context.Background()
	err := inst.Install(ctx, tool)
	if err == nil {
		t.Fatal("expected error for unsupported manager")
	}
}

// TestInstall_StripComponents_Correct verifies strip_components in full flow.
func TestInstall_StripComponents_Correct(t *testing.T) {
	binaryContent := []byte("stripped binary")
	// Archive: prefix/mytool
	tmp := t.TempDir()
	archivePath := filepath.Join(tmp, "fixture.tar.gz")
	f, _ := os.Create(archivePath)
	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)
	tw.WriteHeader(&tar.Header{
		Name: "mytool-1.0.0/mytool",
		Mode: 0o755,
		Size: int64(len(binaryContent)),
	})
	tw.Write(binaryContent)
	tw.Close()
	gw.Close()
	f.Close()

	archive, _ := os.ReadFile(archivePath)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(archive)
	}))
	defer srv.Close()

	installDir := t.TempDir()
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}
	inst := New(installDir, plat)

	tool := &tooldef.Tool{
		Name:            "mytool",
		Version:         "1.0.0",
		Format:          tooldef.FormatTarGz,
		URLTemplate:     srv.URL + "/download",
		StripComponents: 1, // strip "mytool-1.0.0/", leaving "mytool"
	}

	ctx := context.Background()
	if err := inst.Install(ctx, tool); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	binPath := filepath.Join(installDir, "mytool")
	got, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatalf("binary not found: %v", err)
	}
	if !bytes.Equal(got, binaryContent) {
		t.Errorf("content mismatch: got %q", got)
	}
}
