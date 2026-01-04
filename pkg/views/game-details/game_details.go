package gamedetails

import (
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

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

	gs datastore.GameStats

	viewport viewport.Model
	help     help.Model

	backTo tea.Cmd
}

func New(log *slog.Logger, ds datastore.Datastore) *Model {
	return &Model{
		log:      log,
		ds:       ds,
		viewport: viewport.New(0, 20),
		backTo:   common.SwitchViewTo(common.GameDetailsView),
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
		gs            = s.gs
	)

	t1 := common.NewTable().StyleFunc(func(row, col int) lipgloss.Style {
		if col == 0 {
			return common.StyleInactive
		}
		return common.StyleActive
	})
	t1.Row("ID:", gs.ID)
	t1.Row("Type:", fmt.Sprintf("%s (%s, %s)", gs.GameType, gs.Checkin, gs.Checkout))
	t1.Row("Players: ", strings.Join(s.gs.Players, ", "))
	viewportLines = append(viewportLines, t1.Render())

	t2 := common.NewTable().StyleFunc(func(row, col int) lipgloss.Style {
		if col == 0 {
			return common.StyleInactive
		}
		return common.StyleActive
	})
	t2.Row("Rounds:", strconv.Itoa(gs.Rounds))
	t2.Row("Start:", gs.Start.Format(time.DateTime))
	t2.Row("End:", gs.End.Format(time.DateTime))
	t2.Row("Length:", gs.End.Sub(gs.Start).Truncate(time.Millisecond).String())
	viewportLines = append(viewportLines, t2.Render(), "")

	viewportLines = append(viewportLines, common.StyleInactive.Render("Ranks:"))

	type rank struct {
		rank   int
		player string
	}
	var (
		ranks       []rank
		ranksColors = map[int]string{}
	)
	for k, v := range s.gs.Ranks {
		ranks = append(ranks, rank{
			rank:   k,
			player: v,
		})
		ranksColors[k] = ""
	}
	sort.SliceStable(ranks, func(i, j int) bool {
		return ranks[i].rank < ranks[j].rank
	})

	common.DistributeColors(string(common.ColorGreen), string(common.ColorInactive), ranksColors)

	for _, r := range ranks {
		viewportLines = append(viewportLines, "   "+fmt.Sprintf("%s. %s",
			lipgloss.NewStyle().Foreground(lipgloss.Color(ranksColors[r.rank])).Render(strconv.Itoa(r.rank)),
			common.StyleActive.Render(r.player)))
	}
	viewportLines = append(viewportLines, "")

	viewportLines = append(viewportLines, "Moves:")

	t3 := common.NewTable().Headers(
		"Round",
		"Player",
		"Score",
		"Fields",
		"Remaining",
		"Duration",
	).StyleFunc(func(row, col int) lipgloss.Style {
		switch {
		case row == -1:
			return common.StyleInactive
		case col == 0:
			return common.StyleInactive
		default:
			return lipgloss.NewStyle()
		}
	})
	for _, move := range gs.Moves {
		duration := move.Duration
		if d, err := time.ParseDuration(duration); err == nil {
			duration = d.Truncate(time.Millisecond).String()
		}

		t3 = t3.Row(
			strconv.Itoa(move.Round),
			move.Player,
			fmt.Sprintf("%s (%s)", common.StylePink.Render("—"+strconv.Itoa(move.Score.Total)), common.StyleGreen.Render(strconv.Itoa(move.Remaining+move.Score.Total))),
			strings.Join(move.Score.Fields, " → "),
			strconv.Itoa(move.Remaining),
			duration,
		)
	}

	viewportLines = append(viewportLines, t3.Render())

	if s.viewport.Height > 0 { // otherwise it crashes
		s.viewport.SetContent(strings.Join(viewportLines, "\n"))
	}

	lines = append(lines, common.Headline("Game Details"))
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

func (s *Model) SetGameStats(gs datastore.GameStats) {
	s.gs = gs
}
