package screens

import tea "github.com/charmbracelet/bubbletea"

type Welcome struct{}

func NewWelcome() Welcome { return Welcome{} }

func (w Welcome) Init() tea.Cmd { return nil }

func (w Welcome) Update(tea.Msg) (Screen, tea.Cmd) { return w, nil }

func (w Welcome) Title() string {
	return "Willkommen! Bienvenue! Welcome! Добро пожаловать! Welkom!"
}

func (w Welcome) View(width, height int) string {
	return "Use UP, DOWN and ENTER keys to select your language."
}
