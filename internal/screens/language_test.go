package screens

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestLanguage_Title(t *testing.T) {
	title := NewLanguage().Title()
	for _, want := range []string{"Willkommen!", "Bienvenue!", "Welcome!", "Добро пожаловать!", "Welkom!"} {
		assert.Contains(t, title, want)
	}
}

func TestLanguage_DefaultSelectionIsEnglish(t *testing.T) {
	ls := NewLanguage()
	assert.GreaterOrEqual(t, ls.cursor, 0)
	assert.Less(t, ls.cursor, len(ls.visible))
	got := ls.all[ls.visible[ls.cursor]]
	assert.Equal(t, "en_US.UTF-8", got.Code)
}

func TestLanguage_AllLanguagesVisibleInitially(t *testing.T) {
	ls := NewLanguage()
	assert.Len(t, ls.visible, len(ls.all))
}

func TestLanguage_NavigateDown(t *testing.T) {
	ls := NewLanguage()
	start := ls.cursor
	next, _ := ls.Update(tea.KeyMsg{Type: tea.KeyDown})
	ls = next.(*LanguageScreen)
	assert.Equal(t, start+1, ls.cursor)
}

func TestLanguage_NavigateUpClampsAtZero(t *testing.T) {
	ls := NewLanguage()
	ls.cursor = 0
	next, _ := ls.Update(tea.KeyMsg{Type: tea.KeyUp})
	ls = next.(*LanguageScreen)
	assert.Equal(t, 0, ls.cursor)
}

func TestLanguage_NavigateDownClampsAtEnd(t *testing.T) {
	ls := NewLanguage()
	ls.cursor = len(ls.visible) - 1
	next, _ := ls.Update(tea.KeyMsg{Type: tea.KeyDown})
	ls = next.(*LanguageScreen)
	assert.Equal(t, len(ls.visible)-1, ls.cursor)
}

func TestLanguage_TypingFiltersListByNative(t *testing.T) {
	ls := NewLanguage()
	next, _ := ls.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	ls = next.(*LanguageScreen)
	for _, idx := range ls.visible {
		lang := ls.all[idx]
		native := strings.ToLower(lang.Native)
		english := strings.ToLower(lang.English)
		assert.True(t, strings.Contains(native, "d") || strings.Contains(english, "d"),
			"filtered list should only contain matching entries")
	}
}

func TestLanguage_FilterByEnglishName(t *testing.T) {
	ls := NewLanguage()
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
	assert.True(t, found, "filtering by 'german' should include Deutsch")
}

func TestLanguage_FilterCaseInsensitive(t *testing.T) {
	ls := NewLanguage()
	next, _ := ls.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("E")})
	lsUpper := next.(*LanguageScreen)

	ls = NewLanguage()
	next, _ = ls.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	lsLower := next.(*LanguageScreen)

	assert.Len(t, lsUpper.visible, len(lsLower.visible))
}

func TestLanguage_FilterResetsOffset(t *testing.T) {
	ls := NewLanguage()
	ls.offset = 5
	next, _ := ls.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	ls = next.(*LanguageScreen)
	assert.Equal(t, 0, ls.offset)
}

func TestLanguage_BackspaceRemovesLastRune(t *testing.T) {
	ls := NewLanguage()
	ls.filter = "fran"
	ls.rebuildVisible()
	before := len(ls.visible)

	next, _ := ls.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	ls = next.(*LanguageScreen)

	assert.Equal(t, "fra", ls.filter)
	assert.GreaterOrEqual(t, len(ls.visible), before)
}

func TestLanguage_BackspaceOnEmptyFilterIsNoop(t *testing.T) {
	ls := NewLanguage()
	next, _ := ls.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	ls = next.(*LanguageScreen)
	assert.Empty(t, ls.filter)
}

func TestLanguage_EscClearsFilter(t *testing.T) {
	ls := NewLanguage()
	ls.filter = "fran"
	ls.rebuildVisible()
	next, _ := ls.Update(tea.KeyMsg{Type: tea.KeyEsc})
	ls = next.(*LanguageScreen)
	assert.Empty(t, ls.filter)
	assert.Len(t, ls.visible, len(ls.all))
}

func TestLanguage_FilterNoMatchesLeavesEmptyVisible(t *testing.T) {
	ls := NewLanguage()
	for _, r := range "zzzzzzzzz" {
		next, _ := ls.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		ls = next.(*LanguageScreen)
	}
	assert.Empty(t, ls.visible)
}

func TestLanguage_ViewContainsSearchPrompt(t *testing.T) {
	got := NewLanguage().View(80, 21)
	assert.Contains(t, got, "Search:")
}

func TestLanguage_ViewContainsHints(t *testing.T) {
	got := NewLanguage().View(80, 21)
	for _, want := range []string{"Navigate", "Select", "filter", "Clear"} {
		assert.Contains(t, got, want)
	}
}

func TestLanguage_ViewContainsSelectedLanguage(t *testing.T) {
	ls := NewLanguage()
	got := ls.View(80, 21)
	assert.Contains(t, got, "English")
}

func TestLanguage_ViewScrollsToShowCursor(t *testing.T) {
	ls := NewLanguage()
	ls.cursor = len(ls.visible) - 1
	ls.offset = 0
	got := ls.View(80, 21)
	lastName := ls.all[ls.visible[len(ls.visible)-1]]
	assert.Contains(t, got, lastName.English)
}

func TestLanguage_NonKeyMsgIsIgnored(t *testing.T) {
	ls := NewLanguage()
	cursorBefore := ls.cursor
	filterBefore := ls.filter
	next, cmd := ls.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	ls = next.(*LanguageScreen)
	assert.Nil(t, cmd)
	assert.Equal(t, cursorBefore, ls.cursor)
	assert.Equal(t, filterBefore, ls.filter)
}
