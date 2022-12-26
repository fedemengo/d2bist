package main

import (
	"errors"
	"io"
)

func bitsFromReader(r io.Reader) ([]bit, error) {
	bits := []bit{}

	bytes := make([]byte, 8)
	lastCount := -1
	for {
		n, err := r.Read(bytes)
		if errors.Is(err, io.EOF) {
			lastCount = n
			break
		}
		if err != nil {
			return nil, err
		}

		for _, b := range bytes[:n] {
			bitsArray := byteToBits(b)
			//fmt.Printf("%v `%c`\n", bitsArray, b)
			bits = append(bits, bitsArray[0:8]...)
		}
	}

	if lastCount > 0 {
		for _, b := range bytes[:lastCount] {
			bitsArray := byteToBits(b)
			bits = append(bits, bitsArray[0:8]...)
		}
	}

	return bits, nil

}
