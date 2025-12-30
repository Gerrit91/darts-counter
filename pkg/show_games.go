package game

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Gerrit91/darts-counter/pkg/datastore"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	showGamesModel struct {
		log *slog.Logger
		ds  datastore.Datastore

		show   *showGameModel
		cursor int
		stats  []*datastore.GameStats

		viewport viewport.Model
		help     help.Model
		err      error
	}

	deleteGameStatMsg struct{}
)

func deleteGameStat() tea.Msg {
	return deleteGameStatMsg{}
}

func newShowGamesModel(log *slog.Logger, ds datastore.Datastore, show *showGameModel) *showGamesModel {
	return &showGamesModel{
		log:      log,
		viewport: viewport.New(0, 20),
		ds:       ds,
		show:     show,
		help:     newHelp(),
	}
}

func (s *showGamesModel) Init() tea.Cmd {
	var err error
	s.stats, err = s.ds.ListGameStats()
	if err != nil {
		s.err = err
		return nil
	}

	s.cursor = 0
	s.viewport.GotoTop()

	s.log.Info("fetched games from database", "entries", len(s.stats))

	s.show.backTo = switchViewTo(showGames)

	return tea.WindowSize()
}

func (s *showGamesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case deleteGameStatMsg:
		stat := s.stats[s.cursor]

		s.log.Info("deleting game stat", "id", stat.ID)

		err := s.ds.DeleteGameStats(stat.ID)
		if err != nil {
			s.err = err
			return s, nil
		}

		return s, s.Init()
	case tea.WindowSizeMsg:
		headerHeight := 2
		footerHeight := 1
		verticalMarginHeight := headerHeight + footerHeight

		helpWidth := 5
		s.help.Width = msg.Width - helpWidth
		s.viewport.Width = msg.Width
		s.viewport.Height = msg.Height - verticalMarginHeight
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return s, switchViewTo(mainMenuView)
		case "d", "delete":
			return s, switchViewTo(deleteGameStatView)
		case "enter":
			stat := s.stats[s.cursor]
			s.show.gs = *stat
			return s, switchViewTo(showGame)
		case "down":
			s.cursor++
			if s.cursor >= len(s.stats) {
				s.cursor = 0
			}
		case "up":
			s.cursor--
			if s.cursor < 0 {
				s.cursor = len(s.stats) - 1
			}
		}
	}

	var cmd tea.Cmd
	s.viewport, cmd = s.viewport.Update(msg)
	return s, cmd
}

func (s *showGamesModel) View() string {

	var (
		lines         []string
		viewportLines []string
		row           = func(stat *datastore.GameStats, style lipgloss.Style) string {
			var players []string
			players = append(players, stat.Players...)
			for i, p := range players {
				players[i] = fmt.Sprintf("%s (%d.)", p, stat.Ranks.OfPlayer(p))
			}
			s := fill(stat.Start.Format("02.01.2006"), 11) +
				fill(stat.Start.Format(time.TimeOnly), 9) +
				fill(string(stat.GameType), 5) +
				fill(stat.End.Sub(stat.Start).Truncate(time.Second).String(), 9) +
				fill(stat.Ranks[1], 10) +
				strings.Join(players, ", ")
			return style.Render(s)
		}
		header = func() string {
			s := fill("Date", 11) +
				fill("Time", 9) +
				fill("Game", 5) +
				fill("Duration", 9) +
				fill("Winner", 10) +
				"Players"
			return s
		}
	)

	if s.err != nil {
		lines = append(lines, styleError.Render(s.err.Error()))
		lines = append(lines, s.help.ShortHelpView([]key.Binding{
			key.NewBinding(
				key.WithKeys("q", "esc"),
				key.WithHelp("q", "quit"),
			),
		}))
		return strings.Join(lines, "\n")
	}

	viewportLines = append(viewportLines, "   "+header())

	for i, stat := range s.stats {
		if s.cursor == i {
			selection := fill("→", 3)
			viewportLines = append(viewportLines, stylePink.Render(selection)+row(stat, styleActive))
			continue
		}

		selection := fill("", 3)
		viewportLines = append(viewportLines, selection+row(stat, styleInactive))
	}

	s.viewport.SetContent(strings.Join(viewportLines, "\n"))

	lines = append(lines, headline("Game Statistics"), "")
	lines = append(lines, s.viewport.View())
	lines = append(lines, s.help.ShortHelpView([]key.Binding{
		key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "up"),
		),
		key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "down"),
		),
		key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "show details"),
		),
		key.NewBinding(
			key.WithKeys("d", "delete"),
			key.WithHelp("d", "remove entry"),
		),
		key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q", "quit"),
		),
	})+" ("+styleInactive.Render(fmt.Sprintf("%3.f%%", s.viewport.ScrollPercent()*100))+")")

	return strings.Join(lines, "\n")
}
