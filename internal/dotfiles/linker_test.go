package dotfiles

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckStatusMissing(t *testing.T) {
	home := t.TempDir()
	source := t.TempDir()

	// Create source file
	srcFile := filepath.Join(source, "zsh/.zshrc")
	os.MkdirAll(filepath.Dir(srcFile), 0o755)
	os.WriteFile(srcFile, []byte("# zshrc"), 0o644)

	linker := NewLinker(source, home, false)
	m := Mapping{Source: "zsh/.zshrc", Dest: ".zshrc"}

	status := linker.checkStatus(m)
	if status != StatusMissing {
		t.Errorf("expected StatusMissing, got %v", status)
	}
}

func TestCheckStatusLinked(t *testing.T) {
	home := t.TempDir()
	source := t.TempDir()

	srcFile := filepath.Join(source, "zsh/.zshrc")
	os.MkdirAll(filepath.Dir(srcFile), 0o755)
	os.WriteFile(srcFile, []byte("# zshrc"), 0o644)

	destFile := filepath.Join(home, ".zshrc")
	os.Symlink(srcFile, destFile)

	linker := NewLinker(source, home, false)
	m := Mapping{Source: "zsh/.zshrc", Dest: ".zshrc"}

	status := linker.checkStatus(m)
	if status != StatusLinked {
		t.Errorf("expected StatusLinked, got %v", status)
	}
}

func TestCheckStatusConflict(t *testing.T) {
	home := t.TempDir()
	source := t.TempDir()

	srcFile := filepath.Join(source, "zsh/.zshrc")
	os.MkdirAll(filepath.Dir(srcFile), 0o755)
	os.WriteFile(srcFile, []byte("# zshrc"), 0o644)

	// Create a regular file at dest (conflict)
	destFile := filepath.Join(home, ".zshrc")
	os.WriteFile(destFile, []byte("# existing"), 0o644)

	linker := NewLinker(source, home, false)
	m := Mapping{Source: "zsh/.zshrc", Dest: ".zshrc"}

	status := linker.checkStatus(m)
	if status != StatusConflict {
		t.Errorf("expected StatusConflict, got %v", status)
	}
}

func TestLinkCreatesSymlink(t *testing.T) {
	home := t.TempDir()
	source := t.TempDir()

	srcFile := filepath.Join(source, "zsh/.zshrc")
	os.MkdirAll(filepath.Dir(srcFile), 0o755)
	os.WriteFile(srcFile, []byte("# zshrc"), 0o644)

	linker := NewLinker(source, home, false)
	mappings := []Mapping{{Source: "zsh/.zshrc", Dest: ".zshrc"}}

	results := linker.Link(mappings)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Error != nil {
		t.Fatalf("unexpected error: %v", results[0].Error)
	}
	if results[0].Status != StatusLinked {
		t.Errorf("expected StatusLinked, got %v", results[0].Status)
	}

	// Verify symlink exists
	target, err := os.Readlink(filepath.Join(home, ".zshrc"))
	if err != nil {
		t.Fatalf("readlink error: %v", err)
	}
	if target != srcFile {
		t.Errorf("symlink target = %q, want %q", target, srcFile)
	}
}

func TestLinkBacksUpConflict(t *testing.T) {
	home := t.TempDir()
	source := t.TempDir()

	srcFile := filepath.Join(source, "zsh/.zshrc")
	os.MkdirAll(filepath.Dir(srcFile), 0o755)
	os.WriteFile(srcFile, []byte("# new zshrc"), 0o644)

	// Create conflicting file
	destFile := filepath.Join(home, ".zshrc")
	os.WriteFile(destFile, []byte("# old zshrc"), 0o644)

	linker := NewLinker(source, home, false)
	mappings := []Mapping{{Source: "zsh/.zshrc", Dest: ".zshrc"}}

	results := linker.Link(mappings)
	if results[0].Error != nil {
		t.Fatalf("unexpected error: %v", results[0].Error)
	}

	// Check backup exists
	backupPath := filepath.Join(home, ".dotfiles.bak", ".zshrc")
	data, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("backup not found: %v", err)
	}
	if string(data) != "# old zshrc" {
		t.Errorf("backup content = %q, want %q", string(data), "# old zshrc")
	}
}

func TestUnlink(t *testing.T) {
	home := t.TempDir()
	source := t.TempDir()

	srcFile := filepath.Join(source, "zsh/.zshrc")
	os.MkdirAll(filepath.Dir(srcFile), 0o755)
	os.WriteFile(srcFile, []byte("# zshrc"), 0o644)

	// Create symlink
	destFile := filepath.Join(home, ".zshrc")
	os.Symlink(srcFile, destFile)

	linker := NewLinker(source, home, false)
	mappings := []Mapping{{Source: "zsh/.zshrc", Dest: ".zshrc"}}

	results := linker.Unlink(mappings)
	if results[0].Error != nil {
		t.Fatalf("unexpected error: %v", results[0].Error)
	}
	if results[0].Status != StatusMissing {
		t.Errorf("expected StatusMissing after unlink, got %v", results[0].Status)
	}

	// Verify file is gone
	if _, err := os.Lstat(destFile); !os.IsNotExist(err) {
		t.Error("expected symlink to be removed")
	}
}

func TestDryRunDoesNotModify(t *testing.T) {
	home := t.TempDir()
	source := t.TempDir()

	srcFile := filepath.Join(source, "zsh/.zshrc")
	os.MkdirAll(filepath.Dir(srcFile), 0o755)
	os.WriteFile(srcFile, []byte("# zshrc"), 0o644)

	linker := NewLinker(source, home, true)
	mappings := []Mapping{{Source: "zsh/.zshrc", Dest: ".zshrc"}}

	linker.Link(mappings)

	// Verify no symlink was created
	destFile := filepath.Join(home, ".zshrc")
	if _, err := os.Lstat(destFile); !os.IsNotExist(err) {
		t.Error("dry-run should not create files")
	}
}

func TestSummary(t *testing.T) {
	results := []LinkResult{
		{Status: StatusLinked},
		{Status: StatusLinked},
		{Status: StatusConflict},
		{Status: StatusMissing},
	}

	got := Summary(results)
	if got != "2 linked, 1 conflicts, 1 missing" {
		t.Errorf("Summary() = %q", got)
	}
}

func TestNestedDestDir(t *testing.T) {
	home := t.TempDir()
	source := t.TempDir()

	srcFile := filepath.Join(source, "starship/starship.toml")
	os.MkdirAll(filepath.Dir(srcFile), 0o755)
	os.WriteFile(srcFile, []byte("[prompt]"), 0o644)

	linker := NewLinker(source, home, false)
	mappings := []Mapping{{Source: "starship/starship.toml", Dest: ".config/starship.toml"}}

	results := linker.Link(mappings)
	if results[0].Error != nil {
		t.Fatalf("unexpected error: %v", results[0].Error)
	}

	// Verify nested dir was created
	destFile := filepath.Join(home, ".config/starship.toml")
	if _, err := os.Lstat(destFile); err != nil {
		t.Errorf("expected symlink at nested path: %v", err)
	}
}
