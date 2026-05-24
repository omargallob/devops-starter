// Package state manages the persistent state file that tracks which tools
// are installed and at what version. The state file lives at
// ~/.config/devops-starter/state.json (respects XDG_CONFIG_HOME).
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Store represents the on-disk state of all installed tools.
type Store struct {
	Tools map[string]InstalledTool `json:"tools"`
	path  string
}

// InstalledTool records the version and install timestamp for a single tool.
type InstalledTool struct {
	Version     string    `json:"version"`
	InstalledAt time.Time `json:"installed_at"`
}

// StatePath returns the default state file path, respecting XDG_CONFIG_HOME.
func StatePath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "devops-starter", "state.json")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".config", "devops-starter", "state.json")
}

// LoadStore reads the state file from disk. If the file does not exist,
// an empty store is returned (not an error).
func LoadStore(path string) (*Store, error) {
	s := &Store{
		Tools: make(map[string]InstalledTool),
		path:  path,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, fmt.Errorf("reading state file: %w", err)
	}

	if err := json.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("parsing state file %s: %w", path, err)
	}
	if s.Tools == nil {
		s.Tools = make(map[string]InstalledTool)
	}

	return s, nil
}

// Save writes the current state to disk, creating parent directories if needed.
func (s *Store) Save() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating state dir: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}

	return nil
}

// Record updates the installed version for a tool and persists to disk.
func (s *Store) Record(name, version string) error {
	s.Tools[name] = InstalledTool{
		Version:     version,
		InstalledAt: time.Now().UTC(),
	}
	return s.Save()
}

// GetVersion returns the recorded installed version for a tool, or empty string.
func (s *Store) GetVersion(name string) string {
	if t, ok := s.Tools[name]; ok {
		return t.Version
	}
	return ""
}

// Remove deletes a tool from the state store and persists to disk.
func (s *Store) Remove(name string) error {
	delete(s.Tools, name)
	return s.Save()
}
