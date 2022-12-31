package io

import (
	"fmt"
	"io"
	"strings"

	"github.com/fedemengo/d2bist/internal/engine"
	"github.com/fedemengo/d2bist/internal/types"
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

func WriteBits(w io.Writer, bits []types.Bit) error {
	// note: we are safe handling bits grouped in bytes
	// as it's not possible to write anything less than 1 byte https://stackoverflow.com/a/6701236/4712324
	for i := 0; i < len(bits)/8; i++ {
		start := i * 8
		end := min((i+1)*8, len(bits))

		byteVal := [8]types.Bit{}
		copy(byteVal[:], bits[start:end])

		n, err := w.Write([]byte{engine.BitsToByte(byteVal)})
		if err != nil {
			return err
		}
		if n*8 != end-start {
			return fmt.Errorf("%d bytes written - %d expected", n, end-start+1)
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
