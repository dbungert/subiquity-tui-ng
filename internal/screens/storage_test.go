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

func TestStorage_ViewLoadingShowsSpinner(t *testing.T) {
	s := NewStorage("")
	view := s.View(80, 24)
	assert.Contains(t, view, "Waiting for block device probing")
	// Verify it contains one of the spinner frames
	hasSpinner := false
	for _, frame := range spinnerFrames {
		if contains(view, frame) {
			hasSpinner = true
			break
		}
	}
	assert.True(t, hasSpinner, "should display a spinner frame")
}

func TestStorage_ViewLoadedShowsJSON(t *testing.T) {
	jsonData := `{"status":"DONE","targets":[]}`
	s := NewStorage(jsonData)
	view := s.View(80, 24)
	assert.Contains(t, view, jsonData)
}

func TestStorage_InitStartsTickWhenLoading(t *testing.T) {
	s := NewStorage("")
	cmd := s.Init()
	assert.NotNil(t, cmd, "loading screen should start spinner tick")
}

func TestStorage_InitNoTickWhenLoaded(t *testing.T) {
	s := NewStorage(`{"status":"DONE"}`)
	cmd := s.Init()
	assert.Nil(t, cmd, "loaded screen should not start spinner")
}

func TestStorage_UpdateAdvancesFrame(t *testing.T) {
	s := NewStorage("")
	assert.Equal(t, 0, s.frame)
	next, cmd := s.Update(storageSpinnerTickMsg{})
	screen := next.(*StorageScreen)
	assert.Equal(t, 1, screen.frame)
	assert.NotNil(t, cmd, "should return tick command to continue spinning")
}

func TestStorage_UpdateStopsTickWhenLoaded(t *testing.T) {
	s := NewStorage(`{"status":"DONE"}`)
	next, cmd := s.Update(storageSpinnerTickMsg{})
	assert.Equal(t, s, next)
	assert.Nil(t, cmd, "loaded screen should ignore tick and not return cmd")
}

func TestStorage_UpdateIgnoresNonSpinnerMsg(t *testing.T) {
	s := NewStorage("")
	next, cmd := s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	assert.Equal(t, s, next)
	assert.Nil(t, cmd, "non-spinner messages should be ignored")
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
