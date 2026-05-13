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

type capabilityMeta struct {
	name   string
	family string
	desc   string
}

var capabilities = map[string]capabilityMeta{
	"DIRECT":            {"Direct", "Direct", "Standard ext4 filesystem — no encryption, minimal partitions"},
	"LVM":               {"LVM", "LVM", "Logical Volume Manager — flexible partitioning, no encryption"},
	"LVM_LUKS":          {"LVM + Encryption 🔒", "LVM", "LVM with LUKS — same as LVM but passphrase-encrypted"},
	"ZFS":               {"ZFS", "ZFS", "ZFS copy-on-write filesystem — data integrity, no encryption"},
	"ZFS_LUKS_KEYSTORE": {"ZFS + Encryption 🔒", "ZFS", "ZFS with Ubuntu LUKS keystore — same as ZFS but encrypted"},
}

var capabilityOrder = []string{"DIRECT", "LVM", "LVM_LUKS", "ZFS", "ZFS_LUKS_KEYSTORE"}

type StorageItem struct {
	DiskID     string
	Capability string
}

type StorageCapabilitySelectedMsg struct {
	DiskID     string
	Capability string
}

type StorageScreen struct {
	items     []StorageItem
	cursor    int
	loading   bool
	frame     int
	diskLabel string
}

func NewStorageLoading() *StorageScreen {
	return &StorageScreen{
		loading: true,
		frame:   0,
	}
}

func NewStorage(items []StorageItem, diskLabel string) *StorageScreen {
	sortedItems := sortStorageItems(items)
	return &StorageScreen{
		items:     sortedItems,
		loading:   false,
		cursor:    0,
		diskLabel: diskLabel,
	}
}

func sortStorageItems(items []StorageItem) []StorageItem {
	capIndex := make(map[string]int)
	for i, cap := range capabilityOrder {
		capIndex[cap] = i
	}

	sorted := make([]StorageItem, len(items))
	copy(sorted, items)

	sliceSort := func(i, j int) bool {
		iIdx, iOk := capIndex[sorted[i].Capability]
		jIdx, jOk := capIndex[sorted[j].Capability]
		if !iOk {
			return false
		}
		if !jOk {
			return true
		}
		if iIdx != jIdx {
			return iIdx < jIdx
		}
		return sorted[i].DiskID < sorted[j].DiskID
	}

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sliceSort(j, i) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
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
	if s.diskLabel != "" {
		return fmt.Sprintf("Storage Configuration — %s", s.diskLabel)
	}
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
	lines = append(lines, "Choose an installation type:")
	lines = append(lines, "")

	var lastFamily string
	for i, item := range s.items {
		meta := capabilities[item.Capability]

		if meta.family != lastFamily {
			if lastFamily != "" {
				lines = append(lines, "")
			}
			headerLine := fmt.Sprintf(" ─ %s %s", meta.family, strings.Repeat("─", contentWidth-len(meta.family)-4))
			lines = append(lines, langHintStyle.Render(headerLine))
			lastFamily = meta.family
		}

		display := meta.name

		if i == s.cursor {
			display = langSelectedStyle.Width(contentWidth).Render("▶ " + display)
		} else {
			display = langNormalStyle.Width(contentWidth).Render("  " + display)
		}
		lines = append(lines, display)
		lines = append(lines, "    "+langHintStyle.Render(meta.desc))
	}

	lines = append(lines, "")
	lines = append(lines, "↑↓ Navigate   Enter Select")

	return strings.Join(lines, "\n")
}
