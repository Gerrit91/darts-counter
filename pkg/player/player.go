package player

import (
	"container/list"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/Gerrit91/darts-counter/pkg/checkout"
	"github.com/Gerrit91/darts-counter/pkg/util"
)

type Player struct {
	name       string
	c          *util.Console
	out        checkout.CheckoutType
	in         checkout.CheckinType
	remaining  int
	startScore int
	rank       int
	finished   bool
}

type Players []*Player

func (ps Players) Iterator() *PlayerIterator {
	players := list.New()

	for _, p := range ps {
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
		names = append(names, p.name)
	}

	return names
}

func New(name string, console *util.Console, out checkout.CheckoutType, in checkout.CheckinType, remaining int) *Player {
	return &Player{
		name:       name,
		c:          console,
		remaining:  remaining,
		startScore: remaining,
		out:        out,
		in:         in,
	}
}

func (p *Player) Move() {
	var (
		score int
		err   error
	)

	if p.remaining <= 0 {
		p.finished = true
		return
	}

	for {
		p.c.Printf(`enter score ("s" to skip player, "e" to edit player's score): `)

		input := p.c.Read()

		if input == "s" {
			return
		}

		if input == "e" {
			p.c.Printf("enter new remaining score for %s: ", p.GetName())

			remaining, err := strconv.Atoi(p.c.Read())
			if err != nil {
				p.c.Println("unable to parse input (%q), please enter again", err.Error())
				continue
			}

			p.remaining = remaining
			return
		}

		segments := strings.Split(input, ",")

		switch len(segments) {
		case 1:
			s, err := checkout.ParseScore(input)
			if err == nil {
				if p.in == checkout.CheckinTypeDoubleIn && p.remaining == p.startScore {
					if s.GetMultiplier() != checkout.Double {
						p.c.Println("selected game required double-in, but did not start with double")
						return
					}
				}

				score = s.Value()
			} else {
				// user entered summed up score
				score, err = strconv.Atoi(input)
				if err != nil {
					p.c.Println("unable to parse input (%q), please enter again", err.Error())
					continue
				}
			}
		case 2, 3:
			var (
				sum     int
				partial *checkout.Score
				err     error
			)

			for i, segment := range segments {
				partial, err = checkout.ParseScore(segment)
				if err != nil {
					break
				}

				if i == 0 && p.in == checkout.CheckinTypeDoubleIn && p.remaining == p.startScore {
					if partial.GetMultiplier() != checkout.Double {
						p.c.Println("selected game required double-in, but did not start with double")
						return
					}
				}

				sum += partial.Value()
			}

			p.c.Println("summed up partially provided score: %d", sum)

			if err != nil {
				p.c.Println("unable to parse input (%q), please enter again", err.Error())
				continue
			}

			score = sum
		default:
			p.c.Println("no more than three throws are allowed, please enter again")
			continue
		}

		err = validateInput(score, p.remaining, p.out)
		if err == nil {
			break
		}

		p.c.Println(err.Error())
	}

	newScore := p.remaining - score

	if newScore < 0 {
		p.c.Println("%s exceeded the remaining score of %d", p.name, p.remaining)
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

func validateInput(score, remaining int, out checkout.CheckoutType) error {
	if score < 0 {
		return fmt.Errorf("score must be a positive number, please enter again")
	}

	if score > 180 {
		return fmt.Errorf("cannot achieve more than 180 points, please enter again")
	}

	if slices.Contains(checkout.BogeyNumbers(), score) {
		return fmt.Errorf("not possible to achieve %d points in one turn (bogey number), please enter again", score)
	}

	if remaining > 180 {
		return nil
	}

	if remaining-score == 0 && len(checkout.For(score, checkout.NewCheckoutTypeOption(out))) == 0 {
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
