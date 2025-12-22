package checkout

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	BullsEye = 25
)

func BogeyNumbers() []int {
	return []int{169, 168, 166, 165, 163, 162, 159}
}

type Score struct {
	score      int
	multiplier Multiplier
}

func NewScore(score int) *Score {
	return &Score{
		score: score,
	}
}

func ParseScore(input string) (*Score, error) {
	var (
		multiplier Multiplier
	)

	if strings.HasPrefix(input, string(Triple)) {
		multiplier = Triple
		input = strings.TrimPrefix(input, string(Triple))
	} else if strings.HasPrefix(input, string(Double)) {
		multiplier = Double
		input = strings.TrimPrefix(input, string(Double))
	}

	if input == "B" {
		if multiplier == Triple {
			return nil, fmt.Errorf("there is no triple bullseye")
		}
		return NewScore(BullsEye).WithMultiplier(multiplier), nil
	}

	score, err := strconv.Atoi(input)
	if err != nil {
		return nil, err
	}

	switch {
	case score <= 0:
		return nil, fmt.Errorf("score must be greater than 0")
	case score > 20:
		return nil, fmt.Errorf("score must be between 1 and 20 (or B for bullseye)")
	}

	return NewScore(score).WithMultiplier(multiplier), nil
}

func (s *Score) WithMultiplier(m Multiplier) *Score {
	s.multiplier = m
	return s
}

func (s *Score) Value() int {
	value := s.score

	switch s.multiplier {
	case Triple, Double:
		return value * s.multiplier.Value()
	}

	return value
}

func (s *Score) GetMultiplier() Multiplier {
	return s.multiplier
}

func (s *Score) String() string {
	representation := strconv.Itoa(s.score)
	if s.score == BullsEye {
		representation = "B"
	}

	return string(s.multiplier) + representation
}

type Multiplier string

const (
	None   Multiplier = ""
	Triple Multiplier = "T"
	Double Multiplier = "D"
)

func (m Multiplier) Value() int {
	switch m {
	case Triple:
		return 3
	case Double:
		return 2
	default:
		return 1
	}
}

func singles() []*Score {
	var s []*Score

	for i := 20; i > 0; i-- {
		s = append(s, NewScore(i))
	}

	// bullseye is harder to score than other singles, so put it to the end
	s = append(s, NewScore(BullsEye))

	return s
}
