package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestHeader_HasThreeRows(t *testing.T) {
	lines := strings.Split(Header(80, ConstrainedWidth(80), "Title"), "\n")
	assert.Len(t, lines, HeaderHeight)
}

func TestHeader_ContainsTitleAndHelp(t *testing.T) {
	h := Header(80, ConstrainedWidth(80), "Welcome!")
	assert.Contains(t, h, "Welcome!")
	assert.Contains(t, h, "[ Help ]")
}

func TestHeader_RowsAreWidthCellsWide(t *testing.T) {
	const width = 80
	for _, line := range strings.Split(Header(width, ConstrainedWidth(width), "Добро пожаловать!"), "\n") {
		assert.Equal(t, width, lipgloss.Width(line))
	}
}

func TestHeader_NarrowWidthDoesNotPanic(t *testing.T) {
	h := Header(5, ConstrainedWidth(5), "Title")
	assert.NotEmpty(t, h)
}

func TestBody_ProducesRequestedDimensions(t *testing.T) {
	const width, height = 40, 10
	lines := strings.Split(Body(width, height, "hello"), "\n")
	assert.Len(t, lines, height)
	for _, line := range lines {
		assert.Equal(t, width, lipgloss.Width(line))
	}
}

func TestBody_HeightClampsToOne(t *testing.T) {
	lines := strings.Split(Body(40, 0, "hello"), "\n")
	assert.Len(t, lines, 1)
}

func TestRender_TotalHeightMatches(t *testing.T) {
	const width, height = 40, 24
	lines := strings.Split(Render(width, height, "Welcome", "body"), "\n")
	assert.Len(t, lines, height)
}

func TestRender_HeaderAppearsBeforeBody(t *testing.T) {
	out := Render(80, 24, "T-Header", "B-Body")
	hi := strings.Index(out, "T-Header")
	bi := strings.Index(out, "B-Body")
	assert.GreaterOrEqual(t, hi, 0, "expected header in output")
	assert.GreaterOrEqual(t, bi, 0, "expected body in output")
	assert.Less(t, hi, bi, "title should appear before body content")
}

func TestHeader_WideScreenTitleCentered(t *testing.T) {
	const fullWidth = 150
	contentWidth := ConstrainedWidth(fullWidth)
	h := Header(fullWidth, contentWidth, "Title")
	lines := strings.Split(h, "\n")
	// All header lines should be full width
	for _, line := range lines {
		assert.Equal(t, fullWidth, lipgloss.Width(line))
	}
	// Title should be present in the header
	headerStr := strings.Join(lines, "\n")
	assert.Contains(t, headerStr, "Title")
}

func TestConstrainedWidth_NarrowScreenUnconstrained(t *testing.T) {
	assert.Equal(t, 80, ConstrainedWidth(80))
	assert.Equal(t, 100, ConstrainedWidth(100))
}

func TestConstrainedWidth_WideScreenConstrained(t *testing.T) {
	assert.Equal(t, MaxContentWidth, ConstrainedWidth(150))
	assert.Equal(t, MaxContentWidth, ConstrainedWidth(200))
}

func TestRender_NarrowScreenUnchanged(t *testing.T) {
	// On 80-width terminal, output should be 80 chars wide
	out := Render(80, 10, "Test", "Body")
	for _, line := range strings.Split(out, "\n") {
		assert.Equal(t, 80, lipgloss.Width(line))
	}
}

func TestRender_WideScreenHeaderFullBodyCentered(t *testing.T) {
	const totalWidth = 150
	out := Render(totalWidth, 10, "Test", "Body")
	lines := strings.Split(out, "\n")
	assert.Len(t, lines, 10)
	// All lines should be at totalWidth
	for _, line := range lines {
		assert.Equal(t, totalWidth, lipgloss.Width(line))
	}
	// Body lines after header should be centered (have leading space)
	if HeaderHeight < len(lines) {
		bodyLine := lines[HeaderHeight]
		assert.True(t, strings.HasPrefix(bodyLine, " "), "body should be centered with left padding")
	}
}
