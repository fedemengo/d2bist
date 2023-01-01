package engine

import (
	"github.com/fedemengo/d2bist/internal/types"
)

// AnalizeBits count the occurences of bit string of different length
//
// Using a sliding window, bits string up to length = L (4) are counted in O(N), O(L*N) in general
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
			counters[j] &= ^(1 << j) // clear exiting bits
			counters[j] <<= 1        // align all j-1 bits
			counters[j] += int64(c)  // add entering bit

			// window of length j is not full yet
			if j > i {
				continue
			}

			strLenCount[j+1][counters[j]]++
		}
	}

	stats.BitsStrCount = strLenCount

	return stats
}
