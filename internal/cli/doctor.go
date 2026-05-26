package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/omargallob/devops-starter/internal/platform"
)

var doctorFix bool

// newDoctorCmd creates the "doctor" subcommand which runs diagnostic checks
// to verify the system is properly configured for devops-starter.
func newDoctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check system health",
		Long:  "Run diagnostic checks to ensure the system is properly configured.",
		RunE:  runDoctor,
	}

	cmd.Flags().BoolVar(&doctorFix, "fix", false, "automatically fix issues where possible")

	return cmd
}

// doctorEnv holds the environment queries for the doctor command, allowing
// tests to inject fakes.
type doctorEnv struct {
	lookPath    func(name string) (string, error)
	getenv      func(key string) string
	stat        func(path string) (os.FileInfo, error)
	readFile    func(path string) ([]byte, error)
	mkdirAll    func(path string, perm os.FileMode) error
	openFile    func(name string, flag int, perm os.FileMode) (*os.File, error)
	detectPlat  func() (*platform.Info, error)
}

// realDoctorEnv returns the real OS environment for production use.
func realDoctorEnv() doctorEnv {
	return doctorEnv{
		lookPath:   exec.LookPath,
		getenv:     os.Getenv,
		stat:       os.Stat,
		readFile:   os.ReadFile,
		mkdirAll:   os.MkdirAll,
		openFile:   os.OpenFile,
		detectPlat: platform.Detect,
	}
}

// runDoctor is the Cobra entry point that delegates to doDoctor with real env.
func runDoctor(cmd *cobra.Command, args []string) error {
	return doDoctor(os.Stdout, realDoctorEnv(), doctorFix)
}

// doctorPrinter holds styled output functions for doctor checks.
type doctorPrinter struct {
	out  io.Writer
	pass func(string)
	fail func(string)
	warn func(string)
	info func(string)
}

func newDoctorPrinter(out io.Writer) doctorPrinter {
	passStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	failStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	infoStyle := lipgloss.NewStyle().Faint(true)

	return doctorPrinter{
		out: out,
		pass: func(msg string) {
			fmt.Fprintln(out, passStyle.Render(fmt.Sprintf("  ✓ %s", msg)))
		},
		fail: func(msg string) {
			fmt.Fprintln(out, failStyle.Render(fmt.Sprintf("  ✗ %s", msg)))
		},
		warn: func(msg string) {
			fmt.Fprintln(out, warnStyle.Render(fmt.Sprintf("    ℹ %s", msg)))
		},
		info: func(msg string) {
			fmt.Fprintln(out, infoStyle.Render(fmt.Sprintf("    %s", msg)))
		},
	}
}

// checkLocalBinExists checks if ~/.local/bin exists, optionally creating it.
func checkLocalBinExists(env doctorEnv, p doctorPrinter, localBin string, fix bool) (exists, passed bool) {
	if fi, err := env.stat(localBin); err == nil && fi.IsDir() {
		p.pass(fmt.Sprintf("%s exists", localBin))
		return true, true
	}

	p.fail(fmt.Sprintf("%s does not exist", localBin))
	if fix {
		if err := env.mkdirAll(localBin, 0o755); err != nil {
			p.fail(fmt.Sprintf("failed to create %s: %v", localBin, err))
		} else {
			p.pass(fmt.Sprintf("created %s", localBin))
			return true, false
		}
	}
	return false, false
}

// checkLocalBinInPath checks if ~/.local/bin is in the PATH and handles RC file fixes.
func checkLocalBinInPath(env doctorEnv, p doctorPrinter, home, localBin string, fix bool) bool {
	pathEnv := env.getenv("PATH")
	if strings.Contains(pathEnv, localBin) {
		p.pass(fmt.Sprintf("%s is in PATH", localBin))
		return true
	}

	p.fail(fmt.Sprintf("%s is NOT in PATH", localBin))

	zshrcPath := filepath.Join(home, ".zshrc")
	rcStatus := evaluateShellRCWith(env.readFile, env.getenv, zshrcPath, localBin)

	switch rcStatus.state {
	case rcActive:
		p.warn(fmt.Sprintf("%s is configured in %s (line %d) but not active in current shell",
			localBin, zshrcPath, rcStatus.line))
		p.info("Run 'source ~/.zshrc' or open a new terminal to activate")
	case rcCommented, rcAbsent:
		if rcStatus.state == rcCommented {
			p.warn(fmt.Sprintf("Found commented entry in %s (line %d)", zshrcPath, rcStatus.line))
		} else {
			p.warn(fmt.Sprintf("%s has no PATH entry for %s", zshrcPath, localBin))
		}
		fixOrAdvisePathRC(p, fix, zshrcPath)
	}

	return false
}

// fixOrAdvisePathRC either fixes the RC file or prints instructions.
func fixOrAdvisePathRC(p doctorPrinter, fix bool, zshrcPath string) {
	if fix {
		if err := appendPathToRC(zshrcPath); err != nil {
			p.fail(fmt.Sprintf("failed to update %s: %v", zshrcPath, err))
		} else {
			p.pass(fmt.Sprintf("Added 'export PATH=\"$HOME/.local/bin:$PATH\"' to %s", zshrcPath))
			p.info("Run 'source ~/.zshrc' or open a new terminal to activate")
		}
	} else {
		p.info("Run 'devops-starter doctor --fix' to add it, or manually add:")
		p.info("  export PATH=\"$HOME/.local/bin:$PATH\"")
	}
}

// checkCommandAvailable checks if a command is available in PATH.
func checkCommandAvailable(env doctorEnv, p doctorPrinter, cmd string) bool {
	if _, err := env.lookPath(cmd); err == nil {
		p.pass(fmt.Sprintf("%s is available", cmd))
		return true
	}
	p.fail(fmt.Sprintf("%s is NOT available", cmd))
	return false
}

// doDoctor performs a series of health checks. It is extracted for testability.
func doDoctor(out io.Writer, env doctorEnv, fix bool) error {
	p := newDoctorPrinter(out)
	passStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	failStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	fmt.Fprintln(out, "System Health Check")
	fmt.Fprintln(out, "───────────────────")

	allPassed := true
	home := env.getenv("HOME")
	localBin := filepath.Join(home, ".local", "bin")

	localBinExists, passed := checkLocalBinExists(env, p, localBin, fix)
	if !passed {
		allPassed = false
	}

	if !checkLocalBinInPath(env, p, home, localBin, fix) {
		allPassed = false
	}

	// Ensure directory exists if --fix and PATH was just added
	if fix && !localBinExists {
		_ = env.mkdirAll(localBin, 0o755)
	}

	// Check platform detection works
	if pinfo, err := env.detectPlat(); err == nil {
		p.pass(fmt.Sprintf("Platform detected: %s/%s", pinfo.OS, pinfo.Arch))
	} else {
		p.fail(fmt.Sprintf("Platform detection failed: %v", err))
		allPassed = false
	}

	if !checkCommandAvailable(env, p, "git") {
		allPassed = false
	}
	if !checkCommandAvailable(env, p, "curl") {
		allPassed = false
	}

	fmt.Fprintln(out)
	if allPassed {
		fmt.Fprintln(out, passStyle.Render("All checks passed!"))
	} else {
		fmt.Fprintln(out, failStyle.Render("Some checks failed. Please fix the issues above."))
	}

	return nil
}

// rcState represents the state of a PATH entry in a shell RC file.
type rcState int

const (
	rcAbsent    rcState = iota // No entry found
	rcCommented                // Entry exists but is commented out
	rcActive                   // Entry exists and is active
)

// rcResult holds the evaluation result of a shell RC file.
type rcResult struct {
	state rcState
	line  int // 1-indexed line number where entry was found (0 if absent)
}

// evaluateShellRCWith reads the given RC file and checks whether it contains
// a PATH entry that includes the target directory. It accepts function deps
// for testability.
func evaluateShellRCWith(readFile func(string) ([]byte, error), getenv func(string) string, rcPath, targetDir string) rcResult {
	data, err := readFile(rcPath)
	if err != nil {
		return rcResult{state: rcAbsent}
	}

	// Normalize the target for matching — check both expanded and variable forms
	home := getenv("HOME")
	patterns := []string{
		".local/bin",
		targetDir,
	}
	if home != "" {
		patterns = append(patterns, strings.Replace(targetDir, home, "$HOME", 1), strings.Replace(targetDir, home, "~", 1))
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Must reference PATH to be relevant
		if !strings.Contains(trimmed, "PATH") {
			continue
		}

		// Check if any pattern matches
		matched := false
		for _, pat := range patterns {
			if strings.Contains(trimmed, pat) {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}

		// Determine if commented or active
		if strings.HasPrefix(trimmed, "#") {
			return rcResult{state: rcCommented, line: i + 1}
		}
		return rcResult{state: rcActive, line: i + 1}
	}

	return rcResult{state: rcAbsent}
}

// appendPathToRC appends an export PATH line to the given RC file.
// Creates the file if it doesn't exist.
func appendPathToRC(rcPath string) error {
	entry := "\n# Added by devops-starter\nexport PATH=\"$HOME/.local/bin:$PATH\"\n"

	f, err := os.OpenFile(rcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(entry)
	return err
}
