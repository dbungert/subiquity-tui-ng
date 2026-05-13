package screens

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type storageSpinnerTickMsg struct{}

func storageSpinnerTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return storageSpinnerTickMsg{}
	})
}

type StorageScreen struct {
	rawJSON string
	loading bool
	frame   int
}

func NewStorage(rawJSON string) *StorageScreen {
	return &StorageScreen{
		rawJSON: rawJSON,
		loading: rawJSON == "",
		frame:   0,
	}
}

func (s *StorageScreen) Init() tea.Cmd {
	if s.loading {
		return storageSpinnerTick()
	}
	return nil
}

func (s *StorageScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg.(type) {
	case storageSpinnerTickMsg:
		if s.loading {
			s.frame++
			return s, storageSpinnerTick()
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
	return s.rawJSON
}
