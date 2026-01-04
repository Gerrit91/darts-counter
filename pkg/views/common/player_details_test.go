package common

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_DistributeColors(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		from   string
		to     string
		values map[int]string
		want   map[int]string
	}{
		{
			from: "#000000",
			to:   "#ffffff",
			values: map[int]string{
				0:   "",
				50:  "",
				100: "",
			},
			want: map[int]string{
				0:   "#000000",
				50:  "#808080",
				100: "#ffffff",
			},
		},
		{
			from: "#909090",
			to:   "#e3ad0b",
			values: map[int]string{
				0:   "",
				50:  "",
				100: "",
			},
			want: map[int]string{
				0:   "#909090",
				50:  "#ba9f4e",
				100: "#e3ad0b",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DistributeColors(tt.from, tt.to, tt.values)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("diff = %s", diff)
			}
		})
	}
}
