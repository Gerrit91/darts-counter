package player

import (
	"fmt"
	"slices"

	"github.com/Gerrit91/darts-counter/pkg/checkout"
)

var (
	ErrGameFinished = fmt.Errorf("no more players left in the game")
	ErrInvalidInput = fmt.Errorf("invalid input")
)

type (
	Player struct {
		name       string
		out        checkout.CheckoutType
		in         checkout.CheckinType
		remaining  int
		startScore int
		rank       int
		finished   bool
	}

	Players []*Player
)

func (ps Players) Names() []string {
	var names []string

	for _, p := range ps {
		names = append(names, p.name)
	}

	return names
}

func New(name string, out checkout.CheckoutType, in checkout.CheckinType, remaining int) *Player {
	return &Player{
		name:       name,
		remaining:  remaining,
		startScore: remaining,
		out:        out,
		in:         in,
	}
}

func (p *Player) Move(scores []*checkout.Score, total int) error {
	err := p.validateInput(scores, total)
	if err != nil {
		return err
	}

	p.remaining = p.remaining - total

	if p.remaining == 0 {
		p.finished = true
	}

	return nil
}

func (p *Player) Edit(total int) error {
	if total < 0 {
		return fmt.Errorf("%w: unable to set remaining score below zero", ErrInvalidInput)
	}

	if p.remaining == 0 && total > 0 {
		p.finished = false
	}

	p.remaining = total

	if p.remaining == 0 {
		p.finished = true
	}

	return nil
}

func (p *Player) validateInput(scores []*checkout.Score, total int) error {
	if p.in == checkout.CheckinTypeDoubleIn && p.remaining == p.startScore {
		if len(scores) != 0 && scores[0].GetMultiplier() != checkout.Double {
			return fmt.Errorf("selected game requires double-in, but did not start with double")
		}
	}

	if total < 0 {
		return fmt.Errorf("%w: score must be a positive number", ErrInvalidInput)
	}

	if total > 180 {
		return fmt.Errorf("%w: cannot achieve more than 180 points", ErrInvalidInput)
	}

	if slices.Contains(checkout.BogeyNumbers(), total) {
		return fmt.Errorf("%w: not possible to achieve %d points in one turn (bogey number)", ErrInvalidInput, total)
	}

	newScore := p.remaining - total

	if newScore < 0 {
		return fmt.Errorf("%s exceeded the remaining score of %d", p.name, p.remaining)
	}

	if p.out == checkout.CheckoutTypeDoubleOut && newScore == 1 {
		return fmt.Errorf("in double-out games, remaining 1 is considered overshoot")
	}

	if p.out == checkout.CheckoutTypeDoubleOut && newScore == 0 {
		if len(scores) != 0 && scores[len(scores)-1].GetMultiplier() != checkout.Double {
			return fmt.Errorf("selected game requires double-out, but did not checkout with double")
		}
	}

	if p.remaining > 180 {
		// early skip to prevent unnecessary checkout calculation
		return nil
	}

	if p.remaining-total == 0 && len(checkout.For(total, checkout.NewCheckoutTypeOption(p.out))) == 0 {
		return fmt.Errorf("%w: not possible to finish with %d points", ErrInvalidInput, total)
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
