package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestInstallProgress_Title(t *testing.T) {
	s := NewInstallProgress()
	assert.Equal(t, "Installation", s.Title())
}

func TestInstallProgress_ViewDefaultText(t *testing.T) {
	s := NewInstallProgress()
	view := s.View(80, 24)
	assert.Contains(t, view, "Installing...")
}

func TestInstallProgress_ViewShowsState(t *testing.T) {
	s := NewInstallProgress()
	next, _ := s.Update(InstallProgressStateMsg{State: "RUNNING"})
	s = next.(*InstallProgress)

	view := s.View(80, 24)
	assert.Contains(t, view, "RUNNING")
	assert.NotContains(t, view, "Installing...")
}

func TestInstallProgress_UpdateHandlesStateMsg(t *testing.T) {
	s := NewInstallProgress()
	assert.Equal(t, "", s.state)

	next, cmd := s.Update(InstallProgressStateMsg{State: "RUNNING"})
	s = next.(*InstallProgress)
	assert.Equal(t, "RUNNING", s.state)
	assert.Nil(t, cmd)
}

func TestInstallProgress_UpdateIgnoresOtherMessages(t *testing.T) {
	s := NewInstallProgress()
	next, cmd := s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	s = next.(*InstallProgress)
	assert.Equal(t, "", s.state)
	assert.Nil(t, cmd)
}
