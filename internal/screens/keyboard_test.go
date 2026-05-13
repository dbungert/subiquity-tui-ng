package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestKeyboard_Title(t *testing.T) {
	assert.Equal(t, "Keyboard Layout", NewKeyboard().Title())
}

func TestKeyboard_View(t *testing.T) {
	got := NewKeyboard().View(80, 24)
	assert.Contains(t, got, "Keyboard layout selection not yet implemented")
}

func TestKeyboard_Init(t *testing.T) {
	assert.Nil(t, NewKeyboard().Init())
}

func TestKeyboard_UpdateIsIdentity(t *testing.T) {
	k := NewKeyboard()
	next, cmd := k.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	assert.Equal(t, k, next)
	assert.Nil(t, cmd)
}
