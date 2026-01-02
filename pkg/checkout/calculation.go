package checkout

import (
	"fmt"
	"slices"
)

type calculator struct {
	checkouts Checkouts

	opts []option

	limit     int
	maxThrows int

	out CheckoutType
}

func newCalculator(opts ...option) (*calculator, error) {
	c := &calculator{
		opts:      opts,
		limit:     1,
		maxThrows: 3,
		out:       CheckoutTypeDoubleOut,
	}

	for _, o := range opts {
		switch opt := o.(type) {
		case *optionCalcLimit:
			c.limit = opt.calcLimit
		case *optionMaxThrows:
			c.maxThrows = opt.max
		case *optionCheckoutType:
			c.out = opt.out
		default:
			return nil, fmt.Errorf("unknown option: %T", opt)
		}
	}

	return c, nil
}

func (c *calculator) recurse(remaining, throw int) {
	if c.limitReached() {
		return
	}

	if throw > c.maxThrows {
		return
	}

	if remaining <= 0 || remaining > 180 {
		return
	}

	// first look for immediate straight-out wins
	if c.out == CheckoutTypeStraightOut && (remaining <= 20 || remaining == BullsEye) {
		limitReached := c.append(checkout(NewScore(remaining)))
		if limitReached {
			return
		}
	}

	// look for immediate double-out wins
	if remaining%2 == 0 && (remaining <= 40 || remaining == 2*BullsEye) {
		for _, single := range Singles() {
			double := single.WithMultiplier(Double)

			if remaining-double.Value() == 0 {
				limitReached := c.append(checkout(double))
				if limitReached {
					return
				}
			}
		}
	}

	// look for immediate triple-out wins
	if c.out == CheckoutTypeStraightOut && remaining <= 60 {
		for _, single := range Singles() {
			if single.Value() == BullsEye {
				// there is no triple bullseye
				continue
			}

			triple := single.WithMultiplier(Triple)

			if remaining-triple.Value() == 0 {
				limitReached := c.append(checkout(triple))
				if limitReached {
					return
				}
			}
		}
	}

	// start recursing with the ideas:
	// - to prefer short checkouts, so we'll recurse as if we get more and more throws
	// - the less the multiplier the easier it is to score, so start look from the least multiplier
	//   - however, when there is a checkout sequence longer than 2, still try to score the highest first in order
	descendingThrow := c.maxThrows

	for {
		if descendingThrow == throw {
			return
		}

		for _, multiplier := range []Multiplier{None, Double, Triple} {
			singles := Singles()
			slices.Reverse(singles)

			for _, single := range singles {
				if multiplier == Triple && single.Value() == BullsEye {
					// there is no triple bullseye
					continue
				}

				assumedScore := single.WithMultiplier(multiplier)

				rest := remaining - assumedScore.Value()
				if rest < 0 {
					continue
				}

				newCalculator, _ := newCalculator(c.opts...)

				for _, v := range forThrow(rest, descendingThrow, newCalculator) {
					v.prepend(assumedScore)

					if len(v.scores) > 2 {
						v.orderScores(c.out)
					}

					limitReached := c.append(v)
					if limitReached {
						return
					}
				}
			}
		}

		descendingThrow--
	}
}

func (c *calculator) limitReached() bool {
	return len(c.checkouts) >= c.limit
}

func (c *calculator) append(cs *Checkout) bool {
	if slices.ContainsFunc(c.checkouts, func(e *Checkout) bool {
		return e.String() == cs.String()
	}) {
		return false
	}
	c.checkouts = append(c.checkouts, cs)
	return c.limitReached()
}
