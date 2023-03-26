package engine

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/rs/zerolog"

	"github.com/fedemengo/d2bist/internal/types"
)

var ErrEndOfBits = errors.New("end of bits")

func ShannonEntropy(ctx context.Context, bits []types.Bit, chunkSize, symbolLen int) *types.Entropy {
	shannon := types.NewShannonEntropy()
	for i := 0; i < len(bits); i += chunkSize {
		nextBlockSize := min(chunkSize, len(bits)-i)
		chunk := bits[i : i+nextBlockSize]
		entropy := shannonEntropy(ctx, chunk, symbolLen)
		shannon.Values = append(shannon.Values, entropy)
	}

	return shannon
}

func shannonEntropy(ctx context.Context, chunk []types.Bit, symbolLen int) float64 {
	log := zerolog.Ctx(ctx).With().Logger()

	bw := NewBitsWindow(chunk, symbolLen)

	chunkLen := len(chunk)
	counts := make(map[uint64]int, 1<<uint(symbolLen))
	for i := 0; i < chunkLen; i += symbolLen {
		bitsInt := bw.ToInt()
		counts[bitsInt]++
		bw.SlideBy(symbolLen)
	}

	entropy := float64(0)

	symbolsCount := (float64(chunkLen) / float64(symbolLen))
	for intBitsValue, count := range counts {
		pX := float64(count) / symbolsCount
		e := float64(0)
		if pX > 0 {
			e = pX * math.Log2(pX)
		}
		log.Trace().
			Uint64("value", intBitsValue).
			Str("bitstr", fmt.Sprintf("%032b", intBitsValue)).
			Int("count", counts[intBitsValue]).
			Float64("pX", pX).
			Float64("entropy", e).
			Msg("symbol count")

		entropy += e
	}

	// normalize the entropy in the range [0, 1]
	entropy /= float64(symbolLen)

	entropy = -entropy

	log.Debug().Float64("entropy", entropy).Msg("entropy for chunk calculated")

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
