package datastore

import (
	"fmt"
	"time"
)

type (
	PlayerStats struct {
		ID              string
		GamesPlayed     int
		RanksCount      map[int]int
		AverageRank     float64
		TotalMoves      int
		TotalDuration   time.Duration
		AverageDuration time.Duration
		FieldsCount     map[string]int
		HighestScore    Score
		TotalScore      int
		AverageScore    float64

		totalRanks int
	}
)

func ToPlayerStats(stats []*GameStats) ([]*PlayerStats, error) {
	playerMap := map[string]*PlayerStats{}

	for _, s := range stats {
		for _, id := range s.Players {
			p, ok := playerMap[id]
			if !ok {
				p = &PlayerStats{
					ID:          id,
					RanksCount:  map[int]int{},
					FieldsCount: map[string]int{},
				}
			}

			for _, move := range s.Moves {
				if move.Player != id {
					continue
				}

				duration, err := time.ParseDuration(move.Duration)
				if err != nil {
					return nil, fmt.Errorf("unable to parse move duration: %w", err)
				}

				for _, field := range move.Score.Fields {
					p.FieldsCount[field]++
				}
				if p.HighestScore.Total < move.Score.Total {
					p.HighestScore = move.Score
				}
				p.TotalScore += move.Score.Total
				p.TotalDuration += duration
				p.TotalMoves++
			}

			for rank := range s.Ranks {
				// create entries in the ranks count
				p.RanksCount[rank] += 0
			}
			p.GamesPlayed++
			playerMap[id] = p
		}

		for rank, player := range s.Ranks {
			p := playerMap[player]

			p.RanksCount[rank] += 1
			p.totalRanks += rank
		}
	}

	var ps []*PlayerStats
	for _, p := range playerMap {
		p.AverageDuration = time.Duration(int64(p.TotalDuration) / int64(p.TotalMoves))
		p.AverageScore = float64(p.TotalScore) / float64(p.TotalMoves)
		p.AverageRank = float64(p.totalRanks) / float64(p.GamesPlayed)
		ps = append(ps, p)
	}

	return ps, nil
}
