// Command generate-coverage-badge generates a flat-style SVG coverage badge.
//
// Usage:
//
//	generate-coverage-badge <percentage> <output-file>
//	e.g.: generate-coverage-badge 82.5 badges/coverage.svg
package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <percentage> <output-file>\n", os.Args[0])
		os.Exit(1)
	}

	pct, err := strconv.ParseFloat(os.Args[1], 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid percentage %q: %v\n", os.Args[1], err)
		os.Exit(1)
	}

	output := os.Args[2]

	svg := GenerateBadge(pct)

	if err := os.WriteFile(output, []byte(svg), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write %s: %v\n", output, err)
		os.Exit(1)
	}

	color := BadgeColor(pct)
	fmt.Printf("Badge written to %s (%.1f%%, color: %s)\n", output, pct, color)
}

// BadgeColor returns the hex color for a given coverage percentage.
func BadgeColor(pct float64) string {
	switch {
	case pct >= 80:
		return "#4c1" // bright green
	case pct >= 60:
		return "#dfb317" // yellow
	case pct >= 40:
		return "#fe7d37" // orange
	default:
		return "#e05d44" // red
	}
}

// GenerateBadge produces an SVG badge string for the given coverage percentage.
func GenerateBadge(pct float64) string {
	label := "coverage"
	message := fmt.Sprintf("%.1f%%", pct)
	// Drop trailing zero for whole numbers (e.g. "80%" not "80.0%").
	if pct == float64(int(pct)) {
		message = fmt.Sprintf("%d%%", int(pct))
	}

	color := BadgeColor(pct)

	// Approximate character widths at font-size 11 in Verdana.
	// Digits and '%' are wider than lowercase letters, so use 7.5px for the message.
	labelWidth := int(float64(len(label))*6.5 + 10)
	msgWidth := int(float64(len(message))*8.5 + 10)
	totalWidth := labelWidth + msgWidth

	// Center positions scaled x10 for the SVG transform="scale(.1)".
	labelX := labelWidth * 10 / 2
	msgX := (labelWidth + msgWidth/2) * 10

	// textLength values scaled x10.
	labelTL := (labelWidth - 10) * 10
	msgTL := (msgWidth - 10) * 10

	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="20" role="img" aria-label="%s: %s">
  <title>%s: %s</title>
  <linearGradient id="s" x2="0" y2="100%%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <clipPath id="r"><rect width="%d" height="20" rx="3" fill="#fff"/></clipPath>
  <g clip-path="url(#r)">
    <rect width="%d" height="20" fill="#555"/>
    <rect x="%d" width="%d" height="20" fill="%s"/>
    <rect width="%d" height="20" fill="url(#s)"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110">
    <text aria-hidden="true" x="%d" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="%d">%s</text>
    <text x="%d" y="140" transform="scale(.1)" fill="#fff" textLength="%d">%s</text>
    <text aria-hidden="true" x="%d" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="%d">%s</text>
    <text x="%d" y="140" transform="scale(.1)" fill="#fff" textLength="%d">%s</text>
  </g>
</svg>
`,
		totalWidth, label, message,
		label, message,
		totalWidth,
		labelWidth,
		labelWidth, msgWidth, color,
		totalWidth,
		labelX, labelTL, label,
		labelX, labelTL, label,
		msgX, msgTL, message,
		msgX, msgTL, message,
	)
}
