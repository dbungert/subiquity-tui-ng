package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"subiquity-ng/internal/ui"
)

type HostIdentityScreen struct {
	input     string
	showEmpty bool
}

type HostIdentityDoneMsg struct {
	Hostname string
}

func NewHostIdentity() *HostIdentityScreen {
	return &HostIdentityScreen{
		input:     "",
		showEmpty: false,
	}
}

func (h *HostIdentityScreen) Title() string {
	return "System Hostname"
}

func (h *HostIdentityScreen) Init() tea.Cmd {
	return nil
}

func (h *HostIdentityScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if h.input == "" {
				h.showEmpty = true
				return h, nil
			}
			return h, func() tea.Msg {
				return HostIdentityDoneMsg{Hostname: h.input}
			}
		case tea.KeyBackspace, tea.KeyCtrlH:
			if len(h.input) > 0 {
				h.input = h.input[:len(h.input)-1]
			}
			h.showEmpty = false
			return h, nil
		default:
			if len(msg.Runes) > 0 {
				h.input += string(msg.Runes)
				h.showEmpty = false
			}
			return h, nil
		}
	}
	return h, nil
}

func (h *HostIdentityScreen) View(width, height int) string {
	contentWidth := ui.ConstrainedWidth(width)

	lines := make([]string, 0)
	lines = append(lines, "Enter a name for this computer.")
	lines = append(lines, "It will be used to identify it on the network.")
	lines = append(lines, "")

	hostnameDisplay := fmt.Sprintf("  Server name:  %s", h.input)
	lines = append(lines, langNormalStyle.Width(contentWidth).Render(hostnameDisplay))
	lines = append(lines, "")

	if h.showEmpty {
		lines = append(lines, langHintStyle.Render("Server name cannot be empty."))
		lines = append(lines, "")
	}

	lines = append(lines, "Enter Confirm")

	return strings.Join(lines, "\n")
}
