// Package mcp provides the MCP (Model Context Protocol) server for devops-starter.
// It exposes read-only tool and resource handlers that allow AI chat clients
// to query the tool registry, installation status, configuration, and platform info.
package mcp

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/internal/platform"
	"github.com/omargallob/devops-starter/internal/registry"
	"github.com/omargallob/devops-starter/internal/state"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// Version is set at build time via ldflags.
var Version = "dev"

// Deps holds shared dependencies for all MCP tool and resource handlers.
type Deps struct {
	Config   *config.Config
	Registry *registry.Registry
	Store    *state.Store
	Platform *platform.Info
}

// NewDeps initializes the shared dependency bundle by loading config,
// detecting the platform, creating the registry, and loading state.
func NewDeps() (*Deps, error) {
	cfg, err := config.Load(config.Path())
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	info, err := platform.Detect()
	if err != nil {
		return nil, fmt.Errorf("detecting platform: %w", err)
	}

	reg := registry.New(cfg.PluginPaths...)

	store, err := state.LoadStore(state.Path())
	if err != nil {
		return nil, fmt.Errorf("loading state: %w", err)
	}

	return &Deps{
		Config:   cfg,
		Registry: reg,
		Store:    store,
		Platform: info,
	}, nil
}

// NewServer creates a configured MCP server with all tools and resources registered.
func NewServer(deps *Deps) *server.MCPServer {
	s := server.NewMCPServer(
		"devops-starter",
		Version,
		server.WithToolCapabilities(false),
		server.WithResourceCapabilities(false, false),
	)

	registerTools(s, deps)
	registerResources(s, deps)

	return s
}

// Run is the top-level entry point: loads deps, creates the server, and serves stdio.
func Run() error {
	deps, err := NewDeps()
	if err != nil {
		return err
	}

	s := NewServer(deps)
	return server.ServeStdio(s)
}

// groupEnabled checks if a group is enabled in the config.
func groupEnabled(cfg *config.Config, group tooldef.Group) bool {
	switch group {
	case tooldef.GroupLanguages:
		return cfg.Groups.Languages
	case tooldef.GroupContainers:
		return cfg.Groups.Containers
	case tooldef.GroupKubernetes:
		return cfg.Groups.Kubernetes
	case tooldef.GroupInfra:
		return cfg.Groups.Infra
	case tooldef.GroupCloud:
		return cfg.Groups.Cloud
	case tooldef.GroupAnsible:
		return cfg.Groups.Ansible
	case tooldef.GroupRustTools:
		return cfg.Groups.RustTools
	case tooldef.GroupUtilities:
		return cfg.Groups.Utilities
	case tooldef.GroupAI:
		return cfg.Groups.AI
	case tooldef.GroupPackageManagers:
		return cfg.Groups.PackageManagers
	default:
		return false
	}
}

// allGroups returns all group names in display order.
func allGroups() []tooldef.Group {
	return []tooldef.Group{
		tooldef.GroupLanguages,
		tooldef.GroupContainers,
		tooldef.GroupKubernetes,
		tooldef.GroupInfra,
		tooldef.GroupCloud,
		tooldef.GroupAnsible,
		tooldef.GroupRustTools,
		tooldef.GroupUtilities,
		tooldef.GroupAI,
		tooldef.GroupPackageManagers,
	}
}


