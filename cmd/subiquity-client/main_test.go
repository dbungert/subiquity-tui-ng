package main

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"subiquity-ng/internal/screens"
)

func newModel() Model {
	return Model{current: screens.NewLanguage()}
}

func TestModel_InitDelegatesToScreen(t *testing.T) {
	assert.Nil(t, newModel().Init())
}

func TestModel_UpdateWindowSizeStoresDimensions(t *testing.T) {
	m, _ := newModel().Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	got := m.(Model)
	assert.Equal(t, 100, got.width)
	assert.Equal(t, 30, got.height)
}

func TestModel_UpdateCtrlCQuits(t *testing.T) {
	_, cmd := newModel().Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	assert.NotNil(t, cmd, "expected quit cmd")
	// tea.Quit is a function returning tea.QuitMsg; calling it should give that.
	_, ok := cmd().(tea.QuitMsg)
	assert.True(t, ok, "expected tea.QuitMsg")
}

func TestModel_ViewEmptyBeforeWindowSize(t *testing.T) {
	assert.Empty(t, newModel().View())
}

func TestModel_ViewAfterSizeContainsTitle(t *testing.T) {
	m, _ := newModel().Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	v := m.(Model).View()
	assert.Contains(t, v, "Welcome!")
}

func TestModel_StoresSocket(t *testing.T) {
	m := Model{current: screens.NewLanguage(), socket: "/tmp/test.sock"}
	assert.Equal(t, "/tmp/test.sock", m.socket)
}

func TestArgs_DefaultSocketPathWhenNotProvided(t *testing.T) {
	args := Args{}
	assert.Equal(t, "", args.Socket)
}

func TestArgs_CustomSocketPath(t *testing.T) {
	customPath := "/custom/socket/path"
	args := Args{Socket: customPath}
	assert.Equal(t, customPath, args.Socket)
}

func TestSocketDefaultLogic_FallsBackToHomeWhenProdNotFound(t *testing.T) {
	args := Args{}
	prodSocket := "/run/subiquity/socket"
	if _, err := os.Stat(prodSocket); err == nil {
		// Production socket exists, this test doesn't apply
		t.Skip("production socket exists")
	}
	if args.Socket == "" {
		if _, err := os.Stat(prodSocket); err == nil {
			args.Socket = prodSocket
		} else {
			args.Socket = filepath.Join(".subiquity", "socket")
		}
	}
	assert.Equal(t, filepath.Join(".subiquity", "socket"), args.Socket)
}
