package installer

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/omargallob/devops-starter/pkg/tooldef"
)

func TestComputeChecksum(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "test.txt")
	data := []byte("hello world\n")
	os.WriteFile(f, data, 0o644)

	got, err := ComputeChecksum(f)
	if err != nil {
		t.Fatal(err)
	}

	h := sha256.Sum256(data)
	want := hex.EncodeToString(h[:])
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestVerifyChecksum(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "test.txt")
	data := []byte("test data")
	os.WriteFile(f, data, 0o644)

	h := sha256.Sum256(data)
	expected := hex.EncodeToString(h[:])

	if err := VerifyChecksum(f, expected); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if err := VerifyChecksum(f, "badchecksum"); err == nil {
		t.Error("expected error for bad checksum")
	}
}

func TestResolveURL_Template(t *testing.T) {
	tool := &tooldef.Tool{
		Name:        "mytool",
		Version:     "1.2.3",
		Format:      tooldef.FormatTarGz,
		URLTemplate: "https://example.com/{{.Name}}/{{.Version}}/{{.OS}}-{{.Arch}}.{{.Format}}",
	}
	platform := tooldef.Platform{OS: "linux", Arch: "amd64"}

	got, err := ResolveURL(tool, platform)
	if err != nil {
		t.Fatal(err)
	}
	want := "https://example.com/mytool/1.2.3/linux-amd64.tar.gz"
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestResolveURL_Override(t *testing.T) {
	tool := &tooldef.Tool{
		Name:        "mytool",
		Version:     "1.0.0",
		URLTemplate: "https://example.com/default",
		URLs: map[string]string{
			"darwin/arm64": "https://example.com/special-darwin-arm64",
		},
	}
	platform := tooldef.Platform{OS: "darwin", Arch: "arm64"}

	got, err := ResolveURL(tool, platform)
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://example.com/special-darwin-arm64" {
		t.Errorf("got %s, want override URL", got)
	}
}

func TestIsInstalled(t *testing.T) {
	tmp := t.TempDir()
	inst := New(tmp, tooldef.Platform{OS: "linux", Arch: "amd64"})

	tool := &tooldef.Tool{Name: "kubectl"}

	if inst.IsInstalled(tool) {
		t.Error("should not be installed")
	}

	os.WriteFile(filepath.Join(tmp, "kubectl"), []byte("bin"), 0o755)

	if !inst.IsInstalled(tool) {
		t.Error("should be installed")
	}
}

func TestExtractTarGz(t *testing.T) {
	tmp := t.TempDir()

	// Create a tar.gz with a single file: "mybin"
	archivePath := filepath.Join(tmp, "archive.tar.gz")
	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	content := []byte("#!/bin/sh\necho hello")
	tw.WriteHeader(&tar.Header{
		Name: "mybin",
		Mode: 0o755,
		Size: int64(len(content)),
	})
	tw.Write(content)
	tw.Close()
	gw.Close()
	f.Close()

	destDir := filepath.Join(tmp, "out")
	os.MkdirAll(destDir, 0o755)

	if err := Extract(archivePath, destDir, tooldef.FormatTarGz, 0); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(destDir, "mybin"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Errorf("content mismatch")
	}
}

func TestEnsureDir(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "a", "b", "c")

	inst := New(dir, tooldef.Platform{OS: "linux", Arch: "amd64"})
	if err := inst.EnsureDir(); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestNew_Defaults(t *testing.T) {
	plat := tooldef.Platform{OS: "darwin", Arch: "arm64"}
	inst := New("/tmp/test", plat)

	if inst.InstallDir != "/tmp/test" {
		t.Errorf("InstallDir = %s, want /tmp/test", inst.InstallDir)
	}
	if inst.Platform != plat {
		t.Errorf("Platform mismatch")
	}
	if inst.Concurrency != 4 {
		t.Errorf("Concurrency = %d, want 4", inst.Concurrency)
	}
	if inst.DryRun {
		t.Error("DryRun should default to false")
	}
	if inst.StateStore != nil {
		t.Error("StateStore should default to nil")
	}
}

func TestNew_WithOptions(t *testing.T) {
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}
	inst := New("/tmp/test", plat,
		WithDryRun(true),
		WithConcurrency(8),
	)

	if !inst.DryRun {
		t.Error("expected DryRun=true")
	}
	if inst.Concurrency != 8 {
		t.Errorf("Concurrency = %d, want 8", inst.Concurrency)
	}
}

func TestInstall_DryRun(t *testing.T) {
	tmp := t.TempDir()
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}
	inst := New(tmp, plat, WithDryRun(true))

	tool := &tooldef.Tool{
		Name:    "fake-tool",
		Version: "1.0.0",
	}

	ctx := t.Context()
	if err := inst.Install(ctx, tool); err != nil {
		t.Fatalf("dry-run install should not error: %v", err)
	}

	// Binary should NOT be installed
	if inst.IsInstalled(tool) {
		t.Error("dry-run should not actually install")
	}
}

func TestInstallAll_DryRun(t *testing.T) {
	tmp := t.TempDir()
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}
	inst := New(tmp, plat, WithDryRun(true))

	tools := []*tooldef.Tool{
		{Name: "tool-a", Version: "1.0.0"},
		{Name: "tool-b", Version: "2.0.0"},
		{Name: "tool-c", Version: "3.0.0"},
	}

	ctx := t.Context()
	errs := inst.InstallAll(ctx, tools)
	if len(errs) != 0 {
		t.Fatalf("expected no errors in dry-run, got %v", errs)
	}
}

func TestInstall_NoURLTemplate(t *testing.T) {
	tmp := t.TempDir()
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}
	inst := New(tmp, plat)

	tool := &tooldef.Tool{
		Name:    "no-url-tool",
		Version: "1.0.0",
		// No URLTemplate or URLs set
	}

	ctx := t.Context()
	err := inst.Install(ctx, tool)
	if err == nil {
		t.Fatal("expected error when no URL template is set")
	}
}
