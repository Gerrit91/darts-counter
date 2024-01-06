package game

import (
	"fmt"
	"os"

	"github.com/Gerrit91/darts-counter/pkg/checkout"
	"sigs.k8s.io/yaml"
)

type Config struct {
	Game     GameType              `json:"game"`
	Checkout checkout.CheckoutType `json:"checkout"`
	Checkin  checkout.CheckinType  `json:"checkin"`
	Players  []struct {
		Name string `json:"name"`
	} `json:"players"`
}

func ReadConfig() (*Config, error) {
	raw, err := os.ReadFile("config.yaml")
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

	return nil
}
