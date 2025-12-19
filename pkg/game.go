package game

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/Gerrit91/darts-counter/pkg/checkout"
	"github.com/Gerrit91/darts-counter/pkg/config"
	"github.com/Gerrit91/darts-counter/pkg/player"
	"github.com/Gerrit91/darts-counter/pkg/stats"
	"github.com/Gerrit91/darts-counter/pkg/util"
	"github.com/google/uuid"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
)

type Game struct {
	id      string
	c       *util.Console
	t       config.GameType
	out     checkout.CheckoutType
	in      checkout.CheckinType
	players player.Players
	s       stats.Stats
}

func NewGame(console *util.Console, c *config.Config, s stats.Stats) (*Game, error) {
	uuid, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("unable to generate uuid: %w", err)
	}

	game := &Game{
		id:  uuid.String(),
		c:   console,
		t:   c.Game,
		out: c.Checkout,
		in:  c.Checkin,
		s:   s,
	}

	count := 0

	switch gt := game.t; gt {
	case config.GameType101, config.GameType301, config.GameType501, config.GameType701, config.GameType1001:
		count, _ = strconv.Atoi(string(gt))
	default:
		return nil, fmt.Errorf("unknown game: %s", game.t)
	}

	for _, p := range c.Players {
		game.players = append(game.players, player.New(p.Name, game.c, game.out, game.in, count, c.Statistics.Enabled))
	}

	return game, nil
}

func (g *Game) Run() {
	g.c.Println("starting new game of type %q with players: %s", g.t, strings.Join(g.players.Names(), ", "))

	var (
		start = time.Now()
		iter  = g.players.Iterator()
		rank  = 1
		moves []stats.Move
	)

	for {
		p, err := iter.Next()
		if err != nil {
			if errors.Is(err, player.ErrGameFinished) {
				if p != nil {
					p.SetRank(rank)
					g.c.Println("game finished, %s took last place", p.GetName())
				}
				g.showOverview(nil)

				var (
					playerNames []string
					ranks       = map[int]string{}
				)
				for _, p := range g.players {
					ranks[p.GetRank()] = p.GetName()
					playerNames = append(playerNames, p.GetName())
				}

				statsErr := g.s.CreateGameStats(&stats.GameStats{
					ID:       g.id,
					GameType: g.t,
					Checkin:  string(g.in),
					Checkout: string(g.out),
					Players:  playerNames,
					Rounds:   iter.GetRound(),
					Ranks:    ranks,
					Start:    start,
					End:      time.Now(),
					Moves:    moves,
				})
				if statsErr != nil {
					slog.Error("error persisting finished game to database", "error", err)
				}
			} else {
				slog.Error("error getting next player", "error", err)
			}

			break
		}

		g.showOverview(p)

		g.c.Println("round %d, player's turn: %s", iter.GetRound(), p.GetName())

		p.Move()

		score := stats.Score{
			Total: p.LastScore(),
		}
		for _, partial := range p.LastPartials() {
			score.Partials = append(score.Partials, partial.String())
		}

		moves = append(moves, stats.Move{
			Round:     iter.GetRound(),
			Player:    p.GetName(),
			Score:     score,
			Remaining: p.GetRemaining(),
			Duration:  p.MoveDuration().String(),
		})

		if p.HasFinished() {
			p.SetRank(rank)
			g.c.Println("%s took %d. place!", p.GetName(), p.GetRank())
			rank++
		}
	}
}

func (g *Game) showOverview(playerAtTurn *player.Player) {
	printerConfig := &printers.TablePrinterConfig{
		Markdown: true,
	}

	printer := printers.NewTablePrinter(printerConfig)

	printerConfig.ToHeaderAndRows = func(data any, wide bool) ([]string, [][]string, error) {
		players, ok := data.(player.Players)
		if !ok {
			return nil, nil, fmt.Errorf("unexpected type: %T", data)
		}

		header := []string{"turn", "name", "remaining", "rank", "checkout sequences"}
		var rows [][]string

		for _, p := range players {
			rank := ""
			if p.GetRank() > 0 {
				rank = strconv.Itoa(p.GetRank()) + "."
			}

			turn := ""
			if playerAtTurn != nil && p == playerAtTurn {
				turn = "X"
			}

			endingSequence := ""
			if p.GetRemaining() > 0 {
				variants := checkout.For(p.GetRemaining(), checkout.NewCalcLimitOption(3), checkout.NewCheckoutTypeOption(g.out))
				switch len(variants) {
				case 0:
				case 1, 2:
					endingSequence = variants.String()
				default:
					endingSequence = variants[:3].String() + ", ..."
				}
			}

			row := []string{turn, p.GetName(), strconv.Itoa(p.GetRemaining()), rank, endingSequence}

			rows = append(rows, row)
		}

		return header, rows, nil
	}

	printer.Print(g.players)
}
