# MCP Server

devops-starter includes an MCP (Model Context Protocol) server that allows AI chat clients to interact with the tool programmatically. This enables you to query your tool catalog, check installation status, view configuration, and more — all from within your AI coding assistant.

## Build

```sh
make mcp-server
```

This produces a `devops-starter-mcp` binary in the project root.

To install it to `~/.local/bin`:

```sh
make install-mcp
```

## Configuration

### OpenCode

The repository includes an `opencode.json` that configures the MCP server automatically. If you're running opencode from the project root, it will detect and start the server.

For use outside the project, add to your global opencode config (`~/.config/opencode/opencode.json`):

```json
{
  "mcpServers": {
    "devops-starter": {
      "command": "/path/to/devops-starter-mcp"
    }
  }
}
```

### Claude Desktop

Add to `~/.config/claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "devops-starter": {
      "command": "/path/to/devops-starter-mcp"
    }
  }
}
```

### Cursor

Add to your Cursor MCP settings:

```json
{
  "mcpServers": {
    "devops-starter": {
      "command": "/path/to/devops-starter-mcp"
    }
  }
}
```

### Any MCP Client (stdio)

The server communicates over stdin/stdout using JSON-RPC 2.0 (MCP protocol version `2024-11-05`). Point any MCP-compatible client at the binary.

## Available Tools

| Tool | Description | Parameters |
|------|-------------|------------|
| `list_tools` | List all available DevOps tools, optionally filtered by group | `group` (optional): one of `languages`, `containers`, `kubernetes`, `infra`, `cloud`, `ansible`, `rust-tools`, `utilities`, `ai`, `package-managers` |
| `get_tool` | Get full details of a specific tool | `name` (required): tool name (e.g., `kubectl`, `bat`, `terraform`) |
| `get_status` | Get installation status of all tools | `group` (optional), `status_filter` (optional): `missing`, `current`, `outdated`, `disabled`, `detected`, `linked`, `unavailable` |
| `config_show` | Show current configuration as YAML | (none) |
| `detect_platform` | Detect OS, architecture, and Linux distro | (none) |
| `dotfiles_status` | Show symlink status of managed dotfiles | (none) |

All tools are read-only and produce no side effects.

## Available Resources

Resources provide raw data that AI clients can read for context:

| URI | Description | MIME Type |
|-----|-------------|-----------|
| `devops-starter://config` | Current configuration file (YAML) | `text/yaml` |
| `devops-starter://state` | Installed tools state (JSON) | `application/json` |
| `devops-starter://groups` | Tool groups with enabled/disabled status | `application/json` |

## Example Interactions

Once configured, you can ask your AI assistant things like:

- "What DevOps tools are available in the kubernetes group?"
- "Which tools are currently missing from my system?"
- "Show me my devops-starter configuration"
- "What platform am I running on?"
- "Are my dotfiles properly linked?"
- "What version of terraform is configured?"

## Architecture

```
cmd/mcp-server/main.go     Entry point (thin wrapper)
internal/mcp/
  server.go                 Server factory, dependency wiring
  tools.go                  MCP tool handler implementations
  resources.go              MCP resource handler implementations
  helpers.go                Shared utilities
  server_test.go            Contract tests
```

The MCP server is a thin adapter layer over the same internal packages used by the CLI (`registry`, `config`, `state`, `platform`, `dotfiles`). It shares the same tool definitions, configuration loading, and state management — ensuring consistency between CLI and chat interactions.

## Development

Run the contract tests:

```sh
go test -race ./internal/mcp/
```

Smoke test the server manually:

```sh
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | ./devops-starter-mcp
```

## Roadmap (Phase 2)

Write operations will be added in a future release:

- `install_tools` — Install one or more tools (with `dry_run` support)
- `remove_tools` — Remove managed tools
- `adopt_tools` — Adopt system-installed tools into managed state
- `dotfiles_link` / `dotfiles_unlink` — Manage dotfile symlinks
- `doctor` — Run system health checks (with optional `fix`)
- `config_init` — Create default config file
- `install_packages` — Install pip/npm packages from .mise.toml
