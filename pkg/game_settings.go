package game

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/Gerrit91/darts-counter/pkg/checkout"
	"github.com/Gerrit91/darts-counter/pkg/config"
	"github.com/Gerrit91/darts-counter/pkg/datastore"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	gameSettingsModel struct {
		log *slog.Logger
		ds  datastore.Datastore

		settings  *datastore.GameSettings
		choices   []any
		cursor    int
		err       error
		showInput string

		textInput textinput.Model
		help      help.Model
	}

	settingsChoice string
	playerChoice   struct {
		name string
		idx  int
	}
)

var (
	gameTypeSettings           settingsChoice = "game-type"
	checkinSettings            settingsChoice = "check-in"
	checkoutSettings           settingsChoice = "check-out"
	playerSettings             settingsChoice = "player"
	saveSettings               settingsChoice = "save"
	leaveSettingsWithoutSaving settingsChoice = "leave-without-saving"
)

func newGameSettings(log *slog.Logger, ds datastore.Datastore) *gameSettingsModel {
	return &gameSettingsModel{
		log:       log,
		ds:        ds,
		help:      newHelp(),
		textInput: newTextInput(),
	}
}

func (g *gameSettingsModel) Init() tea.Cmd {
	settings, err := g.ds.GetGameSettings()
	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			settings = &datastore.GameSettings{
				Type:     config.GameType301,
				Checkout: checkout.CheckoutTypeStraightOut,
				Checkin:  checkout.CheckinTypeStraightIn,
				Players: []datastore.Player{
					{Name: "Player 1"},
					{Name: "Player 2"},
				},
			}
		} else {
			g.err = err
		}
	}

	g.settings = settings
	g.updateChoices()
	g.cursor = 0

	return g.textInput.Cursor.BlinkCmd()
}

func (g *gameSettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		gameTypes = []config.GameType{
			config.GameType101,
			config.GameType301,
			config.GameType501,
			config.GameType701,
			config.GameType1001,
		}
		gameTypeToggle = func(left bool) {
			idx := slices.IndexFunc(gameTypes, func(gt config.GameType) bool {
				return gt == g.settings.Type
			})

			if idx == -1 {
				idx = 0
			}

			if left {
				to := idx + -1
				if to < 0 {
					to = len(gameTypes) - 1
				}
				g.settings.Type = gameTypes[to%len(gameTypes)]
			} else {
				g.settings.Type = gameTypes[(idx+1)%len(gameTypes)]
			}
		}
		checkoutToggle = func() {
			if g.settings.Checkout == checkout.CheckoutTypeStraightOut {
				g.settings.Checkout = checkout.CheckoutTypeDoubleOut
			} else {
				g.settings.Checkout = checkout.CheckoutTypeStraightOut
			}
		}
		checkinToggle = func() {
			if g.settings.Checkin == checkout.CheckinTypeStraightIn {
				g.settings.Checkin = checkout.CheckinTypeDoubleIn
			} else {
				g.settings.Checkin = checkout.CheckinTypeStraightIn
			}
		}
		rotatePlayers = func() {
			if len(g.settings.Players) > 1 {
				g.settings.Players = append([]datastore.Player{g.settings.Players[len(g.settings.Players)-1]}, g.settings.Players[:len(g.settings.Players)-1]...)
				g.updateChoices()
			}
		}
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		g.err = nil

		if g.showInput != "" {
			switch msg.String() {
			case "esc":
				g.showInput = ""
				return g, nil
			case "enter":
				switch g.choices[g.cursor] {
				case playerSettings:
					{
						g.showInput = ""
						g.settings.Players = append(g.settings.Players, datastore.Player{
							Name: g.textInput.Value(),
						})
						g.updateChoices()
						g.textInput.Reset()
						return g, nil
					}
				default:
					switch choice := g.choices[g.cursor].(type) {
					case playerChoice:
						g.showInput = ""
						g.settings.Players[choice.idx].Name = g.textInput.Value()
						g.updateChoices()
						g.textInput.Reset()
						return g, nil
					}
				}
			}

			var cmd tea.Cmd
			g.textInput, cmd = g.textInput.Update(msg)

			return g, cmd
		}

		switch msg.String() {
		case "esc":
			return g, switchViewTo(mainMenuView) // probably add confirm dialog if a setting was changed
		case "enter":
			switch g.choices[g.cursor] {
			case leaveSettingsWithoutSaving:
				return g, switchViewTo(mainMenuView)
			case saveSettings:
				err := g.ds.UpdateGameSettings(g.settings)
				if err != nil {
					g.err = err
					return g, nil
				}

				return g, switchViewTo(mainMenuView) // if coming from start new game, it should probably go directly to the game
			case gameTypeSettings:
				gameTypeToggle(false)
			case checkinSettings:
				checkinToggle()
			case checkoutSettings:
				checkoutToggle()
			case playerSettings:
				rotatePlayers()
			default:
				switch choice := g.choices[g.cursor].(type) {
				case playerChoice:
					g.showInput = "Edit Player Name:"
					g.textInput.SetValue(choice.name)
					return g, nil
				}
			}
			return g, nil
		case "right":
			switch g.choices[g.cursor] {
			case gameTypeSettings:
				gameTypeToggle(false)
			case checkinSettings:
				checkinToggle()
			case checkoutSettings:
				checkoutToggle()
			}
		case "left":
			switch g.choices[g.cursor] {
			case gameTypeSettings:
				gameTypeToggle(true)
			case checkinSettings:
				checkinToggle()
			case checkoutSettings:
				checkoutToggle()
			}
		case "down":
			g.cursor++
			if g.cursor >= len(g.choices) {
				g.cursor = 0
			}
		case "up":
			g.cursor--
			if g.cursor < 0 {
				g.cursor = len(g.choices) - 1
			}
		case "+":
			switch g.choices[g.cursor] {
			case playerSettings:
				g.showInput = "Enter Player Name:"
				return g, nil
			}
		case "-":
			switch choice := g.choices[g.cursor].(type) {
			case playerChoice:
				g.settings.Players = slices.Delete(g.settings.Players, choice.idx, choice.idx+1)
				g.updateChoices()
				g.cursor--

				return g, nil
			}
		case "pgup":
			switch choice := g.choices[g.cursor].(type) {
			case playerChoice:
				var (
					from = choice.idx
					to   = (choice.idx - 1) % len(g.settings.Players)
				)

				if to == -1 {
					to = len(g.settings.Players) - 1
				}

				g.settings.Players[from], g.settings.Players[to] = g.settings.Players[to], g.settings.Players[from]
				g.updateChoices()

				g.cursor--
				if to == len(g.settings.Players)-1 {
					g.cursor += len(g.settings.Players)
				}

				return g, nil
			}
		case "pgdown":
			switch choice := g.choices[g.cursor].(type) {
			case playerChoice:
				var (
					from = choice.idx
					to   = (choice.idx + 1) % len(g.settings.Players)
				)

				g.settings.Players[from], g.settings.Players[to] = g.settings.Players[to], g.settings.Players[from]
				g.updateChoices()

				g.cursor++
				if to == 0 {
					g.cursor -= len(g.settings.Players)
				}

				return g, nil
			}
		default:
			return g, nil
		}

		return g, nil
	}

	var cmd tea.Cmd
	g.textInput, cmd = g.textInput.Update(msg)

	return g, cmd
}

func (g *gameSettingsModel) View() string {
	var (
		lines   []string
		helpMap = map[any][]key.Binding{
			gameTypeSettings: {
				key.NewBinding(
					key.WithKeys("up", "down"),
					key.WithHelp("↑/↓", "up/down"),
				),
				key.NewBinding(
					key.WithKeys("enter", "left", "right"),
					key.WithHelp("←/→", "toggle"),
				),
			},
			checkinSettings: {
				key.NewBinding(
					key.WithKeys("up", "down"),
					key.WithHelp("↑/↓", "up/down"),
				),
				key.NewBinding(
					key.WithKeys("enter", "left", "right"),
					key.WithHelp("←/→", "toggle"),
				),
			},
			checkoutSettings: {
				key.NewBinding(
					key.WithKeys("up", "down"),
					key.WithHelp("↑/↓", "up/down"),
				),
				key.NewBinding(
					key.WithKeys("enter", "left", "right"),
					key.WithHelp("←/→", "toggle"),
				),
			},
			saveSettings: {
				key.NewBinding(
					key.WithKeys("enter"),
					key.WithHelp("enter", "save"),
				),
			},
			leaveSettingsWithoutSaving: {
				key.NewBinding(
					key.WithKeys("enter"),
					key.WithHelp("enter", "leave without saving"),
				),
			},
			playerSettings: {
				key.NewBinding(
					key.WithKeys("up", "down"),
					key.WithHelp("↑/↓", "up/down"),
				),
				key.NewBinding(
					key.WithKeys("+"),
					key.WithHelp("+", "add player"),
				),
				key.NewBinding(
					key.WithKeys("enter"),
					key.WithHelp("enter", "rotate players"),
				),
			},
		}
		helpKeyBinding []key.Binding
	)

	lines = append(lines, headline("Game Settings"), "")

	if g.settings != nil {
		for i := range len(g.choices) {
			selection := fill("", 3)
			style := styleInactive
			if g.cursor == i {
				selection = stylePink.Render(fill("→", 3))
				style = styleActive
				helpKeyBinding = helpMap[g.choices[i]]
			}

			switch choice := g.choices[i]; choice {
			case gameTypeSettings:
				lines = append(lines, selection+style.Render(fill("Type:", 13)+string(g.settings.Type)))
			case checkinSettings:
				lines = append(lines, selection+style.Render(fill("Check-In:", 12), string(g.settings.Checkin)))
			case checkoutSettings:
				lines = append(lines, selection+style.Render(fill("Check-Out:", 12), string(g.settings.Checkout)))
			case playerSettings:
				lines = append(lines, selection+style.Render("Players:"))
			case saveSettings:
				lines = append(lines, selection+style.Render("Save"))
			case leaveSettingsWithoutSaving:
				lines = append(lines, selection+style.Render("Leave"))
			default:
				switch choice := choice.(type) {
				case playerChoice:
					if g.cursor == i {
						helpKeyBinding = []key.Binding{
							key.NewBinding(
								key.WithKeys("up", "down"),
								key.WithHelp("↑/↓", "up/down"),
							),
							key.NewBinding(
								key.WithKeys("-", "del", "d"),
								key.WithHelp("-", "delete"),
							),
							key.NewBinding(
								key.WithKeys("enter"),
								key.WithHelp("enter", "rename"),
							),
							key.NewBinding(
								key.WithKeys("pgup", "pgdown"),
								key.WithHelp("page up/down", "toggle"),
							),
						}
					}
					lines = append(lines, style.Render(fmt.Sprintf("   %s%d. %s", selection, choice.idx+1, choice.name)))
				}
			}
		}
	}

	if g.showInput != "" {
		lines = append(lines, g.showInput, g.textInput.View())
	}

	if g.err != nil {
		lines = append(lines, "", styleError.Render(g.err.Error()))
	}

	lines = append(lines, "", g.help.ShortHelpView(helpKeyBinding))

	return strings.Join(lines, "\n")
}

func (g *gameSettingsModel) updateChoices() {
	g.choices = []any{
		gameTypeSettings,
		checkinSettings,
		checkoutSettings,
		playerSettings,
	}

	for i, p := range g.settings.Players {
		g.choices = append(g.choices, playerChoice{
			name: p.Name,
			idx:  i,
		})
	}

	g.choices = append(g.choices,
		saveSettings,
		leaveSettingsWithoutSaving,
	)
}
