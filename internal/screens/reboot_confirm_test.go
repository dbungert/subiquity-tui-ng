package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRebootScreen_Title(t *testing.T) {
	r := NewRebootScreen("DONE")
	assert.Equal(t, "Reboot System", r.Title())
}

func TestRebootScreen_ViewShowsFinalState(t *testing.T) {
	r := NewRebootScreen("DONE")
	view := r.View(80, 24)
	assert.Contains(t, view, "Installation complete!")

	r = NewRebootScreen("ERROR")
	view = r.View(80, 24)
	assert.Contains(t, view, "ERROR")
}

func TestRebootScreen_UpdateConfirmEmitsMsg(t *testing.T) {
	r := NewRebootScreen("DONE")
	_, cmd := r.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(RebootConfirmMsg)
	assert.True(t, ok)
}

func TestRebootScreen_UpdateStayEmitsCancelMsg(t *testing.T) {
	r := NewRebootScreen("DONE")
	next, _ := r.Update(tea.KeyMsg{Type: tea.KeyDown})
	r = next.(*RebootScreen)
	assert.Equal(t, 1, r.cursor)

	_, cmd := r.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(RebootCancelMsg)
	assert.True(t, ok)
}

func TestRebootScreen_UpdateEscEmitsCancelMsg(t *testing.T) {
	r := NewRebootScreen("DONE")
	_, cmd := r.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(RebootCancelMsg)
	assert.True(t, ok)
}

func TestRebootScreen_CursorNavigates(t *testing.T) {
	r := NewRebootScreen("DONE")
	assert.Equal(t, 0, r.cursor)

	next, _ := r.Update(tea.KeyMsg{Type: tea.KeyDown})
	r = next.(*RebootScreen)
	assert.Equal(t, 1, r.cursor)

	next, _ = r.Update(tea.KeyMsg{Type: tea.KeyDown})
	r = next.(*RebootScreen)
	assert.Equal(t, 1, r.cursor)

	next, _ = r.Update(tea.KeyMsg{Type: tea.KeyUp})
	r = next.(*RebootScreen)
	assert.Equal(t, 0, r.cursor)

	next, _ = r.Update(tea.KeyMsg{Type: tea.KeyUp})
	r = next.(*RebootScreen)
	assert.Equal(t, 0, r.cursor)
}
