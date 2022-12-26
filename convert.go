package main

func bitsToByte(bits [8]bit) byte {
	b := byte(0)

	for i := range bits {
		b += byte(bits[i] << (7 - i))
	}

	return b
}

func byteToBits(b byte) [8]bit {
	bits := [8]bit{}
	for i := 0; i < 8; i++ {
		if b&(1<<(7-i)) > 0 {
			bits[i] = 1
		} else {
			bits[i] = 0
		}
	}

	return bits
}
