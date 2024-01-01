package game

import (
	"os"

	"github.com/Gerrit91/darts-counter/pkg/checkout"
	"sigs.k8s.io/yaml"
)

type Config struct {
	Game     GameType              `json:"game"`
	Checkout checkout.CheckoutType `json:"checkout"`
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

	return config, nil
}

func (c *Config) Default() {
	if c.Checkout == "" {
		c.Checkout = checkout.CheckoutTypeDoubleOut
	}

	if c.Game == "" {
		c.Game = GameType501
	}
}
