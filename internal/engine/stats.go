package engine

import (
	"sort"

	"github.com/fedemengo/go-data-structures/heap"

	"github.com/fedemengo/d2bist/internal/types"
)

const (
	defaultTopK         = 10
	defaultMaxBlockSize = 8
)

type analysisOpt struct {
	topK         int
	maxBlockSize int
}

type Opt func(*analysisOpt)

func WithTopK(topK int) Opt {
	return func(o *analysisOpt) {
		o.topK = topK
	}
}

func WithMaxBlockSize(maxBlockSize int) Opt {
	return func(o *analysisOpt) {
		o.maxBlockSize = maxBlockSize
	}
}

// AnalizeBits count the occurences of bit string of different length
//
// Using a sliding window, bits string up to length = L (4) are counted in O(N), O(L*N) in general
func AnalizeBits(bits []types.Bit, opts ...Opt) *types.Stats {
	o := &analysisOpt{
		topK:         defaultTopK,
		maxBlockSize: defaultMaxBlockSize,
	}

	for _, opt := range opts {
		opt(o)
	}

	stats := &types.Stats{
		BitsCount: len(bits),
		ByteCount: len(bits) / 8,
	}

	accumulators := make([]string, o.maxBlockSize)
	counterForLen := map[int]map[string]int{}
	for i := range accumulators {
		counterForLen[i+1] = map[string]int{}
	}

	for j := 0; j < len(accumulators); j++ {
		windowSize := j + 1

		// count all bit strings of length windowSize
		for i, b := range bits {
			currAcc := accumulators[j]
			if len(accumulators[j]) > windowSize {
				currAcc = accumulators[j][1:]
			}
			nextAcc := currAcc + string(b.ToByte())
			accumulators[j] = nextAcc

			// window of length windowSize is not full yet
			if i+1 < windowSize {
				continue
			}

			counterForLen[windowSize][nextAcc[:windowSize]]++
		}
	}

	for j := 0; j < len(accumulators); j++ {
		windowSize := j + 1

		subtrs := []string{}
		for substr := range counterForLen[windowSize] {
			subtrs = append(subtrs, substr)
		}
		sort.Strings(subtrs)

		topKSelected := calcTopK(o.topK, counterForLen[windowSize])

		stats.SubstrsCount = append(stats.SubstrsCount, types.SubstrCount{
			Length:        windowSize,
			Total:         len(counterForLen[windowSize]),
			AllCounts:     counterForLen[windowSize],
			SortedSubstrs: subtrs,
			Counts:        topKSelected,
		})
	}

	return stats
}

func calcTopK(k int, counterForLen map[string]int) map[string]int {
	if k <= 0 {
		return counterForLen
	}

	// min heap with negative values
	topK := heap.NewHeap(func(e1, e2 heap.Elem) bool {
		return e1.Val.(int) > e2.Val.(int)
	})

	// select top K bit strings of length j+1
	//
	// by putting all elements in a min heap
	for substr, count := range counterForLen {
		if topK.Size() < k {
			topK.Push(heap.Elem{Key: substr, Val: -count})
		} else if -count < topK.Front().Val.(int) {
			topK.Pop()
			topK.Push(heap.Elem{Key: substr, Val: -count})
		}
	}

	topKSelected := map[string]int{}
	for topK.Size() > 0 {
		e := topK.Pop()
		topKSelected[e.Key.(string)] = -e.Val.(int)
	}

	return topKSelected
}
