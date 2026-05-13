package screens

import (
	"fmt"
	"strings"

	"subiquity-ng/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Language holds display and locale data for one installable system language.
type Language struct {
	Native  string // name in that language, e.g. "Deutsch"
	English string // English name, e.g. "German"
	Code    string // locale code, e.g. "de_DE.UTF-8"
}

// LanguageSelectedMsg is sent when the user confirms language selection via Enter.
type LanguageSelectedMsg struct {
	Code string
}

// allLanguages is sourced from the upstream subiquity languagelist resource,
// sorted by English name.
var allLanguages = []Language{
	{"عربي", "Arabic", "ar_EG.UTF-8"},
	{"Asturianu", "Asturian", "ast_ES.UTF-8"},
	{"Беларуская", "Belarusian", "be_BY.UTF-8"},
	{"Català", "Catalan", "ca_ES.UTF-8"},
	{"中文(简体)", "Chinese (Simplified)", "zh_CN.UTF-8"},
	{"中文(繁體)", "Chinese (Traditional)", "zh_TW.UTF-8"},
	{"Hrvatski", "Croatian", "hr_HR.UTF-8"},
	{"Čeština", "Czech", "cs_CZ.UTF-8"},
	{"Nederlands", "Dutch", "nl_NL.UTF-8"},
	{"English", "English", "en_US.UTF-8"},
	{"English (UK)", "English (UK)", "en_GB.UTF-8"},
	{"Suomi", "Finnish", "fi_FI.UTF-8"},
	{"Français", "French", "fr_FR.UTF-8"},
	{"Galego", "Galician", "gl_ES.UTF-8"},
	{"Deutsch", "German", "de_DE.UTF-8"},
	{"Ελληνικά", "Greek", "el_GR.UTF-8"},
	{"עברית", "Hebrew", "he_IL.UTF-8"},
	{"Magyar", "Hungarian", "hu_HU.UTF-8"},
	{"Bahasa Indonesia", "Indonesian", "id_ID.UTF-8"},
	{"日本語", "Japanese", "ja_JP.UTF-8"},
	{"Taqbaylit", "Kabyle", "kab_DZ.UTF-8"},
	{"Latviski", "Latvian", "lv_LV.UTF-8"},
	{"Lietuviškai", "Lithuanian", "lt_LT.UTF-8"},
	{"Norsk bokmål", "Norwegian Bokmål", "nb"},
	{"Occitan (aprèp 1500)", "Occitan", "oc"},
	{"Polski", "Polish", "pl_PL.UTF-8"},
	{"Português", "Portuguese", "pt_PT.UTF-8"},
	{"Русский", "Russian", "ru_RU.UTF-8"},
	{"Српски", "Serbian", "sr_RS"},
	{"Español", "Spanish", "es_ES.UTF-8"},
	{"Svenska", "Swedish", "sv_SE.UTF-8"},
	{"བོད་ཡིག", "Tibetan", "bo_IN"},
	{"Українська", "Ukrainian", "uk_UA.UTF-8"},
}

var (
	langSelectedStyle = lipgloss.NewStyle().
				Background(ui.UbuntuOrange).
				Foreground(ui.HeaderFg)
	langNormalStyle = lipgloss.NewStyle()
	langHintStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#767676"))
	langCountStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#767676"))
)

// langOverhead is the number of lines View uses outside the scrollable list:
// description + blank + search + blank + blank-before-hints + hints.
const langOverhead = 6

// LanguageScreen is the first screen the user sees; it collects the install language.
type LanguageScreen struct {
	all     []Language
	filter  string
	visible []int // indices into all that match the current filter
	cursor  int   // position within visible
	offset  int   // first rendered row index within visible (for scrolling)
}

// NewLanguage constructs a LanguageScreen with English pre-selected.
func NewLanguage() *LanguageScreen {
	ls := &LanguageScreen{all: allLanguages}
	ls.rebuildVisible()
	for i, idx := range ls.visible {
		if ls.all[idx].Code == "en_US.UTF-8" {
			ls.cursor = i
			break
		}
	}
	return ls
}

func (ls *LanguageScreen) Init() tea.Cmd { return nil }

func (ls *LanguageScreen) Title() string {
	return "Willkommen! Bienvenue! Welcome! Добро пожаловать! Welkom!"
}

func (ls *LanguageScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return ls, nil
	}
	switch key.String() {
	case "up":
		if ls.cursor > 0 {
			ls.cursor--
		}
	case "down":
		if ls.cursor < len(ls.visible)-1 {
			ls.cursor++
		}
	case "pgup":
		ls.cursor = max(0, ls.cursor-10)
	case "pgdown":
		if len(ls.visible) > 0 {
			ls.cursor = min(len(ls.visible)-1, ls.cursor+10)
		}
	case "enter":
		if len(ls.visible) > 0 {
			selected := ls.all[ls.visible[ls.cursor]]
			return ls, func() tea.Msg { return LanguageSelectedMsg{Code: selected.Code} }
		}
	case "esc":
		ls.filter = ""
		ls.rebuildVisible()
	case "backspace", "ctrl+h":
		if len(ls.filter) > 0 {
			runes := []rune(ls.filter)
			ls.filter = string(runes[:len(runes)-1])
			ls.rebuildVisible()
		}
	default:
		if len(key.Runes) == 1 {
			ls.filter += string(key.Runes)
			ls.rebuildVisible()
			ls.cursor = 0
			ls.offset = 0
		}
	}
	return ls, nil
}

func (ls *LanguageScreen) rebuildVisible() {
	ls.visible = ls.visible[:0]
	f := strings.ToLower(ls.filter)
	for i, lang := range ls.all {
		if f == "" ||
			strings.Contains(strings.ToLower(lang.Native), f) ||
			strings.Contains(strings.ToLower(lang.English), f) {
			ls.visible = append(ls.visible, i)
		}
	}
	if ls.cursor >= len(ls.visible) && len(ls.visible) > 0 {
		ls.cursor = len(ls.visible) - 1
	}
}

func (ls *LanguageScreen) View(width, height int) string {
	listH := max(1, height-langOverhead)

	// Keep cursor within the scroll window.
	if ls.cursor < ls.offset {
		ls.offset = ls.cursor
	}
	if ls.cursor >= ls.offset+listH {
		ls.offset = ls.cursor - listH + 1
	}

	var sb strings.Builder

	// Description
	sb.WriteString("Choose the language used during installation and in the installed system.")
	sb.WriteByte('\n')
	sb.WriteByte('\n')

	// Search line with match count on the right
	count := fmt.Sprintf("%d / %d", len(ls.visible), len(ls.all))
	filterDisplay := " Search: " + ls.filter + "█"
	pad := width - lipgloss.Width(filterDisplay) - lipgloss.Width(count) - 1
	if pad < 1 {
		pad = 1
	}
	sb.WriteString(filterDisplay)
	sb.WriteString(strings.Repeat(" ", pad))
	sb.WriteString(langCountStyle.Render(count))
	sb.WriteByte('\n')
	sb.WriteByte('\n')

	// List rows
	for row := 0; row < listH; row++ {
		i := ls.offset + row
		if i < len(ls.visible) {
			lang := ls.all[ls.visible[i]]
			prefix := "  "
			if i == ls.cursor {
				prefix = "> "
			}
			label := prefix + lang.Native
			if lang.Native != lang.English {
				label += " (" + lang.English + ")"
			}
			if i == ls.cursor {
				sb.WriteString(langSelectedStyle.Width(width).Render(label))
			} else {
				sb.WriteString(langNormalStyle.Width(width).Render(label))
			}
		} else {
			sb.WriteString(strings.Repeat(" ", width))
		}
		if row < listH-1 {
			sb.WriteByte('\n')
		}
	}

	sb.WriteByte('\n')
	sb.WriteByte('\n')

	sb.WriteString(langHintStyle.Render(" ↑↓ Navigate   Enter Select   Type to filter   Esc Clear"))

	return sb.String()
}
