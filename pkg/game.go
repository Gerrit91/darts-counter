package game

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Gerrit91/darts-counter/pkg/checkout"
	"github.com/Gerrit91/darts-counter/pkg/config"
	"github.com/Gerrit91/darts-counter/pkg/datastore"
	"github.com/Gerrit91/darts-counter/pkg/player"
	"github.com/google/uuid"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	game struct {
		cfg *config.Config
		log *slog.Logger
		ds  datastore.Datastore

		id            string
		players       player.Players
		currentPlayer *player.Player
		start         time.Time
		startMove     time.Time
		iter          *player.Iterator
		rank          int
		moves         []datastore.Move
		err           error
		msg           string
		finished      bool

		textInput textinput.Model
		help      help.Model
		show      *showGameModel
	}

	undoMoveMsg struct{}
)

func undoMove() tea.Msg {
	return undoMoveMsg{}
}

func newGame(log *slog.Logger, c *config.Config, ds datastore.Datastore, show *showGameModel) (*game, error) {
	uuid, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("unable to generate uuid: %w", err)
	}

	count := 0

	switch gt := c.Game; gt {
	case config.GameType101, config.GameType301, config.GameType501, config.GameType701, config.GameType1001:
		count, _ = strconv.Atoi(string(gt))
	default:
		return nil, fmt.Errorf("unknown game: %s", c.Game)
	}

	var players player.Players
	for _, p := range c.Players {
		players = append(players, player.New(p.Name, c.Checkout, c.Checkin, count, ds.Enabled()))
	}

	playerIterator := players.Iterator()
	currentPlayer, err := playerIterator.Next()
	if err != nil {
		return nil, err
	}

	now := time.Now()

	return &game{
		cfg:           c,
		log:           log,
		ds:            ds,
		id:            uuid.String(),
		players:       players,
		currentPlayer: currentPlayer,
		start:         now,
		startMove:     now,
		iter:          playerIterator,
		rank:          1,
		moves:         nil,
		err:           nil,
		msg:           "",
		finished:      false,
		textInput:     newTextInput(),
		help:          newHelp(),
		show:          show,
	}, nil
}

func (g *game) Init() tea.Cmd {
	g.show.backTo = switchViewTo(gameView)
	return g.textInput.Cursor.BlinkCmd()
}

func (g *game) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		g.help.Width = msg.Width
		return g, nil
	case cursor.BlinkMsg:
		var cmd tea.Cmd
		g.textInput, cmd = g.textInput.Update(msg)
		return g, cmd
	case undoMoveMsg:
		if len(g.moves) == 0 {
			g.err = fmt.Errorf("cannot go back any further, no previous moves")
			return g, nil
		}

		lastIdx := len(g.moves) - 1
		lastMove := g.moves[lastIdx]

		lastPlayer, err := g.iter.SetBackTo(lastMove.Player)
		if err != nil {
			g.err = err
			return g, nil
		}

		err = lastPlayer.Edit(lastPlayer.GetRemaining() + lastMove.Score.Total)
		if err != nil {
			g.err = err
			return g, nil
		}

		g.finished = false
		g.currentPlayer.SetRank(0)
		g.rank = lastPlayer.GetRank()
		lastPlayer.SetRank(0)
		g.moves = g.moves[:lastIdx]
		g.currentPlayer = lastPlayer

		return g, nil
	case tea.KeyMsg:
		g.err = nil
		g.msg = ""

		switch msg.String() {
		case "q", "esc":
			return g, switchViewTo(closeGameDialogView)
		case "v":
			g.show.gs = *g.gameStats()
			return g, switchViewTo(showGame)
		case "u":
			return g, switchViewTo(undoMoveView)
		case "s":
			g.tick(nil, 0)
			return g, nil
		case "enter":
			defer func() {
				g.textInput.Reset()
			}()

			if g.finished {
				if err := g.persist(); err != nil {
					g.log.Error("error persisting finished game to database", "error", err)
				}

				return g, switchViewTo(mainMenuView)
			}

			scores, total, err := g.parseScore(g.textInput.Value())
			if err != nil {
				g.err = err
				return g, nil
			}

			g.tick(scores, total)

			return g, nil
		default:
			var cmd tea.Cmd
			g.textInput, cmd = g.textInput.Update(msg)

			return g, cmd
		}
	}

	return g, nil
}

func (g *game) View() string {
	var (
		lines        []string
		longestName  int
		longestScore int
	)

	for _, p := range g.players {
		if len(p.GetName()) > longestName {
			longestName = len(p.GetName())
		}
		if r := strconv.Itoa(p.GetRemaining()); len(r) > longestScore {
			longestScore = len(r)
		}
	}

	lines = append(lines, headline(fmt.Sprintf("Game %s: Round %d", g.cfg.Game, g.iter.GetRound())))

	lines = append(lines, "")

	for _, p := range g.players {
		var (
			playerName         = p.GetName()
			infos              []string
			currentPlayerArrow = ""

			scoreStyle = styleGreen
		)

		playerStyle := styleInactive
		if g.currentPlayer != nil && p == g.currentPlayer {
			currentPlayerArrow = "→"
			playerStyle = styleActive
		}

		if p.GetRank() > 0 {
			currentPlayerArrow = strconv.Itoa(p.GetRank()) + "."
		}

		if len(g.moves) > 0 {
			var moves []datastore.Move
			moves = append(moves, g.moves...)
			slices.Reverse(moves)

			for _, m := range moves {
				if m.Player == p.GetName() {
					infos = append(infos, stylePink.Render(fmt.Sprintf("(—%d)", m.Score.Total)))
					break
				}
			}
		}

		if p.GetRemaining() > 0 {
			variants := checkout.For(p.GetRemaining(), checkout.NewCalcLimitOption(3), checkout.NewCheckoutTypeOption(g.cfg.Checkout))
			switch len(variants) {
			case 0:
			case 1, 2:
				infos = append(infos, styleInactive.Render(variants.String()))
			default:
				infos = append(infos, styleInactive.Render(variants[:3].String()+", ..."))
			}
		}

		lines = append(lines,
			stylePink.Render(fill(currentPlayerArrow, 3))+
				playerStyle.Render(fill(playerName, longestName+8))+
				scoreStyle.Render(fill(strconv.Itoa(p.GetRemaining()), longestScore+3))+
				strings.Join(infos, " "),
		)
	}

	lines = append(lines, "")

	if g.err != nil {
		lines = append(lines, styleError.Render(g.err.Error()))
	}
	if g.msg != "" {
		lines = append(lines, g.msg)
	}

	if g.finished {
		lines = append(lines, "Game finished.")
		lines = append(lines, g.help.ShortHelpView([]key.Binding{
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "return to main menu"),
			),
			key.NewBinding(
				key.WithKeys("u"),
				key.WithHelp("u", "undo last move"),
			),
			key.NewBinding(
				key.WithKeys("v"),
				key.WithHelp("v", "view move history"),
			),
			key.NewBinding(
				key.WithKeys("q", "esc"),
				key.WithHelp("q", "quit"),
			),
		}))
	} else {
		lines = append(lines, "Enter score:")
		lines = append(lines, g.textInput.View())
		lines = append(lines, g.help.ShortHelpView([]key.Binding{
			key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "skip player"),
			),
			key.NewBinding(
				key.WithKeys("u"),
				key.WithHelp("u", "undo last move"),
			),
			key.NewBinding(
				key.WithKeys("v"),
				key.WithHelp("v", "view move history"),
			),
			key.NewBinding(
				key.WithKeys("q", "esc"),
				key.WithHelp("q", "quit"),
			),
		}))
	}

	return strings.Join(lines, "\n")
}

func (g *game) tick(scores []*checkout.Score, total int) {
	if g.finished {
		return
	}

	p := g.currentPlayer

	err := p.Move(scores, total)
	if err != nil {
		if errors.Is(err, player.ErrInvalidInput) {
			g.err = err
			return
		}

		// on other errors, game can continue
		g.msg = err.Error()
		err = nil
	}

	if p.HasFinished() {
		p.SetRank(g.rank)
		g.msg = fmt.Sprintf("%s took %d. place!", p.GetName(), p.GetRank())
		g.rank++
	}

	statsScore := datastore.Score{
		Total: total,
	}
	for _, score := range scores {
		statsScore.Fields = append(statsScore.Fields, score.String())
	}

	since := time.Since(g.startMove)
	g.startMove = g.startMove.Add(since)

	g.moves = append(g.moves, datastore.Move{
		Round:     g.iter.GetRound(),
		Player:    p.GetName(),
		Score:     statsScore,
		Remaining: p.GetRemaining(),
		Duration:  since.String(),
	})

	p, err = g.iter.Next()
	g.currentPlayer = p

	if errors.Is(err, player.ErrOnlyOnePlayerLeft) {
		if p != nil {
			p.SetRank(g.rank)
		}

		g.finished = true
		return
	}

	if err != nil {
		g.err = err
		return
	}
}

func (g *game) parseScore(input string) ([]*checkout.Score, int, error) {
	var (
		// allow both comma and space separated
		segments = strings.Fields(strings.Join(strings.Split(strings.TrimSpace(input), ","), " "))
		total    int
		scores   []*checkout.Score
	)

	switch len(segments) {
	case 0:
		return nil, 0, fmt.Errorf("no points entered")
	case 1:
		s, err := checkout.ParseScore(input)
		if err == nil {
			total = s.Value()
			scores = append(scores, s)
		} else {
			// user entered summed up score
			total, err = strconv.Atoi(input)
			if err != nil {
				return nil, 0, fmt.Errorf("unable to parse input (%q), please enter again", err.Error())
			}

			if g.ds.Enabled() {
				return nil, 0, fmt.Errorf("when statistics are enabled it's not allowed to enter summed up scores")
			}
		}
	case 2, 3:
		var (
			sum   int
			score *checkout.Score
			err   error
		)

		for _, segment := range segments {
			score, err = checkout.ParseScore(segment)
			if err != nil {
				break
			}

			sum += score.Value()
			scores = append(scores, score)
		}

		if err != nil {
			return nil, 0, fmt.Errorf("unable to parse input (%q), please enter again", err.Error())
		}

		total = sum
	default:
		return nil, 0, fmt.Errorf("no more than three throws are allowed, please enter again")
	}

	return scores, total, nil
}

func (g *game) persist() error {
	err := g.ds.CreateGameStats(g.gameStats())
	if err != nil {
		return err
	}

	g.log.Info("saved game stats to database")

	return nil
}

func (g *game) gameStats() *datastore.GameStats {
	var (
		playerNames []string
		ranks       = map[int]string{}
	)
	for _, p := range g.players {
		if g.finished {
			ranks[p.GetRank()] = p.GetName()
		}
		playerNames = append(playerNames, p.GetName())
	}

	return &datastore.GameStats{
		ID:       g.id,
		GameType: g.cfg.Game,
		Checkin:  string(g.cfg.Checkin),
		Checkout: string(g.cfg.Checkout),
		Players:  playerNames,
		Rounds:   g.iter.GetRound(),
		Ranks:    ranks,
		Start:    g.start,
		End:      time.Now(),
		Moves:    g.moves,
	}
}
