package core

import (
	"context"
	"fmt"
	"io"

	"github.com/fedemengo/f2bist/compression"
	"github.com/fedemengo/f2bist/internal/engine"
	fio "github.com/fedemengo/f2bist/internal/io"
	"github.com/fedemengo/f2bist/internal/types"
)

// Encode receives a bit string in a io.Reader and creates a Result
func Encode(ctx context.Context, r io.Reader, opts ...Opt) (*types.Result, error) {
	c := &Config{
		InMaxBits:         -1,
		InCompressionType: compression.None,

		OutMaxBits:         -1,
		OutCompressionType: compression.None,
	}
	for _, opt := range opts {
		opt(c)
	}

	bits, err := fio.BitsFromBinStrReaderWithCap(ctx, r, c.InMaxBits)
	if err != nil {
		return nil, err
	}

	if c.InCompressionType != compression.None {
		// convert compressed bits to byte reader, no additional compression
		r, err := bitsToReader(ctx, bits, compression.None)
		if err != nil {
			return nil, err
		}

		// read compressed byte from reader, with compression
		cr, err := compression.NewCompressedReader(ctx, r, c.InCompressionType)
		if err != nil {
			return nil, err
		}

		res, err := decode(ctx, cr)
		if err != nil {
			return nil, fmt.Errorf("error decoding from compressed reader: %w", err)
		}

		bits = res.Bits
	}

	if c.OutMaxBits > 0 {
		bitsCap := min(c.OutMaxBits, len(bits))
		bits = bits[:bitsCap]
	}

	stats := engine.AnalizeBits(bits)

	result := &types.Result{
		Bits:  bits,
		Stats: stats,
	}

	if c.OutCompressionType != compression.None {
		cr, err := bitsToReader(ctx, bits, c.OutCompressionType)
		if err != nil {
			return nil, fmt.Errorf("cannot write bytes to compressed reader: %w", err)
		}

		res, err := decodeWithAnalysis(ctx, cr, WithOutBitsCap(c.OutMaxBits))
		if err != nil {
			return nil, fmt.Errorf("error decoding from compressed reader: %w", err)
		}

		result.Stats.CompressionStats = &types.CompressionStats{
			CompressionRatio:     100 - float64(len(res.Bits)*100)/float64(len(result.Bits)),
			CompressionAlgorithm: string(c.OutCompressionType),
			Stats:                res.Stats,
		}
		result.Bits = res.Bits

		result.Bits = res.Bits
	}

	return result, nil
}
