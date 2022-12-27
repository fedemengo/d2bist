package core

import (
	"context"
	"io"

	"github.com/fedemengo/f2bist/internal/engine"
	fio "github.com/fedemengo/f2bist/internal/io"
	"github.com/fedemengo/f2bist/internal/types"
)

func Decode(_ context.Context, r io.Reader, opts ...Opt) (*types.Result, error) {
	c := &Config{
		MaxBits: -1,
	}
	for _, opt := range opts {
		opt(c)
	}

	bits, err := fio.BitsFromByteReaderWithCap(r, c.MaxBits)
	if err != nil {
		return nil, err
	}

	stats := engine.AnalizeBits(bits)

	return &types.Result{
		Bits:  bits,
		Stats: stats,
	}, nil
}
