package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func Render(width, height int, title, body string) string {
	contentWidth := ConstrainedWidth(width)
	header := Header(contentWidth, title)
	bodyContent := Body(contentWidth, height-HeaderHeight, body)

	full := lipgloss.JoinVertical(lipgloss.Left, header, bodyContent)

	// Center on wide terminals
	if width > contentWidth {
		full = centerLines(full, width, contentWidth)
	}

	return full
}

func Header(width int, title string) string {
	var above, below string
	if useHalfBlocks {
		above = HalfBlockStyle.Render(strings.Repeat(LowerHalfBlock, width))
		below = HalfBlockStyle.Render(strings.Repeat(UpperHalfBlock, width))
	} else {
		above = HeaderStyle.Width(width).Render("")
		below = above
	}

	const helpText = "[ Help ]"
	pad := width - lipgloss.Width(title) - lipgloss.Width(helpText) - 2
	if pad < 1 {
		pad = 1
	}
	mid := HeaderStyle.Render(" " + title + strings.Repeat(" ", pad) + helpText + " ")

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
