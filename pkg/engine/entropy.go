package engine

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/rs/zerolog"

	"github.com/fedemengo/d2bist/pkg/types"
)

var ErrEndOfBits = errors.New("end of bits")

func ShannonEntropy(ctx context.Context, bits []types.Bit, chunkSize, symbolLen int) *types.Entropy {
	shannon := types.NewShannonEntropy()
	for i := 0; i < len(bits); i += chunkSize {
		nextBlockSize := min(chunkSize, len(bits)-i)
		chunk := bits[i : i+nextBlockSize]
		entropy := shannonEntropy(ctx, chunk, symbolLen)
		shannon.Values = append(shannon.Values, entropy)
	}

	return shannon
}

func shannonEntropy(ctx context.Context, chunk []types.Bit, symbolLen int) float64 {
	log := zerolog.Ctx(ctx).With().Logger()

	bw := NewBitsWindow(chunk, symbolLen)

	chunkLen := len(chunk)
	counts := make(map[uint64]int, len(chunk)/symbolLen)
	for i := 0; i < chunkLen; i += symbolLen {
		bitsInt := bw.ToInt()
		counts[bitsInt]++
		bw.SlideBy(symbolLen)
	}

	entropy := float64(0)

	symbolsCount := (float64(chunkLen) / float64(symbolLen))
	for intBitsValue, count := range counts {
		pX := float64(count) / symbolsCount
		e := float64(0)
		if pX > 0 {
			e = pX * math.Log2(pX)
		}
		log.Trace().
			Uint64("value", intBitsValue).
			Str("bitstr", fmt.Sprintf("%032b", intBitsValue)).
			Int("count", counts[intBitsValue]).
			Float64("pX", pX).
			Float64("entropy", e).
			Msg("symbol count")

		entropy += e
	}

	// normalize the entropy in the range [0, 1]
	entropy /= math.Log2(symbolsCount)

	entropy = -entropy

	log.Debug().Float64("entropy", entropy).Msg("entropy for chunk calculated")

	return entropy
}
