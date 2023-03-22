package types

import (
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/fedemengo/go-data-structures/heap"
)

var ErrInvalidBit = errors.New("invalid bit")

type Bit uint8

func (b Bit) ToByte() byte {
	switch b {
	case 0:
		return byte('0')
	case 1:
		return byte('1')
	default:
		return byte('-')
	}
}

type SubstrCount struct {
	Total  int
	Length int
	Counts map[string]int

	SortedSubstrs []string
	AllCounts     map[string]int
}

type Stats struct {
	BitsCount int
	ByteCount int

	BitsStrCount map[int]map[int64]int
	SubstrsCount []SubstrCount

	CompressionStats *CompressionStats
}

func (s *Stats) RenderStats(w io.Writer) {
	if s.CompressionStats != nil {
		s.BitsStrCount = map[int]map[int64]int{
			1: s.BitsStrCount[1],
			2: s.BitsStrCount[2],
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "bits:", s.BitsCount)
	fmt.Fprintln(w)

	for l := 0; l < len(s.SubstrsCount); l++ {
		substrGroup := s.SubstrsCount[l]

		total, max := 0, 0
		for _, count := range substrGroup.Counts {
			total += count
			if count > max {
				max = count
			}
		}

		if len(substrGroup.Counts) < len(substrGroup.AllCounts) {
			s.printTopBistrK(max, total, substrGroup, w)
		} else {
			s.printAllBistrK(max, total, substrGroup, w)
		}

		if l < len(s.SubstrsCount)-1 {
			fmt.Fprintln(w)
		}
	}

	if s.CompressionStats != nil {
		fmt.Fprintf(w, `
compression ratio: %.3f
compression algorithm: %s
`, s.CompressionStats.CompressionRatio, s.CompressionStats.CompressionAlgorithm)
		s.CompressionStats.Stats.RenderStats(w)
	} else {
		fmt.Fprintln(w)
	}

}

func (s *Stats) printTopBistrK(max, total int, substrGroup SubstrCount, w io.Writer) {
	topK := heap.NewHeap(func(e1, e2 heap.Elem) bool {
		return e1.Val.(int) > e2.Val.(int)
	})

	for substr, count := range substrGroup.Counts {
		topK.Push(heap.Elem{Key: substr, Val: count})
	}

	for topK.Size() > 0 {
		e := topK.Pop()
		bitStr, v := e.Key.(string), e.Val.(int)
		digits := int(math.Log10(float64(max))) + 1
		count := v

		countStr := fmt.Sprintf("%10d", count)
		if digits > 0 && digits < len(countStr) {
			countStr = countStr[len(countStr)-digits:]
		}
		percentage := float64(count) * 1 / float64(total)

		fmt.Fprintf(w, "%s: %s - %.5f %%\n", bitStr, countStr, percentage)
	}
}

func (s *Stats) printAllBistrK(max, total int, substrGroup SubstrCount, w io.Writer) {
	for _, bitStr := range substrGroup.SortedSubstrs {
		count, ok := substrGroup.Counts[bitStr]
		if !ok {
			continue
		}

		percentage := float64(count) / float64(total)

		digits := int(math.Log10(float64(max))) + 1
		countStr := fmt.Sprintf("%10d", count)
		if digits > 0 && digits < len(countStr) {
			countStr = countStr[len(countStr)-digits:]
		}

		fmt.Fprintf(w, "%s: %s - %.5f %%\n", bitStr, countStr, percentage)
	}
}

type CompressionStats struct {
	CompressionRatio     float64
	CompressionAlgorithm string

	Stats *Stats
}

type Result struct {
	Bits  []Bit
	Stats *Stats
}
