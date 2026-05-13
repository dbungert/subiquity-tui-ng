package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"

	"subiquity-ng/internal/screens"
	"subiquity-ng/internal/ui"
)

type Model struct {
	width, height int
	current       screens.Screen
}

func (m Model) Init() tea.Cmd { return m.current.Init() }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.current, cmd = m.current.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}
	body := m.current.View(m.width, m.height-ui.HeaderHeight)
	return ui.Render(m.width, m.height, m.current.Title(), body)
}

func main() {
	p := tea.NewProgram(
		Model{current: screens.NewLanguage()},
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
