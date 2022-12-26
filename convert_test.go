package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntBitsToByte(t *testing.T) {
	a := assert.New(t)
	for c := 0; c < 256; c++ {
		bits := byteToBits(byte(c))
		b := bitsToByte(bits)

		a.Equal(c, int(b))
	}
}
