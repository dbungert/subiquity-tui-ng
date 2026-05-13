package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestHeader_HasThreeRows(t *testing.T) {
	lines := strings.Split(Header(80, "Title"), "\n")
	if len(lines) != HeaderHeight {
		t.Errorf("want %d lines, got %d", HeaderHeight, len(lines))
	}
}

func TestHeader_ContainsTitleAndHelp(t *testing.T) {
	h := Header(80, "Welcome!")
	if !strings.Contains(h, "Welcome!") {
		t.Errorf("header missing title: %q", h)
	}
	if !strings.Contains(h, "[ Help ]") {
		t.Errorf("header missing [ Help ]: %q", h)
	}
}

func TestHeader_RowsAreWidthCellsWide(t *testing.T) {
	const width = 80
	for _, line := range strings.Split(Header(width, "Добро пожаловать!"), "\n") {
		if got := lipgloss.Width(line); got != width {
			t.Errorf("row width %d != %d for %q", got, width, line)
		}
	}
}

func TestHeader_NarrowWidthDoesNotPanic(t *testing.T) {
	h := Header(5, "Title")
	if h == "" {
		t.Errorf("expected non-empty header for narrow width")
	}
}

func TestBody_ProducesRequestedDimensions(t *testing.T) {
	const width, height = 40, 10
	lines := strings.Split(Body(width, height, "hello"), "\n")
	if len(lines) != height {
		t.Errorf("want %d lines, got %d", height, len(lines))
	}
	for _, line := range lines {
		if got := lipgloss.Width(line); got != width {
			t.Errorf("line width %d != %d for %q", got, width, line)
		}
	}
}

func TestBody_HeightClampsToOne(t *testing.T) {
	lines := strings.Split(Body(40, 0, "hello"), "\n")
	if len(lines) != 1 {
		t.Errorf("want height clamped to 1, got %d lines", len(lines))
	}
}

func TestRender_TotalHeightMatches(t *testing.T) {
	const width, height = 40, 24
	lines := strings.Split(Render(width, height, "Welcome", "body"), "\n")
	if len(lines) != height {
		t.Errorf("want %d total lines, got %d", height, len(lines))
	}
}

func TestRender_HeaderAppearsBeforeBody(t *testing.T) {
	out := Render(80, 24, "T-Header", "B-Body")
	hi := strings.Index(out, "T-Header")
	bi := strings.Index(out, "B-Body")
	if hi < 0 || bi < 0 {
		t.Fatalf("expected both strings in output: %q", out)
	}
	if hi > bi {
		t.Errorf("title should appear before body content")
	}
}
