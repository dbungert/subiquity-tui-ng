package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"subiquity-ng/internal/ui"
)

const maxLogLines = 10

type InstallProgress struct {
	state    string
	logLines []string
}

type InstallProgressStateMsg struct {
	State string
}

type InstallLogLineMsg struct {
	Line string
}

func NewInstallProgress() *InstallProgress {
	return &InstallProgress{
		state:    "",
		logLines: make([]string, 0, maxLogLines),
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
	case InstallLogLineMsg:
		s.logLines = append(s.logLines, msg.Line)
		if len(s.logLines) > maxLogLines {
			s.logLines = s.logLines[len(s.logLines)-maxLogLines:]
		}
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

	for _, l := range s.logLines {
		lines = append(lines, truncateLine(l, contentWidth-2))
	}
	if len(s.logLines) > 0 {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	return langNormalStyle.Width(contentWidth).Render(content)
}

func truncateLine(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= maxRunes {
		return s
	}
	return string(r[:maxRunes-1]) + "…"
}
