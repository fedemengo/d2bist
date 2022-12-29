package engine

import "github.com/fedemengo/f2bist/internal/types"

func BitsToByte(bits [8]types.Bit) byte {
	b := byte(0)

	b += byte(bits[0]) << 7
	b += byte(bits[1]) << 6
	b += byte(bits[2]) << 5
	b += byte(bits[3]) << 4
	b += byte(bits[4]) << 3
	b += byte(bits[5]) << 2
	b += byte(bits[6]) << 1
	b += byte(bits[7]) << 0

	return b
}

func ByteToBits(b byte) [8]types.Bit {
	return [8]types.Bit{
		types.Bit((b & (1 << 7)) >> 7),
		types.Bit((b & (1 << 6)) >> 6),
		types.Bit((b & (1 << 5)) >> 5),
		types.Bit((b & (1 << 4)) >> 4),
		types.Bit((b & (1 << 3)) >> 3),
		types.Bit((b & (1 << 2)) >> 2),
		types.Bit((b & (1 << 1)) >> 1),
		types.Bit((b & (1 << 0)) >> 0),
	}
}
