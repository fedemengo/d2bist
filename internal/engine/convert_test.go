package engine

import (
	"testing"

	"github.com/fedemengo/f2bist/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestIntBitsToByte(t *testing.T) {
	a := assert.New(t)
	for c := 0; c < 256; c++ {
		bits := ByteToBits(byte(c))
		b := BitsToByte(bits)

		a.Equal(c, int(b))
	}

	b := BitsToByte([8]types.Bit{0, 1, 1, 0, 0, 1, 0, 1})
	a.Equal("e", string(b))
}
