package io

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/fedemengo/f2bist/internal/engine"
	"github.com/fedemengo/f2bist/internal/types"
)

func BitsFromBinStrStdin() ([]types.Bit, error) {
	return BitsFromBinStrReader(os.Stdin)
}

func BitsFromBinStrStdinWithCap(maxBits int) ([]types.Bit, error) {
	return BitsFromBinStrReaderWithCap(os.Stdin, maxBits)
}

func BitsFromBinStrReaderWithCap(r io.Reader, maxBits int) ([]types.Bit, error) {
	opts := []opt{
		withMaxBits(maxBits),
		withTranform(func(b byte) ([]types.Bit, error) {
			switch b {
			case '0':
				return []types.Bit{0}, nil
			case '1':
				return []types.Bit{1}, nil
			default:
				return []types.Bit{}, types.ErrInvalidBit
			}
		}),
	}

	return BitsFromReader(r, opts...)

}

func BitsFromBinStrReader(r io.Reader) ([]types.Bit, error) {
	return BitsFromReader(r, withTranform(func(b byte) ([]types.Bit, error) {
		switch b {
		case '0':
			return []types.Bit{0}, nil
		case '1':
			return []types.Bit{1}, nil
		default:
			return []types.Bit{}, types.ErrInvalidBit
		}
	}))
}

func BitsFromByteStdin() ([]types.Bit, error) {
	return BitsFromByteReader(os.Stdin)
}

func BitsFromByteFile(filename string) func() ([]types.Bit, error) {
	return func() ([]types.Bit, error) {
		f, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		return BitsFromByteReader(f)
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

func withTranform(t tranform) opt {
	return func(c *config) {
		c.transform = t
	}
}

func BitsFromReader(r io.Reader, opts ...opt) ([]types.Bit, error) {
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
			return nil, err
		}

		for _, b := range bytes[:n] {
			bitsArray, err := c.transform(b)
			//fmt.Printf("%v `%c`\n", bitsArray, b)
			if errors.Is(err, types.ErrInvalidBit) {
				continue
			} else if err != nil {
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
				//
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

func BitsFromByteReaderWithCap(r io.Reader, maxBits int) ([]types.Bit, error) {
	opts := []opt{
		withMaxBits(maxBits),
		withTranform(func(b byte) ([]types.Bit, error) {
			bits := engine.ByteToBits(b)
			return bits[:], nil
		}),
	}
	return BitsFromReader(r, opts...)
}

func BitsFromByteReader(r io.Reader) ([]types.Bit, error) {
	return BitsFromByteReaderWithCap(r, -1)
}
