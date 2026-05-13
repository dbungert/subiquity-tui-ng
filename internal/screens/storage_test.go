package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestStorage_Title(t *testing.T) {
	s := NewStorage("")
	assert.Equal(t, "Storage Configuration", s.Title())
}

func TestStorage_ViewLoadingWhenEmpty(t *testing.T) {
	s := NewStorage("")
	view := s.View(80, 24)
	assert.Contains(t, view, "Loading")
}

func TestStorage_ViewShowsJSON(t *testing.T) {
	jsonData := `{"status":"DONE","targets":[]}`
	s := NewStorage(jsonData)
	view := s.View(80, 24)
	assert.Contains(t, view, jsonData)
}

func TestStorage_Init(t *testing.T) {
	s := NewStorage("")
	assert.Nil(t, s.Init())
}

func TestStorage_UpdateIgnoresMsg(t *testing.T) {
	s := NewStorage(`{"test":"data"}`)
	next, cmd := s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	assert.Equal(t, s, next)
	assert.Nil(t, cmd)
}
