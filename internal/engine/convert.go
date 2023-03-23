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

func BitsToInt(bits []types.Bit) (int64, error) {
	if len(bits) > 64 {
		return 0, fmt.Errorf("cannot convert %d bits to int64", len(bits))
	}
	v := int64(0)
	for i, b := range bits {
		v += int64(b) << uint(len(bits)-1-i)
	}

	return v, nil
}

func IntToBits(n int64, bitsCount int) ([]types.Bit, error) {
	if n >= 1<<bitsCount {
		return nil, fmt.Errorf("number %d cannot be represented in %d bits", n, bitsCount)
	}

	bits := make([]types.Bit, bitsCount)
	i := bitsCount - 1
	for n > 0 {
		bits[i] = types.Bit(n % 2)
		i--
		n /= 2
	}

	return bits, nil
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

func BitsToString(bits []types.Bit) string {
	s := make([]byte, len(bits))
	for i, b := range bits {
		s[i] = b.ToByte()
	}

	return string(s)
}
