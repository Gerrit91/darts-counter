package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	game "github.com/Gerrit91/darts-counter/pkg"
	"github.com/Gerrit91/darts-counter/pkg/config"
	"github.com/Gerrit91/darts-counter/pkg/stats"
	"github.com/Gerrit91/darts-counter/pkg/util"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v3"
)

func main() {
	if err := run(); err != nil {
		slog.Error("error running darts-counter", "error", err)
		os.Exit(1)
	}
}

func run() error {
	config, err := config.ReadConfig()
	if err != nil {
		return err
	}

	err = config.Validate()
	if err != nil {
		return err
	}

	s, err := stats.New(config.Statistics)
	if err != nil {
		return err
	}

	printerConfig := &printers.TablePrinterConfig{
		Markdown: true,
	}

	printer := printers.NewTablePrinter(printerConfig)

	cmd := &cli.Command{
		Name:  "darts-counter",
		Usage: "Counts remaining scores for a game of darts and shows possible finishes.",
		Action: func(context.Context, *cli.Command) error {
			g, err := game.NewGame(&util.Console{}, config, s)
			if err != nil {
				return fmt.Errorf("error creating new game: %w", err)
			}

			g.Run()

			return nil
		},
		Commands: []*cli.Command{
			{
				Name: "game-stats",
				Commands: []*cli.Command{
					{
						Name: "delete",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "id",
								Required: true,
							},
						},
						Action: func(ctx context.Context, c *cli.Command) error {
							return s.DeleteGameStats(c.String("id"))
						},
					},
					{
						Name: "describe",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "id",
								Required: true,
							},
						},
						Action: func(ctx context.Context, c *cli.Command) error {
							stats, err := s.GetGameStats(stats.IdFilter(c.String("id")))
							if err != nil {
								return err
							}

							return printers.NewJSONPrinter().Print(pointer.FirstOrZero(stats))
						},
					},
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Action: func(ctx context.Context, c *cli.Command) error {
							gs, err := s.GetGameStats()
							if err != nil {
								return err
							}

							printerConfig.ToHeaderAndRows = func(_ any, wide bool) ([]string, [][]string, error) {
								header := []string{"type", "start", "rounds", "took", "ranks", "id"}
								var rows [][]string

								for _, stats := range gs {
									start := ""
									if !stats.Start.IsZero() {
										start = stats.Start.Local().Format("02.01.2006, 15:04:05")
									}

									took := ""
									if !stats.End.IsZero() {
										took = stats.End.Sub(stats.Start).Truncate(time.Second).String()
									}

									var ranks []string
									for rank, name := range stats.Ranks {
										ranks = append(ranks, fmt.Sprintf("%d. %s", rank, name))
									}
									sort.Slice(ranks, func(i, j int) bool {
										return ranks[i] < ranks[j] // TODO: improve
									})

									row := []string{string(stats.GameType), start, strconv.Itoa(stats.Rounds), took, strings.Join(ranks, "\n"), stats.ID}

									rows = append(rows, row)
								}

								printer.MutateTable(func(table *tablewriter.Table) {
									table.SetAutoWrapText(false)
								})

								return header, rows, nil
							}

							return printer.Print(nil)
						},
					},
				},
			},
			{
				Name: "player-stats",
				Commands: []*cli.Command{
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Action: func(ctx context.Context, c *cli.Command) error {
							ps, err := s.GetPlayerStats()
							if err != nil {
								return err
							}

							printerConfig.ToHeaderAndRows = func(_ any, wide bool) ([]string, [][]string, error) {
								header := []string{"id", "games played", "ranks"}
								var rows [][]string

								for _, stats := range ps {
									var ranks []string
									for rank, count := range stats.RanksCount {
										ranks = append(ranks, fmt.Sprintf("%d. (%d times)", rank, count))
									}
									sort.Slice(ranks, func(i, j int) bool {
										return ranks[i] < ranks[j] // TODO: improve
									})

									row := []string{stats.ID, strconv.Itoa(stats.GamesPlayed), strings.Join(ranks, "\n")}

									rows = append(rows, row)
								}

								printer.MutateTable(func(table *tablewriter.Table) {
									table.SetAutoWrapText(false)
								})

								return header, rows, nil
							}

							return printer.Print(ps)
						},
					},
				},
			},
		},
	}

	return cmd.Run(context.Background(), os.Args)
}
