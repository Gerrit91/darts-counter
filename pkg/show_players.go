package game

import (
	"fmt"
	"log/slog"
	"sort"
	"strconv"
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
	showPlayersModel struct {
		log *slog.Logger
		ds  datastore.Datastore

		viewport viewport.Model
		help     help.Model
		err      error
		cursor   int
		stats    []*datastore.PlayerStats
	}
)

func newShowPlayersModel(log *slog.Logger, ds datastore.Datastore) *showPlayersModel {
	return &showPlayersModel{
		log:      log,
		viewport: viewport.New(0, 20),
		ds:       ds,
		help:     newHelp(),
	}
}

func (s *showPlayersModel) Init() tea.Cmd {
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

func (s *showPlayersModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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

func (s *showPlayersModel) View() string {
	var (
		lines         []string
		viewportLines []string
		row           = func(stat *datastore.PlayerStats, style lipgloss.Style) string {
			favField := ""
			favCount := 0
			for field, count := range stat.FieldsCount {
				if count > favCount {
					favField = field
					favCount = count
				}
			}
			favField = fmt.Sprintf("%s (%dx)", favField, favCount)

			s := fill(stat.ID, 15) +
				fill(strconv.Itoa(stat.RanksCount[1]), 5) +
				fill(strconv.Itoa(stat.RanksCount[len(stat.RanksCount)]), 7) +
				fill(strconv.Itoa(stat.GamesPlayed), 6) +
				fill(strconv.FormatFloat(stat.AverageRank, 'f', 3, 64), 7) +
				fill(strconv.FormatFloat(stat.AverageScore, 'f', 1, 64), 8) +
				fill(fmt.Sprintf("%d (%s)", stat.HighestScore.Total, strings.Join(stat.HighestScore.Fields, " → ")), 22) +
				fill(favField, 10) +
				stat.AverageDuration.Truncate(time.Millisecond).String()
			return style.Render(s)
		}
		header = func() string {
			s := fill("Name", 15) +
				fill("Wins", 5) +
				fill("Losses", 7) +
				fill("Games", 6) +
				fill("⌀-Rank", 7) +
				fill("⌀-Score", 8) +
				fill("Max Score", 22) +
				fill("Fav Field", 10) +
				"⌀-Sec/move"
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

	lines = append(lines, headline("Player Statistics"), "")
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
			key.WithKeys("q", "esc"),
			key.WithHelp("q", "quit"),
		),
	})+" ("+styleInactive.Render(fmt.Sprintf("%3.f%%", s.viewport.ScrollPercent()*100))+")")

	return strings.Join(lines, "\n")
}
