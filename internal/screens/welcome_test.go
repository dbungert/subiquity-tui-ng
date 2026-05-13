package screens

import (
	"strings"
	"testing"
)

func TestWelcome_TitleIsMultilingualGreeting(t *testing.T) {
	got := NewWelcome().Title()
	for _, want := range []string{"Willkommen!", "Bienvenue!", "Welcome!", "Добро пожаловать!", "Welkom!"} {
		if !strings.Contains(got, want) {
			t.Errorf("title missing %q: %s", want, got)
		}
	}
}

func TestWelcome_ViewMentionsKeys(t *testing.T) {
	got := NewWelcome().View(80, 20)
	if !strings.Contains(got, "UP") || !strings.Contains(got, "DOWN") || !strings.Contains(got, "ENTER") {
		t.Errorf("view should mention UP/DOWN/ENTER: %q", got)
	}
}

func TestWelcome_InitNoCmd(t *testing.T) {
	if cmd := NewWelcome().Init(); cmd != nil {
		t.Errorf("expected nil cmd")
	}
}

func TestWelcome_UpdateIsIdentity(t *testing.T) {
	w := NewWelcome()
	next, cmd := w.Update(nil)
	if next != Screen(w) {
		t.Errorf("expected same screen back")
	}
	if cmd != nil {
		t.Errorf("expected nil cmd")
	}
}
