package types

import (
	"errors"
	"fmt"
	"io"
	"math"
)

var ErrInvalidBit = errors.New("invalid bit")

type Bit uint8

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

		for j := start; j < end; j++ {
			bitsStr := fmt.Sprintf("%064b", j)
			count := s.BitsStrCount[strLen+1][j]
			percentage := float64(count) * 1 / float64(total)

			digits := int(math.Log10(float64(max))) + 1
			countStr := fmt.Sprintf("%10d", count)
			if digits > 0 && digits < len(countStr) {
				countStr = countStr[len(countStr)-digits:]
			}

			fmt.Fprintf(w, "%s: %s - %.3f %%\n", bitsStr[len(bitsStr)-strLen-1:], countStr, percentage)
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

type CompressionStats struct {
	CompressionRatio     float64
	CompressionAlgorithm string

	Stats *Stats
}

type Result struct {
	Bits  []Bit
	Stats *Stats
}
