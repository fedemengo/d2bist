package engine

import "github.com/fedemengo/f2bist/internal/types"

func BitsToByte(bits [8]types.Bit) byte {
	b := byte(0)

	for i := range bits {
		b += byte(bits[i] << (7 - i))
	}

	return b
}

func ByteToBits(b byte) [8]types.Bit {
	bits := [8]types.Bit{}

	for i := 0; i < 8; i++ {
		if b&(1<<(7-i)) > 0 {
			bits[i] = 1
		} else {
			bits[i] = 0
		}
	}

	return bits
}
