// Package mise provides parsing of .mise.toml configuration files to discover
// tool versions managed by the mise polyglot runtime manager.
package mise

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// ConfigFile is the standard mise configuration filename.
const ConfigFile = ".mise.toml"

// miseConfig represents the top-level structure of a .mise.toml file.
// Only the [tools] section is relevant for version discovery.
type miseConfig struct {
	Tools map[string]interface{} `toml:"tools"`
}

// ToolVersions maps tool names to their version strings as declared in a
// .mise.toml file. For example: {"go": "1.26.3", "python": "3.12", "node": "22"}.
type ToolVersions map[string]string

// ParseFile reads and parses a .mise.toml file at the given path, returning
// the tool name → version mapping from the [tools] section.
func ParseFile(path string) (ToolVersions, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading mise config: %w", err)
	}

	var cfg miseConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing mise config: %w", err)
	}

	if cfg.Tools == nil {
		return ToolVersions{}, nil
	}

	versions := make(ToolVersions, len(cfg.Tools))
	for name, val := range cfg.Tools {
		versions[name] = normalizeVersion(val)
	}

	return versions, nil
}

// Find locates a .mise.toml file by walking up from startDir to the filesystem
// root. Returns the full path to the first .mise.toml found, or empty string
// if none exists.
func Find(startDir string) string {
	dir := startDir
	for {
		candidate := filepath.Join(dir, ConfigFile)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "" // reached filesystem root
		}
		dir = parent
	}
}

// FindAndParse locates and parses the nearest .mise.toml starting from the
// given directory. Returns nil (not an error) if no .mise.toml is found.
func FindAndParse(startDir string) (ToolVersions, error) {
	path := Find(startDir)
	if path == "" {
		return nil, nil
	}
	return ParseFile(path)
}

// normalizeVersion converts the TOML value (which may be a string, number, or
// array) into a version string. Mise supports multiple formats:
//   - Simple string: go = "1.26.3"
//   - Number: node = 22
//   - Array (multiple versions): python = ["3.12", "3.11"]  → uses first
func normalizeVersion(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		// Avoid trailing zeros: 3.12 → "3.12", 22.0 → "22"
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%g", v)
	case []interface{}:
		if len(v) > 0 {
			return normalizeVersion(v[0])
		}
		return ""
	default:
		return fmt.Sprintf("%v", val)
	}
}

// Descriptions provides human-readable descriptions for well-known language
// runtimes managed by mise. Used when dynamically registering tools.
var Descriptions = map[string]string{
	"go":      "Go programming language",
	"python":  "Python programming language",
	"node":    "Node.js JavaScript runtime",
	"ruby":    "Ruby programming language",
	"java":    "Java Development Kit",
	"rust":    "Rust programming language",
	"deno":    "Deno JavaScript/TypeScript runtime",
	"bun":     "Bun JavaScript runtime",
	"erlang":  "Erlang/OTP runtime",
	"elixir":  "Elixir programming language",
	"zig":     "Zig programming language",
	"kotlin":  "Kotlin programming language",
	"scala":   "Scala programming language",
	"lua":     "Lua programming language",
	"perl":    "Perl programming language",
	"php":     "PHP programming language",
	"r":       "R statistical computing language",
	"swift":   "Swift programming language",
	"crystal": "Crystal programming language",
	"nim":     "Nim programming language",
	"julia":   "Julia programming language",
	"haskell": "Haskell programming language",
	"ocaml":   "OCaml programming language",
	"dotnet":  ".NET SDK",
}

// DescriptionFor returns a human-readable description for a mise-managed tool.
// Falls back to a generic description if the tool is not in the known list.
func DescriptionFor(name string) string {
	if desc, ok := Descriptions[name]; ok {
		return desc
	}
	return fmt.Sprintf("%s (managed by mise)", name)
}
