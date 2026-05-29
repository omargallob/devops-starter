package mise

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile_Valid(t *testing.T) {
	content := `[tools]
go = "1.26.3"
python = "3.12"
node = "22"
`
	path := writeTempFile(t, content)
	versions, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile() error: %v", err)
	}
	if len(versions) != 3 {
		t.Fatalf("expected 3 tools, got %d", len(versions))
	}
	tests := map[string]string{
		"go":     "1.26.3",
		"python": "3.12",
		"node":   "22",
	}
	for name, want := range tests {
		got, ok := versions[name]
		if !ok {
			t.Errorf("tool %q not found in parsed versions", name)
			continue
		}
		if got != want {
			t.Errorf("tool %q: got %q, want %q", name, got, want)
		}
	}
}

func TestParseFile_EmptyTools(t *testing.T) {
	content := `[tools]
`
	path := writeTempFile(t, content)
	versions, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile() error: %v", err)
	}
	if len(versions) != 0 {
		t.Fatalf("expected 0 tools, got %d", len(versions))
	}
}

func TestParseFile_NoToolsSection(t *testing.T) {
	content := `[settings]
experimental = true
`
	path := writeTempFile(t, content)
	versions, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile() error: %v", err)
	}
	if len(versions) != 0 {
		t.Fatalf("expected 0 tools, got %d", len(versions))
	}
}

func TestParseFile_ArrayVersion(t *testing.T) {
	content := `[tools]
python = ["3.12", "3.11"]
`
	path := writeTempFile(t, content)
	versions, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile() error: %v", err)
	}
	if got := versions["python"]; got != "3.12" {
		t.Errorf("python version: got %q, want %q", got, "3.12")
	}
}

func TestParseFile_NotFound(t *testing.T) {
	_, err := ParseFile("/nonexistent/.mise.toml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestFind_WalksUp(t *testing.T) {
	// Create temp directory structure: root/.mise.toml and root/sub/deep/
	root := t.TempDir()
	miseFile := filepath.Join(root, ConfigFile)
	if err := os.WriteFile(miseFile, []byte("[tools]\ngo = \"1.21\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	deep := filepath.Join(root, "sub", "deep")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}

	got := Find(deep)
	if got != miseFile {
		t.Errorf("Find(%s) = %q, want %q", deep, got, miseFile)
	}
}

func TestFind_NotFound(t *testing.T) {
	// Use a temp dir with no .mise.toml anywhere above
	dir := t.TempDir()
	got := Find(dir)
	if got != "" {
		t.Errorf("Find(%s) = %q, want empty", dir, got)
	}
}

func TestFindAndParse_NoFile(t *testing.T) {
	dir := t.TempDir()
	versions, err := FindAndParse(dir)
	if err != nil {
		t.Fatalf("FindAndParse() error: %v", err)
	}
	if versions != nil {
		t.Errorf("expected nil versions, got %v", versions)
	}
}

func TestDescriptionFor_Known(t *testing.T) {
	desc := DescriptionFor("go")
	if desc != "Go programming language" {
		t.Errorf("got %q", desc)
	}
}

func TestDescriptionFor_Unknown(t *testing.T) {
	desc := DescriptionFor("obscurelang")
	if desc != "obscurelang (managed by mise)" {
		t.Errorf("got %q", desc)
	}
}

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		input interface{}
		want  string
	}{
		{"1.26.3", "1.26.3"},
		{int64(22), "22"},
		{float64(3.12), "3.12"},
		{float64(22.0), "22"},
		{[]interface{}{"3.12", "3.11"}, "3.12"},
		{[]interface{}{}, ""},
	}
	for _, tt := range tests {
		got := normalizeVersion(tt.input)
		if got != tt.want {
			t.Errorf("normalizeVersion(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ConfigFile)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
