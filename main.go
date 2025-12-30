package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	game "github.com/Gerrit91/darts-counter/pkg"
	"github.com/Gerrit91/darts-counter/pkg/config"
	"github.com/Gerrit91/darts-counter/pkg/datastore"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	config, err := config.ReadConfig()
	if err != nil {
		slog.Error("error reading config", "error", err)
		os.Exit(1)
	}

	log, fileCloser, err := newLogger(config)
	if err != nil {
		slog.Error("error initializing logger", "error", err)
		os.Exit(1)
	}
	if fileCloser != nil {
		defer fileCloser()
	}

	if err := run(config, log); err != nil {
		log.Error("error running darts-counter", "error", err)
		os.Exit(1)
	}
}

func newLogger(config *config.Config) (*slog.Logger, func(), error) {
	if config.Logging.Enabled {
		var level slog.Level

		err := level.UnmarshalText([]byte(config.Logging.Level))
		if err != nil {
			return nil, nil, fmt.Errorf("unable to parse log level: %w", err)
		}

		logFile, err := os.Create(config.Logging.Path)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to open log file: %w", err)
		}

		return slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: level})), func() {
			_ = logFile.Close()
		}, nil
	}

	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})), nil, nil
}

func run(config *config.Config, log *slog.Logger) error {
	ds, err := datastore.New(log, config.Database)
	if err != nil {
		return err
	}

	log.Info("datastore initialized", "db-path", config.Database.Path)

	log.Info("launching main menu")

	var (
		m = game.NewMainMenu(log, config, ds)
		p = tea.NewProgram(m,
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)
	)

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
