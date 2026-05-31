package cli

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/internal/mise"
	"github.com/omargallob/devops-starter/internal/pkgmgr"
)

// newPackagesCmd creates the "packages" subcommand with install and list sub-subcommands.
func newPackagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "packages",
		Short: "Manage globally-installed Python and Node packages",
		Long: `Install and list Python (pip) and Node (npm) packages at the global/user level.

Packages are declared in .mise.toml under [packages.pip] and [packages.npm]:

  [packages.pip]
  "black" = "24.4"
  "ruff"  = "0.4"

  [packages.npm]
  "typescript" = "5.4"
  "prettier"   = "3.2"

Enable package installation in config.yaml:

  packages:
    python:
      enabled: true
    node:
      enabled: true`,
	}

	cmd.AddCommand(newPackagesInstallCmd(), newPackagesListCmd())
	return cmd
}

// newPackagesInstallCmd creates "packages install".
func newPackagesInstallCmd() *cobra.Command {
	var onlyPython, onlyNode bool

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install configured pip and npm packages globally",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPackagesInstall(onlyPython, onlyNode)
		},
	}

	cmd.Flags().BoolVar(&onlyPython, "python", false, "install only Python (pip) packages")
	cmd.Flags().BoolVar(&onlyNode, "node", false, "install only Node (npm) packages")
	return cmd
}

// newPackagesListCmd creates "packages list".
func newPackagesListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List packages configured in .mise.toml",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPackagesList()
		},
	}
}

// runPackagesInstall installs pip and/or npm packages declared in .mise.toml.
func runPackagesInstall(onlyPython, onlyNode bool) error {
	cfg, pkgVersions, err := loadPackageContext()
	if err != nil {
		return err
	}

	if len(pkgVersions) == 0 {
		fmt.Println("No [packages.*] sections found in .mise.toml.")
		fmt.Println("Add [packages.pip] or [packages.npm] sections to declare packages.")
		return nil
	}

	doPython := !onlyNode && (onlyPython || cfg.Packages.Python.Enabled)
	doNode := !onlyPython && (onlyNode || cfg.Packages.Node.Enabled)

	if !doPython && !doNode {
		fmt.Println("Package installation is disabled in config.")
		fmt.Println("Set packages.python.enabled or packages.node.enabled to true,")
		fmt.Println("or use --python / --node to install regardless of config.")
		return nil
	}

	errs := installSelectedPackages(cfg, pkgVersions, doPython, doNode)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "error: %v\n", e)
		}
		return fmt.Errorf("%d package installation(s) failed", len(errs))
	}
	return nil
}

// installSelectedPackages installs the pip and/or npm packages based on flags.
func installSelectedPackages(cfg *config.Config, pkgVersions map[string]map[string]string, doPython, doNode bool) []error {
	ctx := context.Background()
	opts := []pkgmgr.Option{pkgmgr.WithDryRun(dryRun)}
	var errs []error

	if doPython {
		if pipPkgs, ok := pkgVersions["pip"]; ok && len(pipPkgs) > 0 {
			errs = append(errs, installPipPackages(ctx, cfg, pipPkgs, opts)...)
		} else {
			fmt.Println("No [packages.pip] entries found in .mise.toml.")
		}
	}

	if doNode {
		if npmPkgs, ok := pkgVersions["npm"]; ok && len(npmPkgs) > 0 {
			errs = append(errs, installNpmPackages(ctx, cfg, npmPkgs, opts)...)
		} else {
			fmt.Println("No [packages.npm] entries found in .mise.toml.")
		}
	}

	return errs
}

// installPipPackages installs all pip packages via the configured manager.
func installPipPackages(ctx context.Context, cfg *config.Config, pkgs map[string]string, opts []pkgmgr.Option) []error {
	manager := cfg.Packages.Python.Manager
	if manager == "" {
		manager = "pip"
	}

	mgr, err := pkgmgr.New(manager, opts...)
	if err != nil {
		return []error{fmt.Errorf("pip manager unavailable: %w", err)}
	}

	names := sortedKeys(pkgs)
	fmt.Printf("\nInstalling %d Python package(s) via %s:\n", len(names), manager)

	var errs []error
	for _, pkg := range names {
		version := pkgs[pkg]
		fmt.Printf("  • %s", pkg)
		if version != "" {
			fmt.Printf("==%s", version)
		}
		fmt.Println()
		if err := mgr.Install(ctx, pkg, version); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// installNpmPackages installs all npm packages via the configured manager.
func installNpmPackages(ctx context.Context, cfg *config.Config, pkgs map[string]string, opts []pkgmgr.Option) []error {
	manager := cfg.Packages.Node.Manager
	if manager == "" {
		manager = "npm"
	}

	mgr, err := pkgmgr.New(manager, opts...)
	if err != nil {
		return []error{fmt.Errorf("npm manager unavailable: %w", err)}
	}

	names := sortedKeys(pkgs)
	fmt.Printf("\nInstalling %d Node package(s) via %s:\n", len(names), manager)

	var errs []error
	for _, pkg := range names {
		version := pkgs[pkg]
		fmt.Printf("  • %s", pkg)
		if version != "" {
			fmt.Printf("@%s", version)
		}
		fmt.Println()
		if err := mgr.Install(ctx, pkg, version); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// runPackagesList prints all packages declared in .mise.toml.
func runPackagesList() error {
	_, pkgVersions, err := loadPackageContext()
	if err != nil {
		return err
	}

	if len(pkgVersions) == 0 {
		fmt.Println("No [packages.*] sections found in .mise.toml.")
		return nil
	}

	for _, mgr := range []string{"pip", "npm"} {
		pkgs, ok := pkgVersions[mgr]
		if !ok || len(pkgs) == 0 {
			continue
		}
		fmt.Printf("[packages.%s]\n", mgr)
		for _, name := range sortedKeys(pkgs) {
			ver := pkgs[name]
			if ver == "" {
				ver = "latest"
			}
			fmt.Printf("  %-30s %s\n", name, ver)
		}
		fmt.Println()
	}
	return nil
}

// loadPackageContext loads config and parses .mise.toml packages in one step.
func loadPackageContext() (*config.Config, mise.PackageVersions, error) {
	cfgPath := cfgFile
	if cfgPath == "" {
		cfgPath = config.Path()
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, nil, fmt.Errorf("loading config: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, nil, fmt.Errorf("getting working directory: %w", err)
	}

	pkgVersions, err := mise.FindAndParsePackages(cwd)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing .mise.toml packages: %w", err)
	}

	return cfg, pkgVersions, nil
}

// runPackagesFromInstall is called by doInstall when packages are enabled.
// It reuses loadPackageContext and the install helpers.
func runPackagesFromInstall(ctx context.Context, deps installDeps) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	pkgVersions, err := mise.FindAndParsePackages(cwd)
	if err != nil {
		return fmt.Errorf("parsing .mise.toml packages: %w", err)
	}
	if len(pkgVersions) == 0 {
		return nil
	}

	opts := []pkgmgr.Option{pkgmgr.WithDryRun(deps.dryRun)}
	var errs []error

	if deps.cfg.Packages.Python.Enabled {
		if pipPkgs, ok := pkgVersions["pip"]; ok && len(pipPkgs) > 0 {
			errs = append(errs, installPipPackages(ctx, deps.cfg, pipPkgs, opts)...)
		}
	}
	if deps.cfg.Packages.Node.Enabled {
		if npmPkgs, ok := pkgVersions["npm"]; ok && len(npmPkgs) > 0 {
			errs = append(errs, installNpmPackages(ctx, deps.cfg, npmPkgs, opts)...)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%d package installation(s) failed", len(errs))
	}
	return nil
}

// sortedKeys returns map keys in alphabetical order.
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
