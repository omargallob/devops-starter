package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/omargallob/devops-starter/internal/platform"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

func TestDoctorCmd_Exists(t *testing.T) {
	root := NewRootCmd()
	var found bool
	for _, cmd := range root.Commands() {
		if cmd.Use == "doctor" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("doctor command not registered on root")
	}
}

func TestDoDoctor_AllPassing(t *testing.T) {
	home := t.TempDir()
	localBin := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(localBin, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	env := doctorEnv{
		lookPath: func(name string) (string, error) {
			return "/usr/bin/" + name, nil
		},
		getenv: func(key string) string {
			switch key {
			case "HOME":
				return home
			case "PATH":
				return localBin + ":/usr/bin:/bin"
			default:
				return ""
			}
		},
		stat:     os.Stat,
		readFile: os.ReadFile,
		mkdirAll: os.MkdirAll,
		openFile: os.OpenFile,
		detectPlat: func() (*platform.Info, error) {
			return &platform.Info{
				OS:   "darwin",
				Arch: "arm64",
				Platform: tooldef.Platform{
					OS:   "darwin",
					Arch: "arm64",
				},
			}, nil
		},
	}

	var buf bytes.Buffer
	err := doDoctor(&buf, env, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "All checks passed") {
		t.Errorf("expected all checks to pass, got:\n%s", output)
	}
}

func TestDoDoctor_MissingLocalBin(t *testing.T) {
	home := t.TempDir()
	// Don't create .local/bin

	env := doctorEnv{
		lookPath: func(name string) (string, error) {
			return "/usr/bin/" + name, nil
		},
		getenv: func(key string) string {
			switch key {
			case "HOME":
				return home
			case "PATH":
				return "/usr/bin:/bin"
			default:
				return ""
			}
		},
		stat:     os.Stat,
		readFile: os.ReadFile,
		mkdirAll: os.MkdirAll,
		openFile: os.OpenFile,
		detectPlat: func() (*platform.Info, error) {
			return &platform.Info{OS: "darwin", Arch: "arm64", Platform: tooldef.Platform{OS: "darwin", Arch: "arm64"}}, nil
		},
	}

	var buf bytes.Buffer
	_ = doDoctor(&buf, env, false)

	output := buf.String()
	if !strings.Contains(output, "does not exist") {
		t.Errorf("expected 'does not exist' failure, got:\n%s", output)
	}
	if !strings.Contains(output, "Some checks failed") {
		t.Errorf("expected 'Some checks failed', got:\n%s", output)
	}
}

func TestDoDoctor_FixCreatesLocalBin(t *testing.T) {
	home := t.TempDir()
	localBin := filepath.Join(home, ".local", "bin")

	env := doctorEnv{
		lookPath: func(name string) (string, error) {
			return "/usr/bin/" + name, nil
		},
		getenv: func(key string) string {
			switch key {
			case "HOME":
				return home
			case "PATH":
				return localBin + ":/usr/bin:/bin"
			default:
				return ""
			}
		},
		stat:     os.Stat,
		readFile: os.ReadFile,
		mkdirAll: os.MkdirAll,
		openFile: os.OpenFile,
		detectPlat: func() (*platform.Info, error) {
			return &platform.Info{OS: "darwin", Arch: "arm64", Platform: tooldef.Platform{OS: "darwin", Arch: "arm64"}}, nil
		},
	}

	var buf bytes.Buffer
	_ = doDoctor(&buf, env, true)

	// Directory should have been created
	if fi, err := os.Stat(localBin); err != nil || !fi.IsDir() {
		t.Error("expected --fix to create .local/bin")
	}

	output := buf.String()
	if !strings.Contains(output, "created") {
		t.Errorf("expected 'created' message, got:\n%s", output)
	}
}

func TestDoDoctor_MissingGit(t *testing.T) {
	home := t.TempDir()
	localBin := filepath.Join(home, ".local", "bin")
	os.MkdirAll(localBin, 0o755)

	env := doctorEnv{
		lookPath: func(name string) (string, error) {
			if name == "git" {
				return "", fmt.Errorf("not found")
			}
			return "/usr/bin/" + name, nil
		},
		getenv: func(key string) string {
			switch key {
			case "HOME":
				return home
			case "PATH":
				return localBin + ":/usr/bin:/bin"
			default:
				return ""
			}
		},
		stat:     os.Stat,
		readFile: os.ReadFile,
		mkdirAll: os.MkdirAll,
		openFile: os.OpenFile,
		detectPlat: func() (*platform.Info, error) {
			return &platform.Info{OS: "darwin", Arch: "arm64", Platform: tooldef.Platform{OS: "darwin", Arch: "arm64"}}, nil
		},
	}

	var buf bytes.Buffer
	_ = doDoctor(&buf, env, false)

	output := buf.String()
	if !strings.Contains(output, "git is NOT available") {
		t.Errorf("expected git failure message, got:\n%s", output)
	}
}

func TestDoDoctor_PlatformDetectionFailure(t *testing.T) {
	home := t.TempDir()
	localBin := filepath.Join(home, ".local", "bin")
	os.MkdirAll(localBin, 0o755)

	env := doctorEnv{
		lookPath: func(name string) (string, error) {
			return "/usr/bin/" + name, nil
		},
		getenv: func(key string) string {
			switch key {
			case "HOME":
				return home
			case "PATH":
				return localBin + ":/usr/bin:/bin"
			default:
				return ""
			}
		},
		stat:     os.Stat,
		readFile: os.ReadFile,
		mkdirAll: os.MkdirAll,
		openFile: os.OpenFile,
		detectPlat: func() (*platform.Info, error) {
			return nil, fmt.Errorf("unsupported platform: plan9")
		},
	}

	var buf bytes.Buffer
	_ = doDoctor(&buf, env, false)

	output := buf.String()
	if !strings.Contains(output, "Platform detection failed") {
		t.Errorf("expected platform detection failure, got:\n%s", output)
	}
}

func TestDoDoctor_PathNotInShell_RCAbsent(t *testing.T) {
	home := t.TempDir()
	localBin := filepath.Join(home, ".local", "bin")
	os.MkdirAll(localBin, 0o755)

	// Create a .zshrc without PATH entry
	zshrc := filepath.Join(home, ".zshrc")
	os.WriteFile(zshrc, []byte("# Just a comment\nalias ll='ls -la'\n"), 0o644)

	env := doctorEnv{
		lookPath: func(name string) (string, error) {
			return "/usr/bin/" + name, nil
		},
		getenv: func(key string) string {
			switch key {
			case "HOME":
				return home
			case "PATH":
				return "/usr/bin:/bin" // localBin NOT in PATH
			default:
				return ""
			}
		},
		stat:     os.Stat,
		readFile: os.ReadFile,
		mkdirAll: os.MkdirAll,
		openFile: os.OpenFile,
		detectPlat: func() (*platform.Info, error) {
			return &platform.Info{OS: "darwin", Arch: "arm64", Platform: tooldef.Platform{OS: "darwin", Arch: "arm64"}}, nil
		},
	}

	var buf bytes.Buffer
	_ = doDoctor(&buf, env, false)

	output := buf.String()
	if !strings.Contains(output, "is NOT in PATH") {
		t.Errorf("expected PATH warning, got:\n%s", output)
	}
	if !strings.Contains(output, "has no PATH entry") {
		t.Errorf("expected 'has no PATH entry' info, got:\n%s", output)
	}
}

func TestDoDoctor_PathNotInShell_FixAppendsToRC(t *testing.T) {
	home := t.TempDir()
	localBin := filepath.Join(home, ".local", "bin")
	os.MkdirAll(localBin, 0o755)

	zshrc := filepath.Join(home, ".zshrc")
	os.WriteFile(zshrc, []byte("# Just a comment\n"), 0o644)

	env := doctorEnv{
		lookPath: func(name string) (string, error) {
			return "/usr/bin/" + name, nil
		},
		getenv: func(key string) string {
			switch key {
			case "HOME":
				return home
			case "PATH":
				return "/usr/bin:/bin"
			default:
				return ""
			}
		},
		stat:     os.Stat,
		readFile: os.ReadFile,
		mkdirAll: os.MkdirAll,
		openFile: os.OpenFile,
		detectPlat: func() (*platform.Info, error) {
			return &platform.Info{OS: "darwin", Arch: "arm64", Platform: tooldef.Platform{OS: "darwin", Arch: "arm64"}}, nil
		},
	}

	var buf bytes.Buffer
	_ = doDoctor(&buf, env, true)

	// Verify the RC file was updated
	data, err := os.ReadFile(zshrc)
	if err != nil {
		t.Fatalf("failed to read zshrc: %v", err)
	}
	if !strings.Contains(string(data), "export PATH=\"$HOME/.local/bin:$PATH\"") {
		t.Errorf("expected PATH export added to .zshrc, got:\n%s", string(data))
	}
}

func TestEvaluateShellRC_ActiveEntry(t *testing.T) {
	dir := t.TempDir()
	rcPath := filepath.Join(dir, ".zshrc")
	content := "# other stuff\nexport PATH=\"$HOME/.local/bin:$PATH\"\nalias foo=bar\n"
	os.WriteFile(rcPath, []byte(content), 0o644)

	result := evaluateShellRCWith(os.ReadFile, func(key string) string {
		if key == "HOME" {
			return "/home/user"
		}
		return ""
	}, rcPath, "/home/user/.local/bin")

	if result.state != rcActive {
		t.Errorf("expected rcActive, got %d", result.state)
	}
	if result.line != 2 {
		t.Errorf("expected line 2, got %d", result.line)
	}
}

func TestEvaluateShellRC_CommentedEntry(t *testing.T) {
	dir := t.TempDir()
	rcPath := filepath.Join(dir, ".zshrc")
	content := "# export PATH=\"$HOME/.local/bin:$PATH\"\n"
	os.WriteFile(rcPath, []byte(content), 0o644)

	result := evaluateShellRCWith(os.ReadFile, func(key string) string {
		if key == "HOME" {
			return "/home/user"
		}
		return ""
	}, rcPath, "/home/user/.local/bin")

	if result.state != rcCommented {
		t.Errorf("expected rcCommented, got %d", result.state)
	}
}

func TestEvaluateShellRC_Absent(t *testing.T) {
	dir := t.TempDir()
	rcPath := filepath.Join(dir, ".zshrc")
	content := "alias ll='ls -la'\n"
	os.WriteFile(rcPath, []byte(content), 0o644)

	result := evaluateShellRCWith(os.ReadFile, func(key string) string {
		if key == "HOME" {
			return "/home/user"
		}
		return ""
	}, rcPath, "/home/user/.local/bin")

	if result.state != rcAbsent {
		t.Errorf("expected rcAbsent, got %d", result.state)
	}
}

func TestEvaluateShellRC_FileNotFound(t *testing.T) {
	result := evaluateShellRCWith(os.ReadFile, os.Getenv, "/nonexistent/.zshrc", "/home/user/.local/bin")
	if result.state != rcAbsent {
		t.Errorf("expected rcAbsent for missing file, got %d", result.state)
	}
}
