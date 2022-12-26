package types

import "errors"

var ErrInvalidBit = errors.New("invalid bit")

type Bit uint8

type Stats struct {
	ZeroCount int
	OneCount  int

	SizeBits  int
	SizeBytes int

	ZeroStrings  map[int]int
	OneStrings   map[int]int
	MaxStringLen int
}

type Result struct {
	Bits  []Bit
	Stats *Stats
}
