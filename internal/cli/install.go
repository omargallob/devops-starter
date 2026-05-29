package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

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
	reg := registry.New(cfg.PluginPaths...)
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
// It filters the registry, detects conflicts with system binaries, prompts the
// user for conflict resolution, confirms, and runs installations.
func doInstall(deps installDeps, plat tooldef.Platform) error {
	tools := filterToolsForInstall(deps.registry.All(), deps.cfg, plat, deps.only)

	if len(tools) == 0 {
		fmt.Fprintln(deps.out, "No tools to install.")
		return nil
	}

	// Detect conflicts with system-installed binaries.
	conflicts := installer.DetectConflicts(tools, deps.cfg.InstallDir, deps.cfg.Overrides)

	// Resolve conflicts.
	resolutions := make(map[string]string)
	if len(conflicts) > 0 {
		resolved, err := resolveConflicts(deps, conflicts)
		if err != nil {
			return err
		}
		resolutions = resolved
	}

	// Partition tools based on conflict resolutions.
	toInstall, skipped, links := installer.ApplyConflictActions(tools, conflicts, resolutions)

	// Verify existing symlinks for link-mode tools.
	brokenLinks := verifyExistingLinks(deps)

	// Show what will be installed, skipped, and linked.
	printInstallPlan(deps.out, toInstall, skipped, links, brokenLinks, deps.cfg.InstallDir)

	if len(toInstall) == 0 && len(links) == 0 && len(brokenLinks) == 0 {
		fmt.Fprintln(deps.out, "Nothing to do.")
		return nil
	}

	if !deps.dryRun && !deps.autoYes {
		if !confirmAction("Proceed with installation?") {
			fmt.Fprintln(deps.out, "Cancelled.")
			return nil
		}
	}

	// Create symlinks for link-mode tools.
	createLinks(deps, tools, links)

	// Re-create broken symlinks by re-resolving from PATH.
	relinkBroken(deps, tools, brokenLinks)

	// Install tools.
	ctx := context.Background()
	var errs []error
	if len(toInstall) > 0 {
		errs = deps.installer.InstallAll(ctx, toInstall)
	}

	printInstallSummary(deps.out, len(toInstall), len(links), errs)

	// Offer to save conflict preferences.
	if len(conflicts) > 0 && !deps.autoYes && !deps.dryRun {
		saveConflictPreferences(deps, resolutions)
	}

	if len(errs) > 0 {
		return fmt.Errorf("%d installations failed", len(errs))
	}

	return nil
}

// resolveConflicts determines the action for each conflicting tool.
// For non-interactive mode, defaults to "overwrite".
func resolveConflicts(deps installDeps, conflicts []installer.ConflictInfo) (map[string]string, error) {
	resolutions := make(map[string]string)

	// Apply saved preferences first.
	unresolved := installer.UnresolvedConflicts(conflicts)
	for _, c := range conflicts {
		if c.SavedAction != "" {
			resolutions[c.Tool.Name] = c.SavedAction
		}
	}

	// If no unresolved conflicts, we're done.
	if len(unresolved) == 0 {
		return resolutions, nil
	}

	// Non-interactive: default to overwrite.
	if deps.autoYes {
		applyActionToAll(resolutions, unresolved, "overwrite")
		return resolutions, nil
	}

	// Interactive batch prompt.
	printConflictTable(deps.out, unresolved)

	fmt.Fprintln(deps.out)
	fmt.Fprintln(deps.out, "How would you like to handle these conflicts?")
	fmt.Fprintln(deps.out, "  [S] Skip all (keep system versions)")
	fmt.Fprintln(deps.out, "  [O] Overwrite all (install managed versions)")
	fmt.Fprintln(deps.out, "  [C] Customize per-tool")
	fmt.Fprintf(deps.out, "Choice [S/O/C]: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		applyActionToAll(resolutions, unresolved, "overwrite")
		return resolutions, nil
	}

	choice := strings.TrimSpace(strings.ToLower(input))

	switch choice {
	case "s", "skip":
		applyActionToAll(resolutions, unresolved, "skip")
	case "c", "customize":
		for _, c := range unresolved {
			action := promptPerToolConflict(deps.out, c)
			resolutions[c.Tool.Name] = action
		}
	default:
		applyActionToAll(resolutions, unresolved, "overwrite")
	}

	return resolutions, nil
}

// applyActionToAll sets the same conflict action for all unresolved conflicts.
func applyActionToAll(resolutions map[string]string, conflicts []installer.ConflictInfo, action string) {
	for _, c := range conflicts {
		resolutions[c.Tool.Name] = action
	}
}

// promptPerToolConflict asks the user what to do for a single tool conflict.
func promptPerToolConflict(out io.Writer, c installer.ConflictInfo) string {
	versionInfo := ""
	if c.SystemVersion != "" {
		versionInfo = fmt.Sprintf(" (v%s)", c.SystemVersion)
	}
	fmt.Fprintf(out, "  %s%s at %s → wanted %s\n", c.Tool.Name, versionInfo, c.SystemPath, c.WantedVersion)
	fmt.Fprintf(out, "    [S]kip  [O]verwrite  [L]ink: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "s", "skip":
		return "skip"
	case "l", "link":
		return "link"
	default:
		return "overwrite"
	}
}

// printConflictTable displays detected conflicts in a table format.
func printConflictTable(out io.Writer, conflicts []installer.ConflictInfo) {
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, warnStyle.Render(fmt.Sprintf("⚠ Found %d tool(s) already installed on your system:", len(conflicts))))
	fmt.Fprintln(out)
	fmt.Fprintf(out, "  %-16s %-14s %-30s %s\n", "Tool", "System Ver", "Location", "Wanted")
	fmt.Fprintf(out, "  %-16s %-14s %-30s %s\n", "────", "──────────", "────────", "──────")
	for _, c := range conflicts {
		sysVer := c.SystemVersion
		if sysVer == "" {
			sysVer = "unknown"
		}
		fmt.Fprintf(out, "  %-16s %-14s %-30s %s\n", c.Tool.Name, sysVer, c.SystemPath, c.WantedVersion)
	}
}

// printInstallPlan shows what will be installed, skipped, and linked.
func printInstallPlan(out io.Writer, toInstall, skipped []*tooldef.Tool, links map[string]string, brokenLinks []installer.LinkResult, installDir string) {
	if len(toInstall) > 0 {
		fmt.Fprintf(out, "\nThe following %d tool(s) will be installed to %s:\n\n", len(toInstall), installDir)
		for _, t := range toInstall {
			if t.IsMiseManaged() {
				fmt.Fprintf(out, "  • %s %s (via mise)\n", t.Name, t.Version)
			} else {
				fmt.Fprintf(out, "  • %s %s\n", t.Name, t.Version)
			}
		}
	}

	if len(links) > 0 {
		fmt.Fprintf(out, "\nThe following %d tool(s) will be symlinked:\n\n", len(links))
		for name, path := range links {
			fmt.Fprintf(out, "  • %s -> %s\n", name, path)
		}
	}

	if len(skipped) > 0 {
		fmt.Fprintf(out, "\nThe following %d tool(s) will be skipped (using system version):\n\n", len(skipped))
		for _, t := range skipped {
			fmt.Fprintf(out, "  • %s\n", t.Name)
		}
	}

	if len(brokenLinks) > 0 {
		warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
		fmt.Fprintln(out)
		fmt.Fprintln(out, warnStyle.Render(fmt.Sprintf("⚠ %d broken symlink(s) detected (will attempt to re-link):", len(brokenLinks))))
		for _, bl := range brokenLinks {
			fmt.Fprintf(out, "  • %s -> %s (target missing)\n", bl.ToolName, bl.TargetPath)
		}
	}

	fmt.Fprintln(out)
}

// verifyExistingLinks checks all tools with conflict=link for broken symlinks.
func verifyExistingLinks(deps installDeps) []installer.LinkResult {
	overrides := make(map[string]installer.ToolConflictGetter)
	for name, ov := range deps.cfg.Overrides {
		overrides[name] = conflictGetter(ov.Conflict)
	}

	allTools := deps.registry.All()
	results := installer.VerifyLinks(deps.cfg.InstallDir, allTools, overrides)

	// Return only broken links.
	var broken []installer.LinkResult
	for _, r := range results {
		if r.Broken {
			broken = append(broken, r)
		}
	}
	return broken
}

// conflictGetter implements installer.ToolConflictGetter.
type conflictGetter string

func (c conflictGetter) GetConflict() string { return string(c) }

// createLinks creates symlinks for tools resolved with the "link" conflict action.
func createLinks(deps installDeps, tools []*tooldef.Tool, links map[string]string) {
	for toolName, systemPath := range links {
		for _, t := range tools {
			if t.Name == toolName {
				if err := deps.installer.Link(t, systemPath); err != nil {
					fmt.Fprintf(deps.out, "  warning: failed to link %s: %v\n", toolName, err)
				}
				break
			}
		}
	}
}

// relinkBroken re-creates broken symlinks by looking up tools on PATH again.
func relinkBroken(deps installDeps, tools []*tooldef.Tool, brokenLinks []installer.LinkResult) {
	for _, bl := range brokenLinks {
		for _, t := range tools {
			if t.Name == bl.ToolName {
				newPath := lookupSystemBinary(t)
				if newPath != "" {
					if err := deps.installer.Link(t, newPath); err != nil {
						fmt.Fprintf(deps.out, "  warning: failed to re-link %s: %v\n", bl.ToolName, err)
					} else {
						fmt.Fprintf(deps.out, "  re-linked %s -> %s\n", bl.ToolName, newPath)
					}
				} else {
					fmt.Fprintf(deps.out, "  warning: %s has a broken link and is no longer on PATH\n", bl.ToolName)
				}
				break
			}
		}
	}
}

// lookupSystemBinary finds a tool binary on PATH (used for re-linking).
func lookupSystemBinary(tool *tooldef.Tool) string {
	return installer.LookupToolInPath(tool)
}

// saveConflictPreferences offers to persist conflict resolutions to config.
func saveConflictPreferences(deps installDeps, resolutions map[string]string) {
	fmt.Fprintf(deps.out, "Save conflict preferences to config? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	input = strings.TrimSpace(strings.ToLower(input))
	if input != "y" && input != "yes" {
		return
	}

	if deps.cfg.Overrides == nil {
		deps.cfg.Overrides = make(map[string]config.ToolOverride)
	}

	for toolName, action := range resolutions {
		override := deps.cfg.Overrides[toolName]
		override.Conflict = action
		deps.cfg.Overrides[toolName] = override
	}

	cfgPath := cfgFile
	if cfgPath == "" {
		cfgPath = config.Path()
	}

	if err := config.Save(deps.cfg, cfgPath); err != nil {
		fmt.Fprintf(deps.out, "  warning: could not save config: %v\n", err)
	} else {
		fmt.Fprintf(deps.out, "  Preferences saved to %s\n", cfgPath)
	}
}

// printInstallSummary outputs the success/failure summary for install operations.
func printInstallSummary(out io.Writer, total, linked int, errs []error) {
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	installed := total - len(errs)
	fmt.Fprintln(out)
	fmt.Fprintln(out, successStyle.Render(fmt.Sprintf("✓ %d tools installed successfully", installed)))

	if linked > 0 {
		fmt.Fprintln(out, successStyle.Render(fmt.Sprintf("✓ %d tools symlinked", linked)))
	}

	if len(errs) > 0 {
		fmt.Fprintln(out, errorStyle.Render(fmt.Sprintf("✗ %d tools failed:", len(errs))))
		for _, e := range errs {
			fmt.Fprintf(out, "  - %v\n", e)
		}
	}
}
