package io

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fedemengo/f2bist/internal/types"
)

func TestBytesToBit(t *testing.T) {
	testCases := []struct {
		name         string
		data         string
		expectedBits []types.Bit
	}{
		{
			name:         "dead beef",
			data:         "dead beef",
			expectedBits: []types.Bit{0, 1, 1, 0, 0, 1, 0, 0, 0, 1, 1, 0, 0, 1, 0, 1, 0, 1, 1, 0, 0, 0, 0, 1, 0, 1, 1, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 0, 1, 1, 0, 0, 1, 0, 1, 0, 1, 1, 0, 0, 1, 1, 0},
		}, {
			name: "multi line",
			data: `multi line
text`,
			expectedBits: []types.Bit{0, 1, 1, 0, 1, 1, 0, 1, 0, 1, 1, 1, 0, 1, 0, 1, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1, 1, 1, 0, 1, 0, 0, 0, 1, 1, 0, 1, 0, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1, 1, 0, 1, 0, 0, 1, 0, 1, 1, 0, 1, 1, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 0, 0, 0, 0, 1, 0, 1, 0, 0, 1, 1, 1, 0, 1, 0, 0, 0, 1, 1, 0, 0, 1, 0, 1, 0, 1, 1, 1, 1, 0, 0, 0, 0, 1, 1, 1, 0, 1, 0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			a, r := assert.New(tt), require.New(tt)

			reader := bytes.NewReader([]byte(tc.data))

			bits, err := BitsFromByteReader(reader)
			r.NoError(err)

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
			bits, err := BitsFromByteReader(reader)
			r.NoError(err)

			s := ""
			buf := bytes.NewBufferString(s)
			BitsToWriter(buf, bits)
			bstr := buf.String()

			a.Equal(tc.data, bstr)

			hstr := BitsToString(bits)
			expectedBStr := strings.ReplaceAll(tc.expectedString, " ", "")

			a.Equal(expectedBStr, hstr)
		})
	}
}

func TestBitStringToBit(t *testing.T) {
	testCases := []struct {
		name         string
		data         string
		expectedBits []types.Bit
	}{
		{
			name:         "two characters",
			data:         "0110100001100101",
			expectedBits: []types.Bit{0, 1, 1, 0, 1, 0, 0, 0, 0, 1, 1, 0, 0, 1, 0, 1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			a, r := assert.New(tt), require.New(tt)

			reader := bytes.NewReader([]byte(tc.data))

			bits, err := BitsFromBinStrReader(reader)
			r.NoError(err)

			a.Equal(tc.expectedBits, bits)
		})
	}

}
