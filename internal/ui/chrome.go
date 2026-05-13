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
	blank := HeaderStyle.Width(width).Render("")

	const helpText = "[ Help ]"
	pad := width - lipgloss.Width(title) - lipgloss.Width(helpText) - 2
	if pad < 1 {
		pad = 1
	}
	line := HeaderStyle.Render(" " + title + strings.Repeat(" ", pad) + helpText + " ")

	return strings.Join([]string{blank, line, blank}, "\n")
}

func Body(width, height int, content string) string {
	if height < 1 {
		height = 1
	}
	return lipgloss.NewStyle().Width(width).Height(height).Render(content)
}
