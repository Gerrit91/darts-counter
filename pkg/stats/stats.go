package stats

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Gerrit91/darts-counter/pkg/config"
)

type (
	Stats interface {
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

	PlayerStats struct {
		ID              string
		GamesPlayed     int
		RanksCount      map[int]int
		TotalMoves      int
		TotalDuration   time.Duration
		AverageDuration time.Duration
		FieldsCount     map[string]int
		HighestScore    Score
		TotalScore      int
		AverageScore    float64
	}

	stats struct {
		Stats
		c config.StatisticsConfig
	}
)

func New(log *slog.Logger, c *config.StatisticsConfig) (Stats, error) {
	if c == nil {
		return nil, fmt.Errorf("no statistics config defined")
	}

	s := &stats{
		Stats: &noopImpl{},
		c:     *c,
	}

	if c.Enabled {
		b := &boltImpl{c: &s.c, log: log}

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

func (r Ranks) OfPlayer(id string) int {
	for rank, playerID := range r {
		if playerID == id {
			return rank
		}
	}
	return 0
}

func ToPlayerStats(stats []*GameStats) ([]*PlayerStats, error) {
	playerMap := map[string]*PlayerStats{}

	for _, s := range stats {
		for _, id := range s.Players {
			p, ok := playerMap[id]
			if !ok {
				p = &PlayerStats{
					ID:          id,
					RanksCount:  map[int]int{},
					FieldsCount: map[string]int{},
				}
			}

			for _, move := range s.Moves {
				if move.Player != id {
					continue
				}

				duration, err := time.ParseDuration(move.Duration)
				if err != nil {
					return nil, fmt.Errorf("unable to parse move duration: %w", err)
				}

				for _, field := range move.Score.Fields {
					p.FieldsCount[field]++
				}
				if p.HighestScore.Total < move.Score.Total {
					p.HighestScore = move.Score
				}
				p.TotalScore += move.Score.Total
				p.TotalDuration += duration
				p.TotalMoves++
			}

			for rank := range s.Ranks {
				// create entries in the ranks count
				p.RanksCount[rank] += 0
			}
			p.GamesPlayed++
			playerMap[id] = p
		}

		for rank, player := range s.Ranks {
			p := playerMap[player]

			p.RanksCount[rank] += 1
		}
	}

	var ps []*PlayerStats
	for _, p := range playerMap {
		p.AverageDuration = time.Duration(int64(p.TotalDuration) / int64(p.TotalMoves))
		p.AverageScore = float64(p.TotalScore) / float64(p.TotalMoves)
		ps = append(ps, p)
	}

	return ps, nil
}
