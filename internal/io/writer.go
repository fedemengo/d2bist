package io

import (
	"io"
	"strings"

	"github.com/fedemengo/f2bist/internal/engine"
	"github.com/fedemengo/f2bist/internal/types"
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

type Op func(c *b2sConfig)

func WithSep(s rune) Op {
	return func(c *b2sConfig) {
		c.separator = s
	}
}

func WithSepDistance(d int) Op {
	return func(c *b2sConfig) {
		c.distance = d
	}
}

func BitsToString(bits []types.Bit, opts ...Op) string {
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

func BitsToWriter(w io.Writer, bits []types.Bit) {
	for i := 0; i < len(bits)/8; i++ {
		byteVal := [8]types.Bit{}
		copy(byteVal[:], bits[i*8:min((i+1)*8, len(bits)-1)])
		w.Write([]byte{engine.BitsToByte(byteVal)})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
