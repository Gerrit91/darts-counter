package player

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIterator_isFinished(t *testing.T) {
	tests := []struct {
		name     string
		players  Players
		want     *Player
		finished bool
	}{
		{
			name:     "no players",
			players:  Players{},
			want:     nil,
			finished: true,
		},
		{
			name:     "one player",
			players:  Players{{name: "1"}},
			want:     &Player{name: "1"},
			finished: true,
		},
		{
			name:     "two players",
			players:  Players{{name: "1"}, {name: "2"}},
			want:     nil,
			finished: false,
		},
		{
			name:     "two players, one finished",
			players:  Players{{name: "1", finished: true}, {name: "2"}},
			want:     &Player{name: "2"},
			finished: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iterator{
				players: tt.players,
			}
			got, got1 := i.isFinished()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Iterator.isFinished() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.finished {
				t.Errorf("Iterator.isFinished() got1 = %v, want %v", got1, tt.finished)
			}
		})
	}
}

func TestIterator(t *testing.T) {
	players := Players{{name: "1"}, {name: "2"}, {name: "3"}}
	iter := players.Iterator()
	assert.Equal(t, 0, iter.GetRound())

	p, err := iter.Next()
	require.NoError(t, err)
	assert.Equal(t, 1, iter.GetRound())
	require.Equal(t, &Player{name: "1"}, p)

	p, err = iter.Next()
	require.NoError(t, err)
	assert.Equal(t, 1, iter.GetRound())
	require.Equal(t, &Player{name: "2"}, p)

	p, err = iter.Next()
	require.NoError(t, err)
	assert.Equal(t, 1, iter.GetRound())
	require.Equal(t, &Player{name: "3"}, p)

	p, err = iter.Next()
	require.NoError(t, err)
	assert.Equal(t, 2, iter.GetRound())
	require.Equal(t, &Player{name: "1"}, p)

	p.finished = true // 1 finished

	p, err = iter.Next()
	require.NoError(t, err)
	assert.Equal(t, 2, iter.GetRound())
	require.Equal(t, &Player{name: "2"}, p)

	p, err = iter.Next()
	require.NoError(t, err)
	assert.Equal(t, 2, iter.GetRound())
	require.Equal(t, &Player{name: "3"}, p)

	p, err = iter.Next()
	require.NoError(t, err)
	assert.Equal(t, 3, iter.GetRound())
	require.Equal(t, &Player{name: "2"}, p)

	p.finished = true // 2 is finished

	p, err = iter.Next()
	require.Error(t, err)
	require.ErrorIs(t, err, ErrGameFinished)
	assert.Equal(t, 3, iter.GetRound())
	require.Equal(t, &Player{name: "3"}, p)
}

func TestIteratorSetBackTo(t *testing.T) {
	players := Players{{name: "1"}, {name: "2"}, {name: "3"}}
	iter := players.Iterator()
	assert.Equal(t, 0, iter.GetRound())

	p, err := iter.Next()
	require.NoError(t, err)
	assert.Equal(t, 1, iter.GetRound())
	require.Equal(t, &Player{name: "1"}, p)

	p, err = iter.Next()
	require.NoError(t, err)
	assert.Equal(t, 1, iter.GetRound())
	require.Equal(t, &Player{name: "2"}, p)

	p, err = iter.SetBackTo("1")
	require.NoError(t, err)
	assert.Equal(t, 1, iter.GetRound())
	require.Equal(t, &Player{name: "1"}, p)

	p, err = iter.Next()
	require.NoError(t, err)
	assert.Equal(t, 1, iter.GetRound())
	require.Equal(t, &Player{name: "2"}, p)

	p, err = iter.Next()
	require.NoError(t, err)
	assert.Equal(t, 1, iter.GetRound())
	require.Equal(t, &Player{name: "3"}, p)

	p, err = iter.Next()
	require.NoError(t, err)
	assert.Equal(t, 2, iter.GetRound())
	require.Equal(t, &Player{name: "1"}, p)

	p, err = iter.SetBackTo("2")
	require.NoError(t, err)
	assert.Equal(t, 1, iter.GetRound())
	require.Equal(t, &Player{name: "2"}, p)

	p, err = iter.Next()
	require.NoError(t, err)
	assert.Equal(t, 1, iter.GetRound())
	require.Equal(t, &Player{name: "3"}, p)

	p, err = iter.Next()
	require.NoError(t, err)
	assert.Equal(t, 2, iter.GetRound())
	require.Equal(t, &Player{name: "1"}, p)

	p, err = iter.SetBackTo("3")
	require.NoError(t, err)
	assert.Equal(t, 1, iter.GetRound())
	require.Equal(t, &Player{name: "3"}, p)

	p, err = iter.Next()
	require.NoError(t, err)
	assert.Equal(t, 2, iter.GetRound())
	require.Equal(t, &Player{name: "1"}, p)

	p, err = iter.SetBackTo("2")
	require.NoError(t, err)
	assert.Equal(t, 1, iter.GetRound())
	require.Equal(t, &Player{name: "2"}, p)
}
