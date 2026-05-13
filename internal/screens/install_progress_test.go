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

func TestInstallProgress_View(t *testing.T) {
	s := NewInstallProgress()
	view := s.View(80, 24)
	assert.Contains(t, view, "Installing...")
}

func TestInstallProgress_Init(t *testing.T) {
	s := NewInstallProgress()
	cmd := s.Init()
	assert.Nil(t, cmd)
}

func TestInstallProgress_UpdateIsIdentity(t *testing.T) {
	s := NewInstallProgress()
	next, cmd := s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	assert.Equal(t, s, next)
	assert.Nil(t, cmd)
}
