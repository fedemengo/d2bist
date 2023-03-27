package io

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/rs/zerolog"

	"github.com/fedemengo/d2bist/pkg/engine"
	"github.com/fedemengo/d2bist/pkg/types"
)

func bitToRune(b types.Bit) rune {
	if b == 0 {
		return '0'
	}
	return '1'
}

type b2sConfig struct {
	separator rune
	distance  int
}

type Opt func(c *b2sConfig)

func WithSep(s rune) Opt {
	return func(c *b2sConfig) {
		c.separator = s
	}
}

func WithSepDistance(d int) Opt {
	return func(c *b2sConfig) {
		c.distance = d
	}
}

func BitsToString(bits []types.Bit, opts ...Opt) string {
	c := &b2sConfig{
		distance: 8,
	}

	for _, op := range opts {
		op(c)
	}

	sb := strings.Builder{}
	for i, b := range bits {
		if c.separator != rune(0) && i > 0 && i%c.distance == 0 {
			sb.WriteRune(c.separator)
		}
		sb.WriteRune(bitToRune(b))
	}

	return sb.String()
}

func BitsToByteWriter(ctx context.Context, w io.Writer, bits []types.Bit) error {
	log := zerolog.Ctx(ctx)
	// note: we are safe handling bits grouped in bytes
	// as it's not possible to write anything less than 1 byte https://stackoverflow.com/a/6701236/4712324
	bytesCount := len(bits) / 8
	if bytesCount*8 < len(bits) {
		bytesCount++
	}

	for i := 0; i < bytesCount; i++ {
		start := i * 8
		end := min((i+1)*8, len(bits))

		byteVal := [8]types.Bit{}
		bitsStr := [8]byte{}
		for j := start; j < start+8; j++ {
			if j < end {
				byteVal[j-start] = bits[j]
				bitsStr[j-start] = bits[j].ToByte()
			} else {
				bitsStr[j-start] = '-'
			}
		}
		log.Trace().Str("bits", string(bitsStr[:])).Msg("bits to byte conversion")

		n, err := w.Write([]byte{engine.BitsToByte(byteVal)})
		if err != nil {
			return err
		}

		if n*8 < end-start {
			return fmt.Errorf("%d bits written - %d expected", n*8, end-start)
		}
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
