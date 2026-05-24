package state

import (
	"testing"
)

// TestVersionProbeRegexes validates that the regex patterns for each tool
// correctly extract version strings from representative output samples.
func TestVersionProbeRegexes(t *testing.T) {
	tests := []struct {
		tool     string
		output   string
		expected string
	}{
		// languages
		{"mise", "mise 2025.1.6 linux-x64", "2025.1.6"},
		// containers
		{"docker", "Docker version 27.5.1, build abcdef", "27.5.1"},
		{"docker-compose", "Docker Compose version v2.32.4", "2.32.4"},
		{"nerdctl", "nerdctl version 2.0.3", "2.0.3"},
		// kubernetes
		{"kubectl", "  gitVersion: v1.31.4", "1.31.4"},
		{"helm", "v3.16.4+g7d45367", "3.16.4"},
		{"kustomize", "v5.5.0", "5.5.0"},
		{"k9s", "v0.32.7", "0.32.7"},
		{"stern", "version: 1.31.0", "1.31.0"},
		{"argocd", "argocd: v2.13.3+abcdef", "2.13.3"},
		{"flux", "flux version 2.4.0", "2.4.0"},
		{"kind", "kind v0.25.0 go1.23.0 linux/amd64", "0.25.0"},
		// infra
		{"terraform", "Terraform v1.10.4\non linux_amd64", "1.10.4"},
		{"opentofu", "OpenTofu v1.8.0\non linux_amd64", "1.8.0"},
		{"pulumi", "v3.150.0", "3.150.0"},
		{"vault", "Vault v1.18.4 (abc123)", "1.18.4"},
		{"consul", "Consul v1.20.2", "1.20.2"},
		// cloud
		{"aws-cli", "aws-cli/2.22.35 Python/3.12.0 Linux/6.5.0", "2.22.35"},
		{"eksctl", "0.200.0", "0.200.0"},
		// rust-tools
		{"bat", "bat 0.24.0 (abc1234)", "0.24.0"},
		{"ripgrep", "ripgrep 14.1.1 (rev abc1234)", "14.1.1"},
		{"delta", "delta 0.18.2", "0.18.2"},
		{"starship", "starship 1.22.1", "1.22.1"},
		{"bottom", "btm 0.10.2", "0.10.2"},
		// utilities
		{"jq", "jq-1.7.1", "1.7.1"},
		{"fzf", "0.57.0 (abcdef)", "0.57.0"},
		{"age", "v1.2.0", "1.2.0"},
		{"gh", "gh version 2.65.0 (2025-01-20)", "2.65.0"},
		{"neovim", "NVIM v0.10.3\nBuild type: Release", "0.10.3"},
		{"shfmt", "v3.10.0", "3.10.0"},
		{"shellcheck", "version: 0.10.0", "0.10.0"},
		{"direnv", "2.35.0", "2.35.0"},
	}

	for _, tc := range tests {
		t.Run(tc.tool, func(t *testing.T) {
			probe, ok := probes[tc.tool]
			if !ok {
				t.Fatalf("no probe defined for tool %s", tc.tool)
			}

			matches := probe.Regex.FindStringSubmatch(tc.output)
			if len(matches) < 2 {
				t.Fatalf("regex did not match output %q for tool %s", tc.output, tc.tool)
			}

			got := matches[1]
			if got != tc.expected {
				t.Errorf("tool %s: got version %q, want %q (from output %q)", tc.tool, got, tc.expected, tc.output)
			}
		})
	}
}

// TestAllToolsHaveProbes ensures every probe has a valid regex (compile-time check
// already handles this via re(), but this confirms the map is populated).
func TestAllToolsHaveProbes(t *testing.T) {
	if len(probes) < 61 {
		t.Errorf("expected at least 61 probes, got %d", len(probes))
	}

	for name, probe := range probes {
		if probe.Regex == nil {
			t.Errorf("probe for %s has nil regex", name)
		}
		if len(probe.Args) == 0 {
			t.Errorf("probe for %s has no args", name)
		}
	}
}
