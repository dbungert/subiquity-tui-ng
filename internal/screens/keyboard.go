package screens

import tea "github.com/charmbracelet/bubbletea"

type Keyboard struct{}

func NewKeyboard() *Keyboard {
	return &Keyboard{}
}

func (k *Keyboard) Init() tea.Cmd {
	return nil
}

func (k *Keyboard) Update(tea.Msg) (Screen, tea.Cmd) {
	return k, nil
}

func (k *Keyboard) Title() string {
	return "Keyboard Layout"
}

func (k *Keyboard) View(width, height int) string {
	return "Keyboard layout selection not yet implemented."
}
