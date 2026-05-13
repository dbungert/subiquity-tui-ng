package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"subiquity-ng/internal/screens"
)

func newModel() Model {
	return Model{current: screens.NewWelcome()}
}

func TestModel_InitDelegatesToScreen(t *testing.T) {
	if cmd := newModel().Init(); cmd != nil {
		t.Errorf("expected nil cmd from welcome.Init")
	}
}

func TestModel_UpdateWindowSizeStoresDimensions(t *testing.T) {
	m, _ := newModel().Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	got := m.(Model)
	if got.width != 100 || got.height != 30 {
		t.Errorf("got %dx%d, want 100x30", got.width, got.height)
	}
}

func TestModel_UpdateCtrlCQuits(t *testing.T) {
	_, cmd := newModel().Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatalf("expected quit cmd")
	}
	// tea.Quit is a function returning tea.QuitMsg; calling it should give that.
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", cmd())
	}
}

func TestModel_ViewEmptyBeforeWindowSize(t *testing.T) {
	if v := newModel().View(); v != "" {
		t.Errorf("expected empty view pre-sizing, got %q", v)
	}
}

func TestModel_ViewAfterSizeContainsTitle(t *testing.T) {
	m, _ := newModel().Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	v := m.(Model).View()
	if !strings.Contains(v, "Welcome!") {
		t.Errorf("view missing welcome text: %q", v)
	}
}
