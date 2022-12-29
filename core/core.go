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

// the reader contains a binary string, representing data, possibly with compression
// the first run to extract the bits data, should always be performed without compression
// once the raw bits have been read, if they represent compressed data, a run of decompression is in order
func binStrReaderToBits(ctx context.Context, r io.Reader, opts ...Opt) ([]types.Bit, error) {
	c := NewDefaultConfig()
	for _, opt := range opts {
		opt(c)
	}

	bits, err := fio.BitsFromBinStrReaderWithCap(ctx, r, c.InMaxBits)
	if err != nil {
		return nil, err
	}

	// the input data was copressed, use a compressed reader to decompress it
	if c.InCompressionType != compression.None {
		// convert compressed bits to byte reader (of compressed data), no additional compression
		r, err := bitsToReader(ctx, bits, compression.None)
		if err != nil {
			return nil, err
		}

		bits, err = readerToBits(ctx, r, WithInCompression(c.InCompressionType))
		if err != nil {
			return nil, fmt.Errorf("error decoding from compressed reader: %w", err)
		}
	}

	if c.OutMaxBits > 0 {
		bitsCap := min(c.OutMaxBits, len(bits))
		bits = bits[:bitsCap]
	}

	return bits, nil
}

func readerToBits(ctx context.Context, r io.Reader, opts ...Opt) ([]types.Bit, error) {
	c := NewDefaultConfig()
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

	if c.OutMaxBits > 0 {
		bitsCap := min(c.OutMaxBits, len(bits))
		bits = bits[:bitsCap]
	}

	return bits, nil
}

func bitsToReader(ctx context.Context, bits []types.Bit, compType compression.CompressionType) (fio.ReaderWithSize, error) {
	log := zerolog.Ctx(ctx).
		With().
		Str("compression", string(compType)).
		Logger()

	log.Trace().
		Msg("creating writer with compression")
	buf := new(bytes.Buffer)
	cw, err := compression.NewCompressedWriter(ctx, buf, compType)
	if err != nil {
		return nil, fmt.Errorf("cannot get compressed writer: %w", err)
	}

	log.Trace().
		Msg("writing bits to comp writer")
	if err := fio.WriteBits(cw, bits); err != nil {
		return nil, fmt.Errorf("cannot compress bits")
	}

	if err := cw.Close(); err != nil {
		return nil, fmt.Errorf("error when closing writer: %w", err)
	}

	log.Trace().
		Int("bufLen", buf.Len()).
		Int("bits", 8*buf.Len()).
		Msg("bytes written to compression writer")

	cr := bytes.NewReader(buf.Bytes())

	return fio.NewReaderWithSize(cr, buf.Len()), nil
}

func createResult(ctx context.Context, bits []types.Bit, opts ...Opt) (*types.Result, error) {
	log := zerolog.Ctx(ctx)

	c := NewDefaultConfig()
	for _, opt := range opts {
		opt(c)
	}

	log.Trace().
		Int("outBitsCap", c.OutMaxBits).
		Str("outCompression", string(c.OutCompressionType)).
		Msg("creating result")

	result := &types.Result{
		Bits:  bits,
		Stats: engine.AnalizeBits(bits),
	}

	if c.OutCompressionType == compression.None {
		return result, nil
	}

	log.Trace().Msg("output requires compression")

	cr, err := bitsToReader(ctx, bits, c.OutCompressionType)
	if err != nil {
		return nil, fmt.Errorf("cannot write bytes to compressed reader: %w", err)
	}

	log.Trace().Int("bits", 8*cr.Size()).Msg("compressed reader ready")

	compressedBits, err := readerToBits(ctx, cr, WithOutBitsCap(c.OutMaxBits))
	if err != nil {
		return nil, fmt.Errorf("error decoding from compressed reader: %w", err)
	}

	result.Bits = compressedBits
	result.Stats.CompressionStats = &types.CompressionStats{
		CompressionRatio:     100 - float64(len(compressedBits)*100)/float64(len(bits)),
		CompressionAlgorithm: string(c.OutCompressionType),
		Stats:                engine.AnalizeBits(compressedBits),
	}

	return result, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
