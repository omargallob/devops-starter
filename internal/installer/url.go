package installer

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/omargallob/devops-starter/pkg/tooldef"
)

// urlTemplateData holds the variables available in URL templates.
type urlTemplateData struct {
	Name       string
	Version    string
	OS         string
	Arch       string
	Format     string
	BinaryName string
}

// ResolveURL determines the download URL for a tool on the given platform.
// It checks tool.URLs for a platform-specific override first, then renders
// tool.URLTemplate using text/template.
func ResolveURL(tool *tooldef.Tool, platform tooldef.Platform) (string, error) {
	key := platform.String()

	// Check platform-specific override
	if url, ok := tool.URLs[key]; ok {
		return url, nil
	}

	if tool.URLTemplate == "" {
		return "", fmt.Errorf("no URL template or override for tool %s on %s", tool.Name, key)
	}

	tmpl, err := template.New("url").Parse(tool.URLTemplate)
	if err != nil {
		return "", fmt.Errorf("parsing URL template: %w", err)
	}

	data := urlTemplateData{
		Name:       tool.Name,
		Version:    tool.Version,
		OS:         platform.OS,
		Arch:       platform.Arch,
		Format:     string(tool.Format),
		BinaryName: tool.GetBinaryName(),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing URL template: %w", err)
	}

	return buf.String(), nil
}
