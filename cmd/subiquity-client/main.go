package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/alexflint/go-arg"
	tea "github.com/charmbracelet/bubbletea"

	"subiquity-ng/internal/screens"
	"subiquity-ng/internal/ui"
)

type Args struct {
	Socket string `arg:"--socket" help:"Unix socket for subiquity server communication"`
}

type Model struct {
	width, height int
	current       screens.Screen
	socket        string
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
	contentWidth := ui.ConstrainedWidth(m.width)
	body := m.current.View(contentWidth, m.height-ui.HeaderHeight)
	return ui.Render(m.width, m.height, m.current.Title(), body)
}

func main() {
	var args Args
	arg.MustParse(&args)

	if args.Socket == "" {
		prodSocket := "/run/subiquity/socket"
		if _, err := os.Stat(prodSocket); err == nil {
			args.Socket = prodSocket
		} else {
			args.Socket = filepath.Join(".subiquity", "socket")
		}
	}

	p := tea.NewProgram(
		Model{current: screens.NewLanguage(), socket: args.Socket},
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
