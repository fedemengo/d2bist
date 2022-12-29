package io

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fedemengo/f2bist/internal/types"
)

func TestWriteBitStringToWriter(t *testing.T) {
	testCases := []struct {
		name           string
		bits           []types.Bit
		expectedString string
	}{
		{
			name:           "two characters",
			bits:           []types.Bit{0, 1, 1, 0, 1, 0, 0, 0, 0, 1, 1, 0, 0, 1, 0, 1},
			expectedString: "he",
		}, {
			name:           "two characters and some",
			bits:           []types.Bit{0, 1, 1, 0, 1, 0, 0, 0, 0, 1, 1, 0, 0, 1, 0, 1, 1, 0, 1},
			expectedString: "he",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			a, r := assert.New(tt), require.New(tt)

			s := ""
			buf := bytes.NewBufferString(s)
			err := WriteBits(buf, tc.bits)
			r.NoError(err)

			a.Equal(tc.expectedString, buf.String())
		})
	}

}
