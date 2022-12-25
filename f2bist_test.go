package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntBitsToByte(t *testing.T) {
	a := assert.New(t)
	for c := 0; c < 256; c++ {
		bits := byteToBits(byte(c))
		b := bitsToByte(bits)

		a.Equal(c, int(b))
	}
}

func TestBytesToBit(t *testing.T) {
	testCases := []struct {
		name         string
		data         string
		expectedBits []bit
	}{
		{
			name:         "dead beef",
			data:         "dead beef",
			expectedBits: []bit{0, 1, 1, 0, 0, 1, 0, 0, 0, 1, 1, 0, 0, 1, 0, 1, 0, 1, 1, 0, 0, 0, 0, 1, 0, 1, 1, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 0, 1, 1, 0, 0, 1, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0},
		}, {
			name: "multi line",
			data: `multi line
text`,
			expectedBits: []bit{0, 1, 1, 0, 1, 1, 0, 1, 0, 1, 1, 1, 0, 1, 0, 1, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1, 1, 1, 0, 1, 0, 0, 0, 1, 1, 0, 1, 0, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1, 1, 0, 1, 0, 0, 1, 0, 1, 1, 0, 1, 1, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 0, 0, 0, 0, 1, 0, 1, 0, 0, 1, 1, 1, 0, 1, 0, 0, 0, 1, 1, 0, 0, 1, 0, 1, 0, 1, 1, 1, 1, 0, 0, 0, 0, 1, 1, 1, 0, 1, 0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			a, r := assert.New(tt), require.New(tt)

			reader := bytes.NewReader([]byte(tc.data))

			bits, err := bitsFromReader(reader)
			r.NoError(err)

			fmt.Println(len(tc.expectedBits))
			a.Equal(tc.expectedBits, bits)
		})
	}

}

func TestEncodingToBinStr(t *testing.T) {
	testCases := []struct {
		name           string
		data           string
		expectedString string
	}{
		{
			name:           "empty",
			data:           "",
			expectedString: "",
		}, {
			name:           "dead beef",
			data:           "dead beef",
			expectedString: "01100100 01100101 01100001 01100100 00100000 01100010 01100101 01100101 01100110",
		}, {
			name: "multi line",
			data: `multi line
text`,
			expectedString: "01101101 01110101 01101100 01110100 01101001 00100000 01101100 01101001 01101110 01100101 00001010 01110100 01100101 01111000 01110100",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			a, r := assert.New(tt), require.New(tt)

			reader := bytes.NewReader([]byte(tc.data))

			bits, err := bitsFromReader(reader)
			r.NoError(err)

			s := ""
			buf := bytes.NewBufferString(s)
			binaryStringToWriter(buf, bits...)

			bstr := buf.String()

			a.Equal(tc.data, bstr)

			hstr := humanReadableBinStr(bits...)
			hstr = strings.ReplaceAll(hstr, ".", "")

			expectedBStr := strings.ReplaceAll(tc.expectedString, " ", "")

			fmt.Println(len(expectedBStr))
			a.Equal(expectedBStr, hstr)
		})
	}
}
