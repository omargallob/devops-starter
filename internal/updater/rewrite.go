package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// registryDir is the relative path from the project root to registry source files.
const registryDir = "internal/registry"

// groupFileMap maps tool group names to their registry source file.
var groupFileMap = map[string]string{
	"languages":  "languages.go",
	"containers": "containers.go",
	"kubernetes": "kubernetes.go",
	"infra":      "infra.go",
	"cloud":      "cloud.go",
	"rust-tools": "rust_tools.go",
	"utilities":  "utilities.go",
}

// ApplyUpdates rewrites registry Go source files to update pinned versions.
// rootDir is the project root directory (containing internal/registry/).
// Returns the number of tools actually updated and any errors.
func ApplyUpdates(rootDir string, updates []UpdateInfo) (int, []error) {
	// Group updates by file.
	byFile := make(map[string][]UpdateInfo)
	for _, u := range updates {
		filename, ok := groupFileMap[u.Group]
		if !ok {
			continue
		}
		byFile[filename] = append(byFile[filename], u)
	}

	var (
		applied int
		errs    []error
	)

	for filename, fileUpdates := range byFile {
		filePath := filepath.Join(rootDir, registryDir, filename)

		content, err := os.ReadFile(filePath)
		if err != nil {
			errs = append(errs, fmt.Errorf("reading %s: %w", filePath, err))
			continue
		}

		newContent := string(content)
		for _, u := range fileUpdates {
			updated, ok := rewriteToolVersion(newContent, u.ToolName, u.CurrentVersion, u.LatestVersion)
			if ok {
				newContent = updated
				applied++
			} else {
				errs = append(errs, fmt.Errorf("could not locate version %s for %s in %s", u.CurrentVersion, u.ToolName, filename))
			}
		}

		if newContent != string(content) {
			if err := os.WriteFile(filePath, []byte(newContent), 0o644); err != nil {
				errs = append(errs, fmt.Errorf("writing %s: %w", filePath, err))
			}
		}
	}

	return applied, errs
}

// rewriteToolVersion finds the tool's registration block and updates its Version field.
// It uses a targeted approach: find the Name line, then find the nearest Version line.
func rewriteToolVersion(content, toolName, oldVersion, newVersion string) (string, bool) {
	// Build a regex that matches the tool's Name + Version in close proximity.
	// Pattern: Name: "toolName" followed by Version: "oldVersion" within the same struct literal.
	//
	// We use a two-pass approach:
	// 1. Find the struct block containing Name: "toolName"
	// 2. Replace the Version line within that block.

	// Find the line with Name: "toolName"
	namePattern := fmt.Sprintf(`Name:\s*%q`, toolName)
	nameRe := regexp.MustCompile(namePattern)

	nameLoc := nameRe.FindStringIndex(content)
	if nameLoc == nil {
		return content, false
	}

	// Search for the Version field in the next ~500 chars (within same struct).
	searchStart := nameLoc[0]
	searchEnd := searchStart + 500
	if searchEnd > len(content) {
		searchEnd = len(content)
	}

	block := content[searchStart:searchEnd]

	// Find the closing brace to limit our search to this struct.
	braceIdx := strings.Index(block, "})")
	if braceIdx == -1 {
		braceIdx = strings.Index(block, "}")
	}
	if braceIdx > 0 {
		block = block[:braceIdx]
	}

	// Replace the Version line within this block.
	versionPattern := fmt.Sprintf(`Version:\s*%q`, oldVersion)
	versionRe := regexp.MustCompile(versionPattern)

	versionLoc := versionRe.FindStringIndex(block)
	if versionLoc == nil {
		return content, false
	}

	// Calculate absolute positions.
	absStart := searchStart + versionLoc[0]
	absEnd := searchStart + versionLoc[1]

	replacement := fmt.Sprintf("Version:     %q", newVersion)
	// Preserve the original spacing by matching the original format.
	original := content[absStart:absEnd]
	if strings.Contains(original, "Version:     ") {
		replacement = fmt.Sprintf("Version:     %q", newVersion)
	} else if strings.Contains(original, "Version: ") {
		replacement = fmt.Sprintf("Version: %q", newVersion)
	}

	result := content[:absStart] + replacement + content[absEnd:]

	// Also update version strings embedded in URLs if present.
	result = rewriteURLVersions(result, toolName, oldVersion, newVersion)

	return result, true
}

// rewriteURLVersions updates version strings embedded in URL templates and URL maps
// for a specific tool. This handles tools like pulumi where the version appears in
// hardcoded download URLs.
func rewriteURLVersions(content, toolName, oldVersion, newVersion string) string {
	// Find the tool's block again.
	namePattern := fmt.Sprintf(`Name:\s*%q`, toolName)
	nameRe := regexp.MustCompile(namePattern)

	nameLoc := nameRe.FindStringIndex(content)
	if nameLoc == nil {
		return content
	}

	// Find the next tool block or end of function to delimit this tool's block.
	searchStart := nameLoc[0]
	nextTool := regexp.MustCompile(`r\.register\(&tooldef\.Tool\{`)
	remaining := content[searchStart+1:]
	nextLoc := nextTool.FindStringIndex(remaining)

	var blockEnd int
	if nextLoc != nil {
		blockEnd = searchStart + 1 + nextLoc[0]
	} else {
		blockEnd = len(content)
	}

	block := content[searchStart:blockEnd]

	// Replace all occurrences of oldVersion in URLs within this block.
	newBlock := strings.ReplaceAll(block, oldVersion, newVersion)

	if newBlock != block {
		return content[:searchStart] + newBlock + content[blockEnd:]
	}

	return content
}
