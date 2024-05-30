package stats

import (
	"context"
	"sort"

	"github.com/rs/zerolog"

	"github.com/fedemengo/go-data-structures/heap"

	"github.com/fedemengo/d2bist/pkg/compression"
	"github.com/fedemengo/d2bist/pkg/engine"
	"github.com/fedemengo/d2bist/pkg/types"
)

const (
	defaultTopK         = 10
	defaultMaxBlockSize = 8
	defaultSymbolLen    = 2
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
	log := zerolog.Ctx(ctx)

	o := &analysisOpt{
		topKFreq:     defaultTopK,
		maxBlockSize: defaultMaxBlockSize,
		blockSize:    -1,
		symbolLen:    defaultSymbolLen,
	}

	for _, opt := range opts {
		opt(o)
	}

	stats := &types.Stats{
		BitsCount: len(bits),
		ByteCount: len(bits) / 8,
	}

	var windows []int
	var accumulators []uint64
	counterForLen := map[int]map[uint64]int{}

	calculateEntropy := false
	// if blockSize is set, calculate entropy for that window size
	if o.blockSize > 0 {
		calculateEntropy = true
		accumulators = make([]uint64, 1)
		windows = []int{o.blockSize}
		counterForLen[o.blockSize] = make(map[uint64]int, o.blockSize)
	} else {
		accumulators = make([]uint64, o.maxBlockSize)
		for i := range accumulators {
			windows = append(windows, i+1)
			counterForLen[i+1] = map[uint64]int{}
		}
	}

	log.Info().
		Ints("windows", windows).
		Bool("entropyCalc", calculateEntropy).
		Int("symbolLen", o.symbolLen).
		Msg("counting bit strings")

	// count all bit strings of length windowSize
	for j, windowSize := range windows {
		log.Trace().
			Int("windowSize", windowSize).
			Msg("counting bit strings")
		// count all bit strings of length windowSize
		for i, b := range bits {
			if i%8_000 == 0 {
				log.Trace().
					Int("windowSize", windowSize).
					Int("bitsCount", i).
					Int("totalBits", len(bits)).
					Msg("bits processed")
			}

			accumulators[j] &= ^(1 << j) // clear exiting bits
			accumulators[j] <<= 1        // align all j-1 bits
			accumulators[j] |= uint64(b) // add new bit

			// window is not full yet
			if i+1 < windowSize {
				continue
			}

			counterForLen[windowSize][accumulators[j]]++
		}
	}

	// select top K most frequent bit strings
	for _, windowSize := range windows {
		topKSelected := getTopKFreqSubstrs(o.topKFreq, counterForLen[windowSize])

		strAllCount := map[string]int{}
		strTopKSelected := map[string]int{}
		strSubstrs := []string{}

		for substr, count := range counterForLen[windowSize] {
			s, err := engine.IntToBitString(substr, windowSize)
			if err != nil {
				log.Fatal().Err(err).Msg("error converting int to bit string")
			}

			strSubstrs = append(strSubstrs, s)
			strAllCount[s] = count
			if _, ok := topKSelected[substr]; ok {
				strTopKSelected[s] = count
			}
		}
		sort.Strings(strSubstrs)

		stats.SubstrsCount = append(stats.SubstrsCount, types.SubstrCount{
			Length:        windowSize,
			Total:         len(counterForLen[windowSize]),
			AllCounts:     strAllCount,
			SortedSubstrs: strSubstrs,
			Counts:        strTopKSelected,
		})
	}

	if !calculateEntropy {
		return stats
	}

	log.Trace().
		Ints("windows", windows).
		Bool("entropyCalc", calculateEntropy).
		Msg("calculating entropy")

	stats.Entropy = append(
		stats.Entropy,
		CompressionEntropy(ctx, bits, o.blockSize, o.symbolLen, compression.Gzip),
		CompressionEntropy(ctx, bits, o.blockSize, o.symbolLen, compression.Brotli),
		CompressionEntropy(ctx, bits, o.blockSize, o.symbolLen, compression.Bzip2),
		ShannonEntropy(ctx, bits, o.blockSize, o.symbolLen),
	)

	log.Trace().Msg("done bits analysis")

	return stats
}

func getTopKFreqSubstrs(k int, counterForLen map[uint64]int) map[uint64]int {
	if k <= 0 {
		return counterForLen
	}

	// max heap
	topK := heap.NewHeap(func(e1, e2 heap.Elem) bool {
		return e1.Key.(int) < e2.Key.(int)
	})

	// select top K bit strings of length j+1
	//
	// by putting all elements in a max heap
	for substr, count := range counterForLen {
		if topK.Size() < k {
			topK.Push(heap.Elem{Key: count, Val: substr})
		} else if count < topK.Front().Key.(int) {
			topK.Pop()
			topK.Push(heap.Elem{Key: count, Val: substr})
		}
	}

	topKSelected := map[uint64]int{}
	for topK.Size() > 0 {
		e := topK.Pop()
		topKSelected[e.Val.(uint64)] = e.Key.(int)
	}

	return topKSelected
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
