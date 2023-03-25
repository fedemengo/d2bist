package engine

import (
	"errors"
	"math"

	"github.com/fedemengo/d2bist/internal/types"
)

var ErrEndOfBits = errors.New("end of bits")

func ShannonEntropy(chunk []types.Bit, symbolLen int) float64 {
	bw := NewBitsWindow(chunk, symbolLen)

	chunkLen := len(chunk)
	counts := make(map[uint64]int, 1<<uint(symbolLen))
	for i := 0; i < chunkLen; i += symbolLen {
		bitsInt := bw.ToInt()
		bw.SlideBy(symbolLen)
		counts[bitsInt]++
	}

	// opted for a more numerically stable version of the formula
	//
	// symbolsInChunk := chunkLen / symbolLen       // in 128 bits there are 32 symbols of 4 bits each
	// possibleSymbols := 1 << uint(symbolsInChunk) // 2^32 - all possible symbols
	entropy := float64(0)
	for intBitsValue := range counts {
		//pX := float64(counts[intBitsValue]) / float64(possibleSymbols)
		pX := float64(counts[intBitsValue]) / (float64(chunkLen) / float64(symbolLen))
		e := float64(0)
		if pX > 0 {
			e = pX * math.Log2(pX)
		}
		entropy -= e
	}

	entropy /= float64(symbolLen)

	return entropy
}

type BitsWindow interface {
	Slide() error
	SlideBy(n int) error
	ToInt() uint64
	ToString() string
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

	if enter == 1 {
		acc <<= 1
		acc |= uint64(b.bits[b.end-1])
	} else {
		for i := b.lastEnd; i < b.end; i++ {
			acc <<= 1
			acc |= uint64(b.bits[i])
		}
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
