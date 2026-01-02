package common

import (
	"fmt"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type (
	View string

	SwitchViewMsg struct {
		to View
	}
)

const (
	CloseGameDialogView View = "close-game-dialog"
	DeleteGameStatView  View = "delete-game-stat-dialog"
	GameView            View = "game"
	GameSettingsView    View = "game-settings"
	MainMenuView        View = "main-menu"
	GameDetailsView     View = "game-details"
	GameListView        View = "game-list"
	PlayerListView      View = "player-list"
	PlayerDetailsView   View = "player-details"
	UndoMoveView        View = "undo-move-dialog"
)

var (
	StylePink       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF75B7"))
	StyleInactive   = lipgloss.NewStyle().Foreground(lipgloss.Color("#909090"))
	StyleActive     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	StyleGreen      = lipgloss.NewStyle().Foreground(lipgloss.Color("#32CD32"))
	StyleError      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	StyleHelp       = lipgloss.NewStyle().Foreground(lipgloss.Color("#4A4A4A"))
	StyleUnderlined = lipgloss.NewStyle().Underline(true)

	red   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	white = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
)

func NewTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	ti.TextStyle = StylePink
	// TODO: how do suggestions work?
	// ti.SetSuggestions([]string{"1", "2", "3", "4", "5", "B", "DB"})
	// ti.ShowSuggestions = true
	return ti
}

func NewHelp() help.Model {
	h := help.New()
	h.ShortSeparator = ", "
	return h
}

func NewTable() *table.Table {
	var (
		noBorder = lipgloss.Border{
			Top:          "",
			Bottom:       "",
			Left:         "  ",
			Right:        "",
			TopLeft:      "",
			TopRight:     "",
			BottomLeft:   "",
			BottomRight:  "",
			MiddleLeft:   "",
			MiddleRight:  "",
			Middle:       "",
			MiddleTop:    "",
			MiddleBottom: "",
		}
		t = table.New().Border(noBorder).
			BorderLeft(false).
			BorderHeader(false).StyleFunc(func(row, col int) lipgloss.Style {
			return StyleInactive
		})
	)
	return t
}

func SwitchViewTo(v View) tea.Cmd {
	return func() tea.Msg {
		return SwitchViewMsg{
			to: v,
		}
	}
}

func (s *SwitchViewMsg) To() View {
	return s.to
}

func Fill(s string, length int) string {
	for {
		if utf8.RuneCountInString(s) < length {
			s += " "
		} else {
			return s
		}
	}
}

func Headline(s string) string {
	var (
		right = red.Render("»")
		left  = red.Render("«")
		minus = white.Render("—")
	)
	return fmt.Sprintf("%s==%s %s %s==%s", right, minus, s, minus, left)
}

// adjusts a viewport with vertical margin and cursor
func AdjustViewportResize(v *viewport.Model, msg tea.WindowSizeMsg, cursor, headerHeight, footerHeight int) {
	var (
		verticalMarginHeight = headerHeight + footerHeight
		newHeight            = msg.Height - verticalMarginHeight
	)

	// if the window became bigger, we can maybe scroll up a bit top prevent empty lines
	// this is not ideal because PastBottom() does not consider the vertical margin
	if v.PastBottom() && v.YOffset > 0 {
		v.YOffset -= newHeight - v.Height + headerHeight
		if v.YOffset < 0 {
			v.YOffset = 0
		}
	}

	v.Width = msg.Width
	v.Height = newHeight

	// if the window became smaller, the cursor might get out of view
	if lastVisibleLine := v.YOffset - headerHeight + v.VisibleLineCount(); cursor > lastVisibleLine {
		v.YOffset += cursor - lastVisibleLine
	}
}
