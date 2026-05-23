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

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check system health",
		Long:  "Run diagnostic checks to ensure the system is properly configured.",
		RunE:  runDoctor,
	}
}

func runDoctor(cmd *cobra.Command, args []string) error {
	passStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	failStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	pass := func(msg string) {
		fmt.Println(passStyle.Render(fmt.Sprintf("  ✓ %s", msg)))
	}
	fail := func(msg string) {
		fmt.Println(failStyle.Render(fmt.Sprintf("  ✗ %s", msg)))
	}

	fmt.Println("System Health Check")
	fmt.Println("───────────────────")

	allPassed := true

	// Check ~/.local/bin exists
	localBin := filepath.Join(os.Getenv("HOME"), ".local", "bin")
	if info, err := os.Stat(localBin); err == nil && info.IsDir() {
		pass(fmt.Sprintf("%s exists", localBin))
	} else {
		fail(fmt.Sprintf("%s does not exist", localBin))
		allPassed = false
	}

	// Check ~/.local/bin is in PATH
	pathEnv := os.Getenv("PATH")
	if strings.Contains(pathEnv, localBin) {
		pass(fmt.Sprintf("%s is in PATH", localBin))
	} else {
		fail(fmt.Sprintf("%s is NOT in PATH", localBin))
		allPassed = false
	}

	// Check platform detection
	info, err := platform.Detect()
	if err == nil {
		pass(fmt.Sprintf("Platform detected: %s/%s", info.OS, info.Arch))
	} else {
		fail(fmt.Sprintf("Platform detection failed: %v", err))
		allPassed = false
	}

	// Check git
	if _, err := exec.LookPath("git"); err == nil {
		pass("git is available")
	} else {
		fail("git is NOT available")
		allPassed = false
	}

	// Check curl
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
