package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/internal/registry"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available tools",
		Long:  "Show all available tools grouped by category with installation status.",
		RunE:  runList,
	}
}

func runList(cmd *cobra.Command, args []string) error {
	cfgPath := cfgFile
	if cfgPath == "" {
		cfgPath = config.ConfigPath()
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	reg := registry.New()

	groups := []tooldef.Group{
		tooldef.GroupLanguages,
		tooldef.GroupContainers,
		tooldef.GroupKubernetes,
		tooldef.GroupInfra,
		tooldef.GroupCloud,
		tooldef.GroupRustTools,
		tooldef.GroupUtilities,
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4"))
	installedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	dimStyle := lipgloss.NewStyle().Faint(true)

	for _, group := range groups {
		tools := reg.GetByGroup(group)
		if len(tools) == 0 {
			continue
		}

		fmt.Println(headerStyle.Render(fmt.Sprintf("\n[%s]", string(group))))

		for _, t := range tools {
			binPath := filepath.Join(cfg.InstallDir, t.GetInstallName())
			_, statErr := os.Stat(binPath)
			installed := statErr == nil

			status := "  "
			style := dimStyle
			if installed {
				status = "✓ "
				style = installedStyle
			}

			line := fmt.Sprintf("  %s%-20s %-10s %s", status, t.Name, t.Version, t.Description)
			fmt.Println(style.Render(line))
		}
	}

	fmt.Println()
	return nil
}
