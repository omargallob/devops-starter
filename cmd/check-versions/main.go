// check-versions queries GitHub Releases and other APIs to find tools in the
// registry with available version updates. With --apply it rewrites the
// registry source files with the new versions.
//
// Usage:
//
//	bazel run //cmd/check-versions          # print outdated tools
//	bazel run //cmd/check-versions -- --apply  # rewrite registry files
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/omargallob/devops-starter/internal/registry"
	"github.com/omargallob/devops-starter/internal/updater"
)

func main() {
	os.Exit(run())
}

func run() int {
	apply := false
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--apply":
			apply = true
		case "--help", "-h":
			fmt.Println("Usage: check-versions [--apply]")
			fmt.Println()
			fmt.Println("Checks all registry tools for available version updates.")
			fmt.Println("With --apply, rewrites internal/registry/*.go files.")
			return 0
		default:
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n", arg)
			return 1
		}
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}

	opts := updater.DefaultOptions()
	opts.Token = token
	opts.Out = os.Stderr

	// Build registry to get all tools.
	reg := registry.New()
	tools := reg.All()

	fmt.Fprintf(os.Stderr, "Checking %d tools for updates...\n", len(tools))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	result := updater.CheckAll(ctx, tools, opts)

	// Print results.
	if len(result.Updates) == 0 {
		fmt.Println("All tools are up to date.")
		return 0
	}

	fmt.Printf("\nFound %d tool(s) with available updates:\n\n", len(result.Updates))
	fmt.Printf("  %-20s %-14s %s\n", "Tool", "Current", "Latest")
	fmt.Printf("  %-20s %-14s %s\n", "────", "───────", "──────")
	for _, u := range result.Updates {
		fmt.Printf("  %-20s %-14s %s\n", u.ToolName, u.CurrentVersion, u.LatestVersion)
	}
	fmt.Println()

	if len(result.Skipped) > 0 {
		fmt.Fprintf(os.Stderr, "Skipped %d tools (no version source or API error)\n", len(result.Skipped))
	}

	if !apply {
		fmt.Println("Run with --apply to update registry files.")
		return 0
	}

	// Determine project root (the directory containing internal/registry).
	rootDir := findProjectRoot()
	if rootDir == "" {
		fmt.Fprintln(os.Stderr, "error: could not find project root (looking for internal/registry)")
		return 1
	}

	applied, errs := updater.ApplyUpdates(rootDir, result.Updates)
	if len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "\n%d error(s) during rewrite:\n", len(errs))
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  - %v\n", e)
		}
	}

	fmt.Printf("Updated %d tool version(s) in registry source files.\n", applied)

	if len(errs) > 0 {
		return 1
	}
	return 0
}

// findProjectRoot walks up from CWD looking for the internal/registry directory.
func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		candidate := filepath.Join(dir, "internal", "registry")
		if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return ""
}
