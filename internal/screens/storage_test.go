package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorage_Title(t *testing.T) {
	s := NewStorage(nil, "")
	assert.Equal(t, "Storage Configuration", s.Title())
}

func TestStorage_TitleShowsDiskLabel(t *testing.T) {
	items := []StorageItem{
		{DiskID: "disk-sda", Capability: "DIRECT"},
	}
	s := NewStorage(items, "/dev/sda")
	assert.Equal(t, "Storage Configuration — /dev/sda", s.Title())
}

func TestStorage_ViewLoadingShowsSpinner(t *testing.T) {
	s := NewStorageLoading()
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

func TestStorage_InitStartsTickWhenLoading(t *testing.T) {
	s := NewStorageLoading()
	cmd := s.Init()
	assert.NotNil(t, cmd, "loading screen should start spinner tick")
}

func TestStorage_InitNoTickWhenLoaded(t *testing.T) {
	s := NewStorage(nil, "")
	cmd := s.Init()
	assert.Nil(t, cmd, "loaded screen should not start spinner")
}

func TestStorage_ViewShowsItems(t *testing.T) {
	items := []StorageItem{
		{DiskID: "disk-sda", Capability: "DIRECT"},
		{DiskID: "disk-sda", Capability: "LVM"},
	}
	s := NewStorage(items, "")
	view := s.View(80, 24)
	assert.Contains(t, view, "Choose an installation type")
	assert.Contains(t, view, "Direct")
	assert.NotContains(t, view, "disk-sda   ")
}

func TestStorage_ViewShowsCapabilityNames(t *testing.T) {
	items := []StorageItem{
		{DiskID: "disk-sda", Capability: "LVM_LUKS"},
	}
	s := NewStorage(items, "")
	view := s.View(80, 24)
	assert.Contains(t, view, "LVM + Encryption 🔒")
	assert.Contains(t, view, "LVM with LUKS — same as LVM but passphrase-encrypted")
}

func TestStorage_NavigateUpDown(t *testing.T) {
	items := []StorageItem{
		{DiskID: "disk-a", Capability: "DIRECT"},
		{DiskID: "disk-b", Capability: "LVM"},
	}
	s := NewStorage(items, "")
	assert.Equal(t, 0, s.cursor)

	next, _ := s.Update(tea.KeyMsg{Type: tea.KeyDown})
	s = next.(*StorageScreen)
	assert.Equal(t, 1, s.cursor)

	next, _ = s.Update(tea.KeyMsg{Type: tea.KeyUp})
	s = next.(*StorageScreen)
	assert.Equal(t, 0, s.cursor)
}

func TestStorage_NavigateUpDownClamped(t *testing.T) {
	items := []StorageItem{
		{DiskID: "disk-a", Capability: "DIRECT"},
	}
	s := NewStorage(items, "")

	// Can't go up from 0
	next, _ := s.Update(tea.KeyMsg{Type: tea.KeyUp})
	s = next.(*StorageScreen)
	assert.Equal(t, 0, s.cursor)

	// Can't go down past last
	next, _ = s.Update(tea.KeyMsg{Type: tea.KeyDown})
	s = next.(*StorageScreen)
	assert.Equal(t, 0, s.cursor)
}

func TestStorage_EnterEmitsSelectedMsg(t *testing.T) {
	items := []StorageItem{
		{DiskID: "disk-sda", Capability: "DIRECT"},
	}
	s := NewStorage(items, "")
	_, cmd := s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg := cmd()
	selectedMsg, ok := msg.(StorageCapabilitySelectedMsg)
	assert.True(t, ok)
	assert.Equal(t, "disk-sda", selectedMsg.DiskID)
	assert.Equal(t, "DIRECT", selectedMsg.Capability)
}

func TestStorage_UpdateIgnoresNonKey(t *testing.T) {
	items := []StorageItem{
		{DiskID: "disk-a", Capability: "DIRECT"},
	}
	s := NewStorage(items, "")

	next, cmd := s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	assert.Equal(t, s, next)
	assert.Nil(t, cmd)
}

func TestStorage_ViewEmptyWhenNoItems(t *testing.T) {
	s := NewStorage(nil, "")
	view := s.View(80, 24)
	assert.Contains(t, view, "No suitable disks found")
}

func TestStorage_UpdateSpinnerWhenLoading(t *testing.T) {
	s := NewStorageLoading()
	assert.Equal(t, 0, s.frame)
	next, cmd := s.Update(storageSpinnerTickMsg{})
	s = next.(*StorageScreen)
	assert.Equal(t, 1, s.frame)
	assert.NotNil(t, cmd, "should return tick command to continue spinning")
}

func TestStorage_UpdateStopsTickWhenLoaded(t *testing.T) {
	s := NewStorage([]StorageItem{{DiskID: "disk-a", Capability: "DIRECT"}}, "")
	next, cmd := s.Update(storageSpinnerTickMsg{})
	assert.Equal(t, s, next)
	assert.Nil(t, cmd, "loaded screen should ignore tick")
}

func TestStorage_ViewGroupsCapabilitiesByFamily(t *testing.T) {
	items := []StorageItem{
		{DiskID: "disk-a", Capability: "LVM"},
		{DiskID: "disk-a", Capability: "LVM_LUKS"},
	}
	s := NewStorage(items, "")
	view := s.View(80, 24)
	assert.Contains(t, view, " ─ LVM ")
	familyCount := 0
	for i := 0; i < len(view)-(len(" ─ LVM ")); i++ {
		if view[i:i+len(" ─ LVM ")] == " ─ LVM " {
			familyCount++
		}
	}
	assert.Equal(t, 1, familyCount, "LVM family header should appear exactly once")
}

func TestStorage_ViewShowsFriendlyNames(t *testing.T) {
	items := []StorageItem{
		{DiskID: "disk-sda", Capability: "LVM_LUKS"},
		{DiskID: "disk-sda", Capability: "ZFS_LUKS_KEYSTORE"},
	}
	s := NewStorage(items, "")
	view := s.View(80, 24)
	assert.Contains(t, view, "LVM + Encryption 🔒")
	assert.Contains(t, view, "ZFS + Encryption 🔒")
}

func TestStorage_ViewShowsImprovedDescriptions(t *testing.T) {
	items := []StorageItem{
		{DiskID: "disk-sda", Capability: "LVM_LUKS"},
	}
	s := NewStorage(items, "")
	view := s.View(80, 24)
	assert.Contains(t, view, "same as LVM but passphrase-encrypted")
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
