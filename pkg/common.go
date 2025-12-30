package game

import (
	"fmt"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	view string

	switchViewMsg struct {
		to view
	}
)

const (
	closeGameDialogView view = "close-game-dialog"
	deleteGameStatView  view = "delete-game-stat-dialog"
	gameView            view = "game"
	gameSettingsView    view = "game-settings"
	mainMenuView        view = "main-menu"
	showGame            view = "show-game"
	showGames           view = "show-games"
	showPlayers         view = "show-players"
	undoMoveView        view = "undo-move-dialog"
)

var (
	stylePink       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF75B7"))
	styleInactive   = lipgloss.NewStyle().Foreground(lipgloss.Color("#909090"))
	styleActive     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	styleGreen      = lipgloss.NewStyle().Foreground(lipgloss.Color("#32CD32"))
	styleError      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	styleHelp       = lipgloss.NewStyle().Foreground(lipgloss.Color("#4A4A4A"))
	styleUnderlined = lipgloss.NewStyle().Underline(true)
	red             = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	white           = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
)

func newTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	ti.TextStyle = stylePink
	// TODO: how do suggestions work?
	// ti.SetSuggestions([]string{"1", "2", "3", "4", "5", "B", "DB"})
	// ti.ShowSuggestions = true
	return ti
}

func newHelp() help.Model {
	h := help.New()
	h.ShortSeparator = ", "
	return h
}

func switchViewTo(v view) tea.Cmd {
	return func() tea.Msg {
		return switchViewMsg{
			to: v,
		}
	}
}

func fill(s string, length int) string {
	for {
		if utf8.RuneCountInString(s) < length {
			s += " "
		} else {
			return s
		}
	}
}

func headline(s string) string {
	var (
		right = red.Render("»")
		left  = red.Render("«")
		minus = white.Render("—")
	)
	return fmt.Sprintf("%s==%s %s %s==%s", right, minus, s, minus, left)
}
