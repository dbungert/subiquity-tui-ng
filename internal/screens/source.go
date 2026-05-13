package screens

import (
	"fmt"
	"strings"

	"subiquity-ng/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

type SourceItem struct {
	ID          string
	Name        string
	Description string
	Size        int64
}

type SourceSelectedMsg struct {
	ID string
}

type SourceScreen struct {
	items  []SourceItem
	cursor int
}

func NewSource(items []SourceItem, currentID string) *SourceScreen {
	cursor := 0
	for i, item := range items {
		if item.ID == currentID {
			cursor = i
			break
		}
	}
	return &SourceScreen{
		items:  items,
		cursor: cursor,
	}
}

func (s *SourceScreen) Init() tea.Cmd {
	return nil
}

func (s *SourceScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return s, nil
	}

	switch key.String() {
	case "up":
		if s.cursor > 0 {
			s.cursor--
		}
	case "down":
		if s.cursor < len(s.items)-1 {
			s.cursor++
		}
	case "enter":
		if len(s.items) > 0 {
			selected := s.items[s.cursor]
			return s, func() tea.Msg { return SourceSelectedMsg{ID: selected.ID} }
		}
	}
	return s, nil
}

func (s *SourceScreen) Title() string {
	return "Installation Source"
}

func (s *SourceScreen) View(width, height int) string {
	if len(s.items) == 0 {
		return "Loading sources…"
	}

	contentWidth := ui.ConstrainedWidth(width)
	lines := make([]string, 0)
	lines = append(lines, "Choose the installation source for Ubuntu:")
	lines = append(lines, "")

	for i, item := range s.items {
		sizeStr := formatSize(item.Size)
		nameAndSize := fmt.Sprintf("%s %-50s %s", item.Name, "", sizeStr)
		nameAndSize = strings.TrimRight(nameAndSize, " ")

		if i == s.cursor {
			nameAndSize = langSelectedStyle.Width(contentWidth).Render("▶ " + nameAndSize)
		} else {
			nameAndSize = langNormalStyle.Width(contentWidth).Render("  " + nameAndSize)
		}
		lines = append(lines, nameAndSize)
		lines = append(lines, "  "+langHintStyle.Render(item.Description))
	}

	lines = append(lines, "")
	lines = append(lines, "↑↓ Navigate   Enter Select")

	return strings.Join(lines, "\n")
}

func formatSize(bytes int64) string {
	if bytes >= 1e9 {
		return fmt.Sprintf("%.1f GB", float64(bytes)/1e9)
	}
	return fmt.Sprintf("%d MB", bytes/1e6)
}
