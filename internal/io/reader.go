package io

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"

	"github.com/fedemengo/f2bist/internal/engine"
	"github.com/fedemengo/f2bist/internal/types"
)

func BitsFromBinStrReaderWithCap(ctx context.Context, r io.Reader, maxBits int) ([]types.Bit, error) {
	opts := []opt{
		withMaxBits(maxBits),
		withTransform(func(b byte) ([]types.Bit, error) {
			switch b {
			case '0':
				return []types.Bit{0}, nil
			case '1':
				return []types.Bit{1}, nil
			default:
				return []types.Bit{}, fmt.Errorf("cannot handle `%c`: %w", b, types.ErrInvalidBit)
			}
		}),
	}

	return BitsFromReader(ctx, r, opts...)

}

func BitsFromBinStrReader(ctx context.Context, r io.Reader) ([]types.Bit, error) {
	return BitsFromBinStrReaderWithCap(ctx, r, -1)
}

func BitsFromByteStdin(ctx context.Context) ([]types.Bit, error) {
	return BitsFromByteReader(ctx, os.Stdin)
}

func BitsFromByteFile(ctx context.Context, filename string) func() ([]types.Bit, error) {
	return func() ([]types.Bit, error) {
		f, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		return BitsFromByteReader(ctx, f)
	}
}

type tranform func(byte) ([]types.Bit, error)

type config struct {
	maxBits   int
	transform tranform
}

type opt func(c *config)

func withMaxBits(n int) opt {
	return func(c *config) {
		c.maxBits = n
	}
}

func withTransform(t tranform) opt {
	return func(c *config) {
		c.transform = t
	}
}

func BitsFromReader(ctx context.Context, r io.Reader, opts ...opt) ([]types.Bit, error) {
	log := zerolog.Ctx(ctx)

	c := &config{
		maxBits: -1,
	}

	for _, opt := range opts {
		opt(c)
	}

	bits := []types.Bit{}

	bytes := make([]byte, 8)
	lastCount := -1
	for {
		n, err := r.Read(bytes)
		if errors.Is(err, io.EOF) {
			lastCount = n
			break
		}
		if err != nil {
			log.Error().Err(err).Msg("unknown error")
			return nil, fmt.Errorf("unknown error: %w", err)
		}

		for _, b := range bytes[:n] {
			bitsArray, err := c.transform(b)
			//fmt.Printf("%v `%c`\n", bitsArray, b)
			if errors.Is(err, types.ErrInvalidBit) {
				log.Error().Err(err).Msgf("invalid byte to bits: %v", b)
				continue
			} else if err != nil {
				log.Error().Err(err).Msgf("cannot convert byte to bits: %v", b)
				return nil, fmt.Errorf("error when parsing remaining data: %w", err)
			} else {
				bits = append(bits, bitsArray...)
			}
		}

		if c.maxBits > 0 && len(bits) >= c.maxBits {
			lastCount = 0
			break
		}
	}

	if lastCount > 0 {
		for _, b := range bytes[:lastCount] {
			bitsArray, err := c.transform(b)
			if errors.Is(err, types.ErrInvalidBit) {
				log.Error().Err(err).Msgf("invalid byte to bits: %v", b)
			} else if err != nil {
				return nil, fmt.Errorf("error when parsing remaining data: %w", err)
			} else {
				bits = append(bits, bitsArray...)
			}
		}
	}

	if c.maxBits > 0 && c.maxBits < len(bits) {
		bits = bits[:c.maxBits]
	}

	return bits, nil

}

func BitsFromByteReaderWithCap(ctx context.Context, r io.Reader, maxBits int) ([]types.Bit, error) {
	opts := []opt{
		withMaxBits(maxBits),
		withTransform(func(b byte) ([]types.Bit, error) {
			bits := engine.ByteToBits(b)
			return bits[:], nil
		}),
	}
	return BitsFromReader(ctx, r, opts...)
}

func BitsFromByteReader(ctx context.Context, r io.Reader) ([]types.Bit, error) {
	return BitsFromByteReaderWithCap(ctx, r, -1)
}
