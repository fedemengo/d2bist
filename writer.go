package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	isatty "github.com/mattn/go-isatty"
)

func bitToRune(b bit) rune {
	if b == 0 {
		return '0'
	}
	return '1'
}

type b2sConfig struct {
	separator rune
	distance  int
}

type Op func(c *b2sConfig)

func withSep(s rune) Op {
	return func(c *b2sConfig) {
		c.separator = s
	}
}

func withSepDistance(d int) Op {
	return func(c *b2sConfig) {
		c.distance = d
	}
}

func bitsToString(bits []bit, opts ...Op) string {
	c := &b2sConfig{
		distance: 8,
	}

	for _, op := range opts {
		op(c)
	}

	sb := strings.Builder{}
	for i, b := range bits {
		if c.separator != rune(0) && i > 0 && i%c.distance == 0 {
			sb.WriteRune(c.separator)
		}
		sb.WriteRune(bitToRune(b))
	}

	return sb.String()
}

func binaryStringToWriter(w io.Writer, bits []bit) {
	for i := 0; i < len(bits)/8; i++ {
		byteVal := [8]bit{}
		copy(byteVal[:], bits[i*8:min((i+1)*8, len(bits)-1)])
		w.Write([]byte{bitsToByte(byteVal)})
	}

}

func outputBinaryString(bits []bit) {
	if outputUTF8String {
		fmt.Println(bitsToString(bits))
	} else if isatty.IsTerminal(os.Stdout.Fd()) {
		fmt.Println(bitsToString(bits, withSep(' ')))
	} else {
		binaryStringToWriter(os.Stdout, bits)
	}

}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
