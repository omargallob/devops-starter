package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/internal/installer"
	"github.com/omargallob/devops-starter/internal/platform"
	"github.com/omargallob/devops-starter/internal/registry"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// newInstallCmd creates the "install" subcommand which downloads and installs
// tools based on the user's configuration, platform, and optional --only filter.
func newInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install DevOps tools",
		Long:  "Download and install DevOps tools based on configuration and platform detection.",
		RunE:  runInstall,
	}

	cmd.Flags().StringVar(&only, "only", "", "install only tools from this group")

	return cmd
}

// runInstall is the main logic for the install command. It:
// 1. Detects the current platform (OS/arch/distro)
// 2. Loads user configuration (group enables, version overrides)
// 3. Filters the registry to applicable tools
// 4. Runs concurrent installations with progress output
// 5. Prints a summary of successes and failures
func runInstall(cmd *cobra.Command, args []string) error {
	// Detect platform
	info, err := platform.Detect()
	if err != nil {
		return fmt.Errorf("detecting platform: %w", err)
	}

	// Load config
	cfgPath := cfgFile
	if cfgPath == "" {
		cfgPath = config.ConfigPath()
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Get tools from registry
	reg := registry.New()
	allTools := reg.All()

	// Filter tools by group, platform support, and per-tool overrides
	var tools []*tooldef.Tool
	for _, t := range allTools {
		// Filter by --only flag
		if only != "" && string(t.Group) != only {
			continue
		}

		// Filter by config group enables
		if !cfg.IsGroupEnabled(string(t.Group)) {
			continue
		}

		// Check if tool supports this platform
		if !t.SupportsPlatform(info.Platform) {
			continue
		}

		// Check if tool is disabled in overrides
		if override, ok := cfg.Overrides[t.Name]; ok {
			if override.Disabled {
				continue
			}
			// Apply version override
			if override.Version != "" {
				t.Version = override.Version
			}
		}

		tools = append(tools, t)
	}

	if len(tools) == 0 {
		fmt.Println("No tools to install.")
		return nil
	}

	// Show what will be installed and confirm
	fmt.Printf("\nThe following %d tool(s) will be installed to %s:\n\n", len(tools), cfg.InstallDir)
	for _, t := range tools {
		if t.ManagedBy != "" {
			fmt.Printf("  • %s %s (via %s)\n", t.Name, t.Version, t.ManagedBy)
		} else {
			fmt.Printf("  • %s %s\n", t.Name, t.Version)
		}
	}
	fmt.Println()

	if !dryRun && !confirmAction("Proceed with installation?") {
		fmt.Println("Cancelled.")
		return nil
	}

	// Create installer with platform info and dry-run setting
	inst := installer.New(
		cfg.InstallDir,
		info.Platform,
		installer.WithDryRun(dryRun),
	)

	// Install all filtered tools concurrently
	ctx := context.Background()
	errs := inst.InstallAll(ctx, tools)

	// Print summary with coloured output
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	installed := len(tools) - len(errs)
	fmt.Println()
	fmt.Println(successStyle.Render(fmt.Sprintf("✓ %d tools installed successfully", installed)))

	if len(errs) > 0 {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ %d tools failed:", len(errs))))
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  - %v\n", e)
		}
		return fmt.Errorf("%d installations failed", len(errs))
	}

	return nil
}
