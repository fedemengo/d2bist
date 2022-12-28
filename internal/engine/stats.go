package engine

import (
	"github.com/fedemengo/f2bist/internal/types"
)

const (
	maxStringLen = 32
)

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

func AnalizeBits(bits []types.Bit) *types.Stats {
	stats := &types.Stats{
		BitsCount: len(bits),
		ByteCount: len(bits) / 8,
	}

	counters := [4]int64{}
	strLenCount := map[int]map[int64]int{}
	for i := range counters {
		strLenCount[i+1] = map[int64]int{}
	}

	for i, c := range bits {
		for j := 0; j < len(counters); j++ {
			counters[j] &= ^(1 << j)
			counters[j] <<= 1
			counters[j] += int64(c)

			if j > i {
				continue
			}

			strLenCount[j+1][counters[j]]++
		}
	}

	stats.BitsStrCount = strLenCount

	return stats
}
