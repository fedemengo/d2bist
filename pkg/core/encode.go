package core

import (
	"context"
	"io"

	"github.com/rs/zerolog"

	"github.com/fedemengo/d2bist/pkg/types"
)

// Encode receives a bit string in a io.Reader and creates a Result
func Encode(ctx context.Context, r io.Reader, opts ...Opt) (*types.Result, error) {
	log := zerolog.Ctx(ctx)
	bits, err := binStrReaderToBits(ctx, r, opts...)
	if err != nil {
		return nil, err
	}

	log.Trace().
		Int("bits", len(bits)).
		Msg("bits read from input binStr reader")

	return createResult(ctx, bits, opts...)
}
