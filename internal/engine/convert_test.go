package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntBitsToByte(t *testing.T) {
	a := assert.New(t)
	for c := 0; c < 256; c++ {
		bits := ByteToBits(byte(c))
		b := BitsToByte(bits)

		a.Equal(c, int(b))
	}
}
