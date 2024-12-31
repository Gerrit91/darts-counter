package stats

import (
	"fmt"
	"time"

	"github.com/Gerrit91/darts-counter/pkg/config"
)

type (
	Stats interface {
		CreateGameStats(g *GameStats) error
		DeleteGameStats(id string) error
		GetGameStats(filterOpts ...filter) ([]*GameStats, error)
		GetPlayerStats(filterOpts ...filter) ([]*PlayerStats, error)
	}

	filter   interface{}
	idFilter struct{ id string }

	GameStats struct {
		ID       string          `json:"id"`
		GameType config.GameType `json:"type"`
		Checkout string          `json:"checkout"`
		Checkin  string          `json:"checkin"`
		Players  []string        `json:"players"`
		Ranks    map[int]string  `json:"ranks"`
		Rounds   int             `json:"rounds"`
		Start    time.Time       `json:"start"`
		End      time.Time       `json:"end"`
		Moves    []Move          `json:"moves"`
	}

	Move struct {
		Round     int    `json:"round"`
		Player    string `json:"player"`
		Score     int    `json:"score"`
		Remaining int    `json:"remaining"`
	}

	PlayerStats struct {
		ID          string      `json:"id"`
		GamesPlayed int         `json:"games_played"`
		RanksCount  map[int]int `json:"ranks_count"`
	}

	stats struct {
		Stats
		c config.StatisticsConfig
	}
)

func New(c *config.StatisticsConfig) (Stats, error) {
	if c == nil {
		return nil, fmt.Errorf("no statistics config defined")
	}

	s := &stats{
		Stats: &noopImpl{},
		c:     *c,
	}

	if c.Enabled {
		b := &boltImpl{c: &s.c}

		err := b.initializeDatastore()
		if err != nil {
			return nil, fmt.Errorf("unable to initialize datastore: %w", err)
		}

		s.Stats = b
	}

	return s, nil
}

func IdFilter(id string) filter {
	return &idFilter{id: id}
}
