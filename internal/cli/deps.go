package cli

import (
	"context"
	"io"

	"github.com/omargallob/devops-starter/internal/config"
	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// ToolInstaller defines the interface for installing and managing tools.
// This allows commands to be tested with mock implementations.
type ToolInstaller interface {
	Install(ctx context.Context, tool *tooldef.Tool) error
	InstallAll(ctx context.Context, tools []*tooldef.Tool) []error
	IsInstalled(tool *tooldef.Tool) bool
	EnsureDir() error
	Link(tool *tooldef.Tool, systemPath string) error
}

// ToolRegistry defines the interface for looking up tools.
type ToolRegistry interface {
	All() []*tooldef.Tool
	Get(name string) (*tooldef.Tool, bool)
	GetByGroup(group tooldef.Group) []*tooldef.Tool
	Names() []string
}

// StateStore defines the interface for managing installed tool state.
type StateStore interface {
	GetVersion(name string) string
	Record(name, version string) error
	Remove(name string) error
	Save() error
}

// installDeps bundles dependencies for the install command.
type installDeps struct {
	cfg       *config.Config
	registry  ToolRegistry
	installer ToolInstaller
	out       io.Writer
	dryRun    bool
	autoYes   bool
	only      string
}

// removeDeps bundles dependencies for the remove command.
type removeDeps struct {
	cfg        *config.Config
	registry   ToolRegistry
	store      StateStore
	out        io.Writer
	dryRun     bool
	autoYes    bool
	installDir string
}
