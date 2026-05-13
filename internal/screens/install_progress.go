package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"subiquity-ng/internal/ui"
)

type InstallProgress struct {
	state string
}

type InstallProgressStateMsg struct {
	State string
}

func NewInstallProgress() *InstallProgress {
	return &InstallProgress{
		state: "",
	}
}

func (s *InstallProgress) Title() string {
	return "Installation"
}

func (s *InstallProgress) Init() tea.Cmd {
	return nil
}

func (s *InstallProgress) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case InstallProgressStateMsg:
		s.state = msg.State
		return s, nil
	}
	return s, nil
}

func (s *InstallProgress) View(width, height int) string {
	contentWidth := ui.ConstrainedWidth(width)

	lines := make([]string, 0)
	lines = append(lines, "Installing system...")
	lines = append(lines, "")

	if s.state != "" {
		lines = append(lines, fmt.Sprintf("State: %s", s.state))
	} else {
		lines = append(lines, "Installing...")
	}
	lines = append(lines, "")

	content := strings.Join(lines, "\n")
	return langNormalStyle.Width(contentWidth).Render(content)
}
