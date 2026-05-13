package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostIdentity_Title(t *testing.T) {
	h := NewHostIdentity()
	assert.Equal(t, "System Hostname", h.Title())
}

func TestHostIdentity_ViewRendersField(t *testing.T) {
	h := NewHostIdentity()
	view := h.View(80, 24)
	assert.Contains(t, view, "Server name:")
	assert.Contains(t, view, "network")
}

func TestHostIdentity_UpdateAppendsRunes(t *testing.T) {
	h := NewHostIdentity()
	assert.Equal(t, "", h.input)

	next, _ := h.Update(tea.KeyMsg{Runes: []rune{'m', 'y', 'h', 'o', 's', 't'}})
	h = next.(*HostIdentityScreen)
	assert.Equal(t, "myhost", h.input)
}

func TestHostIdentity_UpdateBackspace(t *testing.T) {
	h := NewHostIdentity()
	next, _ := h.Update(tea.KeyMsg{Runes: []rune{'a', 'b', 'c'}})
	h = next.(*HostIdentityScreen)
	assert.Equal(t, "abc", h.input)

	next, _ = h.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	h = next.(*HostIdentityScreen)
	assert.Equal(t, "ab", h.input)
}

func TestHostIdentity_UpdateEnterEmitsMsg(t *testing.T) {
	h := NewHostIdentity()
	next, _ := h.Update(tea.KeyMsg{Runes: []rune{'m', 'y', 'h', 'o', 's', 't'}})
	h = next.(*HostIdentityScreen)

	_, cmd := h.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg := cmd()
	hostMsg, ok := msg.(HostIdentityDoneMsg)
	assert.True(t, ok)
	assert.Equal(t, "myhost", hostMsg.Hostname)
}

func TestHostIdentity_UpdateEnterWithEmptyShowsWarning(t *testing.T) {
	h := NewHostIdentity()
	assert.False(t, h.showEmpty)

	next, cmd := h.Update(tea.KeyMsg{Type: tea.KeyEnter})
	h = next.(*HostIdentityScreen)
	assert.Nil(t, cmd)
	assert.True(t, h.showEmpty)

	view := h.View(80, 24)
	assert.Contains(t, view, "cannot be empty")
}

func TestHostIdentity_UpdateClearsEmptyWarningOnInput(t *testing.T) {
	h := NewHostIdentity()
	next, _ := h.Update(tea.KeyMsg{Type: tea.KeyEnter})
	h = next.(*HostIdentityScreen)
	assert.True(t, h.showEmpty)

	next, _ = h.Update(tea.KeyMsg{Runes: []rune{'a'}})
	h = next.(*HostIdentityScreen)
	assert.False(t, h.showEmpty)
}
