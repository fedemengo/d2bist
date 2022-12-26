package engine

import (
	"fmt"

	"github.com/fedemengo/f2bist/internal/types"
)

func AnalizeBits(bits []types.Bit) {
	zeroC, oneC := 0, 0
	for _, c := range bits {
		if c == 0 {
			zeroC++
		} else {
			oneC++
		}
	}
	fmt.Printf(`
size
    bits: %d
    B:    %d
0: %d
1: %d
`, len(bits), len(bits)/8, zeroC, oneC)
}
