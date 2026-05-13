package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func Render(width, height int, title, body string) string {
	contentWidth := ConstrainedWidth(width)
	// Header spans full width with title positioned in centered area
	header := Header(width, contentWidth, title)

	bodyContent := Body(contentWidth, height-HeaderHeight, body)

	// Center body on wide terminals
	if width > contentWidth {
		bodyContent = centerLines(bodyContent, width, contentWidth)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, bodyContent)
}

// Header renders a 3-line header with orange background spanning fullWidth.
// The title and help text are positioned within contentWidth and centered.
func Header(fullWidth, contentWidth int, title string) string {
	var above, below string
	if useHalfBlocks {
		above = HalfBlockStyle.Render(strings.Repeat(LowerHalfBlock, fullWidth))
		below = HalfBlockStyle.Render(strings.Repeat(UpperHalfBlock, fullWidth))
	} else {
		above = HeaderStyle.Width(fullWidth).Render("")
		below = above
	}

	const helpText = "[ Help ]"
	contentPad := contentWidth - lipgloss.Width(title) - lipgloss.Width(helpText) - 2
	if contentPad < 1 {
		contentPad = 1
	}
	contentStr := " " + title + strings.Repeat(" ", contentPad) + helpText + " "

	// Center content within full width
	leftMargin := (fullWidth - contentWidth) / 2
	rightMargin := fullWidth - leftMargin - contentWidth
	mid := HeaderStyle.Render(
		strings.Repeat(" ", leftMargin) +
			contentStr +
			strings.Repeat(" ", rightMargin),
	)

	return strings.Join([]string{above, mid, below}, "\n")
}

func Body(width, height int, content string) string {
	if height < 1 {
		height = 1
	}
	return lipgloss.NewStyle().Width(width).Height(height).Render(content)
}

// centerLines adds horizontal padding to center lines within the given total width.
// contentWidth is the expected width of the content lines (what they were rendered at).
func centerLines(s string, totalWidth, contentWidth int) string {
	padding := (totalWidth - contentWidth) / 2
	lines := strings.Split(s, "\n")

	for i, line := range lines {
		w := lipgloss.Width(line)
		rightPad := totalWidth - padding - w
		if rightPad < 0 {
			rightPad = 0
		}
		lines[i] = strings.Repeat(" ", padding) + line + strings.Repeat(" ", rightPad)
	}

	return strings.Join(lines, "\n")
}
