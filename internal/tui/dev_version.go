package tui

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:generate cp ../../../.release-please-manifest.json version_manifest.json

//go:embed version_manifest.json
var manifestBytes []byte

// devVersionLabel returns the version string to display when running in dev mode.
// It reads the version from the embedded release-please manifest and appends "-dev".
func devVersionLabel() string {
	var manifest map[string]string
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return "dev"
	}
	if v, ok := manifest["."]; ok && v != "" {
		// Bump patch for dev (next unreleased version)
		parts := strings.SplitN(v, ".", 3)
		if len(parts) == 3 {
			// Parse and increment patch
			patch := 0
			for _, ch := range parts[2] {
				if ch >= '0' && ch <= '9' {
					patch = patch*10 + int(ch-'0')
				} else {
					break
				}
			}
			patch++
			return parts[0] + "." + parts[1] + "." + strings.Repeat("", 0) + itoa(patch) + "-dev"
		}
		return v + "-dev"
	}
	return "dev"
}

// itoa converts a small int to string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
