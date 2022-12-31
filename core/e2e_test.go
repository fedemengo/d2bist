package core

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fedemengo/d2bist/compression"
	fio "github.com/fedemengo/d2bist/internal/io"
	"github.com/fedemengo/d2bist/internal/types"
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

func compressData(data []byte, cType compression.CompressionType) []byte {
	buf := new(bytes.Buffer)
	cw, err := compression.NewCompressedWriter(context.Background(), buf, cType)
	if err != nil {
		return nil
	}

	cw.Write(data)

	if err := cw.Close(); err != nil {
		return nil
	}

	return buf.Bytes()
}

func traceLogger() zerolog.Logger {
	level := zerolog.Disabled
	logWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	}
	return zerolog.New(logWriter).With().
		Timestamp().
		Caller().
		Logger().
		Level(level)
}

func TestE2E(t *testing.T) {
	log := traceLogger()
	ctx := log.WithContext(context.Background())

	testCases := []struct {
		name         string
		data         []byte
		ops          []op
		converters   []converter
		opts         [][]Opt
		expectedData []byte
	}{
		{
			name: "decode/encode",
			data: []byte("dead beef"),
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
			expectedData: []byte("dead beef"),
		}, {
			name: "decode and compress/encode compressed",
			data: []byte("dead beef"),
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
			expectedData: []byte("dead beef"),
		}, {
			name: "decode and cap/encode",
			data: []byte("dead beef"),
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
			expectedData: []byte("dead"),
		}, {
			name: "decode and compress/encode compressed and cap",
			data: []byte("dead beef"),
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
			expectedData: []byte("de"),
		}, {
			name: "decode compressed/encode compressed and compress 2",
			data: compressData([]byte("a longer text so that compression actually does something nice"), compression.Brotli),
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
				{WithInCompression(compression.Brotli), WithOutCompression(compression.Zstd), WithOutBitsCap(16 * 8)},
			},
			expectedData: []byte{0x28, 0xb5, 0x2f, 0xfd, 0x4, 0x0, 0x81, 0x0, 0x0, 0x61, 0x20, 0x6c, 0x6f, 0x6e, 0x67, 0x65},
		}, {
			name: "decode compressed/encode and compress",
			data: compressData([]byte("a longer text so that compression actually does something nice"), compression.Brotli),
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
				{WithOutCompression(compression.Zstd), WithOutBitsCap(16 * 8)},
			},
			expectedData: []byte{0x28, 0xb5, 0x2f, 0xfd, 0x4, 0x0, 0x81, 0x0, 0x0, 0x1b, 0x3d, 0x0, 0xe8, 0xa5, 0xc3, 0xc8},
		}, {
			name: "decode (compressed 1 + 2) and decompress 1/encode (compressed 2) and decompress",
			data: compressData(
				compressData(
					[]byte("a longer text so that compression actually does something nice"),
					compression.S2,
				),
				compression.Brotli,
			),
			ops: []op{
				Decode,
				Encode,
			},
			converters: []converter{
				resultToBinStr,
				basicConverter,
			},
			opts: [][]Opt{
				{WithInCompression(compression.Brotli)},
				{WithInCompression(compression.S2)},
			},
			expectedData: []byte("a longer text so that compression actually does something nice"),
		}, {
			name: "decode (compressed 1 + 2 + 3) and decompress 1/encode (compressed 2) and decompress 2",
			data: compressData(
				compressData(
					compressData(
						[]byte("a longer text so that compression actually does something nice"),
						compression.Gzip,
					),
					compression.Brotli,
				),
				compression.Zstd,
			),
			ops: []op{
				Decode,
				Encode,
			},
			converters: []converter{
				resultToBinStr,
				basicConverter,
			},
			opts: [][]Opt{
				{WithInCompression(compression.Zstd)},
				{WithInCompression(compression.Brotli)},
			},
			expectedData: compressData([]byte("a longer text so that compression actually does something nice"), compression.Gzip),
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
				log.Info().Msgf("step %d", i)
				res, err := tc.ops[i](ctx, reader, tc.opts[i]...)
				r.NoError(err)

				reader, err = tc.converters[i](res)
				r.NoError(err)
			}

			data, err := io.ReadAll(reader)
			r.NoError(err)

			a.Equal(tc.expectedData, data)
		})
	}

}
