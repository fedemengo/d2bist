package f2bist

import (
	"context"
	"io"

	"github.com/fedemengo/f2bist/internal/engine"
	fio "github.com/fedemengo/f2bist/internal/io"
	"github.com/fedemengo/f2bist/internal/types"
)

func Encode(_ context.Context, r io.Reader, opts ...Opt) (*types.Result, error) {
	c := &Config{
		MaxBits: -1,
	}
	for _, opt := range opts {
		opt(c)
	}

	bits, err := fio.BitsFromBinStrReaderWithCap(r, c.MaxBits)
	if err != nil {
		return nil, err
	}

	stats := engine.AnalizeBits(bits)

	return &types.Result{
		Bits:  bits,
		Stats: stats,
	}, nil

}
