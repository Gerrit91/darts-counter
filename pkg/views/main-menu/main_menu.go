package mainmenu

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/Gerrit91/darts-counter/pkg/config"
	"github.com/Gerrit91/darts-counter/pkg/datastore"
	"github.com/Gerrit91/darts-counter/pkg/views/common"
	"github.com/Gerrit91/darts-counter/pkg/views/confirm-dialog"
	"github.com/Gerrit91/darts-counter/pkg/views/game"
	gamedetails "github.com/Gerrit91/darts-counter/pkg/views/game-details"
	gamelist "github.com/Gerrit91/darts-counter/pkg/views/game-list"
	gamesettings "github.com/Gerrit91/darts-counter/pkg/views/game-settings"
	playerdetails "github.com/Gerrit91/darts-counter/pkg/views/player-details"
	playerlist "github.com/Gerrit91/darts-counter/pkg/views/player-list"

	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/davecgh/go-spew/spew"
)

type (
	model struct {
		cfg *config.Config
		log *slog.Logger
		ds  datastore.Datastore

		cursor  int
		choices []mainMenuChoice
		err     error

		currentView      common.View
		views            map[common.View]tea.Model
		gameDetailsModel *gamedetails.Model
	}

	mainMenuChoice string
)

const (
	menuNewGame      mainMenuChoice = "Start New Game"
	menuGameSettings mainMenuChoice = "Game Settings"
	menuShowPlayers  mainMenuChoice = "Show Players"
	menuShowGames    mainMenuChoice = "Show Games"
	menuQuit         mainMenuChoice = "Exit"
)

func New(log *slog.Logger, c *config.Config, ds datastore.Datastore) *model {
	m := &model{
		cfg: c,
		log: log,
		ds:  ds,
		choices: []mainMenuChoice{
			menuNewGame,
			menuGameSettings,
			menuShowPlayers,
			menuShowGames,
			menuQuit,
		},
		currentView:      common.MainMenuView,
		gameDetailsModel: gamedetails.New(log, ds),
	}

	playerDetailsModel := playerdetails.New(log, ds)

	m.views = map[common.View]tea.Model{
		common.MainMenuView:     m,
		common.GameView:         nil,
		common.GameSettingsView: gamesettings.New(log, ds),
		common.DeleteGameStatView: confirm.New(
			log,
			"Are you sure you want to delete this game from the statistics?\nIt cannot be recovered.",
			tea.Sequence(common.SwitchViewTo(common.GameListView), gamelist.DeleteGameStat),
			common.SwitchViewTo(common.GameListView),
		),
		common.CloseGameDialogView: confirm.New(
			log,
			"Are you sure you want to quit a running game?\nAll progress will be lost.",
			common.SwitchViewTo(common.MainMenuView),
			common.SwitchViewTo(common.GameView),
		),
		common.UndoMoveView: confirm.New(
			log,
			"Are you sure you want to undo the last move?",
			tea.Sequence(common.SwitchViewTo(common.GameView), game.UndoMove),
			common.SwitchViewTo(common.GameView),
		),
		common.GameListView:    gamelist.New(log, ds, m.gameDetailsModel),
		common.GameDetailsView: m.gameDetailsModel,
		common.PlayerListView: playerlist.New(
			log,
			ds,
			playerDetailsModel,
		),
		common.PlayerDetailsView: playerDetailsModel,
	}

	return m
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case common.SwitchViewMsg:
		m.log.Info("received switch view message", "to", msg.To())
		m.currentView = msg.To()

		if m.currentView == common.MainMenuView {
			return m, nil
		}

		view := m.views[m.currentView]
		if view == nil {
			m.log.Error("unknown view, falling back to main menu", "to", msg.To())
			return m, common.SwitchViewTo(common.MainMenuView)
		}

		return m, view.Init()
	case tea.KeyMsg:
		m.log.Info("received key message", "msg", spew.Sdump(msg))

		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.QuitMsg:
		m.log.Info("received quit msg, exiting")
		return m, tea.Quit
	case cursor.BlinkMsg, tea.MouseMsg:
		// do not log this
	default:
		m.log.Info("received update message", "msg", spew.Sdump(msg))
	}

	if m.currentView == common.MainMenuView {
		return m.update(msg)
	}

	view := m.views[m.currentView]
	if view == nil {
		m.log.Error("unknown view, falling back to main menu")
		return m, common.SwitchViewTo(common.MainMenuView)
	}

	_, cmd := view.Update(msg)
	return m, cmd
}

func (m *model) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.err = nil

		switch msg.String() {
		case "enter":
			switch m.choices[m.cursor] {
			case menuNewGame:
				g, err := game.New(m.log, m.ds, m.gameDetailsModel)
				if err != nil {
					if errors.Is(err, datastore.ErrNotFound) {
						return m, common.SwitchViewTo(common.GameSettingsView)
					}
					m.err = err
					return m, nil
				}

				m.views[common.GameView] = g

				return m, common.SwitchViewTo(common.GameView)
			case menuGameSettings:
				return m, common.SwitchViewTo(common.GameSettingsView)
			case menuQuit:
				return m, tea.Quit
			case menuShowGames:
				return m, common.SwitchViewTo(common.GameListView)
			case menuShowPlayers:
				return m, common.SwitchViewTo(common.PlayerListView)
			default:

			}
		case "down":
			m.cursor++
			if m.cursor >= len(m.choices) {
				m.cursor = 0
			}
		case "up":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.choices) - 1
			}
		case "q", "esc":
			return m, tea.Quit
		default:
			return m, nil
		}
	}

	return m, nil
}

func (m *model) View() string {
	if m.currentView == common.MainMenuView {
		return m.view()
	}

	view := m.views[m.currentView]
	if view == nil {
		return m.view()
	}

	return view.View()
}

func (m *model) view() string {
	var lines []string

	lines = append(lines, common.Headline("darts-counter"), "")

	for i := range len(m.choices) {
		if m.cursor == i {
			selection := common.Fill("→", 3)
			lines = append(lines, common.StylePink.Render(selection)+common.StyleActive.Render(string(m.choices[i])))
			continue
		}

		selection := common.Fill("", 3)
		lines = append(lines, selection+common.StyleInactive.Render(string(m.choices[i])))
	}

	if m.err != nil {
		lines = append(lines, "", common.StyleError.Render(m.err.Error()))
	}

	return strings.Join(lines, "\n")
}
