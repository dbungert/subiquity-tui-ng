package screens

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"subiquity-ng/internal/ui"
)

type ConfirmAcceptedMsg struct{}
type ConfirmCancelMsg struct{}

type ConfirmScreen struct {
	diskLabel  string
	capability string
	cursor     int // 0 = Yes, 1 = No (default)
}

func NewConfirm(diskLabel, capability string) *ConfirmScreen {
	return &ConfirmScreen{
		diskLabel:  diskLabel,
		capability: capability,
		cursor:     1, // Default to No
	}
}

func (c *ConfirmScreen) Title() string {
	return "Confirm Installation"
}

func (c *ConfirmScreen) Init() tea.Cmd {
	return nil
}

func (c *ConfirmScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if c.cursor > 0 {
				c.cursor--
			}
		case "down":
			if c.cursor < 1 {
				c.cursor++
			}
		case "enter":
			if c.cursor == 0 {
				return c, func() tea.Msg {
					return ConfirmAcceptedMsg{}
				}
			} else {
				return c, func() tea.Msg {
					return ConfirmCancelMsg{}
				}
			}
		case "esc":
			return c, func() tea.Msg {
				return ConfirmCancelMsg{}
			}
		}
	}
	return c, nil
}

func (c *ConfirmScreen) View(width, height int) string {
	contentWidth := ui.ConstrainedWidth(width)
	lines := make([]string, 0)

	lines = append(lines, "⚠  This will ERASE ALL DATA on "+c.diskLabel+".")
	lines = append(lines, "Installation type: "+friendlyCapabilityName(c.capability))
	lines = append(lines, "")
	lines = append(lines, "This action cannot be undone.")
	lines = append(lines, "")

	options := []string{"Yes, erase disk and install", "No, go back"}
	for i, option := range options {
		if i == c.cursor {
			display := langSelectedStyle.Width(contentWidth).Render("▶ " + option)
			lines = append(lines, display)
		} else {
			display := langNormalStyle.Width(contentWidth).Render("  " + option)
			lines = append(lines, display)
		}
	}

	lines = append(lines, "")
	lines = append(lines, "↑↓ Navigate   Enter Select")

	return strings.Join(lines, "\n")
}

func friendlyCapabilityName(capability string) string {
	names := map[string]string{
		"DIRECT":            "Direct",
		"LVM":               "LVM",
		"LVM_LUKS":          "LVM + Encryption",
		"ZFS":               "ZFS",
		"ZFS_LUKS_KEYSTORE": "ZFS + Encryption",
	}
	if name, ok := names[capability]; ok {
		return name
	}
	return capability
}
