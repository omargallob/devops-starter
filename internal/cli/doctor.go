package cli

import (
	"fmt"
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

// runDoctor performs a series of health checks:
// - ~/.local/bin exists and is in $PATH
// - Shell RC file has PATH configured
// - Platform can be detected (supported OS/arch)
// - git and curl are available on $PATH
func runDoctor(cmd *cobra.Command, args []string) error {
	passStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	failStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	infoStyle := lipgloss.NewStyle().Faint(true)

	pass := func(msg string) {
		fmt.Println(passStyle.Render(fmt.Sprintf("  ✓ %s", msg)))
	}
	fail := func(msg string) {
		fmt.Println(failStyle.Render(fmt.Sprintf("  ✗ %s", msg)))
	}
	warn := func(msg string) {
		fmt.Println(warnStyle.Render(fmt.Sprintf("    ℹ %s", msg)))
	}
	info := func(msg string) {
		fmt.Println(infoStyle.Render(fmt.Sprintf("    %s", msg)))
	}

	fmt.Println("System Health Check")
	fmt.Println("───────────────────")

	allPassed := true

	// Check ~/.local/bin exists
	localBin := filepath.Join(os.Getenv("HOME"), ".local", "bin")
	localBinExists := false
	if fi, err := os.Stat(localBin); err == nil && fi.IsDir() {
		pass(fmt.Sprintf("%s exists", localBin))
		localBinExists = true
	} else {
		fail(fmt.Sprintf("%s does not exist", localBin))
		allPassed = false

		if doctorFix {
			if err := os.MkdirAll(localBin, 0o755); err != nil {
				fail(fmt.Sprintf("failed to create %s: %v", localBin, err))
			} else {
				pass(fmt.Sprintf("created %s", localBin))
				localBinExists = true
			}
		}
	}

	// Check ~/.local/bin is in PATH
	pathEnv := os.Getenv("PATH")
	pathOK := strings.Contains(pathEnv, localBin)
	if pathOK {
		pass(fmt.Sprintf("%s is in PATH", localBin))
	} else {
		fail(fmt.Sprintf("%s is NOT in PATH", localBin))
		allPassed = false

		// Evaluate shell RC file
		zshrcPath := filepath.Join(os.Getenv("HOME"), ".zshrc")
		rcStatus := evaluateShellRC(zshrcPath, localBin)

		switch rcStatus.state {
		case rcActive:
			warn(fmt.Sprintf("%s is configured in %s (line %d) but not active in current shell",
				localBin, zshrcPath, rcStatus.line))
			info("Run 'source ~/.zshrc' or open a new terminal to activate")
		case rcCommented:
			warn(fmt.Sprintf("Found commented entry in %s (line %d)", zshrcPath, rcStatus.line))
			if doctorFix {
				if err := appendPathToRC(zshrcPath); err != nil {
					fail(fmt.Sprintf("failed to update %s: %v", zshrcPath, err))
				} else {
					pass(fmt.Sprintf("Added 'export PATH=\"$HOME/.local/bin:$PATH\"' to %s", zshrcPath))
					info("Run 'source ~/.zshrc' or open a new terminal to activate")
				}
			} else {
				info("Run 'devops-starter doctor --fix' to add it, or manually add:")
				info("  export PATH=\"$HOME/.local/bin:$PATH\"")
			}
		case rcAbsent:
			warn(fmt.Sprintf("%s has no PATH entry for %s", zshrcPath, localBin))
			if doctorFix {
				if err := appendPathToRC(zshrcPath); err != nil {
					fail(fmt.Sprintf("failed to update %s: %v", zshrcPath, err))
				} else {
					pass(fmt.Sprintf("Added 'export PATH=\"$HOME/.local/bin:$PATH\"' to %s", zshrcPath))
					info("Run 'source ~/.zshrc' or open a new terminal to activate")
				}
			} else {
				info("Run 'devops-starter doctor --fix' to add it, or manually add:")
				info("  export PATH=\"$HOME/.local/bin:$PATH\"")
			}
		}
	}

	// Ensure directory exists if --fix and PATH was just added
	if doctorFix && !localBinExists {
		_ = os.MkdirAll(localBin, 0o755)
	}

	// Check platform detection works
	pinfo, err := platform.Detect()
	if err == nil {
		pass(fmt.Sprintf("Platform detected: %s/%s", pinfo.OS, pinfo.Arch))
	} else {
		fail(fmt.Sprintf("Platform detection failed: %v", err))
		allPassed = false
	}

	// Check git is available (needed for dotfiles operations)
	if _, err := exec.LookPath("git"); err == nil {
		pass("git is available")
	} else {
		fail("git is NOT available")
		allPassed = false
	}

	// Check curl is available (needed for bootstrap install script)
	if _, err := exec.LookPath("curl"); err == nil {
		pass("curl is available")
	} else {
		fail("curl is NOT available")
		allPassed = false
	}

	fmt.Println()
	if allPassed {
		fmt.Println(passStyle.Render("All checks passed!"))
	} else {
		fmt.Println(failStyle.Render("Some checks failed. Please fix the issues above."))
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

// evaluateShellRC reads the given RC file and checks whether it contains
// a PATH entry that includes the target directory.
func evaluateShellRC(rcPath, targetDir string) rcResult {
	data, err := os.ReadFile(rcPath)
	if err != nil {
		return rcResult{state: rcAbsent}
	}

	// Normalize the target for matching — check both expanded and variable forms
	home := os.Getenv("HOME")
	patterns := []string{
		".local/bin",
		targetDir,
	}
	if home != "" {
		patterns = append(patterns, strings.Replace(targetDir, home, "$HOME", 1))
		patterns = append(patterns, strings.Replace(targetDir, home, "~", 1))
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
