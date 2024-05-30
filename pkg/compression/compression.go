package compression

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/andybalholm/brotli"
	"github.com/dsnet/compress/bzip2"
	"github.com/klauspost/compress/flate"
	"github.com/klauspost/compress/s2"
	"github.com/klauspost/compress/zstd"
	"github.com/rs/zerolog"
)

var ErrAlgorithmNotImplemented = errors.New("algorithm not implemented")

type CompressionType string

const (
	None   = CompressionType("None")
	Zip    = CompressionType("Zip")
	Gzip   = CompressionType("Gzip")
	Brotli = CompressionType("Brotli")
	Zstd   = CompressionType("Zstd")
	S2     = CompressionType("S2")
	Huff   = CompressionType("Huff")
	Bzip2  = CompressionType("Bzip2")
)

func NewCompressedReader(ctx context.Context, r io.Reader, cType CompressionType) (io.Reader, error) {
	log := zerolog.Ctx(ctx)
	switch cType {
	case None:
		log.Trace().Msg("no compression")
		return r, nil
	case Zip:
		log.Error().Msg("zip compression")
		return flate.NewReader(r), nil
	case Gzip:
		log.Trace().Msg("gzip compression")
		return gzip.NewReader(r)
	case Brotli:
		log.Trace().Msg("brotli compression")
		return brotli.NewReader(r), nil
	case Zstd:
		log.Trace().Msg("zstd compression")
		return zstd.NewReader(r)
	case S2:
		log.Trace().Msg("s2 compression")
		return s2.NewReader(r), nil
	case Huff:
		log.Trace().Msg("huffman compression")
		return nil, fmt.Errorf("huffman not implemented: %w", ErrAlgorithmNotImplemented)
	case Bzip2:
		log.Trace().Msg("bzip2 compression")
		return bzip2.NewReader(r, nil)
	default:
		return nil, fmt.Errorf("compression type  %T not supported", cType)
	}
}

type nopWriterCloser struct {
	w io.Writer
}

func NewNopWriterCloser(w io.Writer) io.WriteCloser {
	return nopWriterCloser{w: w}
}

func (w nopWriterCloser) Write(b []byte) (int, error) {
	return w.w.Write(b)
}

func (w nopWriterCloser) Close() error {
	return nil
}

func NewCompressedWriter(ctx context.Context, w io.Writer, cType CompressionType) (io.WriteCloser, error) {
	log := zerolog.Ctx(ctx)
	switch cType {
	case None:
		log.Trace().Msg("no compression")
		return NewNopWriterCloser(w), nil
	case Zip:
		log.Error().Msg("zip compression")
		return flate.NewWriter(w, flate.BestCompression)
	case Gzip:
		log.Trace().Msg("gzip compression")
		return gzip.NewWriterLevel(w, flate.BestCompression)
	case Brotli:
		log.Trace().Msg("brotli compression")
		return brotli.NewWriterLevel(w, brotli.BestCompression), nil
	case Zstd:
		log.Trace().Msg("zstd compression")
		return zstd.NewWriter(w)
	case S2:
		log.Trace().Msg("s2 compression")
		return s2.NewWriter(w, s2.WriterBestCompression()), nil
	case Huff:
		log.Trace().Msg("huffman compression")
		return nil, fmt.Errorf("huffman not implemented: %w", ErrAlgorithmNotImplemented)
	case Bzip2:
		log.Trace().Msg("bzip2 compression")
		return bzip2.NewWriter(w, &bzip2.WriterConfig{Level: bzip2.BestCompression})
	default:
		return nil, fmt.Errorf("compression type  %T not supported", cType)
	}
}
