package checkout

import (
	"sort"
	"strings"
)

const (
	CheckoutTypeStraightOut CheckoutType = "straight-out"
	CheckoutTypeDoubleOut   CheckoutType = "double-out"

	CheckinTypeStraightIn CheckinType = "straight-in"
	CheckinTypeDoubleIn   CheckinType = "double-in"
)

type (
	CheckoutType string
	CheckinType  string

	Checkout struct {
		scores []*Score
	}

	Checkouts []*Checkout
)

func For(score int, opts ...option) Checkouts {
	s, err := newCalculator(opts...)
	if err != nil {
		panic(err) // TODO: too harsh
	}

	return forThrow(score, 1, s)
}

func forThrow(remaining, throw int, c *calculator) Checkouts {
	c.recurse(remaining, throw)
	return c.checkouts
}

func checkout(scores ...*Score) *Checkout {
	return &Checkout{
		scores: scores,
	}
}

func (c *Checkout) prepend(score *Score) {
	c.scores = append([]*Score{score}, c.scores...)
}

func (c *Checkout) orderScores(out CheckoutType) {
	switch out {
	case CheckoutTypeStraightOut:
		sort.Slice(c.scores, func(i, j int) bool {
			return c.scores[j].Value() < c.scores[i].Value()
		})
	case CheckoutTypeDoubleOut:
		withoutLast := c.scores[:len(c.scores)-1]

		sort.Slice(withoutLast, func(i, j int) bool {
			return withoutLast[j].Value() < withoutLast[i].Value()
		})

		withoutLast = append(withoutLast, c.scores[len(c.scores)-1])

		c.scores = withoutLast
	}
}

func (c *Checkout) String() string {
	var scores []string
	for _, s := range c.scores {
		scores = append(scores, s.String())
	}
	return strings.Join(scores, " → ")
}

func (cs Checkouts) String() string {
	var res []string
	for _, c := range cs {
		res = append(res, c.String())
	}
	return strings.Join(res, ", ")
}
