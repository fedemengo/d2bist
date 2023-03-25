package engine_test

import (
	"math"
	"testing"

	"github.com/fedemengo/d2bist/internal/engine"
	"github.com/fedemengo/d2bist/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBitsWindow(t *testing.T) {
	testCases := []struct {
		name           string
		bits           []types.Bit
		windowSize     int
		slide          int
		expectedBefore uint64
		expectedAfter  uint64
		expectedToErr  bool
	}{
		{
			name:           "window size 3",
			bits:           []types.Bit{1, 1, 1, 0},
			windowSize:     3,
			slide:          1,
			expectedBefore: 7,
			expectedAfter:  6,
		}, {
			name:           "window size 3 slide 2",
			bits:           []types.Bit{1, 1, 1, 0, 0, 1},
			windowSize:     3,
			slide:          2,
			expectedBefore: 7,
			expectedAfter:  4,
		}, {
			name:           "window size 3 slide 4",
			bits:           []types.Bit{1, 1, 1, 0, 0, 1, 1},
			windowSize:     3,
			slide:          4,
			expectedBefore: 7,
			expectedAfter:  3,
		}, {
			name:           "slide outside window",
			bits:           []types.Bit{1, 1, 1, 0, 0, 1, 1},
			windowSize:     3,
			slide:          6,
			expectedBefore: 7,
			expectedToErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			bw := engine.NewBitsWindow(tc.bits, tc.windowSize)

			assert.Equalf(tc.expectedBefore, bw.ToInt(), "before sliding")

			err := bw.SlideBy(tc.slide)
			if tc.expectedToErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equalf(tc.expectedAfter, bw.ToInt(), "after sliding")
		})
	}
}

func TestShannongEntropy(t *testing.T) {
	testCases := []struct {
		name          string
		bits          []types.Bit
		lenSymbol     int
		expectedValue float64
	}{
		{
			name:          "zero entropy",
			bits:          nZeros(100),
			lenSymbol:     2,
			expectedValue: 0,
		}, {
			name:          "max entropy",
			bits:          []types.Bit{0, 0, 0, 1, 1, 0, 1, 1},
			lenSymbol:     2,
			expectedValue: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			e := engine.ShannonEntropy(tc.bits, tc.lenSymbol)

			assert.LessOrEqual(math.Abs(tc.expectedValue-e), 1e-3)
		})
	}
}

func nZeros(n int) []types.Bit {
	return make([]types.Bit, n)
}
