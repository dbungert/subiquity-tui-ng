package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiskSelection_Title(t *testing.T) {
	d := NewDiskSelection(nil)
	assert.Equal(t, "Select Disk", d.Title())
}

func TestDiskSelection_ViewShowsDiskPaths(t *testing.T) {
	items := []DiskItem{
		{DiskID: "disk-sda", Path: "/dev/sda", Allowed: []string{"DIRECT", "LVM"}},
		{DiskID: "disk-sdb", Path: "/dev/sdb", Allowed: []string{"DIRECT"}},
	}
	d := NewDiskSelection(items)
	view := d.View(80, 24)
	assert.Contains(t, view, "Choose a disk to install on")
	assert.Contains(t, view, "/dev/sda")
	assert.Contains(t, view, "/dev/sdb")
	assert.NotContains(t, view, "disk-sda")
	assert.NotContains(t, view, "disk-sdb")
}

func TestDiskSelection_ViewShowsFamilyHint(t *testing.T) {
	items := []DiskItem{
		{DiskID: "sda", Allowed: []string{"DIRECT", "LVM", "LVM_LUKS"}},
	}
	d := NewDiskSelection(items)
	view := d.View(80, 24)
	assert.Contains(t, view, "Supported:")
	assert.Contains(t, view, "Direct")
	assert.Contains(t, view, "LVM")
}

func TestDiskSelection_NavigateUpDown(t *testing.T) {
	items := []DiskItem{
		{DiskID: "sda", Allowed: []string{"DIRECT"}},
		{DiskID: "sdb", Allowed: []string{"DIRECT"}},
	}
	d := NewDiskSelection(items)
	assert.Equal(t, 0, d.cursor)

	next, _ := d.Update(tea.KeyMsg{Type: tea.KeyDown})
	d = next.(*DiskSelectionScreen)
	assert.Equal(t, 1, d.cursor)

	next, _ = d.Update(tea.KeyMsg{Type: tea.KeyUp})
	d = next.(*DiskSelectionScreen)
	assert.Equal(t, 0, d.cursor)
}

func TestDiskSelection_NavigateClamped(t *testing.T) {
	items := []DiskItem{
		{DiskID: "sda", Allowed: []string{"DIRECT"}},
	}
	d := NewDiskSelection(items)

	next, _ := d.Update(tea.KeyMsg{Type: tea.KeyUp})
	d = next.(*DiskSelectionScreen)
	assert.Equal(t, 0, d.cursor)

	next, _ = d.Update(tea.KeyMsg{Type: tea.KeyDown})
	d = next.(*DiskSelectionScreen)
	assert.Equal(t, 0, d.cursor)
}

func TestDiskSelection_EnterEmitsDiskSelectedMsg(t *testing.T) {
	items := []DiskItem{
		{DiskID: "sda", Allowed: []string{"DIRECT"}},
	}
	d := NewDiskSelection(items)
	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg := cmd()
	selectedMsg, ok := msg.(DiskSelectedMsg)
	assert.True(t, ok)
	assert.Equal(t, "sda", selectedMsg.DiskID)
}

func TestDiskSelection_ViewFallsBackToDiskID(t *testing.T) {
	items := []DiskItem{
		{DiskID: "disk-sda", Path: "", Allowed: []string{"DIRECT"}},
	}
	d := NewDiskSelection(items)
	view := d.View(80, 24)
	assert.Contains(t, view, "disk-sda")
}

func TestDiskSelection_SingleItem(t *testing.T) {
	items := []DiskItem{
		{DiskID: "disk-sda", Path: "/dev/sda", Allowed: []string{"DIRECT", "LVM"}},
	}
	d := NewDiskSelection(items)
	view := d.View(80, 24)
	assert.Contains(t, view, "/dev/sda")
}
