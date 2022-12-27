package core

import (
	"context"
	"fmt"
	"io"

	"github.com/rs/zerolog"

	"github.com/fedemengo/f2bist/compression"
	"github.com/fedemengo/f2bist/internal/types"
)

func Decode(ctx context.Context, r io.Reader, opts ...Opt) (*types.Result, error) {
	log := zerolog.Ctx(ctx)

	c := &Config{
		InMaxBits:         -1,
		InCompressionType: compression.None,

		OutMaxBits:         -1,
		OutCompressionType: compression.None,
	}
	for _, opt := range opts {
		opt(c)
	}

	res, err := decode(ctx, r, opts...)
	if err != nil {
		return nil, err
	}

	result := &types.Result{
		Bits:  res.Bits,
		Stats: res.Stats,
	}

	if c.OutCompressionType != compression.None {
		log.Trace().Msg("output requires compression")

		cr, err := bitsToReader(ctx, res.Bits, c.OutCompressionType)
		if err != nil {
			return nil, fmt.Errorf("cannot write bytes to compressed reader: %w", err)
		}

		log.Trace().Int("bits", 8*cr.Size()).Msg("compressed reader ready")

		res, err := decode(ctx, cr, WithOutBitsCap(c.OutMaxBits))
		if err != nil {
			return nil, fmt.Errorf("error decoding from compressed reader: %w", err)
		}

		result.Stats.CompressionStats = &types.CompressionStats{
			CompressionRatio:     100 - float64(len(res.Bits)*100)/float64(len(result.Bits)),
			CompressionAlgorithm: string(c.OutCompressionType),
			Stats:                res.Stats,
		}
		result.Bits = res.Bits
	}

	return result, nil
}
