package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"

	"github.com/fedemengo/d2bist/pkg/core"
	"github.com/fedemengo/d2bist/pkg/flags"
	"github.com/fedemengo/d2bist/pkg/image"
	iio "github.com/fedemengo/d2bist/pkg/io"
	"github.com/fedemengo/d2bist/pkg/types"
)

var (
	outputString = false
	printStats   = false
	topKOutput   = -1
	maxBlockSize = 8
	blockSize    = -1
	symbolLen    = 2

	readDataCap   = ""
	compressionIn = ""

	writeDataCap   = ""
	compressionOut = ""

	pngFileName   = ""
	pixelLen      = 1
	separatorRune = rune(0)
	count         = 8
)

var app *cli.App

func init() {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:        "wcap",
			Usage:       "cap the amount of data to write after processing",
			Destination: &writeDataCap,
		}, &cli.StringFlag{
			Name:        "compression",
			Aliases:     []string{"c"},
			Usage:       "specify the compression algorithm to compress the output data",
			DefaultText: "auto",
			Destination: &compressionOut,
		}, &cli.StringFlag{
			Name:        "png",
			Usage:       "write bit string to png file",
			Destination: &pngFileName,
		}, &cli.IntFlag{
			Name:        "plen",
			Usage:       "length of a pixel in bits",
			DefaultText: "1",
			Destination: &pixelLen,
		}, &cli.StringFlag{
			Name:        "sep",
			Usage:       "separator to make the bin string more readable",
			DefaultText: "none",
			Action: func(_ *cli.Context, s string) error {
				if len(s) > 0 && len(s) != 1 {
					return fmt.Errorf("d2bist: bad separator `%s`", s)
				}
				separatorRune = rune(s[0])

				return nil
			},
		}, &cli.IntFlag{
			Name:        "count",
			Aliases:     []string{"b"},
			Usage:       "add a separator after #count bits",
			Destination: &count,
			DefaultText: "8",
		}, &cli.IntFlag{
			Name:        "topk",
			Aliases:     []string{"k"},
			Usage:       "output the top k most frequent substrings",
			Destination: &topKOutput,
			DefaultText: "all",
		}, &cli.IntFlag{
			Name:        "maxchunk",
			Usage:       "max chunk size of bits consider when counting substrings",
			Destination: &maxBlockSize,
			DefaultText: "8",
		}, &cli.IntFlag{
			Name:        "chunk",
			Usage:       "exact chunk size to consider when counting substrings",
			Destination: &blockSize,
		}, &cli.IntFlag{
			Name:        "slen",
			Usage:       "length of unitary symbol used when calculating data entropy",
			Destination: &symbolLen,
		}, &cli.BoolFlag{
			Name:        "stats",
			Aliases:     []string{"s"},
			Usage:       "output bits distribution stats",
			Destination: &printStats,
		}, &cli.BoolFlag{
			Name:        "str",
			Usage:       "the output will be a string of 0s and 1s",
			Destination: &outputString,
		},
	}

	app = &cli.App{
		Suggest:              true,
		EnableBashCompletion: true,
		Name:                 "d2bist",
		Description:          "Handle data as binary strings",
		Usage:                "decode and encode data to bit strings",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "rcap",
				Usage:       "cap the amount of data to read before processing",
				Destination: &readDataCap,
			}, &cli.StringFlag{
				Name:        "compression",
				Aliases:     []string{"c"},
				Usage:       "specify the compression algorithm to decompress the input data",
				DefaultText: "auto",
				Destination: &compressionIn,
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "decode",
				Aliases: []string{"d"},
				Usage:   "Decode data to corresponding binary string",
				Flags:   flags,
				Action:  decode,
			},
			{
				Name:    "encode",
				Aliases: []string{"e"},
				Usage:   "Encode a binary string to the corresponding data",
				Flags:   flags,
				Action:  encode,
			},
		},
	}
}

func Run() {
	log := logger()
	ctx := log.WithContext(context.Background())

	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Fatal().Err(err).Msg("execution failed")
		os.Exit(1)
	}
}

func logger() zerolog.Logger {
	level := zerolog.ErrorLevel
	l, err := zerolog.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err == nil {
		level = l
	}

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

func OptsFromFlags(_ context.Context) ([]core.Opt, error) {
	options := []core.Opt{}

	if maxBits, err := flags.ParseDataCapToBitsCount(readDataCap); err != nil {
		return nil, fmt.Errorf("cannot parse data cap flag")
	} else if maxBits > 0 {
		options = append(options, core.WithInBitsCap(maxBits))
	}
	if maxBits, err := flags.ParseDataCapToBitsCount(writeDataCap); err != nil {
		return nil, fmt.Errorf("cannot parse data cap flag")
	} else if maxBits > 0 {
		options = append(options, core.WithOutBitsCap(maxBits))
	}

	cInType := flags.ParseCompressionFlag(compressionIn)
	options = append(options, core.WithInCompression(cInType))

	cOutType := flags.ParseCompressionFlag(compressionOut)
	options = append(options, core.WithOutCompression(cOutType))

	if blockSize > 0 {
		options = append(options, core.WithStatsBlockSize(blockSize))
		if symbolLen > blockSize {
			return nil, fmt.Errorf("entropy chunk size cannot be greater than block size")
		}
		if symbolLen > 0 && blockSize%symbolLen != 0 {
			return nil, fmt.Errorf("entropy chunk size must be a multiple of block size")
		}
		options = append(options, core.WithStatsSymbolLen(symbolLen))
	} else {
		options = append(options, core.WithStatsMaxBlockSize(maxBlockSize))
	}

	if topKOutput > 0 {
		options = append(options, core.WithStatsTopK(topKOutput))
	}

	options = append(options, core.WithEntropyPlotName(fmt.Sprintf("entropy-%d", time.Now().Unix())))

	return options, nil
}

type operation func(context.Context, io.Reader, ...core.Opt) (*types.Result, error)

// decode read data and decodes it to the binary string
func decode(cliCtx *cli.Context) error {
	log := zerolog.Ctx(cliCtx.Context).With().Str("command", "decode").Logger()
	ctx := log.WithContext(cliCtx.Context)

	return process(ctx, cliCtx.Args().First(), core.Decode)
}

// encode read a binary string and encodes it to the equivalent data
func encode(cliCtx *cli.Context) error {
	log := zerolog.Ctx(cliCtx.Context).With().Str("command", "encode").Logger()
	ctx := log.WithContext(cliCtx.Context)

	return process(ctx, cliCtx.Args().First(), core.Encode)
}

func process(ctx context.Context, filename string, op operation) error {
	log := zerolog.Ctx(ctx)

	r := os.Stdin
	if len(filename) != 0 {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}

		defer f.Close()
		r = f
	}

	opts, err := OptsFromFlags(ctx)
	if err != nil {
		return fmt.Errorf("error parsing input flags: %w", err)
	}

	res, err := op(ctx, r, opts...)
	if err != nil {
		return err
	}

	err = outputBinaryString(ctx, res.Bits)
	if err != nil {
		return err
	}

	log.Trace().Int("bits", len(res.Bits)).Msg("encoded bits")

	if printStats {
		res.Stats.RenderStats(os.Stderr)
	}

	if len(pngFileName) > 0 {
		if pixelLen == 0 {
			pixelLen = 1
		}
		return image.WriteToPNG(res.Bits, pngFileName, pixelLen)
	}

	return nil
}

func outputBinaryString(ctx context.Context, bits []types.Bit) error {
	var err error
	if outputString || isatty.IsTerminal(os.Stdout.Fd()) {
		opts := []iio.Opt{
			iio.WithSep(separatorRune),
			iio.WithSepDistance(count),
		}
		fmt.Fprintln(os.Stdout, iio.BitsToString(bits, opts...))
	} else {
		err = iio.BitsToByteWriter(ctx, os.Stdout, bits)
	}

	return err
}
