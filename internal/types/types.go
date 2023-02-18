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

type Stats struct {
	BitsCount int
	ByteCount int

	BitsStrCount map[int]map[int64]int

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

	for i := 0; i < len(s.BitsStrCount); i++ {
		strLen := i
		start := int64(0)
		end := int64(1 << (strLen + 1))

		total, max := 0, 0
		for j := start; j < end; j++ {
			count := s.BitsStrCount[strLen+1][j]
			total += count
			if count > max {
				max = count
			}
		}

		// if not all keys are there we're doing topk
		if len(s.BitsStrCount[strLen+1]) < int(end-start+1) {
			s.printTopBistrK(strLen+1, total, max, w)
		} else {
			s.printAllBistrK(strLen+1, total, max, start, end, w)
		}

		if i < len(s.BitsStrCount)-1 {
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

func (s *Stats) printTopBistrK(kval, total, max int, w io.Writer) {
	topK := heap.NewHeap(func(e1, e2 heap.Elem) bool {
		return e1.Val.(int) > e2.Val.(int)
	})

	for k, v := range s.BitsStrCount[kval] {
		topK.Push(heap.Elem{Key: k, Val: v})
	}

	for topK.Size() > 0 {
		e := topK.Pop()
		k, v := e.Key.(int64), e.Val.(int)
		bitsStr := fmt.Sprintf("%064b", k)
		digits := int(math.Log10(float64(max))) + 1
		count := v

		countStr := fmt.Sprintf("%10d", count)
		if digits > 0 && digits < len(countStr) {
			countStr = countStr[len(countStr)-digits:]
		}
		percentage := float64(count) * 1 / float64(total)

		fmt.Fprintf(w, "%s: %s - %.5f %%\n", bitsStr[len(bitsStr)-kval-2:], countStr, percentage)
	}
}

func (s *Stats) printAllBistrK(kval, total, max int, start, end int64, w io.Writer) {
	for j := start; j < end; j++ {
		bitsStr := fmt.Sprintf("%064b", j)
		count := s.BitsStrCount[kval][j]
		percentage := float64(count) * 1 / float64(total)

		digits := int(math.Log10(float64(max))) + 1
		countStr := fmt.Sprintf("%10d", count)
		if digits > 0 && digits < len(countStr) {
			countStr = countStr[len(countStr)-digits:]
		}

		fmt.Fprintf(w, "%s: %s - %.5f %%\n", bitsStr[len(bitsStr)-kval-2:], countStr, percentage)
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
