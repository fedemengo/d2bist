package engine

import (
	"log"

	"github.com/fedemengo/f2bist/internal/types"
)

const (
	maxStringLen = 32
)

type Stats struct {
	ZeroCount int
	OneCount  int

	SizeBits  int
	SizeBytes int

	ZeroStrings  map[int]int
	OneStrings   map[int]int
	MaxStringLen int
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func AnalizeBits(bits []types.Bit) *Stats {
	zeroC, oneC := 0, 0
	zeroL, oneL := 0, 0

	stats := &Stats{
		ZeroStrings: make(map[int]int),
		OneStrings:  make(map[int]int),
	}

	for _, c := range bits {
		switch c {
		case 0:
			zeroC++

			zeroL++
			if oneL > 0 && oneL < maxStringLen {
				stats.OneStrings[oneL]++
			}
			oneL = 0
		case 1:
			oneC++

			oneL++
			if zeroL > 0 && zeroL < maxStringLen {
				stats.ZeroStrings[zeroL]++
			}
			zeroL = 0
		default:
			log.Fatalf("digit %c should not be here", c)
		}

		stats.MaxStringLen = min(max(stats.MaxStringLen, max(zeroL, oneL)), maxStringLen)
	}

	if zeroL > 0 && zeroL < maxStringLen {
		stats.ZeroStrings[zeroL]++
	}
	if oneL > 0 && oneL < maxStringLen {
		stats.OneStrings[oneL]++
	}

	stats.MaxStringLen = min(max(stats.MaxStringLen, max(zeroL, oneL)), maxStringLen)

	stats.ZeroCount = zeroC
	stats.OneCount = oneC
	stats.SizeBits = len(bits)
	stats.SizeBytes = len(bits) / 8

	return stats
}
