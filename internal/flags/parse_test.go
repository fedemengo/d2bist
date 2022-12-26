package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataCapParsing(t *testing.T) {
	testCases := []struct {
		name            string
		dataCap         string
		expectedMaxBits int
		expectedToFail  bool
	}{
		{
			name:            "bits",
			dataCap:         "1234",
			expectedMaxBits: 1234,
		}, {
			name:            "bytes",
			dataCap:         "12B",
			expectedMaxBits: 12 * 8,
		}, {
			name:            "kilo byte",
			dataCap:         "13kB",
			expectedMaxBits: 13 * 8 * 1000,
		}, {
			name:            "kibi byte",
			dataCap:         "14K",
			expectedMaxBits: 14 * 8 * 1024,
		}, {
			name:           "bad format",
			dataCap:        "B",
			expectedToFail: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			a, r := assert.New(tt), require.New(tt)

			count, err := ParseDataCapToBitsCount(tc.dataCap)
			if tc.expectedToFail {
				r.Error(err)
				return
			}

			r.NoError(err)
			a.Equal(tc.expectedMaxBits, count)
		})
	}

}
