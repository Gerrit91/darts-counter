package game

import (
	"log/slog"
	"strings"

	"github.com/Gerrit91/darts-counter/pkg/config"
	"github.com/Gerrit91/darts-counter/pkg/datastore"

	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/davecgh/go-spew/spew"
)

type (
	mainMenu struct {
		cfg *config.Config
		log *slog.Logger
		ds  datastore.Datastore

		cursor  int
		choices []mainMenuChoice

		currentView   view
		views         map[view]tea.Model
		showGameModel *showGameModel
	}

	mainMenuChoice string
)

const (
	menuNewGame mainMenuChoice = "Start New Game"
	// "Game Configuration"
	// --> allow selecting game (type, check-in, players, ...)
	// --> redirect from start new game on first start
	menuShowPlayers mainMenuChoice = "Show Players"
	menuShowGames   mainMenuChoice = "Show Games"
	menuQuit        mainMenuChoice = "Exit"
)

func NewMainMenu(log *slog.Logger, c *config.Config, ds datastore.Datastore) *mainMenu {
	m := &mainMenu{
		cfg: c,
		log: log,
		ds:  ds,
		choices: []mainMenuChoice{
			menuNewGame,
			menuShowPlayers,
			menuShowGames,
			menuQuit,
		},
		currentView:   mainMenuView,
		showGameModel: newShowGameModel(log, ds),
	}

	m.views = map[view]tea.Model{
		mainMenuView: m,
		gameView:     nil,
		deleteGameStatView: newConfirmDialog(
			log,
			"Are you sure you want to delete this game from the statistics?\nIt cannot be recovered.",
			tea.Sequence(switchViewTo(showGames), deleteGameStat),
			switchViewTo(showGames),
		),
		closeGameDialogView: newConfirmDialog(
			log,
			"Are you sure you want to quit a running game?\nAll progress will be lost.",
			switchViewTo(mainMenuView),
			switchViewTo(gameView),
		),
		undoMoveView: newConfirmDialog(
			log,
			"Are you sure you want to undo the last move?",
			tea.Sequence(switchViewTo(gameView), undoMove),
			switchViewTo(gameView),
		),
		showGames:   newShowGamesModel(log, ds, m.showGameModel),
		showGame:    m.showGameModel,
		showPlayers: newShowPlayersModel(log, ds),
	}

	return m
}

func (m *mainMenu) Init() tea.Cmd {
	return nil
}

func (m *mainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case switchViewMsg:
		m.log.Info("received switch view message", "to", msg.to)
		m.currentView = msg.to

		if m.currentView == mainMenuView {
			return m, nil
		}

		view := m.views[m.currentView]
		if view == nil {
			m.log.Error("unknown view, falling back to main menu", "to", msg.to)
			return m, switchViewTo(mainMenuView)
		}

		return m, view.Init()
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.QuitMsg:
		m.log.Info("received quit msg, exiting")
		return m, tea.Quit
	case cursor.BlinkMsg:
		// do not log this
	default:
		m.log.Info("received update message", "msg", spew.Sdump(msg))
	}

	if m.currentView == mainMenuView {
		return m.update(msg)
	}

	view := m.views[m.currentView]
	if view == nil {
		m.log.Error("unknown view, falling back to main menu")
		return m, switchViewTo(mainMenuView)
	}

	_, cmd := view.Update(msg)
	return m, cmd
}

func (m *mainMenu) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			switch m.choices[m.cursor] {
			case menuNewGame:
				g, err := newGame(m.log, m.cfg, m.ds, m.showGameModel)
				if err != nil {
					panic(err)
				}

				m.views[gameView] = g

				return m, switchViewTo(gameView)
			case menuQuit:
				return m, tea.Quit
			case menuShowGames:
				return m, switchViewTo(showGames)
			case menuShowPlayers:
				return m, switchViewTo(showPlayers)
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

func (m *mainMenu) View() string {
	if m.currentView == mainMenuView {
		return m.view()
	}

	view := m.views[m.currentView]
	if view == nil {
		return m.view()
	}

	return view.View()
}

func (m *mainMenu) view() string {
	var lines []string

	lines = append(lines, headline("darts-counter"), "")

	for i := 0; i < len(m.choices); i++ {
		if m.cursor == i {
			selection := fill("→", 3)
			lines = append(lines, stylePink.Render(selection)+styleActive.Render(string(m.choices[i])))
			continue
		}

		selection := fill("", 3)
		lines = append(lines, selection+styleInactive.Render(string(m.choices[i])))
	}

	return strings.Join(lines, "\n")
}
