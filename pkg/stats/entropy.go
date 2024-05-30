package stats

import (
	"context"
	"fmt"
	"math"

	"github.com/rs/zerolog"

	"github.com/fedemengo/d2bist/pkg/compression"
	"github.com/fedemengo/d2bist/pkg/engine"
	iio "github.com/fedemengo/d2bist/pkg/io"
	"github.com/fedemengo/d2bist/pkg/types"
)

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

	bw := engine.NewBitsWindow(chunk, symbolLen)

	chunkLen := len(chunk)
	log.Debug().Int("chunkLen", chunkLen).Int("symbolLen", symbolLen).Msg("calculating entropy")
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

func CompressionEntropy(ctx context.Context, bits []types.Bit, chunkSize, _ int, cType compression.CompressionType) *types.Entropy {
	compr := types.NewCompressionEntropy(cType)

	for i := 0; i < len(bits); i += chunkSize {
		nextBlockSize := min(chunkSize, len(bits)-i)
		chunk := bits[i : i+nextBlockSize]

		e := float64(0)
		cr, err := iio.BitsToReader(ctx, chunk, cType)
		if err == nil {
			e = float64(cr.Size()*8) / float64(len(chunk))
		}
		if e > 1 {
			e = 1
		}

		compr.Values = append(compr.Values, e)
	}

	return compr
}
