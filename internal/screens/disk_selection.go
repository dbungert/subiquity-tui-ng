package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"subiquity-ng/internal/ui"
)

type DiskItem struct {
	DiskID  string
	Allowed []string
}

type DiskSelectionScreen struct {
	items  []DiskItem
	cursor int
}

type DiskSelectedMsg struct {
	DiskID string
}

func NewDiskSelection(items []DiskItem) *DiskSelectionScreen {
	return &DiskSelectionScreen{
		items:  items,
		cursor: 0,
	}
}

func (d *DiskSelectionScreen) Title() string {
	return "Select Disk"
}

func (d *DiskSelectionScreen) Init() tea.Cmd {
	return nil
}

func (d *DiskSelectionScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if len(d.items) > 0 {
			switch msg.String() {
			case "up":
				if d.cursor > 0 {
					d.cursor--
				}
			case "down":
				if d.cursor < len(d.items)-1 {
					d.cursor++
				}
			case "enter":
				selected := d.items[d.cursor]
				return d, func() tea.Msg {
					return DiskSelectedMsg{DiskID: selected.DiskID}
				}
			}
		}
	}
	return d, nil
}

func (d *DiskSelectionScreen) View(width, height int) string {
	if len(d.items) == 0 {
		return "No suitable disks found."
	}

	contentWidth := ui.ConstrainedWidth(width)
	lines := make([]string, 0)
	lines = append(lines, "Choose a disk to install on:")
	lines = append(lines, "")

	for i, item := range d.items {
		families := extractFamilies(item.Allowed)
		hint := fmt.Sprintf("Supported: %s", strings.Join(families, ", "))

		display := item.DiskID

		if i == d.cursor {
			display = langSelectedStyle.Width(contentWidth).Render("▶ " + display)
		} else {
			display = langNormalStyle.Width(contentWidth).Render("  " + display)
		}
		lines = append(lines, display)
		lines = append(lines, "    "+langHintStyle.Render(hint))
	}

	lines = append(lines, "")
	lines = append(lines, "↑↓ Navigate   Enter Select")

	return strings.Join(lines, "\n")
}

func extractFamilies(capList []string) []string {
	families := make(map[string]bool)
	for _, cap := range capList {
		meta, ok := capabilities[cap]
		if ok {
			families[meta.family] = true
		}
	}

	familyList := make([]string, 0, len(families))
	for family := range families {
		familyList = append(familyList, family)
	}
	return familyList
}
