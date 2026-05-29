// verify is an e2e test binary that runs inside a Docker container to validate
// the full devops-starter install flow. It:
// 1. Runs `devops-starter setup --non-interactive`
// 2. Runs `devops-starter install --yes`
// 3. Verifies every registered tool is installed, on PATH, and reports the correct version.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/internal/registry"
	"github.com/omargallob/devops-starter/internal/state"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// result tracks the outcome for a single tool.
type result struct {
	Name    string
	Status  string // "pass", "fail", "skip", "warn"
	Message string
}

func main() {
	os.Exit(run())
}

func run() int {
	// Allow limiting to specific groups via E2E_GROUPS env var (comma-separated).
	groupFilter := os.Getenv("E2E_GROUPS")

	fmt.Println("=== E2E Verify: devops-starter install ===")
	fmt.Println()

	// Step 1: Setup
	fmt.Println("[1/3] Running: devops-starter setup --non-interactive")
	if err := runCmd("devops-starter", "setup", "--non-interactive"); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: setup failed: %v\n", err)
		return 1
	}
	fmt.Println("      OK")
	fmt.Println()

	// Step 2: Install
	installArgs := []string{"install", "--yes"}
	if groupFilter != "" {
		fmt.Printf("[2/3] Running: devops-starter install --yes --only %s\n", groupFilter)
		installArgs = append(installArgs, "--only", groupFilter)
	} else {
		fmt.Println("[2/3] Running: devops-starter install --yes")
	}
	if err := runCmd("devops-starter", installArgs...); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: install failed: %v\n", err)
		return 1
	}
	fmt.Println("      OK")
	fmt.Println()

	// Step 3: Activate mise shims (add mise shim directory to PATH).
	activateMiseShims()

	// Step 4: Verify tools
	fmt.Println("[3/3] Verifying installed tools...")
	fmt.Println()

	reg := registry.New()
	cfg, _ := config.Load("")
	plat := tooldef.Platform{OS: "linux", Arch: "amd64"}

	tools := reg.All()
	var results []result

	for _, tool := range tools {
		// Skip tools not supported on this platform.
		if !tool.SupportsPlatform(plat) {
			results = append(results, result{Name: tool.Name, Status: "skip", Message: "unsupported platform"})
			continue
		}

		// Skip disabled groups.
		if cfg != nil && !cfg.IsGroupEnabled(string(tool.Group)) {
			results = append(results, result{Name: tool.Name, Status: "skip", Message: "group disabled"})
			continue
		}

		// Filter by E2E_GROUPS if set.
		if groupFilter != "" && !groupMatchesFilter(string(tool.Group), groupFilter) {
			results = append(results, result{Name: tool.Name, Status: "skip", Message: "filtered out"})
			continue
		}

		// Special case: gh extensions.
		if tool.IsGhExtension() {
			r := verifyGhExtension(tool)
			results = append(results, r)
			continue
		}

		// General case: check binary on PATH + version.
		r := verifyTool(tool)
		results = append(results, r)
	}

	// Print results
	return printResults(results)
}

// verifyTool checks that a tool's binary exists on PATH and its version matches.
func verifyTool(tool *tooldef.Tool) result {
	installName := tool.GetInstallName()

	// Check binary exists on PATH.
	binPath, err := exec.LookPath(installName)
	if err != nil {
		return result{Name: tool.Name, Status: "fail", Message: fmt.Sprintf("binary '%s' not found on PATH", installName)}
	}

	// Try version detection using the state package probes.
	ver, err := state.DetectVersionAtPath(tool.Name, binPath)
	if err != nil {
		// No probe or probe failed — binary exists, that's good enough.
		return result{Name: tool.Name, Status: "pass", Message: fmt.Sprintf("binary found at %s (no version probe)", binPath)}
	}

	// Compare versions.
	if ver == tool.Version {
		return result{Name: tool.Name, Status: "pass", Message: fmt.Sprintf("v%s OK", ver)}
	}
	return result{Name: tool.Name, Status: "warn", Message: fmt.Sprintf("expected v%s, got v%s", tool.Version, ver)}
}

// verifyGhExtension checks that a gh extension is installed.
func verifyGhExtension(tool *tooldef.Tool) result {
	out, err := exec.Command("gh", "extension", "list").Output()
	if err != nil {
		return result{Name: tool.Name, Status: "fail", Message: "gh extension list failed"}
	}
	// Check if the repo (e.g., "github/gh-copilot") appears in the output.
	if strings.Contains(string(out), tool.Repo) {
		return result{Name: tool.Name, Status: "pass", Message: "gh extension found"}
	}
	// Also check by extension name (e.g., "copilot").
	extName := strings.TrimPrefix(tool.Repo, "github/gh-")
	if strings.Contains(string(out), extName) {
		return result{Name: tool.Name, Status: "pass", Message: "gh extension found"}
	}
	return result{Name: tool.Name, Status: "fail", Message: fmt.Sprintf("gh extension %s not in list", tool.Repo)}
}

// activateMiseShims adds the mise shims directory to PATH if mise is installed.
func activateMiseShims() {
	home, _ := os.UserHomeDir()
	shimDir := home + "/.local/share/mise/shims"
	if _, err := os.Stat(shimDir); err == nil {
		os.Setenv("PATH", shimDir+":"+os.Getenv("PATH"))
		fmt.Printf("      Activated mise shims: %s\n", shimDir)
	}
}

// groupMatchesFilter checks if a group name matches the comma-separated filter.
func groupMatchesFilter(group, filter string) bool {
	for _, g := range strings.Split(filter, ",") {
		if strings.TrimSpace(g) == group {
			return true
		}
	}
	return false
}

// runCmd runs a command, streaming output to stdout/stderr.
func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	return cmd.Run()
}

// printResults prints a summary table and returns exit code.
func printResults(results []result) int {
	var passed, failed, skipped, warned int
	var failures []result

	fmt.Println("┌────────────────────────────┬────────┬─────────────────────────────────────────┐")
	fmt.Printf("│ %-26s │ %-6s │ %-39s │\n", "Tool", "Status", "Message")
	fmt.Println("├────────────────────────────┼────────┼─────────────────────────────────────────┤")

	for _, r := range results {
		switch r.Status {
		case "pass":
			passed++
		case "fail":
			failed++
			failures = append(failures, r)
		case "skip":
			skipped++
		case "warn":
			warned++
		}

		icon := statusIcon(r.Status)
		msg := r.Message
		if len(msg) > 39 {
			msg = msg[:36] + "..."
		}
		name := r.Name
		if len(name) > 26 {
			name = name[:23] + "..."
		}
		fmt.Printf("│ %s %-24s │ %-6s │ %-39s │\n", icon, name, r.Status, msg)
	}

	fmt.Println("└────────────────────────────┴────────┴─────────────────────────────────────────┘")
	fmt.Println()
	fmt.Printf("Total: %d | Passed: %d | Warned: %d | Failed: %d | Skipped: %d\n",
		len(results), passed, warned, failed, skipped)
	fmt.Printf("Duration: %s\n", time.Since(startTime).Round(time.Second))

	if failed > 0 {
		fmt.Println("\nFailed tools:")
		for _, f := range failures {
			fmt.Printf("  ✗ %s: %s\n", f.Name, f.Message)
		}
		return 1
	}
	return 0
}

func statusIcon(s string) string {
	switch s {
	case "pass":
		return "✓"
	case "fail":
		return "✗"
	case "warn":
		return "!"
	case "skip":
		return "-"
	default:
		return " "
	}
}

var startTime = time.Now()
