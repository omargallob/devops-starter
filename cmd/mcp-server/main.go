// Command mcp-server runs the devops-starter MCP server over stdio.
// It exposes tool listing, status queries, configuration, platform detection,
// and dotfiles status as MCP tools and resources for AI chat clients.
package main

import (
	"log"

	mcpserver "github.com/omargallob/devops-starter/internal/mcp"
)

func main() {
	if err := mcpserver.Run(); err != nil {
		log.Fatal(err)
	}
}
