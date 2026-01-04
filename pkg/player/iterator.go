package player

import (
	"fmt"
	"slices"
)

type (
	Iterator struct {
		round   int
		players Players
		nextIdx int

		singlePlayerGame bool
	}
)

func (ps Players) Iterator() *Iterator {
	return &Iterator{
		players:          ps,
		nextIdx:          0,
		round:            0,
		singlePlayerGame: len(ps) == 1,
	}
}

func (i *Iterator) Next() (*Player, error) {
	if p, finished := i.isFinished(); finished {
		return p, ErrGameFinished
	}

	for idx := range len(i.players) {
		var (
			nextIdx    = (i.nextIdx + idx) % len(i.players)
			nextPlayer = i.players[nextIdx]
		)

		if nextIdx == 0 {
			i.round++
		}

		if nextPlayer.finished {
			continue
		}

		i.nextIdx = nextIdx + 1

		return nextPlayer, nil
	}

	return nil, ErrGameFinished
}

func (i *Iterator) SetBackTo(name string) (*Player, error) {
	playerIdx := slices.IndexFunc(i.players, func(p *Player) bool {
		return p.name == name
	})

	if playerIdx < 0 {
		return nil, fmt.Errorf("no player found with name %q", name)
	}

	if playerIdx >= i.nextIdx {
		i.round--
	}

	i.nextIdx = playerIdx + 1

	return i.players[playerIdx], nil
}

func (i *Iterator) GetRound() int {
	return i.round
}

func (i *Iterator) isFinished() (*Player, bool) {
	var (
		stillPlaying []*Player
	)

	for _, p := range i.players {
		if p.finished {
			continue
		}

		stillPlaying = append(stillPlaying, p)
	}

	switch len(stillPlaying) {
	case 0:
		return nil, true
	case 1:
		finished := true
		if i.singlePlayerGame {
			finished = false
		}
		return stillPlaying[0], finished
	default:
		return nil, false
	}
}
