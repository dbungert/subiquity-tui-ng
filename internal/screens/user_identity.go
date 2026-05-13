package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"subiquity-ng/internal/ui"
)

type UserIdentityScreen struct {
	inputs    [3]string
	focused   int
	showEmpty bool
}

type UserIdentityDoneMsg struct {
	Realname string
	Username string
	Password string
}

func NewUserIdentity() *UserIdentityScreen {
	return &UserIdentityScreen{
		inputs:    [3]string{"", "", ""},
		focused:   0,
		showEmpty: false,
	}
}

func (u *UserIdentityScreen) Title() string {
	return "User Identity"
}

func (u *UserIdentityScreen) Init() tea.Cmd {
	return nil
}

func (u *UserIdentityScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if u.inputs[0] == "" || u.inputs[1] == "" || u.inputs[2] == "" {
				u.showEmpty = true
				return u, nil
			}
			return u, func() tea.Msg {
				return UserIdentityDoneMsg{
					Realname: u.inputs[0],
					Username: u.inputs[1],
					Password: u.inputs[2],
				}
			}
		case tea.KeyDown:
			if u.focused < 2 {
				u.focused++
				u.showEmpty = false
			}
			return u, nil
		case tea.KeyUp:
			if u.focused > 0 {
				u.focused--
				u.showEmpty = false
			}
			return u, nil
		case tea.KeyBackspace, tea.KeyCtrlH:
			if len(u.inputs[u.focused]) > 0 {
				u.inputs[u.focused] = u.inputs[u.focused][:len(u.inputs[u.focused])-1]
			}
			u.showEmpty = false
			return u, nil
		default:
			if len(msg.Runes) > 0 {
				u.inputs[u.focused] += string(msg.Runes)
				u.showEmpty = false
			}
			return u, nil
		}
	}
	return u, nil
}

func (u *UserIdentityScreen) View(width, height int) string {
	contentWidth := ui.ConstrainedWidth(width)
	labels := [3]string{"Your name:", "Username:", "Password:"}

	lines := make([]string, 0)
	lines = append(lines, "Enter your identity for the new system.")
	lines = append(lines, "")

	for i := 0; i < 3; i++ {
		label := labels[i]
		value := u.inputs[i]

		if i == 2 {
			value = strings.Repeat("●", len(u.inputs[i]))
		}

		display := fmt.Sprintf("%s %s", label, value)
		if i == u.focused {
			lines = append(lines, langSelectedStyle.Width(contentWidth).Render("▶ " + display))
		} else {
			lines = append(lines, langNormalStyle.Width(contentWidth).Render("  " + display))
		}
		lines = append(lines, "")
	}

	if u.showEmpty {
		lines = append(lines, langHintStyle.Render("All fields are required."))
		lines = append(lines, "")
	}

	lines = append(lines, "Enter Confirm   Up/Down Navigate")

	return strings.Join(lines, "\n")
}
