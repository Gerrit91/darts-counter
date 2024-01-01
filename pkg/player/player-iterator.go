package player

import (
	"container/list"
	"fmt"
)

type (
	PlayerIterator struct {
		players       *list.List
		currentPlayer *list.Element
	}
)

var ErrGameFinished = fmt.Errorf("only one player left in the game")

func (pi *PlayerIterator) Next() (*Player, error) {
	if p, finished := pi.isFinished(); finished {
		return p, ErrGameFinished
	}

	p := pi.currentPlayer.Value.(*Player)

	next := pi.currentPlayer.Next()
	if next == nil {
		next = pi.players.Front()
	}
	pi.currentPlayer = next

	if !p.finished {
		return p, nil
	}

	return pi.Next()
}

func (pi *PlayerIterator) isFinished() (*Player, bool) {
	if pi.players.Len() == 0 {
		return nil, true
	}

	if pi.players.Len() == 1 {
		p := pi.players.Front().Value.(*Player)

		return nil, p.finished
	}

	var (
		lastPlayer   *Player
		stillPlaying int
	)

	for e := pi.players.Front(); e != nil; e = e.Next() {
		e := e

		p := e.Value.(*Player)

		if !p.finished {
			lastPlayer = p
			stillPlaying++
		}
	}

	return lastPlayer, stillPlaying < 2
}
