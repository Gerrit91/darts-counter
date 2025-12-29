package datastore

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Gerrit91/darts-counter/pkg/config"
)

type (
	Datastore interface {
		CreateGameStats(g *GameStats) error
		DeleteGameStats(id string) error
		ListGameStats(filterOpts ...filter) ([]*GameStats, error)
		Close()
		Enabled() bool
	}

	filter   any
	idFilter struct{ id string }

	GameStats struct {
		ID       string          `json:"id"`
		GameType config.GameType `json:"type"`
		Checkout string          `json:"checkout"`
		Checkin  string          `json:"checkin"`
		Players  []string        `json:"players"`
		Ranks    Ranks           `json:"ranks"`
		Rounds   int             `json:"rounds"`
		Start    time.Time       `json:"start"`
		End      time.Time       `json:"end"`
		Moves    []Move          `json:"moves"`
	}

	Ranks map[int]string

	Move struct {
		Round     int    `json:"round"`
		Player    string `json:"player"`
		Score     Score  `json:"score"`
		Remaining int    `json:"remaining"`
		Duration  string `json:"duration"`
	}

	Score struct {
		Total  int      `json:"total"`
		Fields []string `json:"partials"`
	}

	stats struct {
		Datastore
		c config.StatisticsConfig
	}
)

func New(log *slog.Logger, c *config.StatisticsConfig) (Datastore, error) {
	if c == nil {
		return nil, fmt.Errorf("no statistics config defined")
	}

	s := &stats{
		Datastore: &noopImpl{},
		c:         *c,
	}

	if c.Enabled {
		b := &boltImpl{c: &s.c, log: log}

		err := b.initializeDatastore()
		if err != nil {
			return nil, fmt.Errorf("unable to initialize datastore: %w", err)
		}

		s.Datastore = b
	}

	return s, nil
}

func IdFilter(id string) filter {
	return &idFilter{id: id}
}

func (r Ranks) OfPlayer(id string) int {
	for rank, playerID := range r {
		if playerID == id {
			return rank
		}
	}
	return 0
}
