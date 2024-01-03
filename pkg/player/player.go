package player

import (
	"container/list"
	"fmt"
	"slices"

	"github.com/Gerrit91/darts-counter/pkg/checkout"
	"github.com/Gerrit91/darts-counter/pkg/util"
)

type Player struct {
	name      string
	c         *util.Console
	out       checkout.CheckoutType
	remaining int
	rank      int
	finished  bool
}

type Players []*Player

func (ps Players) Iterator() *PlayerIterator {
	players := list.New()

	for _, p := range ps {
		p := p
		players.PushBack(p)
	}

	return &PlayerIterator{
		players:       players,
		currentPlayer: players.Front(),
	}
}

func (ps Players) Names() []string {
	var names []string

	for _, p := range ps {
		p := p
		names = append(names, p.name)
	}

	return names
}

func New(name string, console *util.Console, out checkout.CheckoutType, remaining int) *Player {
	return &Player{
		name:      name,
		c:         console,
		remaining: remaining,
		out:       out,
	}
}

func (p *Player) Move() {
	var score int
	for {
		score = p.c.AskForScore()

		err := ValidateScore(score, p.remaining, p.out)
		if err == nil {
			break
		}

		p.c.Println(err.Error())
	}

	newScore := p.remaining - score

	if newScore < 0 {
		p.c.Println("%s has exceeded the remaining score of %d", p.name, p.remaining)
		return
	}

	if p.out == checkout.CheckoutTypeDoubleOut && newScore == 1 {
		p.c.Println("in double-out, remaining 1 is considered overshoot")
		return
	}

	p.remaining = newScore

	if p.remaining == 0 {
		p.finished = true
	}
}

func ValidateScore(score, remaining int, out checkout.CheckoutType) error {
	if score < 0 {
		return fmt.Errorf("score must be a positive number, please enter again")
	}

	if score > 180 {
		return fmt.Errorf("cannot achieve more than 180 points, please enter again")
	}

	if slices.Contains(checkout.BogeyNumbers(), score) {
		return fmt.Errorf("not possible to achieve %d points in one turn, please enter again", score)
	}

	if remaining > 180 {
		return nil
	}

	if len(checkout.For(score, checkout.NewCheckoutTypeOption(out))) == 0 {
		return fmt.Errorf("not possible to finish with %d points, please enter again", score)
	}

	return nil
}

func (p *Player) GetName() string {
	return p.name
}

func (p *Player) GetRank() int {
	return p.rank
}

func (p *Player) GetRemaining() int {
	return p.remaining
}

func (p *Player) HasFinished() bool {
	return p.finished
}

func (p *Player) SetRank(rank int) {
	p.rank = rank
}
