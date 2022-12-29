package core

import (
	"context"
	"io"

	"github.com/rs/zerolog"

	"github.com/fedemengo/f2bist/internal/types"
)

// Decode receives byte data in a io.Reader and creates a Result
func Decode(ctx context.Context, r io.Reader, opts ...Opt) (*types.Result, error) {
	log := zerolog.Ctx(ctx)
	bits, err := readerToBits(ctx, r, opts...)
	if err != nil {
		return nil, err
	}

	log.Trace().
		Int("bits", len(bits)).
		Msg("bits read from input reader")

	return createResult(ctx, bits, opts...)
}
