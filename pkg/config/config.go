package config

import (
	"os"

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
	Database *DatabaseConfig `json:"database"`
	Logging  *LoggingConfig  `json:"logging"`
}

type LoggingConfig struct {
	Enabled bool   `json:"enabled"`
	Path    string `json:"path"`
	Level   string `json:"level"`
}

type DatabaseConfig struct {
	Path string `json:"path"`
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
	if c.Database == nil {
		c.Database = &DatabaseConfig{
			Path: "darts-counter.db",
		}
	}

	if c.Logging == nil {
		c.Logging = &LoggingConfig{
			Enabled: true,
			Path:    "darts-counter.log",
			Level:   "info",
		}
	}
}
