package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPassphrase_Title(t *testing.T) {
	p := NewPassphrase("disk-sda", "LVM_LUKS")
	assert.Equal(t, "Encryption Passphrase", p.Title())
}

func TestPassphrase_ViewShowsDiskAndCapability(t *testing.T) {
	p := NewPassphrase("disk-sda", "LVM_LUKS")
	view := p.View(80, 24)
	assert.Contains(t, view, "disk-sda")
	assert.Contains(t, view, "LVM + Encryption 🔒")
}

func TestPassphrase_UpdateAppendsRunes(t *testing.T) {
	p := NewPassphrase("disk-sda", "LVM_LUKS")
	assert.Equal(t, "", p.input)

	next, _ := p.Update(tea.KeyMsg{Runes: []rune{'a'}})
	p = next.(*PassphraseScreen)
	assert.Equal(t, "a", p.input)

	next, _ = p.Update(tea.KeyMsg{Runes: []rune{'b', 'c'}})
	p = next.(*PassphraseScreen)
	assert.Equal(t, "abc", p.input)
}

func TestPassphrase_UpdateBackspace(t *testing.T) {
	p := NewPassphrase("disk-sda", "LVM_LUKS")
	next, _ := p.Update(tea.KeyMsg{Runes: []rune{'a', 'b', 'c'}})
	p = next.(*PassphraseScreen)
	assert.Equal(t, "abc", p.input)

	next, _ = p.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	p = next.(*PassphraseScreen)
	assert.Equal(t, "ab", p.input)

	next, _ = p.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	p = next.(*PassphraseScreen)
	assert.Equal(t, "a", p.input)

	next, _ = p.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	p = next.(*PassphraseScreen)
	assert.Equal(t, "", p.input)

	next, _ = p.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	p = next.(*PassphraseScreen)
	assert.Equal(t, "", p.input)
}

func TestPassphrase_UpdateEnterEmitsMsg(t *testing.T) {
	p := NewPassphrase("disk-sda", "LVM_LUKS")
	next, _ := p.Update(tea.KeyMsg{Runes: []rune{'t', 'e', 's', 't'}})
	p = next.(*PassphraseScreen)

	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg := cmd()
	passphraseMsg, ok := msg.(PassphraseEnteredMsg)
	assert.True(t, ok)
	assert.Equal(t, "disk-sda", passphraseMsg.DiskID)
	assert.Equal(t, "LVM_LUKS", passphraseMsg.Capability)
	assert.Equal(t, "test", passphraseMsg.Passphrase)
}

func TestPassphrase_UpdateEnterWithEmptyShowsWarning(t *testing.T) {
	p := NewPassphrase("disk-sda", "LVM_LUKS")
	assert.False(t, p.showEmpty)

	next, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	p = next.(*PassphraseScreen)
	assert.Nil(t, cmd)
	assert.True(t, p.showEmpty)

	view := p.View(80, 24)
	assert.Contains(t, view, "cannot be empty")
}

func TestPassphrase_UpdateEscEmitsCancelMsg(t *testing.T) {
	p := NewPassphrase("disk-sda", "LVM_LUKS")
	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(PassphraseCancelMsg)
	assert.True(t, ok)
}

func TestPassphrase_ViewMasksInput(t *testing.T) {
	p := NewPassphrase("disk-sda", "LVM_LUKS")
	next, _ := p.Update(tea.KeyMsg{Runes: []rune{'s', 'e', 'c', 'r', 'e', 't'}})
	p = next.(*PassphraseScreen)

	view := p.View(80, 24)
	assert.NotContains(t, view, "secret")
	assert.Contains(t, view, "●●●●●●")
}

func TestPassphrase_UpdateClearsEmptyWarningOnInput(t *testing.T) {
	p := NewPassphrase("disk-sda", "LVM_LUKS")
	next, _ := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	p = next.(*PassphraseScreen)
	assert.True(t, p.showEmpty)

	next, _ = p.Update(tea.KeyMsg{Runes: []rune{'a'}})
	p = next.(*PassphraseScreen)
	assert.False(t, p.showEmpty)
}
