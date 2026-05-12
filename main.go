package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// 1. THE MODEL: This stores your application state.
type model struct {
	choices  []string         // Items in our list
	cursor   int              // Which item our cursor is pointing at
	selected map[int]struct{} // Which items are selected
}

func initialModel() model {
	return model{
		choices:  []string{"Buy Milk", "Drink Coffee", "Build TUI"},
		selected: make(map[int]struct{}),
	}
}

// 2. THE INIT: This returns an initial command to run (like a socket connection).
func (m model) Init() tea.Cmd {
	return nil // No I/O for now
}

// 3. THE UPDATE: This handles all events (key presses, socket data, etc.)
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:
		switch msg.String() {

		// Standard "Quit" keys
		case "ctrl+c", "q":
			return m, tea.Quit

		// Move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// Move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// Toggle selection
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	// Return the updated model to the framework
	return m, nil
}

// 4. THE VIEW: This renders the UI as a string every time the model changes.
func (m model) View() string {
	s := "What should we do today?\n\n"

	for i, choice := range m.choices {
		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected!
		}

		// Render the row
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	s += "\nPress q to quit.\n"

	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
