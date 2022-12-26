package f2bist

import (
	"context"
	"io"

	"github.com/fedemengo/f2bist/internal/engine"
	fio "github.com/fedemengo/f2bist/internal/io"
	"github.com/fedemengo/f2bist/internal/types"
)

type Result struct {
	Bits  []types.Bit
	Stats *types.Stats
}

type config struct {
	maxBits int
}

type Opt func(c *config)

func WithBitsCap(maxBits int) Opt {
	return func(c *config) {
		c.maxBits = maxBits
	}
}

func Decode(ctx context.Context, r io.Reader, opts ...Opt) (*Result, error) {
	c := &config{
		maxBits: -1,
	}
	for _, opt := range opts {
		opt(c)
	}

	bits, err := fio.BitsFromReaderWithCap(r, c.maxBits)
	if err != nil {
		return nil, err
	}

	stats := engine.AnalizeBits(bits)

	return &Result{
		Bits:  bits,
		Stats: stats,
	}, nil
}
