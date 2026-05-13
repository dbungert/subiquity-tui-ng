package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"subiquity-ng/internal/ui"
)

type RebootConfirmMsg struct{}
type RebootCancelMsg struct{}

type RebootScreen struct {
	finalState string
	cursor     int // 0 = Reboot, 1 = Stay (default is 0 for reboot)
}

func NewRebootScreen(finalState string) *RebootScreen {
	return &RebootScreen{
		finalState: finalState,
		cursor:     0, // Default to Reboot now
	}
}

func (r *RebootScreen) Title() string {
	return "Reboot System"
}

func (r *RebootScreen) Init() tea.Cmd {
	return nil
}

func (r *RebootScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if r.cursor > 0 {
				r.cursor--
			}
		case "down":
			if r.cursor < 1 {
				r.cursor++
			}
		case "enter":
			if r.cursor == 0 {
				return r, func() tea.Msg {
					return RebootConfirmMsg{}
				}
			} else {
				return r, func() tea.Msg {
					return RebootCancelMsg{}
				}
			}
		case "esc":
			return r, func() tea.Msg {
				return RebootCancelMsg{}
			}
		}
	}
	return r, nil
}

func (r *RebootScreen) View(width, height int) string {
	contentWidth := ui.ConstrainedWidth(width)
	lines := make([]string, 0)

	if r.finalState == "DONE" {
		lines = append(lines, "Installation complete!")
		lines = append(lines, "The system is ready to use. Reboot to start using your new system.")
	} else {
		lines = append(lines, fmt.Sprintf("Installation finished with state: %s", r.finalState))
	}
	lines = append(lines, "")

	options := []string{"Reboot now", "Stay here"}
	for i, option := range options {
		if i == r.cursor {
			display := langSelectedStyle.Width(contentWidth).Render("▶ " + option)
			lines = append(lines, display)
		} else {
			display := langNormalStyle.Width(contentWidth).Render("  " + option)
			lines = append(lines, display)
		}
	}

	lines = append(lines, "")
	lines = append(lines, "↑↓ Navigate   Enter Select")

	return strings.Join(lines, "\n")
}
