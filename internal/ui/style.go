package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// ConstrainedWidth returns the width to use for content, clamped to MaxContentWidth.
func ConstrainedWidth(w int) int {
	if w > MaxContentWidth {
		return MaxContentWidth
	}
	return w
}

const (
	HeaderHeight    = 3
	MaxContentWidth = 120 // content narrower than this on wide terminals
)

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
