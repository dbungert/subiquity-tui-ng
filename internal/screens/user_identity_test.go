package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserIdentity_Title(t *testing.T) {
	u := NewUserIdentity()
	assert.Equal(t, "User Identity", u.Title())
}

func TestUserIdentity_ViewRendersAllFields(t *testing.T) {
	u := NewUserIdentity()
	view := u.View(80, 24)
	assert.Contains(t, view, "Your name:")
	assert.Contains(t, view, "Username:")
	assert.Contains(t, view, "Password:")
}

func TestUserIdentity_UpdateAppendsRunesToFocusedField(t *testing.T) {
	u := NewUserIdentity()
	assert.Equal(t, "", u.inputs[0])

	next, _ := u.Update(tea.KeyMsg{Runes: []rune{'J', 'o', 'h', 'n'}})
	u = next.(*UserIdentityScreen)
	assert.Equal(t, "John", u.inputs[0])
	assert.Equal(t, "", u.inputs[1])
	assert.Equal(t, "", u.inputs[2])
}

func TestUserIdentity_UpdateDownNavigatesFields(t *testing.T) {
	u := NewUserIdentity()
	assert.Equal(t, 0, u.focused)

	next, _ := u.Update(tea.KeyMsg{Type: tea.KeyDown})
	u = next.(*UserIdentityScreen)
	assert.Equal(t, 1, u.focused)

	next, _ = u.Update(tea.KeyMsg{Type: tea.KeyDown})
	u = next.(*UserIdentityScreen)
	assert.Equal(t, 2, u.focused)

	next, _ = u.Update(tea.KeyMsg{Type: tea.KeyDown})
	u = next.(*UserIdentityScreen)
	assert.Equal(t, 2, u.focused)
}

func TestUserIdentity_UpdateUpNavigatesFields(t *testing.T) {
	u := NewUserIdentity()
	u.focused = 2

	next, _ := u.Update(tea.KeyMsg{Type: tea.KeyUp})
	u = next.(*UserIdentityScreen)
	assert.Equal(t, 1, u.focused)

	next, _ = u.Update(tea.KeyMsg{Type: tea.KeyUp})
	u = next.(*UserIdentityScreen)
	assert.Equal(t, 0, u.focused)

	next, _ = u.Update(tea.KeyMsg{Type: tea.KeyUp})
	u = next.(*UserIdentityScreen)
	assert.Equal(t, 0, u.focused)
}

func TestUserIdentity_UpdateBackspace(t *testing.T) {
	u := NewUserIdentity()
	next, _ := u.Update(tea.KeyMsg{Runes: []rune{'a', 'b', 'c'}})
	u = next.(*UserIdentityScreen)
	assert.Equal(t, "abc", u.inputs[0])

	next, _ = u.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	u = next.(*UserIdentityScreen)
	assert.Equal(t, "ab", u.inputs[0])
}

func TestUserIdentity_UpdateEnterEmitsMsg(t *testing.T) {
	u := NewUserIdentity()
	next, _ := u.Update(tea.KeyMsg{Runes: []rune{'J', 'o', 'h', 'n'}})
	u = next.(*UserIdentityScreen)

	next, _ = u.Update(tea.KeyMsg{Type: tea.KeyDown})
	u = next.(*UserIdentityScreen)

	next, _ = u.Update(tea.KeyMsg{Runes: []rune{'j', 'd', 'o', 'e'}})
	u = next.(*UserIdentityScreen)

	next, _ = u.Update(tea.KeyMsg{Type: tea.KeyDown})
	u = next.(*UserIdentityScreen)

	next, _ = u.Update(tea.KeyMsg{Runes: []rune{'s', 'e', 'c', 'r', 'e', 't'}})
	u = next.(*UserIdentityScreen)

	_, cmd := u.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg := cmd()
	idMsg, ok := msg.(UserIdentityDoneMsg)
	assert.True(t, ok)
	assert.Equal(t, "John", idMsg.Realname)
	assert.Equal(t, "jdoe", idMsg.Username)
	assert.Equal(t, "secret", idMsg.Password)
}

func TestUserIdentity_UpdateEnterWithEmptyShowsWarning(t *testing.T) {
	u := NewUserIdentity()
	assert.False(t, u.showEmpty)

	next, cmd := u.Update(tea.KeyMsg{Type: tea.KeyEnter})
	u = next.(*UserIdentityScreen)
	assert.Nil(t, cmd)
	assert.True(t, u.showEmpty)

	view := u.View(80, 24)
	assert.Contains(t, view, "required")
}

func TestUserIdentity_ViewMasksPassword(t *testing.T) {
	u := NewUserIdentity()
	u.inputs[2] = "secret"
	u.focused = 2

	view := u.View(80, 24)
	assert.NotContains(t, view, "secret")
	assert.Contains(t, view, "●●●●●●")
}

func TestUserIdentity_UpdateClearsEmptyWarningOnInput(t *testing.T) {
	u := NewUserIdentity()
	next, _ := u.Update(tea.KeyMsg{Type: tea.KeyEnter})
	u = next.(*UserIdentityScreen)
	assert.True(t, u.showEmpty)

	next, _ = u.Update(tea.KeyMsg{Runes: []rune{'a'}})
	u = next.(*UserIdentityScreen)
	assert.False(t, u.showEmpty)
}
