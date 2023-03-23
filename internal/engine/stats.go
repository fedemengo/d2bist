package engine

import (
	"math"
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
	blockSize    int
	entropyChunk int
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

func WithBlockSize(blockSize int) Opt {
	return func(o *analysisOpt) {
		o.blockSize = blockSize
	}
}

func WithEntropyChunk(chunkSize int) Opt {
	return func(o *analysisOpt) {
		o.entropyChunk = chunkSize
	}
}

// AnalizeBits count the occurences of bit string of different length
//
// Using a sliding window, bits string up to length = L (4) are counted in O(N), O(L*N) in general
func AnalizeBits(bits []types.Bit, opts ...Opt) *types.Stats {
	o := &analysisOpt{
		topK:         defaultTopK,
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

	for j := 0; j < len(accumulators); j++ {
		windowSize := windows[j]

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
		windowSize := windows[j]

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

	if !calculateEntropy {
		return stats
	}

	for i := 0; i < len(bits); i += o.blockSize {
		chunk := bits[i : i+o.blockSize]
		entropy := calcEntropy(chunk, o.entropyChunk)
		stats.Entropy = append(stats.Entropy, entropy)
	}

	return stats
}

func calcEntropy(chunk []types.Bit, eChunk int) float64 {
	return shannonEntropy(chunk, eChunk)
}

func shannonEntropy(chunk []types.Bit, eChunk int) float64 {

	allBitsSubstr := []int{}
	for i := 0; i < 1<<eChunk; i += eChunk {
		nBits, _ := IntToBits(int64(i), eChunk)
		nDec, _ := BitsToInt(nBits)
		allBitsSubstr = append(allBitsSubstr, int(nDec))
	}

	counts := map[int]int{}
	for i := 0; i < len(chunk); i += eChunk {
		nDec, _ := BitsToInt(chunk[i : i+eChunk])
		counts[int(nDec)]++
	}

	entropy := float64(0)
	for _, bitStr := range allBitsSubstr {
		pX := float64(counts[bitStr]) / float64(len(chunk))
		if pX > 0 {
			entropy -= pX * math.Log2(pX)
		}
	}

	return entropy
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
