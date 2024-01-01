package main

import (
	"log/slog"
	"os"

	game "github.com/Gerrit91/darts-counter/pkg"
	"github.com/Gerrit91/darts-counter/pkg/util"
)

func main() {
	c := &util.Console{}

	g, err := game.NewGame(c)
	if err != nil {
		slog.Error("error creating new game", "error", err)
		os.Exit(1)
	}

	g.Run()
}
