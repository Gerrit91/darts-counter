package checkout

import (
	"fmt"
	"reflect"
	"testing"
)

func Test_For(t *testing.T) {
	tests := []struct {
		score        int
		limit        int
		wantStraight string
		wantDouble   string
	}{
		{
			score:        -1,
			limit:        1,
			wantStraight: "",
			wantDouble:   "",
		},
		{
			score:        0,
			limit:        1,
			wantStraight: "",
			wantDouble:   "",
		},
		{
			score:        1,
			limit:        1,
			wantStraight: "1",
			wantDouble:   "",
		},
		{
			score:        2,
			limit:        1,
			wantStraight: "2",
			wantDouble:   "D1",
		},
		{
			score:        3,
			limit:        10,
			wantStraight: "3, T1, 1 → 2, 1 → D1, 2 → 1, D1 → 1, 1 → 1 → 1",
			wantDouble:   "1 → D1",
		},
		{
			score:        4,
			limit:        1,
			wantStraight: "4",
			wantDouble:   "D2",
		},
		{
			score:        20,
			limit:        1,
			wantStraight: "20",
			wantDouble:   "D10",
		},
		{
			score:        21,
			limit:        2,
			wantStraight: "T7, 1 → 20",
			wantDouble:   "1 → D10, 3 → D9",
		},
		{
			score:        25,
			limit:        1,
			wantStraight: "B",
			wantDouble:   "1 → D12",
		},
		{
			score:        40,
			limit:        1,
			wantStraight: "D20",
			wantDouble:   "D20",
		},
		{
			score:        50,
			limit:        1,
			wantStraight: "DB",
			wantDouble:   "DB",
		},
		{
			score:        60,
			limit:        2,
			wantStraight: "T20, 3 → T19",
			wantDouble:   "10 → DB, 20 → D20",
		},
		{
			score:        61,
			limit:        2,
			wantStraight: "B → D18, B → T12",
			wantDouble:   "B → D18, 11 → DB",
		},
		{
			score:        85,
			limit:        2,
			wantStraight: "B → T20, D14 → T19",
			wantDouble:   "T15 → D20, T17 → D17",
		},
		{
			score:        100,
			limit:        2,
			wantStraight: "DB → DB, D20 → T20",
			wantDouble:   "DB → DB, T20 → D20",
		},
		{
			score:        107,
			limit:        1,
			wantStraight: "DB → T19",
			wantDouble:   "T19 → DB",
		},
		{
			score:        119,
			limit:        1,
			wantStraight: "T20 → D17 → B",
			wantDouble:   "T18 → B → D20",
		},
		{
			score:        120,
			limit:        2,
			wantStraight: "T20 → T20, DB → T15 → B",
			wantDouble:   "T15 → B → DB, T19 → B → D19",
		},
		{
			score:        121,
			limit:        1,
			wantStraight: "T20 → D18 → B",
			wantDouble:   "T20 → B → D18",
		},
		{
			score:        170,
			limit:        2,
			wantStraight: "T20 → T20 → DB",
			wantDouble:   "T20 → T20 → DB",
		},
		{
			score:        179,
			limit:        1,
			wantStraight: "",
			wantDouble:   "",
		},
		{
			score:        180,
			limit:        1,
			wantStraight: "T20 → T20 → T20",
			wantDouble:   "",
		},
		{
			score:        181,
			limit:        1,
			wantStraight: "",
			wantDouble:   "",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("score_%d_%s_limit_%d", tt.score, CheckoutTypeStraightOut, tt.limit), func(t *testing.T) {
			if got := For(tt.score, NewCalcLimitOption(tt.limit), NewCheckoutTypeOption(CheckoutTypeStraightOut)); !reflect.DeepEqual(got.String(), tt.wantStraight) {
				t.Errorf("%v, want %v", got, tt.wantStraight)
			}
		})
		t.Run(fmt.Sprintf("score_%d_%s_limit_%d", tt.score, CheckoutTypeDoubleOut, tt.limit), func(t *testing.T) {
			if got := For(tt.score, NewCalcLimitOption(tt.limit), NewCheckoutTypeOption(CheckoutTypeDoubleOut)); !reflect.DeepEqual(got.String(), tt.wantDouble) {
				t.Errorf("%v, want %v", got, tt.wantDouble)
			}
		})
	}
}
