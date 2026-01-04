package datastore

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Gerrit91/darts-counter/pkg/checkout"
	"github.com/Gerrit91/darts-counter/pkg/config"
)

type (
	Datastore interface {
		CreateGameStats(g *GameStats) error
		DeleteGameStats(id string) error
		ListGameStats(filterOpts ...filter) ([]*GameStats, error)
		GetGameSettings() (*GameSettings, error)
		UpdateGameSettings(s *GameSettings) error
		Close()
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

	GameSettings struct {
		Type            config.GameType       `json:"game_type"`
		Checkout        checkout.CheckoutType `json:"checkout"`
		Checkin         checkout.CheckinType  `json:"checkin"`
		Players         []Player              `json:"players"`
		SaveGameToStats bool                  `json:"save_game_to_stats"`
	}

	Player struct {
		Name string `json:"name"`
	}
)

var (
	ErrNotFound = fmt.Errorf("not found")
)

func New(log *slog.Logger, c *config.DatabaseConfig) (Datastore, error) {
	if c == nil {
		return nil, fmt.Errorf("no statistics config defined")
	}

	b := &boltImpl{c: c, log: log}

	err := b.initializeDatastore()
	if err != nil {
		return nil, fmt.Errorf("unable to initialize datastore: %w", err)
	}

	return b, nil
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

func validateGameSettings(g *GameSettings) error {
	switch gt := g.Type; gt {
	case config.GameType101, config.GameType301, config.GameType501, config.GameType701, config.GameType1001:
		// noop
	default:
		return fmt.Errorf("unknown game type: %s", gt)
	}

	switch g.Checkin {
	case checkout.CheckinTypeDoubleIn, checkout.CheckinTypeStraightIn:
		// noop
	default:
		return fmt.Errorf("unknown check-in type: %s", g.Checkin)
	}

	switch g.Checkout {
	case checkout.CheckoutTypeDoubleOut, checkout.CheckoutTypeStraightOut:
		// noop
	default:
		return fmt.Errorf("unknown check-out type: %s", g.Checkout)
	}

	if len(g.Players) < 1 {
		return fmt.Errorf("a game needs at least one player")
	}

	names := map[string]bool{}
	for _, p := range g.Players {
		_, ok := names[p.Name]
		if ok {
			return fmt.Errorf("player names must be unique")
		}

		names[p.Name] = true
	}

	return nil
}
