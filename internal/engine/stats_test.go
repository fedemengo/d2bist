package engine

import (
	"testing"

	"github.com/fedemengo/f2bist/internal/types"
)

func TestStats(t *testing.T) {
	testCases := []struct {
		name            string
		bits            []types.Bit
		expectedZeroMap map[int]int
		expectedOneMap  map[int]int
	}{
		{
			name:            "only single digit",
			bits:            []types.Bit{0, 1, 0, 1, 0, 1, 0, 1},
			expectedZeroMap: map[int]int{1: 4},
			expectedOneMap:  map[int]int{1: 4},
		}, {
			name:            "only two digit",
			bits:            []types.Bit{0, 0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1},
			expectedZeroMap: map[int]int{2: 4},
			expectedOneMap:  map[int]int{2: 4},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			//a := assert.New(tt)

			//stats := AnalizeBits(tc.bits)

			//a.EqualValues(tc.expectedZeroMap, stats.ZeroStrings)
			//a.EqualValues(tc.expectedOneMap, stats.OneStrings)

		})
	}

}
