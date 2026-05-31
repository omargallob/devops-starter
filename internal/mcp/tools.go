package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"gopkg.in/yaml.v3"

	"github.com/omargallob/devops-starter/internal/dotfiles"
	"github.com/omargallob/devops-starter/internal/state"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// registerTools registers all read-only MCP tools on the server.
func registerTools(s *server.MCPServer, deps *Deps) {
	s.AddTool(listToolsDef(), listToolsHandler(deps))
	s.AddTool(getToolDef(), getToolHandler(deps))
	s.AddTool(getStatusDef(), getStatusHandler(deps))
	s.AddTool(configShowDef(), configShowHandler(deps))
	s.AddTool(detectPlatformDef(), detectPlatformHandler(deps))
	s.AddTool(dotfilesStatusDef(), dotfilesStatusHandler(deps))
}

// --- list_tools ---

func listToolsDef() mcp.Tool {
	return mcp.NewTool("list_tools",
		mcp.WithDescription("List all available DevOps tools in the registry, optionally filtered by group"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithString("group",
			mcp.Description("Filter by tool group"),
			mcp.Enum(
				"languages", "containers", "kubernetes", "infra",
				"cloud", "ansible", "rust-tools", "utilities",
				"ai", "package-managers",
			),
		),
	)
}

func listToolsHandler(deps *Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		group := request.GetString("group", "")

		var tools []*tooldef.Tool
		if group != "" {
			tools = deps.Registry.GetByGroup(tooldef.Group(group))
		} else {
			tools = deps.Registry.All()
		}

		type toolSummary struct {
			Name        string `json:"name"`
			Version     string `json:"version"`
			Group       string `json:"group"`
			Description string `json:"description"`
			InstallMode string `json:"install_mode"`
		}

		summaries := make([]toolSummary, 0, len(tools))
		for _, t := range tools {
			summaries = append(summaries, toolSummary{
				Name:        t.Name,
				Version:     t.Version,
				Group:       string(t.Group),
				Description: t.Description,
				InstallMode: string(t.EffectiveInstallMode()),
			})
		}

		data, err := json.MarshalIndent(summaries, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("marshal error: %v", err)), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}

// --- get_tool ---

func getToolDef() mcp.Tool {
	return mcp.NewTool("get_tool",
		mcp.WithDescription("Get full details of a specific tool by name (version, platforms, install mode, repo, etc.)"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("The tool name (e.g., kubectl, bat, terraform)"),
		),
	)
}

func getToolHandler(deps *Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("parameter 'name' is required"), nil
		}

		tool, ok := deps.Registry.Get(name)
		if !ok {
			return mcp.NewToolResultError(fmt.Sprintf("tool %q not found in registry", name)), nil
		}

		// Serialize the full tool definition.
		type toolDetail struct {
			Name            string            `json:"name"`
			Version         string            `json:"version"`
			Description     string            `json:"description"`
			Group           string            `json:"group"`
			Subgroup        string            `json:"subgroup,omitempty"`
			InstallMode     string            `json:"install_mode"`
			Repo            string            `json:"repo,omitempty"`
			Asset           string            `json:"asset,omitempty"`
			URLTemplate     string            `json:"url_template,omitempty"`
			BinaryName      string            `json:"binary_name,omitempty"`
			InstallName     string            `json:"install_name,omitempty"`
			Format          string            `json:"format,omitempty"`
			Platforms       []string          `json:"platforms,omitempty"`
			Dependencies    []string          `json:"dependencies,omitempty"`
			MiseBackend     string            `json:"mise_backend,omitempty"`
			Checksums       map[string]string `json:"checksums,omitempty"`
			StripComponents int               `json:"strip_components,omitempty"`
		}

		var platforms []string
		for _, p := range tool.Platforms {
			platforms = append(platforms, p.String())
		}

		detail := toolDetail{
			Name:            tool.Name,
			Version:         tool.Version,
			Description:     tool.Description,
			Group:           string(tool.Group),
			Subgroup:        tool.Subgroup,
			InstallMode:     string(tool.EffectiveInstallMode()),
			Repo:            tool.Repo,
			Asset:           tool.Asset,
			URLTemplate:     tool.URLTemplate,
			BinaryName:      tool.BinaryName,
			InstallName:     tool.InstallName,
			Format:          string(tool.Format),
			Platforms:       platforms,
			Dependencies:    tool.Dependencies,
			MiseBackend:     tool.MiseBackend,
			Checksums:       tool.Checksums,
			StripComponents: tool.StripComponents,
		}

		data, err := json.MarshalIndent(detail, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("marshal error: %v", err)), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}

// --- get_status ---

func getStatusDef() mcp.Tool {
	return mcp.NewTool("get_status",
		mcp.WithDescription("Get installation status of all tools (installed version, desired version, status: missing/current/outdated/disabled)"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithString("group",
			mcp.Description("Filter by tool group"),
			mcp.Enum(
				"languages", "containers", "kubernetes", "infra",
				"cloud", "ansible", "rust-tools", "utilities",
				"ai", "package-managers",
			),
		),
		mcp.WithString("status_filter",
			mcp.Description("Filter by status"),
			mcp.Enum("missing", "current", "outdated", "disabled", "detected", "linked", "unavailable"),
		),
	)
}

func getStatusHandler(deps *Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		groupFilter := request.GetString("group", "")
		statusFilter := request.GetString("status_filter", "")

		groupStates := state.ResolveAll(deps.Config, deps.Store, deps.Platform.Platform)

		type statusEntry struct {
			Name             string `json:"name"`
			Group            string `json:"group"`
			DesiredVersion   string `json:"desired_version"`
			InstalledVersion string `json:"installed_version,omitempty"`
			Status           string `json:"status"`
			Source           string `json:"source,omitempty"`
		}

		var entries []statusEntry
		for _, gs := range groupStates {
			if groupFilter != "" && gs.Name != groupFilter {
				continue
			}
			for _, ts := range gs.Tools {
				if statusFilter != "" && ts.Status.String() != statusFilter {
					continue
				}
				entries = append(entries, statusEntry{
					Name:             ts.Name,
					Group:            ts.Group,
					DesiredVersion:   ts.DesiredVersion,
					InstalledVersion: ts.InstalledVersion,
					Status:           ts.Status.String(),
					Source:           string(ts.Source),
				})
			}
		}

		if len(entries) == 0 {
			filters := []string{}
			if groupFilter != "" {
				filters = append(filters, fmt.Sprintf("group=%s", groupFilter))
			}
			if statusFilter != "" {
				filters = append(filters, fmt.Sprintf("status=%s", statusFilter))
			}
			msg := "No tools found"
			if len(filters) > 0 {
				msg += " matching filters: " + strings.Join(filters, ", ")
			}
			return mcp.NewToolResultText(msg), nil
		}

		data, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("marshal error: %v", err)), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}

// --- config_show ---

func configShowDef() mcp.Tool {
	return mcp.NewTool("config_show",
		mcp.WithDescription("Show the current devops-starter configuration as YAML (install directory, enabled groups, overrides, packages)"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func configShowHandler(deps *Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := yaml.Marshal(deps.Config)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("marshal error: %v", err)), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}

// --- detect_platform ---

func detectPlatformDef() mcp.Tool {
	return mcp.NewTool("detect_platform",
		mcp.WithDescription("Detect the current operating system, architecture, and Linux distribution"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func detectPlatformHandler(deps *Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result := map[string]string{
			"os":   deps.Platform.OS,
			"arch": deps.Platform.Arch,
		}
		if deps.Platform.Distro != "" {
			result["distro"] = string(deps.Platform.Distro)
		}

		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("marshal error: %v", err)), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}

// --- dotfiles_status ---

func dotfilesStatusDef() mcp.Tool {
	return mcp.NewTool("dotfiles_status",
		mcp.WithDescription("Show the symlink status of all managed dotfiles (linked, missing, conflict, broken)"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func dotfilesStatusHandler(deps *Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		homeDir, err := homeDirectory()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("cannot determine home directory: %v", err)), nil
		}

		// Use the dotfiles directory relative to the working directory.
		sourceDir := dotfilesSourceDir()
		linker := dotfiles.NewLinker(sourceDir, homeDir, true)
		mappings := dotfiles.DefaultMappings()
		results := linker.Status(mappings)

		type dotfileEntry struct {
			Source string `json:"source"`
			Dest   string `json:"dest"`
			Status string `json:"status"`
		}

		entries := make([]dotfileEntry, 0, len(results))
		for _, r := range results {
			entries = append(entries, dotfileEntry{
				Source: r.Mapping.Source,
				Dest:   r.Mapping.Dest,
				Status: r.Status.String(),
			})
		}

		data, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("marshal error: %v", err)), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}
