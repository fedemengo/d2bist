package stats_test

import (
	"context"
	"math"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fedemengo/d2bist/pkg/engine"
	"github.com/fedemengo/d2bist/pkg/stats"
	"github.com/fedemengo/d2bist/pkg/types"
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

	bits32 := make([][]types.Bit, 32)
	for i := range bits32 {
		bits32[i], _ = engine.IntToBits(uint64(i), 5)
	}

	testCases := []struct {
		name          string
		bits          []types.Bit
		expectedLen   int
		lenSymbol     int
		expectedValue float64
	}{
		{
			name:          "zero entropy",
			bits:          nZeros(100),
			expectedLen:   100,
			lenSymbol:     2,
			expectedValue: 0,
		}, {
			name:          "max entropy",
			bits:          []types.Bit{0, 0, 0, 1, 1, 0, 1, 1},
			expectedLen:   8,
			lenSymbol:     2,
			expectedValue: 1,
		}, {
			name: "analytical test, 128 bits, 8 bit symbol",
			bits: append(
				[]types.Bit{0, 0, 0, 0, 0, 0, 0, 1},
				nZeros(120)...,
			),
			expectedLen:   128,
			lenSymbol:     8,
			expectedValue: 0.0843,
		}, {
			name:          "analytical max test, 160 bits, 5 bit symbol",
			bits:          flattenBits(bits32),
			expectedLen:   160,
			lenSymbol:     5,
			expectedValue: 1,
		}, {
			name:          "analytical not max test, 9 bits, 3 bit symbol",
			bits:          []types.Bit{0, 0, 0, 0, 0, 1, 0, 1, 1},
			expectedLen:   9,
			lenSymbol:     3,
			expectedValue: 1,
		}, {
			name:          "analytical max test, 24 bits, 3 bit symbol",
			bits:          []types.Bit{0, 0, 0, 0, 0, 1, 0, 1, 0, 0, 1, 1, 1, 0, 0, 1, 0, 1, 1, 1, 0, 1, 1, 1},
			expectedLen:   24,
			lenSymbol:     3,
			expectedValue: 1,
		},
	}

	ctx := context.Background()
	log := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
	ctx = log.WithContext(ctx)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(tc.expectedLen, len(tc.bits))

			e := stats.ShannonEntropy(ctx, tc.bits, len(tc.bits), tc.lenSymbol)
			require.Len(t, e.Values, 1)

			assert.LessOrEqualf(math.Abs(tc.expectedValue-e.Values[0]), 1e-3, "expected %.4f got %.4f", tc.expectedValue, e.Values[0])
		})
	}
}

func flattenBits(bits [][]types.Bit) []types.Bit {
	var res []types.Bit

	for _, b := range bits {
		res = append(res, b...)
	}

	return res
}

func nZeros(n int) []types.Bit {
	return make([]types.Bit, n)
}
