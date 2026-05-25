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
	"github.com/omargallob/devops-starter/internal/state"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

var adoptAll bool

// newAdoptCmd creates the "adopt" subcommand which installs managed versions
// of tools that are already detected on the system (e.g., via Homebrew or
// system packages) so that devops-starter manages them instead.
func newAdoptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "adopt [tool...]",
		Short: "Adopt system-installed tools into managed versions",
		Long: `Install managed versions of tools that are already present on the system
but not managed by devops-starter. This downloads the desired version to the
managed install directory so it takes precedence over the system binary.

Use --all-detected to adopt all tools found on the system at once.`,
		Args: cobra.ArbitraryArgs,
		RunE: runAdopt,
	}

	cmd.Flags().BoolVar(&adoptAll, "all-detected", false, "adopt all detected (unmanaged) tools")

	return cmd
}

// adoptDeps bundles dependencies for the adopt command.
type adoptDeps struct {
	cfg       *config.Config
	registry  ToolRegistry
	installer ToolInstaller
	store     StateStore
	out       io.Writer
	dryRun    bool
	autoYes   bool
	adoptAll  bool
	platform  tooldef.Platform
	// detected is a precomputed map of tools detected on the system
	detected map[string]state.ToolState
}

// runAdopt is the entry point that constructs real dependencies and delegates
// to doAdopt for the actual logic.
func runAdopt(cmd *cobra.Command, args []string) error {
	if len(args) == 0 && !adoptAll {
		return fmt.Errorf("specify one or more tool names, or use --all-detected")
	}

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

	// Load state store
	store, err := state.LoadStore(state.StatePath())
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	// Resolve current state to find detected tools
	groups := state.ResolveAll(cfg, store, info.Platform)

	// Build a map of detected tools
	detected := make(map[string]state.ToolState)
	for _, g := range groups {
		for _, ts := range g.Tools {
			if ts.Status == state.StatusDetected {
				detected[ts.Name] = ts
			}
		}
	}

	reg := registry.New()
	inst := installer.New(
		cfg.InstallDir,
		info.Platform,
		installer.WithStateStore(store),
	)

	deps := adoptDeps{
		cfg:       cfg,
		registry:  reg,
		installer: inst,
		store:     store,
		out:       os.Stdout,
		dryRun:    dryRun,
		autoYes:   autoYes,
		adoptAll:  adoptAll,
		platform:  info.Platform,
		detected:  detected,
	}

	return doAdopt(deps, args)
}

// doAdopt contains the testable adopt logic, separated from Cobra wiring.
func doAdopt(deps adoptDeps, args []string) error {
	// Determine which tools to adopt
	var toAdopt []string
	if deps.adoptAll {
		for name := range deps.detected {
			toAdopt = append(toAdopt, name)
		}
	} else {
		toAdopt = append(toAdopt, args...)
	}

	if len(toAdopt) == 0 {
		fmt.Fprintln(deps.out, "No detected tools to adopt.")
		return nil
	}

	// Validate and resolve tools from registry
	var toolsToInstall []*tooldef.Tool
	var warnings []string

	for _, name := range toAdopt {
		tool, ok := deps.registry.Get(name)
		if !ok {
			warnings = append(warnings, fmt.Sprintf("unknown tool: %s", name))
			continue
		}

		// Check platform support
		if !tool.SupportsPlatform(deps.platform) {
			warnings = append(warnings, fmt.Sprintf("%s: not supported on %s", name, deps.platform))
			continue
		}

		// Apply version override from config
		if override, ok := deps.cfg.Overrides[name]; ok {
			if override.Disabled {
				warnings = append(warnings, fmt.Sprintf("%s: disabled in config", name))
				continue
			}
			if override.Version != "" {
				tool.Version = override.Version
			}
		}

		// Check if already managed (current or outdated means we already manage it)
		if _, isDetected := deps.detected[name]; !isDetected {
			// Check if it's already managed
			if ver := deps.store.GetVersion(name); ver != "" {
				warnings = append(warnings, fmt.Sprintf("%s: already managed (version %s)", name, ver))
				continue
			}
		}

		toolsToInstall = append(toolsToInstall, tool)
	}

	// Print warnings
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	for _, w := range warnings {
		fmt.Fprintln(deps.out, warnStyle.Render("  ⚠ "+w))
	}

	if len(toolsToInstall) == 0 {
		fmt.Fprintln(deps.out, "No tools to adopt.")
		return nil
	}

	// Show what we're about to do
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	fmt.Fprintln(deps.out, infoStyle.Render(fmt.Sprintf("\nAdopting %d tool(s) into %s:\n", len(toolsToInstall), deps.cfg.InstallDir)))

	for _, t := range toolsToInstall {
		detail := ""
		if ts, ok := deps.detected[t.Name]; ok && ts.DetectedPath != "" {
			detail = fmt.Sprintf(" (replacing system: %s", ts.DetectedPath)
			if ts.DetectedVersion != "" {
				detail += " v" + ts.DetectedVersion
			}
			detail += ")"
		}
		fmt.Fprintf(deps.out, "  • %s %s%s\n", t.Name, t.Version, detail)
	}
	fmt.Fprintln(deps.out)

	if deps.dryRun {
		fmt.Fprintln(deps.out, "[dry-run] No changes made.")
		return nil
	}

	if !deps.autoYes {
		if !confirmAction("Proceed with adoption?") {
			fmt.Fprintln(deps.out, "Cancelled.")
			return nil
		}
	}

	// Install tools
	ctx := context.Background()
	errs := deps.installer.InstallAll(ctx, toolsToInstall)

	// Save state
	if err := deps.store.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to save state: %v\n", err)
	}

	// Print summary
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	adopted := len(toolsToInstall) - len(errs)
	if adopted > 0 {
		fmt.Fprintln(deps.out, successStyle.Render(fmt.Sprintf("✓ %d tool(s) adopted successfully", adopted)))
	}

	if len(errs) > 0 {
		fmt.Fprintln(deps.out, errorStyle.Render(fmt.Sprintf("✗ %d tool(s) failed:", len(errs))))
		for _, e := range errs {
			fmt.Fprintf(deps.out, "  - %v\n", e)
		}
		return fmt.Errorf("%d adoptions failed", len(errs))
	}

	return nil
}
