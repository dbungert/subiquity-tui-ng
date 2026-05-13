package screens

import tea "github.com/charmbracelet/bubbletea"

type Screen interface {
	Init() tea.Cmd
	Update(tea.Msg) (Screen, tea.Cmd)
	View(width, height int) string
	Title() string
}
