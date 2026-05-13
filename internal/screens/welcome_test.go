package screens

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWelcome_TitleIsMultilingualGreeting(t *testing.T) {
	got := NewWelcome().Title()
	for _, want := range []string{"Willkommen!", "Bienvenue!", "Welcome!", "Добро пожаловать!", "Welkom!"} {
		assert.Contains(t, got, want)
	}
}

func TestWelcome_ViewMentionsKeys(t *testing.T) {
	got := NewWelcome().View(80, 20)
	assert.Contains(t, got, "UP")
	assert.Contains(t, got, "DOWN")
	assert.Contains(t, got, "ENTER")
}

func TestWelcome_InitNoCmd(t *testing.T) {
	assert.Nil(t, NewWelcome().Init())
}

func TestWelcome_UpdateIsIdentity(t *testing.T) {
	w := NewWelcome()
	next, cmd := w.Update(nil)
	assert.Equal(t, w, next.(Welcome))
	assert.Nil(t, cmd)
}
