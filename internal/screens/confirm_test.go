package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfirm_Title(t *testing.T) {
	c := NewConfirm("disk-sda", "LVM_LUKS")
	assert.Equal(t, "Confirm Installation", c.Title())
}

func TestConfirm_DefaultCursorIsNo(t *testing.T) {
	c := NewConfirm("disk-sda", "DIRECT")
	assert.Equal(t, 1, c.cursor, "default cursor should be on No (index 1)")
}

func TestConfirm_EnterOnNoEmitsCancelMsg(t *testing.T) {
	c := NewConfirm("disk-sda", "DIRECT")
	// cursor is already 1 (No), so just press enter
	_, cmd := c.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(ConfirmCancelMsg)
	assert.True(t, ok, "expected ConfirmCancelMsg")
}

func TestConfirm_EnterOnYesEmitsAcceptedMsg(t *testing.T) {
	c := NewConfirm("disk-sda", "DIRECT")
	// Move cursor up to Yes (index 0)
	next, _ := c.Update(tea.KeyMsg{Type: tea.KeyUp})
	c = next.(*ConfirmScreen)
	assert.Equal(t, 0, c.cursor)

	// Press enter on Yes
	_, cmd := c.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(ConfirmAcceptedMsg)
	assert.True(t, ok, "expected ConfirmAcceptedMsg")
}

func TestConfirm_EscEmitsCancelMsg(t *testing.T) {
	c := NewConfirm("disk-sda", "DIRECT")
	_, cmd := c.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(ConfirmCancelMsg)
	assert.True(t, ok, "expected ConfirmCancelMsg on Esc")
}

func TestConfirm_ViewShowsDiskLabel(t *testing.T) {
	c := NewConfirm("my-test-disk", "DIRECT")
	view := c.View(80, 24)
	assert.Contains(t, view, "my-test-disk")
}

func TestConfirm_ViewShowsCapability(t *testing.T) {
	c := NewConfirm("disk-sda", "LVM_LUKS")
	view := c.View(80, 24)
	assert.Contains(t, view, "LVM + Encryption")
}

func TestConfirm_NavigateUpDown(t *testing.T) {
	c := NewConfirm("disk-sda", "DIRECT")
	assert.Equal(t, 1, c.cursor)

	// Move up to Yes
	next, _ := c.Update(tea.KeyMsg{Type: tea.KeyUp})
	c = next.(*ConfirmScreen)
	assert.Equal(t, 0, c.cursor)

	// Move down to No
	next, _ = c.Update(tea.KeyMsg{Type: tea.KeyDown})
	c = next.(*ConfirmScreen)
	assert.Equal(t, 1, c.cursor)
}

func TestConfirm_NavigateClamped(t *testing.T) {
	c := NewConfirm("disk-sda", "DIRECT")
	// At No, can't go down further
	next, _ := c.Update(tea.KeyMsg{Type: tea.KeyDown})
	c = next.(*ConfirmScreen)
	assert.Equal(t, 1, c.cursor)

	// Move to Yes
	next, _ = c.Update(tea.KeyMsg{Type: tea.KeyUp})
	c = next.(*ConfirmScreen)
	// At Yes, can't go up further
	next, _ = c.Update(tea.KeyMsg{Type: tea.KeyUp})
	c = next.(*ConfirmScreen)
	assert.Equal(t, 0, c.cursor)
}
