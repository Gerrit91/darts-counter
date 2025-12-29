package datastore

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/Gerrit91/darts-counter/pkg/config"

	bolt "go.etcd.io/bbolt"
	berrors "go.etcd.io/bbolt/errors"
)

var (
	gamesBucket = []byte("games")
)

type boltImpl struct {
	log *slog.Logger
	c   *config.StatisticsConfig
	db  *bolt.DB
}

func (b *boltImpl) ListGameStats(filterOpts ...filter) ([]*GameStats, error) {
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

		err := b.ForEach(func(k, v []byte) error {
			var s *GameStats
			err := json.Unmarshal(v, &s)
			if err != nil {
				return err
			}

			gs = append(gs, s)

			return nil
		})
		if err != nil {
			return err
		}

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

func (b *boltImpl) Close() {
	if b.db != nil {
		if err := b.db.Close(); err != nil {
			b.log.Error("error closing data store", "error", err)
		}
	}
}

func (*boltImpl) Enabled() bool {
	return true
}

func (b *boltImpl) initializeDatastore() error {
	db, err := bolt.Open(b.c.Path, 0600, &bolt.Options{
		Timeout: 2 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("unable to open db: %w", err)
	}

	b.db = db

	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucket(gamesBucket)
		if err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}

		return nil
	})
	if err != nil && !errors.Is(err, berrors.ErrBucketExists) {
		return err
	}

	return nil
}
