package screens

import tea "github.com/charmbracelet/bubbletea"

type InstallProgress struct{}

func NewInstallProgress() *InstallProgress {
	return &InstallProgress{}
}

func (s *InstallProgress) Title() string {
	return "Installation"
}

func (s *InstallProgress) Init() tea.Cmd {
	return nil
}

func (s *InstallProgress) Update(tea.Msg) (Screen, tea.Cmd) {
	return s, nil
}

func (s *InstallProgress) View(width, height int) string {
	return "Installing..."
}
