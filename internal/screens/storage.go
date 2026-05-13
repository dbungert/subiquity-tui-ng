package screens

import (
	tea "github.com/charmbracelet/bubbletea"
)

type StorageScreen struct {
	rawJSON string
}

func NewStorage(rawJSON string) *StorageScreen {
	return &StorageScreen{rawJSON: rawJSON}
}

func (s *StorageScreen) Init() tea.Cmd {
	return nil
}

func (s *StorageScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	return s, nil
}

func (s *StorageScreen) Title() string {
	return "Storage Configuration"
}

func (s *StorageScreen) View(width, height int) string {
	if s.rawJSON == "" {
		return "Loading…"
	}
	return s.rawJSON
}
