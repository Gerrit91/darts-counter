package playerdetails

import (
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Gerrit91/darts-counter/pkg/checkout"
	"github.com/Gerrit91/darts-counter/pkg/datastore"
	"github.com/Gerrit91/darts-counter/pkg/views/common"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	log *slog.Logger
	ds  datastore.Datastore

	ps *datastore.PlayerStats

	viewport viewport.Model
	help     help.Model

	backTo tea.Cmd
}

func New(log *slog.Logger, ds datastore.Datastore) *Model {
	return &Model{
		log:      log,
		ds:       ds,
		viewport: viewport.New(0, 20),
		backTo:   common.SwitchViewTo(common.PlayerListView),
		help:     common.NewHelp(),
	}
}

func (s *Model) Init() tea.Cmd {
	s.viewport.GotoTop()
	return tea.WindowSize()
}

func (s *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return s, s.backTo
		case "g":
			s.viewport.GotoTop()
		case "G":
			s.viewport.GotoBottom()
		}
	case tea.WindowSizeMsg:
		headerHeight := 2
		footerHeight := 1
		verticalMarginHeight := headerHeight + footerHeight
		s.viewport.Width = msg.Width
		s.viewport.Height = msg.Height - verticalMarginHeight
	}

	var cmd tea.Cmd
	s.viewport, cmd = s.viewport.Update(msg)

	return s, cmd
}

func (s *Model) View() string {
	var (
		lines         []string
		viewportLines []string
		ps            = s.ps
	)

	ranksTable := common.NewTable()
	var orderedRanks []int
	for rank := range ps.RanksCount {
		orderedRanks = append(orderedRanks, rank)
	}
	sort.SliceStable(orderedRanks, func(i, j int) bool {
		return orderedRanks[i] < orderedRanks[j]
	})
	for _, rank := range orderedRanks {
		ranksTable.Row(fmt.Sprintf("%d. %d", rank, ps.RanksCount[rank]))
	}

	fieldsTable := common.NewTable().StyleFunc(func(row, col int) lipgloss.Style {
		switch {
		case col == 0:
			return common.StylePink
		case row == -1:
			return common.StylePink
		}
		return common.StyleInactive
	})

	headers := []string{" "}
	for _, score := range checkout.Singles() {
		headers = append(headers, common.Fill(score.String(), 2))
	}
	fieldsTable.Headers(headers...)

	for _, m := range []checkout.Multiplier{checkout.None, checkout.Double, checkout.Triple} {
		row := []string{string(m)}
		for _, score := range checkout.Singles() {
			if m == checkout.Triple && score.Value() == checkout.BullsEye {
				row = append(row, "")
				continue
			}
			row = append(row, strconv.Itoa(ps.FieldsCount[score.WithMultiplier(m).String()]))
		}
		fieldsTable.Row(row...)
	}

	infoTable := common.NewTable()
	infoTable.Row("Games Played:", strconv.Itoa(ps.GamesPlayed))
	infoTable.Row("Total Moves:", strconv.Itoa(ps.TotalMoves))
	infoTable.Row("Total Move Time:", ps.TotalDuration.Truncate(time.Second).String())
	infoTable.Row("⌀-Duration per Move:", ps.AverageDuration.Truncate(time.Millisecond).String())

	viewportLines = append(viewportLines, infoTable.Render())

	viewportLines = append(viewportLines, "", fmt.Sprintf("Ranks (⌀ %s):", strconv.FormatFloat(ps.AverageRank, 'f', 3, 64)))
	viewportLines = append(viewportLines, ranksTable.Render())

	viewportLines = append(viewportLines, "", "Field Counts:")
	viewportLines = append(viewportLines, fieldsTable.Render())
	viewportLines = append(viewportLines, "⌀-Score: "+strconv.FormatFloat(ps.AverageScore, 'f', 1, 64))
	viewportLines = append(viewportLines, "Highest Score: "+fmt.Sprintf("%d (%s)", ps.HighestScore.Total, strings.Join(ps.HighestScore.Fields, " → ")))

	if s.viewport.Height > 0 { // otherwise it crashes
		s.viewport.SetContent(strings.Join(viewportLines, "\n"))
	}

	lines = append(lines, common.Headline(ps.ID))
	lines = append(lines, s.viewport.View())

	lines = append(lines, s.help.ShortHelpView([]key.Binding{
		key.NewBinding(
			key.WithKeys("up", "down"),
			key.WithHelp("↑/↓", "up/down"),
		),

		key.NewBinding(
			key.WithKeys("pgup", "pgdown"),
			key.WithHelp("page up/down", "page up/down"),
		),
		key.NewBinding(
			key.WithKeys("g", "G"),
			key.WithHelp("g/G", "top/bottom"),
		),
		key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
	})+common.StyleHelp.Render(fmt.Sprintf(" (%3.f%%)", s.viewport.ScrollPercent()*100)))

	return strings.Join(lines, "\n")
}

func (s *Model) SetBackTo(cmd tea.Cmd) {
	s.backTo = cmd
}

func (s *Model) SetPlayerStats(ps *datastore.PlayerStats) {
	s.ps = ps
}
