package config

import (
	"fmt"
	"os"

	"github.com/Gerrit91/darts-counter/pkg/checkout"
	"sigs.k8s.io/yaml"
)

type GameType string

const (
	GameType101  GameType = "101"
	GameType301  GameType = "301"
	GameType501  GameType = "501"
	GameType701  GameType = "701"
	GameType1001 GameType = "1001"
)

type Config struct {
	Game     GameType              `json:"game"`
	Checkout checkout.CheckoutType `json:"checkout"`
	Checkin  checkout.CheckinType  `json:"checkin"`
	Players  []struct {
		Name string `json:"name"`
	} `json:"players"`
	Statistics *StatisticsConfig `json:"statistics"`
	Logging    *LoggingConfig    `json:"logging"`
}

type LoggingConfig struct {
	Enabled bool   `json:"enabled"`
	Path    string `json:"path"`
	Level   string `json:"level"`
}

type StatisticsConfig struct {
	Enabled bool   `json:"enabled"`
	Path    string `json:"path"`
}

func ReadConfig() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	raw, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	config := &Config{}

	err = yaml.Unmarshal(raw, config)
	if err != nil {
		return nil, err
	}

	config.Default()

	return config, nil
}

func (c *Config) Default() {
	if c.Checkout == "" {
		c.Checkout = checkout.CheckoutTypeDoubleOut
	}

	if c.Checkin == "" {
		c.Checkin = checkout.CheckinTypeStraightIn
	}

	if c.Game == "" {
		c.Game = GameType501
	}

	if c.Statistics == nil {
		c.Statistics = &StatisticsConfig{
			Enabled: true,
			Path:    "darts-counter.db",
		}
	}

	if c.Logging == nil {
		c.Logging = &LoggingConfig{
			Enabled: true,
			Path:    "darts-counter.log",
			Level:   "DEBUG",
		}
	}
}

func (c *Config) Validate() error {
	switch gt := c.Game; gt {
	case GameType101, GameType301, GameType501, GameType701, GameType1001:
		// noop
	default:
		return fmt.Errorf("unknown game type: %s", gt)
	}

	switch c.Checkin {
	case checkout.CheckinTypeDoubleIn, checkout.CheckinTypeStraightIn:
		// noop
	default:
		return fmt.Errorf("unknown check-in type: %s", c.Checkin)
	}

	switch c.Checkout {
	case checkout.CheckoutTypeDoubleOut, checkout.CheckoutTypeStraightOut:
		// noop
	default:
		return fmt.Errorf("unknown check-out type: %s", c.Checkout)
	}

	names := map[string]bool{}
	for _, p := range c.Players {
		_, ok := names[p.Name]
		if ok {
			return fmt.Errorf("player names must be unique")
		}

		names[p.Name] = true
	}

	return nil
}
