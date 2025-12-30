package game

import (
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Gerrit91/darts-counter/pkg/datastore"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"sigs.k8s.io/yaml"
)

type showGameModel struct {
	log *slog.Logger
	ds  datastore.Datastore

	viewport viewport.Model
	gs       datastore.GameStats

	backTo tea.Cmd
}

func newShowGameModel(log *slog.Logger, ds datastore.Datastore) *showGameModel {
	return &showGameModel{
		log:      log,
		ds:       ds,
		viewport: viewport.New(0, 20),
		backTo:   switchViewTo(showGames),
	}
}

func (s *showGameModel) Init() tea.Cmd {
	s.viewport.GotoTop()
	return tea.WindowSize()
}

func (s *showGameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return s, s.backTo
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

func (s *showGameModel) View() string {
	var (
		lines         []string
		viewportLines []string
	)

	viewportLines = append(viewportLines, fill("ID:", 6)+s.gs.ID)
	viewportLines = append(viewportLines, fill("Type:", 6)+fmt.Sprintf("%s (%s, %s)", s.gs.GameType, s.gs.Checkin, s.gs.Checkout))
	viewportLines = append(viewportLines, "")

	viewportLines = append(viewportLines, fill("Rounds:", 8)+strconv.Itoa(s.gs.Rounds))
	viewportLines = append(viewportLines, fill("Start:", 8)+s.gs.Start.Format(time.DateTime))
	viewportLines = append(viewportLines, fill("End:", 8)+s.gs.End.Format(time.DateTime))
	viewportLines = append(viewportLines, fill("Length:", 8)+s.gs.End.Sub(s.gs.Start).Truncate(time.Millisecond).String())
	viewportLines = append(viewportLines, "")

	viewportLines = append(viewportLines, "Players: "+strings.Join(s.gs.Players, ", "))
	viewportLines = append(viewportLines, "Ranks:")
	type rank struct {
		rank   int
		player string
	}
	var ranks []rank
	for k, v := range s.gs.Ranks {
		ranks = append(ranks, rank{
			rank:   k,
			player: v,
		})
	}
	sort.SliceStable(ranks, func(i, j int) bool {
		return ranks[i].rank < ranks[j].rank
	})
	for _, r := range ranks {
		viewportLines = append(viewportLines, "   "+fmt.Sprintf("%d. %s", r.rank, r.player))
	}
	viewportLines = append(viewportLines, "")

	viewportLines = append(viewportLines, "Moves:")
	rawMoves, _ := yaml.Marshal(s.gs.Moves)
	viewportLines = append(viewportLines, string(rawMoves))

	s.viewport.SetContent(strings.Join(viewportLines, "\n"))

	lines = append(lines, headline("Game Details"), "")
	lines = append(lines, s.viewport.View())
	lines = append(lines, styleInactive.Render(fmt.Sprintf("%3.f%%", s.viewport.ScrollPercent()*100)))

	return strings.Join(lines, "\n")
}
