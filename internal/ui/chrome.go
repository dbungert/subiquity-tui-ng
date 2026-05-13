package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func Render(width, height int, title, body string) string {
	return lipgloss.JoinVertical(lipgloss.Left,
		Header(width, title),
		Body(width, height-HeaderHeight, body),
	)
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
