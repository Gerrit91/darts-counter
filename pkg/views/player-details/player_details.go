package playerdetails

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/Gerrit91/darts-counter/pkg/datastore"
	"github.com/Gerrit91/darts-counter/pkg/views/common"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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
