package core

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/rs/zerolog"

	"github.com/fedemengo/d2bist/pkg/compression"
	"github.com/fedemengo/d2bist/pkg/engine"
	iio "github.com/fedemengo/d2bist/pkg/io"
	"github.com/fedemengo/d2bist/pkg/types"
)

// the reader contains a binary string, representing data, possibly with compression
// the first run to extract the bits data, should always be performed without compression
// once the raw bits have been read, if they represent compressed data, a run of decompression is in order
func binStrReaderToBits(ctx context.Context, r io.Reader, opts ...Opt) ([]types.Bit, error) {
	log := zerolog.Ctx(ctx)
	c := NewDefaultConfig()
	for _, opt := range opts {
		opt(c)
	}

	bits, err := iio.BitsFromBinStrReaderWithCap(ctx, r, c.InMaxBits)
	if err != nil {
		return nil, err
	}

	log.Trace().Msgf("read %d bits", len(bits))

	// the input data was copressed, use a compressed reader to decompress it
	if c.InCompressionType != compression.None {
		log.Trace().
			Str("compression", string(c.InCompressionType)).
			Msg("bits requires decompression")

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

	bits, err := iio.BitsFromByteReaderWithCap(ctx, cr, c.InMaxBits)
	if err != nil {
		return nil, fmt.Errorf("cannot read bits from reader: %w", err)
	}

	if c.OutMaxBits > 0 {
		bitsCap := min(c.OutMaxBits, len(bits))
		bits = bits[:bitsCap]
	}

	return bits, nil
}

func bitsToReader(ctx context.Context, bits []types.Bit, compType compression.CompressionType) (iio.ReaderWithSize, error) {
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
	if err := iio.BitsToByteWriter(ctx, cw, bits); err != nil {
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

	return iio.NewReaderWithSize(cr, buf.Len()), nil
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

	engineOpts := []engine.Opt{
		engine.WithMaxBlockSize(c.StatsMaxBlockSize),
		engine.WithTopKFreq(c.StatsTopK),
		engine.WithSymbolLen(c.StatsSymbolLen),
	}

	if c.StatsBlockSize > 0 {
		engineOpts = append(engineOpts, engine.WithBlockSize(c.StatsBlockSize))
	}

	log.Trace().
		Int("statsBlockSize", c.StatsBlockSize).
		Int("statsMaxBlockSize", c.StatsMaxBlockSize).
		Int("statsTopK", c.StatsTopK).
		Msg("analizing bits")

	stats := engine.AnalizeBits(ctx, bits, engineOpts...)
	stats.EntropyPlotName = c.EntropyPlotName

	result := &types.Result{
		Bits:  bits,
		Stats: stats,
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
		Stats:                engine.AnalizeBits(ctx, compressedBits),
	}

	return result, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
