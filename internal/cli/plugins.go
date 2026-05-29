package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/internal/registry"
)

func newPluginsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugins",
		Short: "Manage tool plugins",
		Long:  "List loaded plugin tools or validate a plugin YAML file.",
	}
	cmd.AddCommand(
		newPluginsListCmd(),
		newPluginsValidateCmd(),
	)
	return cmd
}

func newPluginsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List loaded plugin tools",
		Long:  "Show all tool plugins currently loaded, grouped by source file.",
		RunE:  runPluginsList,
	}
}

func newPluginsValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <file>",
		Short: "Validate a plugin YAML file",
		Long:  "Check a plugin YAML file for errors without installing anything.",
		Args:  cobra.ExactArgs(1),
		RunE:  runPluginsValidate,
	}
}

func runPluginsList(cmd *cobra.Command, args []string) error {
	cfgPath := cfgFile
	if cfgPath == "" {
		cfgPath = config.Path()
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	reg := registry.New(cfg.PluginPaths...)
	entries := reg.PluginEntries()

	if len(entries) == 0 {
		fmt.Println("No plugin tools loaded.")
		fmt.Println()
		fmt.Println("Add plugin files to one of:")
		fmt.Printf("  Project-local:  .devops-starter/plugins/*.yaml\n")
		fmt.Printf("  User-global:    %s\n", filepath.Join(pluginsUserConfigDir(), "devops-starter", "plugins"))
		return nil
	}

	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4"))
	dim := lipgloss.NewStyle().Faint(true)

	// Group entries by source file, preserving load order.
	type fileGroup struct {
		path    string
		entries []registry.PluginEntry
	}
	var groups []fileGroup
	indexByFile := make(map[string]int)

	for _, e := range entries {
		if idx, ok := indexByFile[e.FilePath]; ok {
			groups[idx].entries = append(groups[idx].entries, e)
		} else {
			indexByFile[e.FilePath] = len(groups)
			groups = append(groups, fileGroup{path: e.FilePath, entries: []registry.PluginEntry{e}})
		}
	}

	for _, g := range groups {
		fmt.Println(header.Render(fmt.Sprintf("\n[%s]", g.path)))
		for _, e := range g.entries {
			t := e.Tool
			line := fmt.Sprintf("  %-22s %-10s %s", t.Name, t.Version, t.Description)
			line += dim.Render(fmt.Sprintf("  (%s)", t.Group))
			fmt.Println(line)
		}
	}
	fmt.Println()
	return nil
}

func runPluginsValidate(cmd *cobra.Command, args []string) error {
	path := args[0]
	entries, errs := registry.LoadPluginFile(path)

	ok := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("✓")
	fail := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("✗")

	fmt.Printf("Validating %s\n", path)

	if len(errs) == 0 && len(entries) == 0 {
		fmt.Println("  (no tools defined)")
		return nil
	}

	for _, e := range entries {
		fmt.Printf("  %s %s\n", ok, e.Tool.Name)
	}
	for _, e := range errs {
		fmt.Printf("  %s %v\n", fail, e)
	}

	if len(errs) > 0 {
		return fmt.Errorf("%d validation error(s)", len(errs))
	}
	return nil
}

func pluginsUserConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config")
}
