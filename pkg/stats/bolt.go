package stats

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/Gerrit91/darts-counter/pkg/config"

	bolt "go.etcd.io/bbolt"
)

var (
	gamesBucket = []byte("games")
)

type boltImpl struct {
	c  *config.StatisticsConfig
	db *bolt.DB
}

func (b *boltImpl) GetPlayerStats(filterOpts ...filter) ([]*PlayerStats, error) {
	playerMap := map[string]*PlayerStats{}

	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(gamesBucket)

		b.ForEach(func(_, v []byte) error {
			var s *GameStats
			err := json.Unmarshal(v, &s)
			if err != nil {
				return err
			}

			for _, id := range s.Players {
				p, ok := playerMap[id]
				if !ok {
					p = &PlayerStats{
						ID:         id,
						RanksCount: map[int]int{},
					}
				}

				p.GamesPlayed++

				playerMap[id] = p
			}

			for rank, player := range s.Ranks {
				p := playerMap[player]

				p.RanksCount[rank] += 1
			}

			return nil
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	var ps []*PlayerStats
	for _, p := range playerMap {
		ps = append(ps, p)
	}

	return ps, nil
}

func (b *boltImpl) GetGameStats(filterOpts ...filter) ([]*GameStats, error) {
	var gs []*GameStats

	for _, opt := range filterOpts {
		switch o := opt.(type) {
		case *idFilter:
			err := b.db.View(func(tx *bolt.Tx) error {
				b := tx.Bucket(gamesBucket)

				v := b.Get([]byte(o.id))
				if v == nil {
					return fmt.Errorf("game with id %q not found", o.id)
				}

				var s *GameStats
				err := json.Unmarshal(v, &s)
				if err != nil {
					return err
				}

				gs = append(gs, s)

				return nil
			})
			if err != nil {
				return nil, err
			}

			return gs, nil
		default:
			return nil, fmt.Errorf("internal error: unsupported option %T", o)
		}
	}

	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(gamesBucket)

		b.ForEach(func(k, v []byte) error {
			var s *GameStats
			err := json.Unmarshal(v, &s)
			if err != nil {
				return err
			}

			gs = append(gs, s)

			return nil
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(gs, func(i, j int) bool {
		return gs[i].Start.Before(gs[j].Start)
	})

	return gs, nil
}

func (b *boltImpl) CreateGameStats(gameStats *GameStats) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(gamesBucket)

		buf, err := json.Marshal(gameStats)
		if err != nil {
			return err
		}

		return b.Put([]byte(gameStats.ID), buf)
	})
}

func (b *boltImpl) DeleteGameStats(id string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(gamesBucket)

		return b.Delete([]byte(id))
	})
}

func (b *boltImpl) initializeDatastore() error {
	db, err := bolt.Open(b.c.Path, 0600, nil)
	if err != nil {
		return fmt.Errorf("unable to open db: %w", err)
	}

	b.db = db

	db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucket(gamesBucket)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		return nil
	})

	return nil
}
