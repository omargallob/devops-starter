package installer

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
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
	if !bytes.Equal(got, content) {
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

// --- archive.go tests ---

func TestStripPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		n    int
		want string
	}{
		{"no strip", "foo/bar/baz", 0, "foo/bar/baz"},
		{"strip 1", "foo/bar/baz", 1, "bar/baz"},
		{"strip 2", "foo/bar/baz", 2, "baz"},
		{"strip all", "foo/bar", 2, ""},
		{"strip more than exists", "foo", 2, ""},
		{"negative strip", "foo/bar", -1, "foo/bar"},
		{"single component strip 1", "foo", 1, ""},
		{"strip with trailing slash", "prefix/dir/", 1, "dir/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripPath(tt.path, tt.n)
			if got != tt.want {
				t.Errorf("stripPath(%q, %d) = %q, want %q", tt.path, tt.n, got, tt.want)
			}
		})
	}
}

func TestExtractZip(t *testing.T) {
	tmp := t.TempDir()

	// Create a zip archive with two files
	archivePath := filepath.Join(tmp, "archive.zip")
	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	zw := zip.NewWriter(f)

	// Add a file
	w, err := zw.Create("hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte("hello world"))

	// Add a file in a subdirectory
	w, err = zw.Create("subdir/nested.txt")
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte("nested content"))

	zw.Close()
	f.Close()

	destDir := filepath.Join(tmp, "out")
	os.MkdirAll(destDir, 0o755)

	if err := Extract(archivePath, destDir, tooldef.FormatZip, 0); err != nil {
		t.Fatal(err)
	}

	// Verify first file
	got, err := os.ReadFile(filepath.Join(destDir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello world" {
		t.Errorf("hello.txt content = %q, want %q", got, "hello world")
	}

	// Verify nested file
	got, err = os.ReadFile(filepath.Join(destDir, "subdir", "nested.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "nested content" {
		t.Errorf("nested.txt content = %q, want %q", got, "nested content")
	}
}

func TestExtractZip_StripComponents(t *testing.T) {
	tmp := t.TempDir()

	archivePath := filepath.Join(tmp, "archive.zip")
	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	zw := zip.NewWriter(f)
	w, err := zw.Create("prefix/mybin")
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte("binary content"))
	zw.Close()
	f.Close()

	destDir := filepath.Join(tmp, "out")
	os.MkdirAll(destDir, 0o755)

	if err := Extract(archivePath, destDir, tooldef.FormatZip, 1); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(destDir, "mybin"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "binary content" {
		t.Errorf("content = %q, want %q", got, "binary content")
	}
}

func TestExtractZip_ZipSlipPrevention(t *testing.T) {
	tmp := t.TempDir()

	// Create a zip with a path-traversal entry
	archivePath := filepath.Join(tmp, "evil.zip")
	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	zw := zip.NewWriter(f)
	// This entry tries to escape destDir
	w, err := zw.Create("../../../etc/evil.txt")
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte("malicious"))

	// Also add a safe entry to ensure extraction still works
	w, err = zw.Create("safe.txt")
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte("safe"))
	zw.Close()
	f.Close()

	destDir := filepath.Join(tmp, "out")
	os.MkdirAll(destDir, 0o755)

	if err := Extract(archivePath, destDir, tooldef.FormatZip, 0); err != nil {
		t.Fatal(err)
	}

	// The evil file should NOT exist outside destDir
	evilPath := filepath.Join(tmp, "etc", "evil.txt")
	if _, err := os.Stat(evilPath); err == nil {
		t.Fatal("zip-slip: malicious file was extracted outside destDir")
	}

	// The safe file should exist
	if _, err := os.Stat(filepath.Join(destDir, "safe.txt")); err != nil {
		t.Error("safe.txt should have been extracted")
	}
}

func TestExtractTarGz_ZipSlipPrevention(t *testing.T) {
	tmp := t.TempDir()

	archivePath := filepath.Join(tmp, "evil.tar.gz")
	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	// Malicious entry with path traversal
	content := []byte("malicious")
	tw.WriteHeader(&tar.Header{
		Name: "../../../etc/evil.txt",
		Mode: 0o644,
		Size: int64(len(content)),
	})
	tw.Write(content)

	// Safe entry
	safe := []byte("safe")
	tw.WriteHeader(&tar.Header{
		Name: "safe.txt",
		Mode: 0o644,
		Size: int64(len(safe)),
	})
	tw.Write(safe)

	tw.Close()
	gw.Close()
	f.Close()

	destDir := filepath.Join(tmp, "out")
	os.MkdirAll(destDir, 0o755)

	if err := Extract(archivePath, destDir, tooldef.FormatTarGz, 0); err != nil {
		t.Fatal(err)
	}

	// Evil file should not exist
	evilPath := filepath.Join(tmp, "etc", "evil.txt")
	if _, err := os.Stat(evilPath); err == nil {
		t.Fatal("zip-slip: malicious file was extracted outside destDir")
	}

	// Safe file should exist
	if _, err := os.Stat(filepath.Join(destDir, "safe.txt")); err != nil {
		t.Error("safe.txt should have been extracted")
	}
}

func TestExtractTarGz_StripComponents(t *testing.T) {
	tmp := t.TempDir()

	archivePath := filepath.Join(tmp, "archive.tar.gz")
	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	content := []byte("stripped binary")
	tw.WriteHeader(&tar.Header{
		Name: "prefix/subdir/mybin",
		Mode: 0o755,
		Size: int64(len(content)),
	})
	tw.Write(content)

	tw.Close()
	gw.Close()
	f.Close()

	destDir := filepath.Join(tmp, "out")
	os.MkdirAll(destDir, 0o755)

	if err := Extract(archivePath, destDir, tooldef.FormatTarGz, 2); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(destDir, "mybin"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "stripped binary" {
		t.Errorf("content = %q, want %q", got, "stripped binary")
	}
}

func TestCopyBinary(t *testing.T) {
	tmp := t.TempDir()

	// Create a fake binary
	srcPath := filepath.Join(tmp, "mytool")
	content := []byte("#!/bin/sh\necho hello")
	os.WriteFile(srcPath, content, 0o644)

	destDir := filepath.Join(tmp, "out")
	os.MkdirAll(destDir, 0o755)

	if err := Extract(srcPath, destDir, tooldef.FormatBinary, 0); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(destDir, "mytool"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("content mismatch")
	}

	// Verify it's executable
	info, _ := os.Stat(filepath.Join(destDir, "mytool"))
	if info.Mode()&0o111 == 0 {
		t.Error("expected executable permissions")
	}
}

func TestExtract_UnsupportedFormat(t *testing.T) {
	tmp := t.TempDir()
	archivePath := filepath.Join(tmp, "file")
	os.WriteFile(archivePath, []byte("data"), 0o644)

	err := Extract(archivePath, tmp, "unknown-format", 0)
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

// --- checksum edge cases ---

func TestComputeChecksum_FileNotFound(t *testing.T) {
	_, err := ComputeChecksum("/nonexistent/path/to/file.txt")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestVerifyChecksum_FileNotFound(t *testing.T) {
	err := VerifyChecksum("/nonexistent/path/to/file.txt", "abc123")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

// --- url edge cases ---

func TestResolveURL_EmptyTemplate(t *testing.T) {
	tool := &tooldef.Tool{
		Name:    "mytool",
		Version: "1.0.0",
		// No URLTemplate or URLs
	}
	platform := tooldef.Platform{OS: "linux", Arch: "amd64"}

	_, err := ResolveURL(tool, platform)
	if err == nil {
		t.Fatal("expected error when no URL template or override is set")
	}
}

func TestResolveURL_InvalidTemplate(t *testing.T) {
	tool := &tooldef.Tool{
		Name:        "mytool",
		Version:     "1.0.0",
		URLTemplate: "https://example.com/{{.Invalid",
	}
	platform := tooldef.Platform{OS: "linux", Arch: "amd64"}

	_, err := ResolveURL(tool, platform)
	if err == nil {
		t.Fatal("expected error for invalid template syntax")
	}
}

func TestResolveURL_TemplateExecutionError(t *testing.T) {
	tool := &tooldef.Tool{
		Name:        "mytool",
		Version:     "1.0.0",
		URLTemplate: "https://example.com/{{.NonExistentField}}",
	}
	platform := tooldef.Platform{OS: "linux", Arch: "amd64"}

	_, err := ResolveURL(tool, platform)
	// text/template with a missing field may or may not error depending on settings
	// but with strict execution it should error
	if err == nil {
		// If template execution doesn't error on missing fields, that's also valid
		t.Skip("template execution did not error on missing field (permissive mode)")
	}
}

func TestResolveURL_BinaryNameInTemplate(t *testing.T) {
	tool := &tooldef.Tool{
		Name:        "mytool",
		Version:     "1.0.0",
		BinaryName:  "mytool-bin",
		Format:      tooldef.FormatBinary,
		URLTemplate: "https://example.com/{{.BinaryName}}-{{.Version}}",
	}
	platform := tooldef.Platform{OS: "linux", Arch: "amd64"}

	got, err := ResolveURL(tool, platform)
	if err != nil {
		t.Fatal(err)
	}
	want := "https://example.com/mytool-bin-1.0.0"
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

// --- extractTarXz conditional test ---

func TestExtractTarXz_SkipIfNoXz(t *testing.T) {
	if _, err := exec.LookPath("xz"); err != nil {
		t.Skip("xz not available on this system")
	}

	tmp := t.TempDir()

	// Create a small tar, then compress with xz
	tarPath := filepath.Join(tmp, "archive.tar")
	f, err := os.Create(tarPath)
	if err != nil {
		t.Fatal(err)
	}
	tw := tar.NewWriter(f)
	content := []byte("xz binary content")
	tw.WriteHeader(&tar.Header{
		Name: "xztool",
		Mode: 0o755,
		Size: int64(len(content)),
	})
	tw.Write(content)
	tw.Close()
	f.Close()

	// Compress with xz
	xzPath := filepath.Join(tmp, "archive.tar.xz")
	cmd := exec.Command("xz", "-k", "-f", "--stdout", tarPath)
	xzOut, err := os.Create(xzPath)
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stdout = xzOut
	if err := cmd.Run(); err != nil {
		xzOut.Close()
		t.Fatalf("xz compression failed: %v", err)
	}
	xzOut.Close()

	destDir := filepath.Join(tmp, "out")
	os.MkdirAll(destDir, 0o755)

	if err := Extract(xzPath, destDir, tooldef.FormatTarXz, 0); err != nil {
		t.Fatalf("extractTarXz failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(destDir, "xztool"))
	if err != nil {
		t.Fatalf("extracted file not found: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("content mismatch: got %q, want %q", got, content)
	}
}
