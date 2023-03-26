package engine_test

import (
	"context"
	"math"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fedemengo/d2bist/internal/engine"
	"github.com/fedemengo/d2bist/internal/types"
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
	bits := make([][]types.Bit, 16)
	bits[0], _ = engine.IntToBits(0, 8)
	bits[1], _ = engine.IntToBits(1, 8)
	bits[2], _ = engine.IntToBits(2, 8)
	bits[3], _ = engine.IntToBits(3, 8)
	bits[4], _ = engine.IntToBits(4, 8)
	bits[5], _ = engine.IntToBits(5, 8)
	bits[6], _ = engine.IntToBits(6, 8)
	bits[7], _ = engine.IntToBits(7, 8)
	bits[8], _ = engine.IntToBits(8, 8)
	bits[9], _ = engine.IntToBits(9, 8)
	bits[10], _ = engine.IntToBits(10, 8)
	bits[11], _ = engine.IntToBits(11, 8)
	bits[12], _ = engine.IntToBits(12, 8)
	bits[13], _ = engine.IntToBits(13, 8)
	bits[14], _ = engine.IntToBits(14, 8)
	bits[15], _ = engine.IntToBits(15, 8)

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
		}, {
			name: "analytical test, 128 bits, 8 bit symbol",
			bits: append(
				[]types.Bit{0, 0, 0, 0, 0, 0, 0, 1},
				nZeros(120)...,
			),
			lenSymbol:     8,
			expectedValue: 0.337,
		}, {
			name:          "analytical max test, 128 bits, 8 bit symbol",
			bits:          flattenBits(bits),
			lenSymbol:     8,
			expectedValue: 1,
		},
	}

	ctx := context.Background()
	log := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
	ctx = log.WithContext(ctx)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			e := engine.ShannonEntropy(ctx, tc.bits, tc.lenSymbol)

			assert.LessOrEqualf(math.Abs(tc.expectedValue-e), 1e-3, "expected %.4f got %.4f", tc.expectedValue, e)
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
