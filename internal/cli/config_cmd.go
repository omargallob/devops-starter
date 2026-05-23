package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/omargallob/devops-starter/internal/config"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  "Initialize or display the devops-starter configuration.",
	}

	cmd.AddCommand(
		newConfigInitCmd(),
		newConfigShowCmd(),
	)

	return cmd
}

func newConfigInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create default configuration file",
		RunE:  runConfigInit,
	}
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	path := cfgFile
	if path == "" {
		path = config.ConfigPath()
	}

	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config file already exists at %s", path)
	}

	cfg := config.DefaultConfig()
	if err := config.Save(cfg, path); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Configuration created at %s\n", path)
	return nil
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Display current configuration",
		RunE:  runConfigShow,
	}
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	path := cfgFile
	if path == "" {
		path = config.ConfigPath()
	}

	cfg, err := config.Load(path)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	fmt.Printf("# Config file: %s\n", path)
	fmt.Print(string(data))
	return nil
}
