package engine

import (
	"fmt"

	"github.com/fedemengo/d2bist/internal/types"
)

// BitsToByte convers 8 bits to a byte, treating bits as follow
//
// 00010101 = 2^4 + 2^2 + 2+0 =
//
//	= bits[3] * 2^4 +
//	  bits[5] * 2^2 +
//	  bits[7] * 2^0   = byte(21)
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

func ByteToBit(b byte) (types.Bit, error) {
	switch b {
	case '0':
		return 0, nil
	case '1':
		return 1, nil
	default:
		return 0, fmt.Errorf("cannot handle `%c`: %w", b, types.ErrInvalidBit)
	}
}
