package engine

import (
	"fmt"
	"math"

	"github.com/fedemengo/d2bist/pkg/types"
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

func BitsToInt(bits []types.Bit) (uint64, error) {
	if len(bits) > 64 {
		return 0, fmt.Errorf("cannot convert %d bits to int64", len(bits))
	}
	v := uint64(0)
	for i, b := range bits {
		v += uint64(b) << uint(len(bits)-1-i)
	}

	return v, nil
}

func IntToBits(n uint64, bitsCount int) ([]types.Bit, error) {
	if n > 0 && int(math.Log2(float64(n))) > bitsCount {
		return nil, fmt.Errorf("number %d cannot be represented in %d bits", n, bitsCount)
	}

	bits := make([]types.Bit, bitsCount)
	for i := bitsCount - 1; n > 0; i-- {
		bits[i] = types.Bit(n % 2)
		n >>= 1
	}

	return bits, nil
}

func IntToBitString(n uint64, bitsCount int) (string, error) {
	bits, err := IntToBits(n, bitsCount)
	if err != nil {
		return "", err
	}

	return BitsToStringN(bits, bitsCount), nil
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

func BitsToStringN(bits []types.Bit, n int) string {
	if n < len(bits) {
		n = len(bits)
	}

	s := make([]byte, n)
	start := n - len(bits)
	for i, b := range bits {
		s[start+i] = b.ToByte()
	}

	return string(s)
}

func BitsToString(bits []types.Bit) string {
	s := make([]byte, len(bits))
	for i, b := range bits {
		s[i] = b.ToByte()
	}

	return string(s)
}

type BitsWindow interface {
	Slide() error
	SlideBy(n int) error
	ToInt() uint64
	ToString() string

	ToIntSlide() (uint64, error)
}

var _ BitsWindow = (*bitsWindow)(nil)

func NewBitsWindow(bits []types.Bit, windowSize int) BitsWindow {
	return &bitsWindow{
		start:      0,
		end:        windowSize,
		bits:       bits,
		windowSize: windowSize,
		lastStart:  -1,
		lastEnd:    -1,
		lastStr:    "",
	}
}

type bitsWindow struct {
	start, end int
	bits       []types.Bit
	windowSize int

	lastStart, lastEnd int
	lastInt            uint64
	lastStr            string
}

func (b *bitsWindow) Slide() error {
	return b.SlideBy(1)
}

func (b *bitsWindow) SlideBy(n int) error {
	if b.end+n > len(b.bits) {
		return ErrEndOfBits
	}

	b.start += n
	b.end += n

	return nil
}

func (b *bitsWindow) ToInt() uint64 {
	// if it's the first time, calculate in full
	if b.lastStart == -1 || b.lastEnd == -1 {
		return b.fullToInt()
	}

	// if we are completely outside the last window, calculate in full
	if b.start >= b.lastEnd {
		return b.fullToInt()
	}

	exit := b.start - b.lastStart
	enter := b.end - b.lastEnd

	return b.slideToInt(exit, enter)
}

func (b *bitsWindow) fullToInt() uint64 {
	acc := uint64(0)
	for i := b.start; i < b.end; i++ {
		acc <<= 1
		acc |= uint64(b.bits[i])
	}

	b.lastStart = b.start
	b.lastEnd = b.end
	b.lastInt = acc

	return acc
}

// slideToInt returns the int representation of the bits window
//
// given :  0010010100100011001
// diff  :   ---        ~~~~~~
// before:   ^          ^       	(lastStart, lastEnd]
// after :      ^             ^ 	(start, end]
// so 3 bits are exiting: 010
// and 6 bits are entering: 001100
func (b *bitsWindow) slideToInt(exit, enter int) uint64 {
	acc := b.lastInt

	mask := uint64(1<<(b.windowSize-exit) - 1)
	acc &= mask

	for i := b.lastEnd; i < b.end; i++ {
		acc <<= 1
		acc |= uint64(b.bits[i])
	}

	b.lastStart = b.start
	b.lastEnd = b.end
	b.lastInt = acc

	return acc
}

func (b *bitsWindow) ToString() string {
	s := b.lastStr
	exit := b.start - b.lastStart
	enter := b.end - b.lastEnd

	s = s[exit:]

	bytes := make([]byte, enter)
	for i := range bytes {
		bytes[i] = b.bits[b.lastEnd+i].ToByte()
	}

	newStr := s + string(bytes)

	b.lastStart = b.start
	b.lastEnd = b.end
	b.lastStr = newStr

	return newStr
}

func (b *bitsWindow) ToIntSlide() (uint64, error) {
	n := b.ToInt()
	err := b.Slide()

	return n, err
}
