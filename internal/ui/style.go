package ui

import "github.com/charmbracelet/lipgloss"

const HeaderHeight = 3

var (
	UbuntuOrange = lipgloss.Color("#E95420")
	HeaderFg     = lipgloss.Color("#FFFFFF")

	HeaderStyle = lipgloss.NewStyle().Background(UbuntuOrange).Foreground(HeaderFg)
)
