package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBadgeColor(t *testing.T) {
	tests := []struct {
		pct  float64
		want string
	}{
		{100, "#4c1"},
		{80, "#4c1"},
		{79.9, "#dfb317"},
		{60, "#dfb317"},
		{59.9, "#fe7d37"},
		{40, "#fe7d37"},
		{39.9, "#e05d44"},
		{0, "#e05d44"},
	}

	for _, tc := range tests {
		got := BadgeColor(tc.pct)
		if got != tc.want {
			t.Errorf("BadgeColor(%v) = %q, want %q", tc.pct, got, tc.want)
		}
	}
}

func TestGenerateBadge_ContainsExpectedElements(t *testing.T) {
	svg := GenerateBadge(82.5)

	checks := []string{
		`aria-label="coverage: 82.5%"`,
		`<title>coverage: 82.5%</title>`,
		`fill="#4c1"`, // green for >= 80
		`82.5%`,
	}

	for _, c := range checks {
		if !strings.Contains(svg, c) {
			t.Errorf("GenerateBadge(82.5) missing expected substring %q", c)
		}
	}
}

func TestGenerateBadge_WholeNumber(t *testing.T) {
	svg := GenerateBadge(80)

	// Should render "80%" not "80.0%"
	if strings.Contains(svg, "80.0%") {
		t.Error("GenerateBadge(80) should render '80%%' not '80.0%%'")
	}
	if !strings.Contains(svg, "80%") {
		t.Error("GenerateBadge(80) should contain '80%%'")
	}
}

func TestGenerateBadge_LowCoverage(t *testing.T) {
	svg := GenerateBadge(15.3)

	if !strings.Contains(svg, `fill="#e05d44"`) {
		t.Error("GenerateBadge(15.3) should use red color")
	}
	if !strings.Contains(svg, "15.3%") {
		t.Error("GenerateBadge(15.3) should contain '15.3%%'")
	}
}

func TestGenerateBadge_ValidSVG(t *testing.T) {
	svg := GenerateBadge(50)

	if !strings.HasPrefix(svg, "<svg") {
		t.Error("badge should start with <svg")
	}
	if !strings.HasSuffix(strings.TrimSpace(svg), "</svg>") {
		t.Error("badge should end with </svg>")
	}
}

func TestGenerateBadge_WritesToFile(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "badge.svg")

	svg := GenerateBadge(75.0)
	err := os.WriteFile(out, []byte(svg), 0o644)
	if err != nil {
		t.Fatalf("failed to write badge: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("failed to read badge: %v", err)
	}

	if string(data) != svg {
		t.Error("written file content does not match generated SVG")
	}
}

func TestGenerateBadge_Widths(t *testing.T) {
	// "coverage" = 8 chars -> 8*6.5+10 = 62
	// "82.5%" = 5 chars -> 5*8.5+10 = 52 (truncated from 52.5)
	// total = 114
	svg := GenerateBadge(82.5)

	if !strings.Contains(svg, `width="114"`) {
		t.Errorf("expected total width 114 in SVG, got:\n%s", svg[:200])
	}
}
