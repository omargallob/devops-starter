package cli

import (
	"context"
	"fmt"
	"io"
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

// runInstall is the entry point that constructs real dependencies and delegates
// to doInstall for the actual logic.
func runInstall(cmd *cobra.Command, args []string) error {
	// Detect platform
	info, err := platform.Detect()
	if err != nil {
		return fmt.Errorf("detecting platform: %w", err)
	}

	// Load config
	cfgPath := cfgFile
	if cfgPath == "" {
		cfgPath = config.Path()
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Create real dependencies
	reg := registry.New()
	inst := installer.New(
		cfg.InstallDir,
		info.Platform,
		installer.WithDryRun(dryRun),
	)

	deps := installDeps{
		cfg:       cfg,
		registry:  reg,
		installer: inst,
		out:       os.Stdout,
		dryRun:    dryRun,
		autoYes:   autoYes,
		only:      only,
	}

	return doInstall(deps, info.Platform)
}

// filterToolsForInstall selects tools eligible for installation based on group,
// platform, and config overrides.
func filterToolsForInstall(allTools []*tooldef.Tool, cfg *config.Config, plat tooldef.Platform, onlyGroup string) []*tooldef.Tool {
	var tools []*tooldef.Tool
	for _, t := range allTools {
		if onlyGroup != "" && string(t.Group) != onlyGroup {
			continue
		}
		if !cfg.IsGroupEnabled(string(t.Group)) {
			continue
		}
		if !t.SupportsPlatform(plat) {
			continue
		}
		if override, ok := cfg.Overrides[t.Name]; ok {
			if override.Disabled {
				continue
			}
			if override.Version != "" {
				t.Version = override.Version
			}
		}
		tools = append(tools, t)
	}
	return tools
}

// doInstall contains the testable install logic, separated from Cobra wiring.
// It filters the registry, confirms with the user, and runs installations.
func doInstall(deps installDeps, plat tooldef.Platform) error {
	tools := filterToolsForInstall(deps.registry.All(), deps.cfg, plat, deps.only)

	if len(tools) == 0 {
		fmt.Fprintln(deps.out, "No tools to install.")
		return nil
	}

	// Show what will be installed and confirm
	fmt.Fprintf(deps.out, "\nThe following %d tool(s) will be installed to %s:\n\n", len(tools), deps.cfg.InstallDir)
	for _, t := range tools {
		if t.ManagedBy != "" {
			fmt.Fprintf(deps.out, "  • %s %s (via %s)\n", t.Name, t.Version, t.ManagedBy)
		} else {
			fmt.Fprintf(deps.out, "  • %s %s\n", t.Name, t.Version)
		}
	}
	fmt.Fprintln(deps.out)

	if !deps.dryRun && !deps.autoYes {
		if !confirmAction("Proceed with installation?") {
			fmt.Fprintln(deps.out, "Cancelled.")
			return nil
		}
	}

	ctx := context.Background()
	errs := deps.installer.InstallAll(ctx, tools)

	printInstallSummary(deps.out, len(tools), errs)

	if len(errs) > 0 {
		return fmt.Errorf("%d installations failed", len(errs))
	}

	return nil
}

// printInstallSummary outputs the success/failure summary for install operations.
func printInstallSummary(out io.Writer, total int, errs []error) {
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	installed := total - len(errs)
	fmt.Fprintln(out)
	fmt.Fprintln(out, successStyle.Render(fmt.Sprintf("✓ %d tools installed successfully", installed)))

	if len(errs) > 0 {
		fmt.Fprintln(out, errorStyle.Render(fmt.Sprintf("✗ %d tools failed:", len(errs))))
		for _, e := range errs {
			fmt.Fprintf(out, "  - %v\n", e)
		}
	}
}
