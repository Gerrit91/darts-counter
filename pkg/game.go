package game

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/Gerrit91/darts-counter/pkg/checkout"
	"github.com/Gerrit91/darts-counter/pkg/player"
	"github.com/Gerrit91/darts-counter/pkg/util"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/olekukonko/tablewriter"
)

type GameType string

const (
	GameType301  GameType = "301"
	GameType501  GameType = "501"
	GameType701  GameType = "701"
	GameType1001 GameType = "1001"
)

type Game struct {
	c       *util.Console
	t       GameType
	out     checkout.CheckoutType
	players player.Players
}

func NewGame(c *util.Console) (*Game, error) {
	config, err := ReadConfig()
	if err != nil {
		return nil, err
	}

	game := &Game{
		c:   c,
		t:   config.Game,
		out: config.Checkout,
	}

	count := 0

	switch gt := game.t; gt {
	case GameType301, GameType501, GameType701, GameType1001:
		count, _ = strconv.Atoi(string(gt))
	default:
		return nil, fmt.Errorf("unknown game: %s", game.t)
	}

	for _, p := range config.Players {
		game.players = append(game.players, player.New(p.Name, game.c, game.out, count))
	}

	return game, nil
}

func (g *Game) Run() {
	g.c.Println("starting new game of type %q with players: %s", g.t, strings.Join(g.players.Names(), ", "))

	iter := g.players.Iterator()
	rank := 1

	for {
		p, err := iter.Next()
		if err != nil {
			if errors.Is(err, player.ErrGameFinished) {
				if p != nil {
					p.SetRank(rank)
					g.c.Println("game finished, %s took last place", p.GetName())
				}
				g.showOverview(nil)
			} else {
				slog.Error("error getting next player", "error", err)
			}

			break
		}

		g.showOverview(p)

		g.c.Println("player's turn: %s", p.GetName())

		p.Move()

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
			p := p

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
					endingSequence = variants[:2].String() + ", ..."
				}
			}

			row := []string{turn, p.GetName(), strconv.Itoa(p.GetRemaining()), rank, endingSequence}

			rows = append(rows, row)
		}

		printer.MutateTable(func(table *tablewriter.Table) {
			table.SetAutoWrapText(false)
		})

		return header, rows, nil
	}

	printer.Print(g.players)
}
