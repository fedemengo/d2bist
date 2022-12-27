package core

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/rs/zerolog"

	"github.com/fedemengo/f2bist/compression"
	"github.com/fedemengo/f2bist/internal/engine"
	fio "github.com/fedemengo/f2bist/internal/io"
	"github.com/fedemengo/f2bist/internal/types"
)

func decode(ctx context.Context, r io.Reader, opts ...Opt) (*types.Result, error) {
	c := &Config{
		InMaxBits:         -1,
		InCompressionType: compression.None,
	}
	for _, opt := range opts {
		opt(c)
	}

	cr, err := compression.NewCompressedReader(ctx, r, c.InCompressionType)
	if err != nil {
		return nil, err
	}

	bits, err := fio.BitsFromByteReaderWithCap(ctx, cr, c.InMaxBits)
	if err != nil {
		return nil, fmt.Errorf("cannot read bits from reader: %w", err)
	}

	stats := engine.AnalizeBits(bits)

	return &types.Result{
		Bits:  bits,
		Stats: stats,
	}, nil
}

func bitsToReader(ctx context.Context, bits []types.Bit, compType compression.CompressionType) (fio.ReaderWithSize, error) {
	log := zerolog.Ctx(ctx)

	buf := new(bytes.Buffer)
	cw, err := compression.NewCompressedWriter(ctx, buf, compType)
	if err != nil {
		return nil, fmt.Errorf("cannot get compressed writer: %w", err)
	}

	if err := fio.BitsToWriter(cw, bits); err != nil {
		return nil, fmt.Errorf("cannot compress bits")
	}

	log.Trace().
		Int("bufLen", buf.Len()).
		Int("bits", 8*buf.Len()).
		Msg("bytes written to compression writer")

	if err := cw.Close(); err != nil {
		return nil, fmt.Errorf("error when closing writer: %w", err)
	}

	cr := bytes.NewReader(buf.Bytes())

	return fio.NewReaderWithSize(cr, buf.Len()), nil

}
