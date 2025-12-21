package game

import (
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	confirmDialogModel struct {
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

func newConfirmDialog(log *slog.Logger, view string, yes, no tea.Cmd) *confirmDialogModel {
	return &confirmDialogModel{
		log:    log,
		yes:    yes,
		no:     no,
		phrase: view,
		choices: []confirmDialogChoice{
			confirmNo,
			confirmYes,
		},
		help: newHelp(),
	}
}

func (c *confirmDialogModel) Init() tea.Cmd {
	c.cursor = 0
	return nil
}

func (c *confirmDialogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (c *confirmDialogModel) View() string {
	var lines []string

	lines = append(lines, "⚠️  "+styleUnderlined.Render("Please confirm"), "")

	if c.phrase != "" {
		lines = append(lines, c.phrase, "")
	} else {
		lines = append(lines, "Are you sure?", "")
	}

	for i := 0; i < len(c.choices); i++ {
		if c.cursor == i {
			selection := fill("→", 3)
			lines = append(lines, stylePink.Render(selection)+styleActive.Render(string(c.choices[i])))
			continue
		}

		selection := fill("", 3)
		lines = append(lines, selection+styleInactive.Render(string(c.choices[i])))
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
