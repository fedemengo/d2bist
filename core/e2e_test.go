package core

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fedemengo/f2bist/compression"
	fio "github.com/fedemengo/f2bist/internal/io"
	"github.com/fedemengo/f2bist/internal/types"
)

type op func(context.Context, io.Reader, ...Opt) (*types.Result, error)

type converter func(*types.Result) (io.Reader, error)

func basicConverter(res *types.Result) (io.Reader, error) {
	r, err := bitsToReader(context.Background(), res.Bits, compression.None)
	return r, err
}

func resultToBinStr(res *types.Result) (io.Reader, error) {
	s := fio.BitsToString(res.Bits)
	return bytes.NewReader([]byte(s)), nil
}

func TestE2E(t *testing.T) {

	ctx := context.Background()

	testCases := []struct {
		name         string
		data         string
		ops          []op
		converters   []converter
		opts         [][]Opt
		expectedData string
	}{
		{
			name: "decode/encode",
			data: "dead beef",
			ops: []op{
				Decode,
				Encode,
			},
			converters: []converter{
				resultToBinStr,
				basicConverter,
			},
			opts: [][]Opt{
				{},
				{},
			},
			expectedData: "dead beef",
		}, {
			name: "decode compress/encode decompress",
			data: "dead beef",
			ops: []op{
				Decode,
				Encode,
			},
			converters: []converter{
				resultToBinStr,
				basicConverter,
			},
			opts: [][]Opt{
				{WithOutCompression(compression.Brotli)},
				{WithInCompression(compression.Brotli)},
			},
			expectedData: "dead beef",
		}, {
			name: "decode and cap/encode",
			data: "dead beef",
			ops: []op{
				Decode,
				Encode,
			},
			converters: []converter{
				resultToBinStr,
				basicConverter,
			},
			opts: [][]Opt{
				{WithOutBitsCap(32)},
				{},
			},
			expectedData: "dead",
		}, {
			name: "decode compress/encode decompress and cap",
			data: "dead beef",
			ops: []op{
				Decode,
				Encode,
			},
			converters: []converter{
				resultToBinStr,
				basicConverter,
			},
			opts: [][]Opt{
				{WithOutCompression(compression.Brotli)},
				{WithInCompression(compression.Brotli), WithOutBitsCap(16)},
			},
			expectedData: "de",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			a, r := assert.New(tt), require.New(tt)

			opsLen := len(tc.ops)
			r.Len(tc.converters, opsLen)
			r.Len(tc.opts, opsLen)

			var reader io.Reader
			reader = bytes.NewReader([]byte(tc.data))
			for i := 0; i < opsLen; i++ {
				res, err := tc.ops[i](ctx, reader, tc.opts[i]...)
				r.NoError(err)

				reader, err = tc.converters[i](res)
				r.NoError(err)
			}

			data, err := io.ReadAll(reader)
			r.NoError(err)

			a.Equal(tc.expectedData, string(data))
		})
	}

}
