package core

import (
	"context"
	"fmt"
	"io"

	"github.com/rs/zerolog"

	"github.com/fedemengo/d2bist/pkg/compression"
	iio "github.com/fedemengo/d2bist/pkg/io"
	"github.com/fedemengo/d2bist/pkg/stats"
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
		r, err := iio.BitsToReader(ctx, bits, compression.None)
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

func createResult(ctx context.Context, bits []types.Bit, opts ...Opt) (*types.Result, error) {
	log := zerolog.Ctx(ctx)

	c := NewDefaultConfig()
	for _, opt := range opts {
		opt(c)
	}

	log.Debug().
		Int("outBitsCap", c.OutMaxBits).
		Str("outCompression", string(c.OutCompressionType)).
		Int("symbolLen", c.StatsSymbolLen).
		Msg("creating result")

	statsOpts := []stats.Opt{
		stats.WithMaxBlockSize(c.StatsMaxBlockSize),
		stats.WithTopKFreq(c.StatsTopK),
		stats.WithSymbolLen(c.StatsSymbolLen),
	}

	if c.StatsBlockSize > 0 {
		statsOpts = append(statsOpts, stats.WithBlockSize(c.StatsBlockSize))
	}

	log.Trace().
		Int("statsBlockSize", c.StatsBlockSize).
		Int("statsMaxBlockSize", c.StatsMaxBlockSize).
		Int("statsTopK", c.StatsTopK).
		Msg("analizing bits")

	bitsStats := stats.AnalizeBits(ctx, bits, statsOpts...)
	bitsStats.EntropyPlotName = c.EntropyPlotName

	result := &types.Result{
		Bits:  bits,
		Stats: bitsStats,
	}

	if c.OutCompressionType == compression.None {
		return result, nil
	}

	log.Trace().Msg("output requires compression")

	cr, err := iio.BitsToReader(ctx, bits, c.OutCompressionType)
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
		Stats:                stats.AnalizeBits(ctx, compressedBits),
	}

	return result, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
