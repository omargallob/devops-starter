// detect.go provides version detection by executing installed binaries and
// parsing their version output. Each tool has a probe defining the command,
// arguments, and a regex to extract the version string.
package state

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// probeTimeout is the maximum time to wait for a version command to complete.
// Some tools (vault, consul) may try to connect to a server on `version`.
const probeTimeout = 2 * time.Second

// VersionProbe defines how to detect the installed version of a single tool.
type VersionProbe struct {
	// BinName is the binary filename (may differ from tool name, e.g., "nvim" for neovim).
	BinName string
	// Args are the command-line arguments to invoke (e.g., ["--version"]).
	Args []string
	// Regex extracts the version from the combined stdout+stderr output.
	// The first capturing group is used as the version string.
	Regex *regexp.Regexp
}

// probes maps tool names to their version detection configuration.
// BinName is only set when it differs from the tool name.
var probes = map[string]VersionProbe{
	// languages
	"mise":   {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"go":     {Args: []string{"version"}, Regex: re(`go(\d+\.\d+\.\d+)`)},
	"python": {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"node":   {Args: []string{"--version"}, Regex: re(`v?(\d+\.\d+\.\d+)`)},
	"ruby":   {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"java":   {Args: []string{"-version"}, Regex: re(`"(\d+\.\d+\.\d+)"`)},
	"rust":   {BinName: "rustc", Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"deno":   {Args: []string{"--version"}, Regex: re(`deno (\d+\.\d+\.\d+)`)},
	"bun":    {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"erlang": {BinName: "erl", Args: []string{"-eval", "io:format(\"~s~n\", [erlang:system_info(otp_release)]), halt()."}, Regex: re(`(\d+)`)},
	"elixir": {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"zig":    {Args: []string{"version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"kotlin": {BinName: "kotlin", Args: []string{"-version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"scala":  {Args: []string{"-version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"lua":    {Args: []string{"-v"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"perl":   {Args: []string{"--version"}, Regex: re(`v(\d+\.\d+\.\d+)`)},
	"php":    {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"swift":  {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"nim":    {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"julia":  {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"dotnet": {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},

	// containers
	"docker":         {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"docker-compose": {Args: []string{"version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"nerdctl":        {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},

	// kubernetes
	"kubectl":   {Args: []string{"version", "--client", "--output=yaml"}, Regex: re(`gitVersion: v(\d+\.\d+\.\d+)`)},
	"helm":      {Args: []string{"version", "--short"}, Regex: re(`v(\d+\.\d+\.\d+)`)},
	"kustomize": {Args: []string{"version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"k9s":       {Args: []string{"version", "--short"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"kubectx":   {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"kubens":    {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"stern":     {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"argocd":    {Args: []string{"version", "--client"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"flux":      {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"istioctl":  {Args: []string{"version", "--remote=false"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"cilium":    {Args: []string{"version", "--client"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"kind":      {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"kubeseal":  {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"velero":    {Args: []string{"version", "--client-only"}, Regex: re(`(\d+\.\d+\.\d+)`)},

	// infra
	"terraform": {Args: []string{"version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"opentofu":  {BinName: "tofu", Args: []string{"version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"pulumi":    {Args: []string{"version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"packer":    {Args: []string{"version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"vault":     {Args: []string{"version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"consul":    {Args: []string{"version"}, Regex: re(`(\d+\.\d+\.\d+)`)},

	// cloud
	"aws-cli": {BinName: "aws", Args: []string{"--version"}, Regex: re(`aws-cli/(\d+\.\d+\.\d+)`)},
	"eksctl":  {Args: []string{"version"}, Regex: re(`(\d+\.\d+\.\d+)`)},

	// rust-tools
	"bat":       {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"eza":       {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"fd":        {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"ripgrep":   {BinName: "rg", Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"delta":     {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"zoxide":    {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"starship":  {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"tokei":     {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"hyperfine": {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"procs":     {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"bottom":    {BinName: "btm", Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"gitui":     {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"dust":      {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"bandwhich": {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"sd":        {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"tealdeer":  {BinName: "tldr", Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"xh":        {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"yazi":      {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"atuin":     {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"zellij":    {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"just":      {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"watchexec": {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},

	// utilities
	"jq":         {Args: []string{"--version"}, Regex: re(`jq-(\d+\.\d+\.\d*)`)},
	"yq":         {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"fzf":        {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"direnv":     {Args: []string{"version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"age":        {Args: []string{"--version"}, Regex: re(`v(\d+\.\d+\.\d+)`)},
	"sops":       {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"gh":         {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"trivy":      {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"lazygit":    {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"shellcheck": {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"shfmt":      {Args: []string{"--version"}, Regex: re(`v(\d+\.\d+\.\d+)`)},
	"task":       {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"neovim":     {BinName: "nvim", Args: []string{"--version"}, Regex: re(`NVIM v(\d+\.\d+\.\d+)`)},

	// ai
	"ollama":     {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"claude-code": {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"aider":      {Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"openai-cli": {BinName: "openai", Args: []string{"--version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
	"copilot-cli": {BinName: "gh", Args: []string{"copilot", "version"}, Regex: re(`(\d+\.\d+\.\d+)`)},
}

// re is a helper to compile a regexp at init time; panics on invalid patterns.
func re(pattern string) *regexp.Regexp {
	return regexp.MustCompile(pattern)
}

// DetectVersion runs the version probe for the named tool and returns the
// detected version string. The binary is looked up at installDir/binName.
// Returns empty string and an error if the probe fails or the tool is not
// installed.
func DetectVersion(toolName, installDir string) (string, error) {
	probe, ok := probes[toolName]
	if !ok {
		return "", fmt.Errorf("no version probe defined for %s", toolName)
	}

	binName := probe.BinName
	if binName == "" {
		binName = toolName
	}

	binPath := filepath.Join(installDir, binName)
	if _, err := os.Stat(binPath); err != nil {
		return "", fmt.Errorf("binary not found: %s", binPath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, binPath, probe.Args...)
	// Prevent interactive prompts and server connections
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	// Capture both stdout and stderr (some tools print version to stderr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Some tools exit non-zero for --version (rare but possible)
		// Still try to parse if we got output
		if len(output) == 0 {
			return "", fmt.Errorf("running %s %s: %w", binPath, strings.Join(probe.Args, " "), err)
		}
	}

	matches := probe.Regex.FindSubmatch(output)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not parse version from output of %s: %q", binPath, string(output))
	}

	return string(matches[1]), nil
}

// LookupInPath checks if the tool's binary exists anywhere in $PATH using
// exec.LookPath. No subprocess is spawned. Returns the resolved absolute path
// if found, or empty string if not. Uses the probe's BinName mapping (e.g.,
// neovim→nvim, ripgrep→rg) when available.
func LookupInPath(toolName string) string {
	binName := toolName
	if probe, ok := probes[toolName]; ok && probe.BinName != "" {
		binName = probe.BinName
	}
	path, err := exec.LookPath(binName)
	if err != nil {
		return ""
	}
	return path
}

// LookupToolInPath is like LookupInPath but accepts a Tool struct, enabling
// detection of tools whose binary name differs from the tool name via the
// InstallName field (e.g., openai-cli→openai, aws-cli→aws). The lookup order
// is: probe BinName → Tool.GetInstallName() → tool name.
func LookupToolInPath(tool *tooldef.Tool) string {
	// First try the probe-aware lookup by tool name.
	if path := LookupInPath(tool.Name); path != "" {
		return path
	}
	// Fall back to InstallName if it differs from the tool name.
	installName := tool.GetInstallName()
	if installName != tool.Name {
		if path, err := exec.LookPath(installName); err == nil {
			return path
		}
	}
	return ""
}

// DetectVersionAtPath runs the version probe for the named tool against an
// arbitrary binary path (e.g., a system-installed binary found in PATH).
// Returns the detected version string or an error.
func DetectVersionAtPath(toolName, binPath string) (string, error) {
	probe, ok := probes[toolName]
	if !ok {
		return "", fmt.Errorf("no version probe defined for %s", toolName)
	}

	if _, err := os.Stat(binPath); err != nil {
		return "", fmt.Errorf("binary not found: %s", binPath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, binPath, probe.Args...)
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	output, err := cmd.CombinedOutput()
	if err != nil && len(output) == 0 {
		return "", fmt.Errorf("running %s %s: %w", binPath, strings.Join(probe.Args, " "), err)
	}

	matches := probe.Regex.FindSubmatch(output)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not parse version from output of %s: %q", binPath, string(output))
	}

	return string(matches[1]), nil
}

// VerifyAll runs version detection for all tools that have a binary present
// in installDir, updating the store with detected versions. Tools without
// a probe or without a binary are skipped silently.
func VerifyAll(store *Store, installDir string) {
	for toolName := range probes {
		version, err := DetectVersion(toolName, installDir)
		if err != nil {
			continue // skip tools that aren't installed or can't be probed
		}
		store.Tools[toolName] = InstalledTool{
			Version:     version,
			InstalledAt: store.Tools[toolName].InstalledAt, // preserve original install time
		}
	}
}

// VerifyOne runs version detection for a single tool, updating the store.
// Returns the detected version or an error.
func VerifyOne(store *Store, toolName, installDir string) (string, error) {
	version, err := DetectVersion(toolName, installDir)
	if err != nil {
		return "", err
	}
	existing := store.Tools[toolName]
	existing.Version = version
	store.Tools[toolName] = existing
	return version, nil
}
