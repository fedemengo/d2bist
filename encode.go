package f2bist

import (
	"context"

	"github.com/fedemengo/f2bist/internal/engine"
	fio "github.com/fedemengo/f2bist/internal/io"
	"github.com/fedemengo/f2bist/internal/types"
)

func Encode(ctx context.Context, opts ...Opt) (*types.Result, error) {
	c := &Config{
		MaxBits: -1,
	}
	for _, opt := range opts {
		opt(c)
	}

	bits, err := fio.BitsFromBinStrStdinWithCap(c.MaxBits)
	if err != nil {
		return nil, err
	}

	stats := engine.AnalizeBits(bits)

	return &types.Result{
		Bits:  bits,
		Stats: stats,
	}, nil

}
