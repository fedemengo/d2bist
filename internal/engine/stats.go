package engine

import (
	"context"

	"github.com/fedemengo/go-data-structures/heap"

	"github.com/fedemengo/d2bist/internal/types"
)

const (
	defaultTopK         = 10
	defaultMaxBlockSize = 8
)

type analysisOpt struct {
	topKFreq     int
	maxBlockSize int
	blockSize    int
	symbolLen    int
}

type Opt func(*analysisOpt)

func WithTopKFreq(topKFreq int) Opt {
	return func(o *analysisOpt) {
		o.topKFreq = topKFreq
	}
}

func WithMaxBlockSize(maxBlockSize int) Opt {
	return func(o *analysisOpt) {
		o.maxBlockSize = maxBlockSize
	}
}

func WithBlockSize(blockSize int) Opt {
	return func(o *analysisOpt) {
		o.blockSize = blockSize
	}
}

func WithSymbolLen(chunkSize int) Opt {
	return func(o *analysisOpt) {
		o.symbolLen = chunkSize
	}
}

// AnalizeBits count the occurences of bit string of different length
//
// Using a sliding window, bits string up to length = L (4) are counted in O(N), O(L*N) in general
func AnalizeBits(ctx context.Context, bits []types.Bit, opts ...Opt) *types.Stats {
	o := &analysisOpt{
		topKFreq:     defaultTopK,
		maxBlockSize: defaultMaxBlockSize,
		blockSize:    -1,
	}

	for _, opt := range opts {
		opt(o)
	}

	stats := &types.Stats{
		BitsCount: len(bits),
		ByteCount: len(bits) / 8,
	}

	var windows []int
	var accumulators []string
	counterForLen := map[int]map[string]int{}

	calculateEntropy := false
	// if blockSize is set, calculate entropy for that window size
	if o.blockSize > 0 {
		calculateEntropy = true
		accumulators = make([]string, 1)
		windows = []int{o.blockSize}
		counterForLen[o.blockSize] = map[string]int{}
	} else {
		accumulators = make([]string, o.maxBlockSize)
		for i := range accumulators {
			windows = append(windows, i+1)
			counterForLen[i+1] = map[string]int{}
		}
	}

	// count all bit strings of length windowSize
	for j, windowSize := range windows {
		// count all bit strings of length windowSize
		for i, b := range bits {
			currAcc := accumulators[j]
			if len(accumulators[j]) > windowSize {
				currAcc = accumulators[j][1:] // exit the window
			}
			nextAcc := currAcc + string(b.ToByte()) // enter the window
			accumulators[j] = nextAcc

			// window is not full yet
			if i+1 < windowSize {
				continue
			}

			counterForLen[windowSize][nextAcc[:windowSize]]++
		}
	}

	// select top K most frequent bit strings
	for _, windowSize := range windows {
		topKSelected := getTopKFreqSubstrs(o.topKFreq, counterForLen[windowSize])

		subtrs := []string{}
		for substr := range counterForLen[windowSize] {
			subtrs = append(subtrs, substr)
		}
		//sort.Strings(subtrs)

		stats.SubstrsCount = append(stats.SubstrsCount, types.SubstrCount{
			Length:        windowSize,
			Total:         len(counterForLen[windowSize]),
			AllCounts:     counterForLen[windowSize],
			SortedSubstrs: subtrs,
			Counts:        topKSelected,
		})
	}

	if !calculateEntropy {
		return stats
	}

	shannon := types.NewShannonEntropy()
	for i := 0; i < len(bits); i += o.blockSize {
		nextBlockSize := min(o.blockSize, len(bits)-i)
		chunk := bits[i : i+nextBlockSize]
		entropy := ShannonEntropy(chunk, o.symbolLen)
		shannon.Values = append(shannon.Values, entropy)
	}

	stats.Entropy = append(stats.Entropy, shannon)

	return stats
}

func getTopKFreqSubstrs(k int, counterForLen map[string]int) map[string]int {
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
