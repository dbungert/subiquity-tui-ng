package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"subiquity-ng/internal/ui"
)

type PassphraseScreen struct {
	diskID     string
	capability string
	input      string
	showEmpty  bool
}

type PassphraseEnteredMsg struct {
	DiskID     string
	Capability string
	Passphrase string
}

type PassphraseCancelMsg struct{}

func NewPassphrase(diskID, capability string) *PassphraseScreen {
	return &PassphraseScreen{
		diskID:     diskID,
		capability: capability,
		input:      "",
		showEmpty:  false,
	}
}

func (p *PassphraseScreen) Title() string {
	return "Encryption Passphrase"
}

func (p *PassphraseScreen) Init() tea.Cmd {
	return nil
}

func (p *PassphraseScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if p.input == "" {
				p.showEmpty = true
				return p, nil
			}
			return p, func() tea.Msg {
				return PassphraseEnteredMsg{
					DiskID:     p.diskID,
					Capability: p.capability,
					Passphrase: p.input,
				}
			}
		case tea.KeyEsc:
			return p, func() tea.Msg {
				return PassphraseCancelMsg{}
			}
		case tea.KeyBackspace, tea.KeyCtrlH:
			if len(p.input) > 0 {
				p.input = p.input[:len(p.input)-1]
			}
			p.showEmpty = false
			return p, nil
		default:
			if len(msg.Runes) > 0 {
				p.input += string(msg.Runes)
				p.showEmpty = false
			}
			return p, nil
		}
	}
	return p, nil
}

func (p *PassphraseScreen) View(width, height int) string {
	contentWidth := ui.ConstrainedWidth(width)
	meta := capabilities[p.capability]

	lines := make([]string, 0)
	lines = append(lines, fmt.Sprintf("Enter a passphrase to encrypt %s with %s", p.diskID, meta.name))
	lines = append(lines, "")

	masked := strings.Repeat("●", len(p.input))
	passphraseDisplay := fmt.Sprintf("  Passphrase:  %s", masked)
	lines = append(lines, langNormalStyle.Width(contentWidth).Render(passphraseDisplay))
	lines = append(lines, "")

	if p.showEmpty {
		lines = append(lines, langHintStyle.Render("Passphrase cannot be empty."))
		lines = append(lines, "")
	}

	lines = append(lines, "Enter Confirm   Esc Back")

	return strings.Join(lines, "\n")
}
