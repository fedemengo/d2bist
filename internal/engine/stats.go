package engine

import (
	"github.com/fedemengo/go-data-structures/heap"

	"github.com/fedemengo/d2bist/internal/types"
)

const (
	K = 10
)

// AnalizeBits count the occurences of bit string of different length
//
// Using a sliding window, bits string up to length = L (4) are counted in O(N), O(L*N) in general
func AnalizeBits(bits []types.Bit) *types.Stats {
	stats := &types.Stats{
		BitsCount: len(bits),
		ByteCount: len(bits) / 8,
	}

	counters := [8]int64{}
	strLenCount := map[int]map[int64]int{}
	for i := range counters {
		strLenCount[i+1] = map[int64]int{}
	}

	// min heap with negative values
	topK := heap.NewHeap(func(e1, e2 heap.Elem) bool {
		return e1.Val.(int) > e2.Val.(int)
	})
	for j := 0; j < len(counters); j++ {

		for i, c := range bits {
			counters[j] &= ^(1 << j) // clear exiting bits
			counters[j] <<= 1        // align all j-1 bits
			counters[j] += int64(c)  // add entering bit

			// window of length j is not full yet
			if j > i {
				continue
			}

			strLenCount[j+1][counters[j]]++
		}

		for kmer, count := range strLenCount[j+1] {
			if topK.Size() < K {
				topK.Push(heap.Elem{Key: kmer, Val: -count})
			} else if -count < topK.Front().Val.(int) {
				topK.Pop()
				topK.Push(heap.Elem{Key: kmer, Val: -count})
			}
		}

		topKSelected := map[int64]int{}
		for topK.Size() > 0 {
			e := topK.Pop()
			topKSelected[e.Key.(int64)] = -e.Val.(int)
		}
		strLenCount[j+1] = topKSelected
	}

	stats.BitsStrCount = strLenCount

	return stats
}
