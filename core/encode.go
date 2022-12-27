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
	}
	for _, opt := range opts {
		opt(c)
	}

	bits, err := fio.BitsFromBinStrReaderWithCap(ctx, r, c.InMaxBits)
	if err != nil {
		return nil, err
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

		res, err := decode(ctx, cr, WithOutBitsCap(c.OutMaxBits))
		if err != nil {
			return nil, fmt.Errorf("error decoding from compressed reader: %w", err)
		}

		result.Bits = res.Bits
	}

	return result, nil
}
