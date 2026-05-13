package ui

import "github.com/charmbracelet/lipgloss"

const HeaderHeight = 3

const (
	UpperHalfBlock = "▀"
	LowerHalfBlock = "▄"
)

var (
	UbuntuOrange = lipgloss.Color("#E95420")
	HeaderFg     = lipgloss.Color("#FFFFFF")

	HeaderStyle    = lipgloss.NewStyle().Background(UbuntuOrange).Foreground(HeaderFg)
	HalfBlockStyle = lipgloss.NewStyle().Foreground(UbuntuOrange)
)
