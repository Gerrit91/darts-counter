package checkout_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/Gerrit91/darts-counter/pkg/checkout"
	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
)

func TestParseScore(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *checkout.Score
		wantErr error
	}{
		{
			name:    "empty input",
			input:   "",
			wantErr: &strconv.NumError{Func: "Atoi", Err: fmt.Errorf("invalid syntax")},
		},
		{
			name:    "invalid input",
			input:   "a",
			wantErr: &strconv.NumError{Func: "Atoi", Num: "a", Err: fmt.Errorf("invalid syntax")},
		},
		{
			name:    "disallow zero",
			input:   "0",
			wantErr: fmt.Errorf("score must be greater than 0"),
		},
		{
			name:    "disallow number below zero",
			input:   "-5",
			wantErr: fmt.Errorf("score must be greater than 0"),
		},
		{
			name:  "valid number",
			input: "1",
			want:  checkout.NewScore(1),
		},
		{
			name:  "valid number with double multiplier",
			input: "D2",
			want:  checkout.NewScore(2).WithMultiplier(checkout.Double),
		},
		{
			name:  "valid number with triple multiplier",
			input: "T20",
			want:  checkout.NewScore(20).WithMultiplier(checkout.Triple),
		},
		{
			name:  "bullseye",
			input: "B",
			want:  checkout.NewScore(checkout.BullsEye),
		},
		{
			name:  "double bullseye",
			input: "DB",
			want:  checkout.NewScore(checkout.BullsEye).WithMultiplier(checkout.Double),
		},
		{
			name:    "disallow too high number",
			input:   "21",
			wantErr: fmt.Errorf("score must be between 1 and 20 (or B for bullseye)"),
		},
		{
			name:    "disallow triple bullseye",
			input:   "TB",
			wantErr: fmt.Errorf("there is no triple bullseye"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := checkout.ParseScore(tt.input)
			if diff := cmp.Diff(gotErr, tt.wantErr, testcommon.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff: %s", diff)
			}

			if tt.wantErr != nil {
				return
			}

			if diff := cmp.Diff(got.String(), tt.want.String()); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}
