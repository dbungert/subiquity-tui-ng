package screens

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestInstallProgress_Title(t *testing.T) {
	s := NewInstallProgress()
	assert.Equal(t, "Installation", s.Title())
}

func TestInstallProgress_ViewDefaultText(t *testing.T) {
	s := NewInstallProgress()
	view := s.View(80, 24)
	assert.Contains(t, view, "Installing...")
}

func TestInstallProgress_ViewShowsState(t *testing.T) {
	s := NewInstallProgress()
	next, _ := s.Update(InstallProgressStateMsg{State: "RUNNING"})
	s = next.(*InstallProgress)

	view := s.View(80, 24)
	assert.Contains(t, view, "RUNNING")
	assert.NotContains(t, view, "Installing...")
}

func TestInstallProgress_UpdateHandlesStateMsg(t *testing.T) {
	s := NewInstallProgress()
	assert.Equal(t, "", s.state)

	next, cmd := s.Update(InstallProgressStateMsg{State: "RUNNING"})
	s = next.(*InstallProgress)
	assert.Equal(t, "RUNNING", s.state)
	assert.Nil(t, cmd)
}

func TestInstallProgress_UpdateIgnoresOtherMessages(t *testing.T) {
	s := NewInstallProgress()
	next, cmd := s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	s = next.(*InstallProgress)
	assert.Equal(t, "", s.state)
	assert.Nil(t, cmd)
}

func TestInstallProgress_LogLineAppearsInView(t *testing.T) {
	s := NewInstallProgress()
	next, _ := s.Update(InstallLogLineMsg{Line: "unpacking base-files"})
	s = next.(*InstallProgress)

	view := s.View(80, 24)
	assert.Contains(t, view, "unpacking base-files")
}

func TestInstallProgress_MultipleLogLinesAllAppear(t *testing.T) {
	s := NewInstallProgress()
	next, _ := s.Update(InstallLogLineMsg{Line: "line-1"})
	s = next.(*InstallProgress)
	next, _ = s.Update(InstallLogLineMsg{Line: "line-2"})
	s = next.(*InstallProgress)
	next, _ = s.Update(InstallLogLineMsg{Line: "line-3"})
	s = next.(*InstallProgress)

	view := s.View(80, 24)
	assert.Contains(t, view, "line-1")
	assert.Contains(t, view, "line-2")
	assert.Contains(t, view, "line-3")
}

func TestInstallProgress_LogLinesCappedAtMaxLogLines(t *testing.T) {
	s := NewInstallProgress()
	// Send maxLogLines + 3 lines (0 through 12 for maxLogLines=10)
	for i := 0; i < maxLogLines+3; i++ {
		next, _ := s.Update(InstallLogLineMsg{Line: fmt.Sprintf("line-%d", i)})
		s = next.(*InstallProgress)
	}

	// Buffer should only contain the last 10 lines (3-12)
	assert.Equal(t, maxLogLines, len(s.logLines))
	assert.Equal(t, "line-3", s.logLines[0])
	assert.Equal(t, "line-12", s.logLines[9])

	view := s.View(80, 24)
	assert.Contains(t, view, "line-12")
}

func TestInstallProgress_LongLogLineTruncated(t *testing.T) {
	s := NewInstallProgress()
	longLine := strings.Repeat("a", 200)
	next, _ := s.Update(InstallLogLineMsg{Line: longLine})
	s = next.(*InstallProgress)

	view := s.View(80, 24)
	assert.Contains(t, view, "…")
}

func TestInstallProgress_UpdateHandlesLogLineMsg(t *testing.T) {
	s := NewInstallProgress()
	next, cmd := s.Update(InstallLogLineMsg{Line: "test"})
	s = next.(*InstallProgress)
	assert.Equal(t, 1, len(s.logLines))
	assert.Nil(t, cmd)
}

func TestInstallProgress_NoLogLinesShowsInstalling(t *testing.T) {
	s := NewInstallProgress()
	view := s.View(80, 24)
	assert.Contains(t, view, "Installing...")
}
