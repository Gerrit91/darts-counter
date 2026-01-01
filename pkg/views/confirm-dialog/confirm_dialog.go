package confirm

import (
	"log/slog"
	"strings"

	"github.com/Gerrit91/darts-counter/pkg/views/common"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	model struct {
		log *slog.Logger

		yes    tea.Cmd
		no     tea.Cmd
		phrase string

		cursor  int
		choices []confirmDialogChoice
		help    help.Model
	}

	confirmDialogChoice string
)

const (
	confirmYes confirmDialogChoice = "Yes"
	confirmNo  confirmDialogChoice = "No"
)

func New(log *slog.Logger, view string, yes, no tea.Cmd) *model {
	return &model{
		log:    log,
		yes:    yes,
		no:     no,
		phrase: view,
		choices: []confirmDialogChoice{
			confirmNo,
			confirmYes,
		},
		help: common.NewHelp(),
	}
}

func (c *model) Init() tea.Cmd {
	c.cursor = 0
	return nil
}

func (c *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "n":
			return c, c.no
		case "y":
			return c, c.yes
		case "enter":
			switch c.choices[c.cursor] {
			case confirmYes:
				return c, c.yes
			default:
				return c, c.no
			}
		case "down":
			c.cursor++
			if c.cursor >= len(c.choices) {
				c.cursor = 0
			}
		case "up":
			c.cursor--
			if c.cursor < 0 {
				c.cursor = len(c.choices) - 1
			}
		default:
			return c, nil
		}
	}

	return c, nil
}

func (c *model) View() string {
	var lines []string

	lines = append(lines, "⚠️  "+common.StyleUnderlined.Render("Please confirm"), "")

	if c.phrase != "" {
		lines = append(lines, c.phrase, "")
	} else {
		lines = append(lines, "Are you sure?", "")
	}

	for i := range len(c.choices) {
		if c.cursor == i {
			selection := common.Fill("→", 3)
			lines = append(lines, common.StylePink.Render(selection)+common.StyleActive.Render(string(c.choices[i])))
			continue
		}

		selection := common.Fill("", 3)
		lines = append(lines, selection+common.StyleInactive.Render(string(c.choices[i])))
	}

	lines = append(lines, "", c.help.ShortHelpView([]key.Binding{
		key.NewBinding(
			key.WithKeys("y", "yes"),
			key.WithHelp("y", "yes"),
		),
		key.NewBinding(
			key.WithKeys("n", "no"),
			key.WithHelp("n", "no"),
		),
	}))

	return strings.Join(lines, "\n")
}
