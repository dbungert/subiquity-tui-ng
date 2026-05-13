package screens

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"subiquity-ng/internal/ui"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type storageSpinnerTickMsg struct{}

func storageSpinnerTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return storageSpinnerTickMsg{}
	})
}

var capabilityNames = map[string]string{
	"DIRECT":            "Direct (ext4)",
	"LVM":               "LVM",
	"LVM_LUKS":          "LVM + LUKS",
	"ZFS":               "ZFS",
	"ZFS_LUKS_KEYSTORE": "ZFS + LUKS Keystore",
}

var capabilityDescriptions = map[string]string{
	"DIRECT":            "ext4, minimal partitions",
	"LVM":               "LVM, ext4 filesystem",
	"LVM_LUKS":          "LVM with LUKS full-disk encryption",
	"ZFS":               "ZFS, unencrypted",
	"ZFS_LUKS_KEYSTORE": "ZFS with Ubuntu LUKS keystore encryption",
}

type StorageItem struct {
	DiskID     string
	Capability string
}

type StorageCapabilitySelectedMsg struct {
	DiskID     string
	Capability string
}

type StorageScreen struct {
	items   []StorageItem
	cursor  int
	loading bool
	frame   int
}

func NewStorageLoading() *StorageScreen {
	return &StorageScreen{
		loading: true,
		frame:   0,
	}
}

func NewStorage(items []StorageItem) *StorageScreen {
	return &StorageScreen{
		items:   items,
		loading: false,
		cursor:  0,
	}
}

func (s *StorageScreen) Init() tea.Cmd {
	if s.loading {
		return storageSpinnerTick()
	}
	return nil
}

func (s *StorageScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case storageSpinnerTickMsg:
		if s.loading {
			s.frame++
			return s, storageSpinnerTick()
		}
	case tea.KeyMsg:
		if !s.loading && len(s.items) > 0 {
			switch msg.String() {
			case "up":
				if s.cursor > 0 {
					s.cursor--
				}
			case "down":
				if s.cursor < len(s.items)-1 {
					s.cursor++
				}
			case "enter":
				selected := s.items[s.cursor]
				return s, func() tea.Msg {
					return StorageCapabilitySelectedMsg(selected)
				}
			}
		}
	}
	return s, nil
}

func (s *StorageScreen) Title() string {
	return "Storage Configuration"
}

func (s *StorageScreen) View(width, height int) string {
	if s.loading {
		return fmt.Sprintf("%s  Waiting for block device probing…", spinnerFrames[s.frame%len(spinnerFrames)])
	}

	if len(s.items) == 0 {
		return "No suitable disks found."
	}

	contentWidth := ui.ConstrainedWidth(width)
	lines := make([]string, 0)
	lines = append(lines, "Choose a disk and installation type:")
	lines = append(lines, "")

	for i, item := range s.items {
		capName := capabilityNames[item.Capability]
		if capName == "" {
			capName = item.Capability
		}

		display := fmt.Sprintf("%s   %s", item.DiskID, capName)

		if i == s.cursor {
			display = langSelectedStyle.Width(contentWidth).Render("▶ " + display)
		} else {
			display = langNormalStyle.Width(contentWidth).Render("  " + display)
		}
		lines = append(lines, display)

		desc := capabilityDescriptions[item.Capability]
		if desc != "" {
			lines = append(lines, "    "+langHintStyle.Render(desc))
		}
	}

	lines = append(lines, "")
	lines = append(lines, "↑↓ Navigate   Enter Select")

	return strings.Join(lines, "\n")
}
