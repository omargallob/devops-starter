package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"gopkg.in/yaml.v3"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/internal/state"
)

// registerResources registers all MCP resources on the server.
func registerResources(s *server.MCPServer, deps *Deps) {
	// Config resource
	s.AddResource(
		mcp.NewResource(
			"devops-starter://config",
			"Configuration",
			mcp.WithResourceDescription("Current devops-starter configuration (YAML)"),
			mcp.WithMIMEType("text/yaml"),
		),
		configResourceHandler(deps),
	)

	// State resource
	s.AddResource(
		mcp.NewResource(
			"devops-starter://state",
			"Installed State",
			mcp.WithResourceDescription("Current installed tools state (JSON)"),
			mcp.WithMIMEType("application/json"),
		),
		stateResourceHandler(),
	)

	// Groups resource
	s.AddResource(
		mcp.NewResource(
			"devops-starter://groups",
			"Tool Groups",
			mcp.WithResourceDescription("Available tool groups with enabled/disabled status"),
			mcp.WithMIMEType("application/json"),
		),
		groupsResourceHandler(deps),
	)
}

func configResourceHandler(deps *Deps) server.ResourceHandlerFunc {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		// Try to read the raw config file from disk first.
		cfgPath := config.Path()
		data, err := os.ReadFile(cfgPath)
		if err != nil {
			// Fall back to marshaling current in-memory config.
			data, err = yaml.Marshal(deps.Config)
			if err != nil {
				return nil, fmt.Errorf("marshaling config: %w", err)
			}
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "devops-starter://config",
				MIMEType: "text/yaml",
				Text:     string(data),
			},
		}, nil
	}
}

func stateResourceHandler() server.ResourceHandlerFunc {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		statePath := state.Path()
		data, err := os.ReadFile(statePath)
		if err != nil {
			if os.IsNotExist(err) {
				return []mcp.ResourceContents{
					mcp.TextResourceContents{
						URI:      "devops-starter://state",
						MIMEType: "application/json",
						Text:     `{"tools":{}}`,
					},
				}, nil
			}
			return nil, fmt.Errorf("reading state file: %w", err)
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "devops-starter://state",
				MIMEType: "application/json",
				Text:     string(data),
			},
		}, nil
	}
}

func groupsResourceHandler(deps *Deps) server.ResourceHandlerFunc {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		type groupInfo struct {
			Name    string `json:"name"`
			Enabled bool   `json:"enabled"`
		}

		groups := make([]groupInfo, 0, len(allGroups()))
		for _, g := range allGroups() {
			groups = append(groups, groupInfo{
				Name:    string(g),
				Enabled: groupEnabled(deps.Config, g),
			})
		}

		data, err := json.MarshalIndent(groups, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("marshaling groups: %w", err)
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "devops-starter://groups",
				MIMEType: "application/json",
				Text:     string(data),
			},
		}, nil
	}
}
