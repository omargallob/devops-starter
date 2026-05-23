// Package dotfiles manages symlinking configuration files from the repo to the user's home.
package dotfiles

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LinkStatus represents the state of a dotfile symlink.
type LinkStatus int

const (
	StatusLinked    LinkStatus = iota // Symlink exists and points to our file
	StatusConflict                    // File exists but is not our symlink
	StatusMissing                     // No file at destination
	StatusBroken                      // Symlink exists but target is gone
)

func (s LinkStatus) String() string {
	switch s {
	case StatusLinked:
		return "linked"
	case StatusConflict:
		return "conflict"
	case StatusMissing:
		return "missing"
	case StatusBroken:
		return "broken"
	default:
		return "unknown"
	}
}

// Mapping defines a source→destination dotfile mapping.
type Mapping struct {
	Source string // Path relative to repo dotfiles/ dir
	Dest   string // Path relative to $HOME
}

// LinkResult holds the result of a link operation.
type LinkResult struct {
	Mapping Mapping
	Status  LinkStatus
	Error   error
}

// Linker manages dotfile symlinks between a source directory and $HOME.
type Linker struct {
	// SourceDir is the absolute path to the dotfiles directory in the repo.
	SourceDir string
	// HomeDir is the target home directory (usually $HOME).
	HomeDir string
	// DryRun if true, only reports what would happen.
	DryRun bool
	// BackupDir is where conflicting files are moved. Defaults to ~/.dotfiles.bak
	BackupDir string
}

// NewLinker creates a new Linker instance.
func NewLinker(sourceDir, homeDir string, dryRun bool) *Linker {
	return &Linker{
		SourceDir: sourceDir,
		HomeDir:   homeDir,
		DryRun:    dryRun,
		BackupDir: filepath.Join(homeDir, ".dotfiles.bak"),
	}
}

// DefaultMappings returns the standard dotfile mappings.
func DefaultMappings() []Mapping {
	return []Mapping{
		// ZSH
		{Source: "zsh/.zshrc", Dest: ".zshrc"},
		{Source: "zsh/.zprofile", Dest: ".zprofile"},
		// Bash
		{Source: "bash/.bashrc", Dest: ".bashrc"},
		{Source: "bash/.bash_profile", Dest: ".bash_profile"},
		// Git
		{Source: "git/.gitconfig", Dest: ".gitconfig"},
		{Source: "git/.gitignore_global", Dest: ".gitignore_global"},
		// Tmux
		{Source: "tmux/.tmux.conf", Dest: ".tmux.conf"},
		// Starship
		{Source: "starship/starship.toml", Dest: ".config/starship.toml"},
		// Neovim
		{Source: "nvim/init.lua", Dest: ".config/nvim/init.lua"},
		{Source: "nvim/lua/plugins.lua", Dest: ".config/nvim/lua/plugins.lua"},
		{Source: "nvim/lua/options.lua", Dest: ".config/nvim/lua/options.lua"},
	}
}

// Status checks the current state of all mappings.
func (l *Linker) Status(mappings []Mapping) []LinkResult {
	results := make([]LinkResult, 0, len(mappings))
	for _, m := range mappings {
		results = append(results, LinkResult{
			Mapping: m,
			Status:  l.checkStatus(m),
		})
	}
	return results
}

// Link creates symlinks for all mappings, backing up conflicts.
func (l *Linker) Link(mappings []Mapping) []LinkResult {
	results := make([]LinkResult, 0, len(mappings))
	for _, m := range mappings {
		result := l.linkOne(m)
		results = append(results, result)
	}
	return results
}

// Unlink removes symlinks that point to our source files.
func (l *Linker) Unlink(mappings []Mapping) []LinkResult {
	results := make([]LinkResult, 0, len(mappings))
	for _, m := range mappings {
		result := l.unlinkOne(m)
		results = append(results, result)
	}
	return results
}

func (l *Linker) sourcePath(m Mapping) string {
	return filepath.Join(l.SourceDir, m.Source)
}

func (l *Linker) destPath(m Mapping) string {
	return filepath.Join(l.HomeDir, m.Dest)
}

func (l *Linker) checkStatus(m Mapping) LinkStatus {
	dest := l.destPath(m)
	source := l.sourcePath(m)

	info, err := os.Lstat(dest)
	if os.IsNotExist(err) {
		return StatusMissing
	}
	if err != nil {
		return StatusConflict
	}

	// Check if it's a symlink
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(dest)
		if err != nil {
			return StatusBroken
		}
		// Resolve relative symlinks
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(dest), target)
		}
		if target == source {
			return StatusLinked
		}
		// Symlink to somewhere else
		if _, err := os.Stat(target); os.IsNotExist(err) {
			return StatusBroken
		}
		return StatusConflict
	}

	// Regular file or directory - conflict
	return StatusConflict
}

func (l *Linker) linkOne(m Mapping) LinkResult {
	source := l.sourcePath(m)
	dest := l.destPath(m)
	status := l.checkStatus(m)

	result := LinkResult{Mapping: m, Status: status}

	switch status {
	case StatusLinked:
		// Already linked correctly
		return result

	case StatusMissing, StatusBroken:
		if l.DryRun {
			result.Status = StatusLinked
			return result
		}
		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			result.Error = fmt.Errorf("creating parent dir: %w", err)
			return result
		}
		// Remove broken symlink if present
		if status == StatusBroken {
			os.Remove(dest)
		}
		if err := os.Symlink(source, dest); err != nil {
			result.Error = fmt.Errorf("creating symlink: %w", err)
			return result
		}
		result.Status = StatusLinked

	case StatusConflict:
		if l.DryRun {
			return result
		}
		// Backup existing file
		if err := l.backup(dest, m); err != nil {
			result.Error = fmt.Errorf("backing up: %w", err)
			return result
		}
		// Create symlink
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			result.Error = fmt.Errorf("creating parent dir: %w", err)
			return result
		}
		if err := os.Symlink(source, dest); err != nil {
			result.Error = fmt.Errorf("creating symlink: %w", err)
			return result
		}
		result.Status = StatusLinked
	}

	return result
}

func (l *Linker) unlinkOne(m Mapping) LinkResult {
	dest := l.destPath(m)
	status := l.checkStatus(m)
	result := LinkResult{Mapping: m, Status: status}

	if status != StatusLinked {
		// Only remove symlinks we own
		return result
	}

	if l.DryRun {
		result.Status = StatusMissing
		return result
	}

	if err := os.Remove(dest); err != nil {
		result.Error = fmt.Errorf("removing symlink: %w", err)
		return result
	}
	result.Status = StatusMissing
	return result
}

func (l *Linker) backup(path string, m Mapping) error {
	backupPath := filepath.Join(l.BackupDir, m.Dest)
	if err := os.MkdirAll(filepath.Dir(backupPath), 0o755); err != nil {
		return err
	}
	// Remove old backup if exists
	os.Remove(backupPath)
	return os.Rename(path, backupPath)
}

// Summary returns a human-readable summary of link results.
func Summary(results []LinkResult) string {
	var linked, conflicts, missing, errors int
	for _, r := range results {
		if r.Error != nil {
			errors++
			continue
		}
		switch r.Status {
		case StatusLinked:
			linked++
		case StatusConflict:
			conflicts++
		case StatusMissing:
			missing++
		}
	}

	parts := []string{}
	if linked > 0 {
		parts = append(parts, fmt.Sprintf("%d linked", linked))
	}
	if conflicts > 0 {
		parts = append(parts, fmt.Sprintf("%d conflicts", conflicts))
	}
	if missing > 0 {
		parts = append(parts, fmt.Sprintf("%d missing", missing))
	}
	if errors > 0 {
		parts = append(parts, fmt.Sprintf("%d errors", errors))
	}
	return strings.Join(parts, ", ")
}
