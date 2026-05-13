package screens

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestLanguage_Init(t *testing.T) {
	if cmd := NewLanguage().Init(); cmd != nil {
		t.Error("expected nil cmd from Init")
	}
}

func TestLanguage_Title(t *testing.T) {
	title := NewLanguage().Title()
	for _, want := range []string{"Willkommen!", "Bienvenue!", "Welcome!", "Добро пожаловать!", "Welkom!"} {
		if !strings.Contains(title, want) {
			t.Errorf("title missing %q: %s", want, title)
		}
	}
}

func TestLanguage_DefaultSelectionIsEnglish(t *testing.T) {
	ls := NewLanguage()
	if ls.cursor < 0 || ls.cursor >= len(ls.visible) {
		t.Fatalf("cursor %d out of range [0, %d)", ls.cursor, len(ls.visible))
	}
	got := ls.all[ls.visible[ls.cursor]]
	if got.Code != "en_US.UTF-8" {
		t.Errorf("default selection: got %q (%s), want en_US.UTF-8", got.Native, got.Code)
	}
}

func TestLanguage_AllLanguagesVisibleInitially(t *testing.T) {
	ls := NewLanguage()
	if len(ls.visible) != len(ls.all) {
		t.Errorf("initially visible=%d, want %d", len(ls.visible), len(ls.all))
	}
}

func TestLanguage_NavigateDown(t *testing.T) {
	ls := NewLanguage()
	start := ls.cursor
	next, _ := ls.Update(tea.KeyMsg{Type: tea.KeyDown})
	ls = next.(*LanguageScreen)
	if ls.cursor != start+1 {
		t.Errorf("cursor after Down: got %d, want %d", ls.cursor, start+1)
	}
}

func TestLanguage_NavigateUpClampsAtZero(t *testing.T) {
	ls := NewLanguage()
	ls.cursor = 0
	next, _ := ls.Update(tea.KeyMsg{Type: tea.KeyUp})
	ls = next.(*LanguageScreen)
	if ls.cursor != 0 {
		t.Errorf("cursor should stay at 0 when pressing Up at top, got %d", ls.cursor)
	}
}

func TestLanguage_NavigateDownClampsAtEnd(t *testing.T) {
	ls := NewLanguage()
	ls.cursor = len(ls.visible) - 1
	next, _ := ls.Update(tea.KeyMsg{Type: tea.KeyDown})
	ls = next.(*LanguageScreen)
	if ls.cursor != len(ls.visible)-1 {
		t.Errorf("cursor should stay at end when pressing Down at bottom, got %d", ls.cursor)
	}
}

func TestLanguage_TypingFiltersListByNative(t *testing.T) {
	ls := NewLanguage()
	next, _ := ls.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	ls = next.(*LanguageScreen)
	for _, idx := range ls.visible {
		lang := ls.all[idx]
		native := strings.ToLower(lang.Native)
		english := strings.ToLower(lang.English)
		if !strings.Contains(native, "d") && !strings.Contains(english, "d") {
			t.Errorf("filtered list contains non-matching entry: %q / %q", lang.Native, lang.English)
		}
	}
}

func TestLanguage_FilterByEnglishName(t *testing.T) {
	ls := NewLanguage()
	// Type "german" - matches Deutsch by English name
	for _, r := range "german" {
		next, _ := ls.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		ls = next.(*LanguageScreen)
	}
	found := false
	for _, idx := range ls.visible {
		if ls.all[idx].Code == "de_DE.UTF-8" {
			found = true
		}
	}
	if !found {
		t.Error("filtering by 'german' should include Deutsch")
	}
}

func TestLanguage_FilterCaseInsensitive(t *testing.T) {
	ls := NewLanguage()
	next, _ := ls.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("E")})
	lsUpper := next.(*LanguageScreen)

	ls = NewLanguage()
	next, _ = ls.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	lsLower := next.(*LanguageScreen)

	if len(lsUpper.visible) != len(lsLower.visible) {
		t.Errorf("case should not matter: 'E' matched %d, 'e' matched %d",
			len(lsUpper.visible), len(lsLower.visible))
	}
}

func TestLanguage_FilterResetsOffset(t *testing.T) {
	ls := NewLanguage()
	ls.offset = 5
	next, _ := ls.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	ls = next.(*LanguageScreen)
	if ls.offset != 0 {
		t.Errorf("typing should reset offset to 0, got %d", ls.offset)
	}
}

func TestLanguage_BackspaceRemovesLastRune(t *testing.T) {
	ls := NewLanguage()
	ls.filter = "fran"
	ls.rebuildVisible()
	before := len(ls.visible)

	next, _ := ls.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	ls = next.(*LanguageScreen)

	if ls.filter != "fra" {
		t.Errorf("after backspace filter=%q, want %q", ls.filter, "fra")
	}
	if len(ls.visible) < before {
		t.Errorf("removing a filter char should not decrease matches: before=%d after=%d", before, len(ls.visible))
	}
}

func TestLanguage_BackspaceOnEmptyFilterIsNoop(t *testing.T) {
	ls := NewLanguage()
	next, _ := ls.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	ls = next.(*LanguageScreen)
	if ls.filter != "" {
		t.Errorf("backspace on empty filter should leave it empty, got %q", ls.filter)
	}
}

func TestLanguage_EscClearsFilter(t *testing.T) {
	ls := NewLanguage()
	ls.filter = "fran"
	ls.rebuildVisible()
	next, _ := ls.Update(tea.KeyMsg{Type: tea.KeyEsc})
	ls = next.(*LanguageScreen)
	if ls.filter != "" {
		t.Errorf("Esc should clear filter, got %q", ls.filter)
	}
	if len(ls.visible) != len(ls.all) {
		t.Errorf("after Esc all languages should be visible: got %d, want %d", len(ls.visible), len(ls.all))
	}
}

func TestLanguage_FilterNoMatchesLeavesEmptyVisible(t *testing.T) {
	ls := NewLanguage()
	for _, r := range "zzzzzzzzz" {
		next, _ := ls.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		ls = next.(*LanguageScreen)
	}
	if len(ls.visible) != 0 {
		t.Errorf("nonsense filter should match nothing, got %d matches", len(ls.visible))
	}
}

func TestLanguage_ViewContainsSearchPrompt(t *testing.T) {
	got := NewLanguage().View(80, 21)
	if !strings.Contains(got, "Search:") {
		t.Error("View should contain 'Search:'")
	}
}

func TestLanguage_ViewContainsHints(t *testing.T) {
	got := NewLanguage().View(80, 21)
	for _, want := range []string{"Navigate", "Select", "filter", "Clear"} {
		if !strings.Contains(got, want) {
			t.Errorf("View hints missing %q", want)
		}
	}
}

func TestLanguage_ViewContainsSelectedLanguage(t *testing.T) {
	ls := NewLanguage()
	got := ls.View(80, 21)
	if !strings.Contains(got, "English") {
		t.Error("View should contain selected language 'English'")
	}
}

func TestLanguage_ViewScrollsToShowCursor(t *testing.T) {
	ls := NewLanguage()
	// Move cursor to the last entry
	ls.cursor = len(ls.visible) - 1
	ls.offset = 0
	got := ls.View(80, 21)
	// The last language should be visible
	lastName := ls.all[ls.visible[len(ls.visible)-1]]
	if !strings.Contains(got, lastName.English) {
		t.Errorf("View should scroll to show cursor at bottom; last lang %q not found", lastName.English)
	}
}

func TestLanguage_NonKeyMsgIsIgnored(t *testing.T) {
	ls := NewLanguage()
	cursorBefore := ls.cursor
	filterBefore := ls.filter
	next, cmd := ls.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	ls = next.(*LanguageScreen)
	if cmd != nil {
		t.Error("non-key msg should produce nil cmd")
	}
	if ls.cursor != cursorBefore || ls.filter != filterBefore {
		t.Error("non-key msg should not change cursor or filter")
	}
}
