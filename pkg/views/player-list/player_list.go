package showplayers

import (
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Gerrit91/darts-counter/pkg/datastore"
	"github.com/Gerrit91/darts-counter/pkg/views/common"
	playerdetails "github.com/Gerrit91/darts-counter/pkg/views/player-details"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type (
	model struct {
		log *slog.Logger
		ds  datastore.Datastore

		viewport      viewport.Model
		table         *table.Table
		help          help.Model
		err           error
		cursor        int
		stats         []*datastore.PlayerStats
		playerDetails *playerdetails.Model
	}
)

func New(log *slog.Logger, ds datastore.Datastore, playerDetails *playerdetails.Model) *model {
	return &model{
		log:           log,
		ds:            ds,
		viewport:      viewport.New(0, 20),
		help:          common.NewHelp(),
		table:         common.NewTable(),
		playerDetails: playerDetails,
	}
}

func (s *model) Init() tea.Cmd {
	gameStats, err := s.ds.ListGameStats()
	if err != nil {
		s.err = err
		return nil
	}

	s.stats, err = datastore.ToPlayerStats(gameStats)
	if err != nil {
		s.err = err
		return nil
	}

	sort.SliceStable(s.stats, func(i, j int) bool {
		return s.stats[i].RanksCount[1] > s.stats[j].RanksCount[1]
	}) // sort by wins

	s.cursor = 0
	s.viewport.GotoTop()

	s.log.Info("fetched players from database", "entries", len(s.stats))

	return tea.WindowSize()
}

func (s *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		common.AdjustViewportResize(&s.viewport, msg, s.cursor, 2, 1)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return s, common.SwitchViewTo(common.MainMenuView)
		case "down":
			s.cursor++
			if s.cursor >= len(s.stats) {
				s.cursor = 0
				s.viewport.GotoTop()
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
		case "enter":
			ps := s.stats[s.cursor]
			s.playerDetails.SetPlayerStats(ps)
			return s, common.SwitchViewTo(common.PlayerDetailsView)
		}
	}

	var cmd tea.Cmd
	s.viewport, cmd = s.viewport.Update(msg)

	return s, cmd
}

func (s *model) View() string {
	var (
		lines []string
		row   = func(stat *datastore.PlayerStats) []string {
			var (
				favField = ""
				favCount = 0
				wins     = ""
				losses   = ""
			)

			for field, count := range stat.FieldsCount {
				if count > favCount {
					favField = field
					favCount = count
				}
			}
			favField = fmt.Sprintf("%s (%dx)", favField, favCount)

			if len(stat.RanksCount) > 1 {
				wins = strconv.Itoa(stat.RanksCount[1])
				losses = strconv.Itoa(stat.RanksCount[len(stat.RanksCount)])
			}

			return []string{
				stat.ID,
				wins,
				losses,
				strconv.Itoa(stat.GamesPlayed),
				strconv.FormatFloat(stat.AverageRank, 'f', 3, 64),
				strconv.FormatFloat(stat.AverageScore, 'f', 1, 64),
				fmt.Sprintf("%d (%s)", stat.HighestScore.Total, strings.Join(stat.HighestScore.Fields, " → ")),
				favField,
				stat.AverageDuration.Truncate(time.Millisecond).String(),
			}
		}
		header = func() []string {
			return []string{
				"",
				"Name",
				"Wins",
				"Losses",
				"Games",
				"⌀-Rank",
				"⌀-Score",
				"Max Score",
				"Fav Field",
				"⌀-Sec/move",
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

	t = t.Headers(header()...)

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

	lines = append(lines, common.Headline("Player Statistics"))
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
			key.WithKeys("q", "esc"),
			key.WithHelp("q", "quit"),
		),
	})+common.StyleHelp.Render(fmt.Sprintf(" (%3.f%%)", s.viewport.ScrollPercent()*100)))

	return strings.Join(lines, "\n")
}
