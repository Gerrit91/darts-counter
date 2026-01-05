package gamelist

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Gerrit91/darts-counter/pkg/datastore"
	"github.com/Gerrit91/darts-counter/pkg/views/common"
	gamedetails "github.com/Gerrit91/darts-counter/pkg/views/game-details"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	model struct {
		log *slog.Logger
		ds  datastore.Datastore

		gameDetails *gamedetails.Model
		cursor      int
		stats       []*datastore.GameStats
		toDelete    *datastore.GameStats

		viewport viewport.Model
		help     help.Model
		err      error
	}

	deleteGameStatMsg struct{}
)

func DeleteGameStat() tea.Msg {
	return deleteGameStatMsg{}
}

func New(log *slog.Logger, ds datastore.Datastore, gameDetails *gamedetails.Model) *model {
	return &model{
		log:         log,
		viewport:    viewport.New(0, 20),
		ds:          ds,
		gameDetails: gameDetails,
		help:        common.NewHelp(),
	}
}

func (s *model) Init() tea.Cmd {
	var err error
	s.stats, err = s.ds.ListGameStats()
	if err != nil {
		s.err = err
		return nil
	}

	s.cursor = 0
	s.viewport.GotoTop()

	s.log.Info("fetched games from database", "entries", len(s.stats))

	s.gameDetails.SetBackTo(common.SwitchViewTo(common.GameListView))

	return tea.WindowSize()
}

func (s *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case deleteGameStatMsg:
		if s.toDelete == nil {
			s.log.Error("no game stat marked for deletion")
			return s, s.Init()
		}

		s.log.Info("deleting game stat", "id", s.toDelete.ID)

		err := s.ds.DeleteGameStats(s.toDelete.ID)
		if err != nil {
			s.err = err
			return s, nil
		}

		return s, s.Init()
	case tea.WindowSizeMsg:
		common.AdjustViewportResize(&s.viewport, msg, s.cursor, 2, 1)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return s, common.SwitchViewTo(common.MainMenuView)
		case "d", "delete":
			s.toDelete = s.stats[s.cursor]
			return s, common.SwitchViewTo(common.DeleteGameStatView)
		case "enter":
			stat := s.stats[s.cursor]
			s.gameDetails.SetGameStats(*stat)
			return s, common.SwitchViewTo(common.GameDetailsView)
		case "down":
			s.cursor++
			if s.cursor >= len(s.stats) {
				s.cursor = 0
				_ = s.viewport.GotoTop()
			}
		case "up":
			s.cursor--
			if s.cursor < 0 {
				s.cursor = len(s.stats) - 1
				s.viewport.GotoBottom()
			}
		case "g":
			s.cursor = 0
			s.viewport.GotoTop()
		case "G":
			s.cursor = len(s.stats) - 1
			s.viewport.GotoBottom()
		}
	}

	var cmd tea.Cmd
	s.viewport, cmd = s.viewport.Update(msg)

	return s, cmd
}

func (s *model) View() string {
	var (
		lines []string
		row   = func(stat *datastore.GameStats) []string {
			var players []string
			players = append(players, stat.Players...)
			for i, p := range players {
				players[i] = fmt.Sprintf("%s (%d.)", p, stat.Ranks.OfPlayer(p))
			}

			return []string{
				stat.Start.Format("02.01.2006"),
				stat.Start.Format(time.TimeOnly),
				string(stat.GameType),
				stat.End.Sub(stat.Start).Truncate(time.Second).String(),
				stat.Ranks[1],
				strings.Join(players, ", "),
			}
		}
		header = func() []string {
			return []string{
				"",
				"Date",
				"Time",
				"Game",
				"Duration",
				"Winner",
				"Players",
			}
		}
	)

	if s.err != nil {
		lines = append(lines, common.StyleError.Render(s.err.Error()))
		lines = append(lines, s.help.ShortHelpView([]key.Binding{
			key.NewBinding(
				key.WithKeys("q", "esc"),
				key.WithHelp("q", "quit"),
			),
		}))
		return strings.Join(lines, "\n")
	}

	t := common.NewTable().StyleFunc(func(row, col int) lipgloss.Style {
		switch {
		case col == 0:
			return common.StylePink
		case row == s.cursor:
			return common.StyleActive
		default:
			return common.StyleInactive
		}
	})

	t.Headers(header()...)

	for i, stat := range s.stats {
		selection := ""
		if s.cursor == i {
			selection = "→"
		}

		t = t.Row(append([]string{selection}, row(stat)...)...)
	}

	if s.viewport.Height > 0 { // otherwise it crashes
		s.viewport.SetContent(t.Render())
	}

	lines = append(lines, common.Headline("Game Statistics"))
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
			key.WithKeys("g", "G"),
			key.WithHelp("g/G", "top/bottom"),
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
	})+common.StyleHelp.Render(fmt.Sprintf(" (%3.f%%)", s.viewport.ScrollPercent()*100)))

	return strings.Join(lines, "\n")
}
