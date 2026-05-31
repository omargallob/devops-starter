package mcp

import (
	"os"
	"path/filepath"
)

// homeDirectory returns the user's home directory.
func homeDirectory() (string, error) {
	return os.UserHomeDir()
}

// dotfilesSourceDir returns the absolute path to the dotfiles directory
// in the repository (relative to working directory).
func dotfilesSourceDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "dotfiles"
	}
	return filepath.Join(wd, "dotfiles")
}
