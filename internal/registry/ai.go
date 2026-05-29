// ai.go registers AI-related developer tools: LLM runners, AI coding
// assistants, and AI CLI interfaces. Tools in this group use a mix of
// install modes: eget for native binaries, mise backends for npm/pip
// distributed tools, and gh-extension for GitHub CLI extensions.
package registry

import "github.com/omargallob/devops-starter/pkg/tooldef"

// registerAI adds AI tools to the registry.
// Includes ollama, claude-code, aider, openai-cli, and copilot-cli.
func registerAI(r *Registry) {
	r.register(&tooldef.Tool{
		Name:        "ollama",
		Version:     "0.9.0",
		Description: "Local LLM runner",
		Group:       tooldef.GroupAI,
		Subgroup:    "Runtimes",
		InstallMode: tooldef.InstallModeEget,
		Repo:        "ollama/ollama",
	})

	r.register(&tooldef.Tool{
		Name:         "claude-code",
		Version:      "1.0.3",
		Description:  "Anthropic Claude AI coding assistant",
		Group:        tooldef.GroupAI,
		Subgroup:     "Coding Assistants",
		InstallMode:  tooldef.InstallModeMise,
		MiseBackend:  "npm:@anthropic-ai/claude-code",
		Dependencies: []string{"mise"},
	})

	r.register(&tooldef.Tool{
		Name:         "aider",
		Version:      "0.82.3",
		Description:  "AI pair programming in your terminal",
		Group:        tooldef.GroupAI,
		Subgroup:     "Coding Assistants",
		InstallMode:  tooldef.InstallModeMise,
		MiseBackend:  "pipx:aider-chat",
		Dependencies: []string{"mise"},
	})

	r.register(&tooldef.Tool{
		Name:         "openai-cli",
		Version:      "1.82.0",
		Description:  "OpenAI CLI and Python library",
		Group:        tooldef.GroupAI,
		Subgroup:     "Platform CLIs",
		InstallMode:  tooldef.InstallModeMise,
		MiseBackend:  "pipx:openai",
		InstallName:  "openai",
		Dependencies: []string{"mise"},
	})

	r.register(&tooldef.Tool{
		Name:         "copilot-cli",
		Version:      "1.0.5",
		Description:  "GitHub Copilot in the CLI",
		Group:        tooldef.GroupAI,
		Subgroup:     "Platform CLIs",
		InstallMode:  tooldef.InstallModeGhExtension,
		Repo:         "github/gh-copilot",
		Dependencies: []string{"gh"},
	})
}
