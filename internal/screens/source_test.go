package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestSource_Title(t *testing.T) {
	s := NewSource(nil, "")
	assert.Equal(t, "Installation Source", s.Title())
}

func TestSource_ViewShowsNames(t *testing.T) {
	items := []SourceItem{
		{ID: "server", Name: "Ubuntu Server", Description: "Full install", Size: 2500000000},
		{ID: "minimal", Name: "Ubuntu Minimal", Description: "Minimal install", Size: 1500000000},
	}
	s := NewSource(items, "")
	view := s.View(80, 24)
	assert.Contains(t, view, "Ubuntu Server")
	assert.Contains(t, view, "Ubuntu Minimal")
}

func TestSource_DefaultsToCurrentID(t *testing.T) {
	items := []SourceItem{
		{ID: "server", Name: "Ubuntu Server", Description: "Full", Size: 2500000000},
		{ID: "minimal", Name: "Ubuntu Minimal", Description: "Minimal", Size: 1500000000},
	}
	s := NewSource(items, "minimal")
	assert.Equal(t, 1, s.cursor)
}

func TestSource_NavigateUpDown(t *testing.T) {
	items := []SourceItem{
		{ID: "a", Name: "Option A", Description: "A", Size: 1000},
		{ID: "b", Name: "Option B", Description: "B", Size: 2000},
	}
	s := NewSource(items, "a")
	assert.Equal(t, 0, s.cursor)

	next, _ := s.Update(tea.KeyMsg{Type: tea.KeyDown})
	s = next.(*SourceScreen)
	assert.Equal(t, 1, s.cursor)

	next, _ = s.Update(tea.KeyMsg{Type: tea.KeyUp})
	s = next.(*SourceScreen)
	assert.Equal(t, 0, s.cursor)
}

func TestSource_NavigateUpDownClamped(t *testing.T) {
	items := []SourceItem{{ID: "a", Name: "A", Description: "A", Size: 1000}}
	s := NewSource(items, "")

	// Can't go up from 0
	next, _ := s.Update(tea.KeyMsg{Type: tea.KeyUp})
	s = next.(*SourceScreen)
	assert.Equal(t, 0, s.cursor)

	// Can't go down past last
	next, _ = s.Update(tea.KeyMsg{Type: tea.KeyDown})
	s = next.(*SourceScreen)
	assert.Equal(t, 0, s.cursor)
}

func TestSource_EnterEmitsSelectedMsg(t *testing.T) {
	items := []SourceItem{
		{ID: "server", Name: "Server", Description: "Full", Size: 2500000000},
	}
	s := NewSource(items, "")
	_, cmd := s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotNil(t, cmd)

	msg := cmd()
	selectedMsg, ok := msg.(SourceSelectedMsg)
	assert.True(t, ok)
	assert.Equal(t, "server", selectedMsg.ID)
}

func TestSource_Init(t *testing.T) {
	s := NewSource(nil, "")
	assert.Nil(t, s.Init())
}

func TestSource_UpdateIgnoresNonKey(t *testing.T) {
	items := []SourceItem{{ID: "a", Name: "A", Description: "A", Size: 1000}}
	s := NewSource(items, "")

	next, cmd := s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	assert.Equal(t, s, next)
	assert.Nil(t, cmd)
}

func TestSource_ViewLoadingWhenEmpty(t *testing.T) {
	s := NewSource(nil, "")
	view := s.View(80, 24)
	assert.Contains(t, view, "Loading")
}

func TestSource_ViewShowsSize(t *testing.T) {
	items := []SourceItem{
		{ID: "a", Name: "A", Description: "A", Size: 2500000000},
	}
	s := NewSource(items, "")
	view := s.View(80, 24)
	assert.Contains(t, view, "2.5 GB")
}
