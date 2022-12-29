package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/fedemengo/f2bist/core"
	"github.com/fedemengo/f2bist/internal/flags"
	"github.com/fedemengo/f2bist/internal/image"
	fio "github.com/fedemengo/f2bist/internal/io"
	"github.com/fedemengo/f2bist/internal/types"
)

var (
	outputString = false
	printStats   = false

	readDataCap   = ""
	compressionIn = ""

	writeDataCap   = ""
	compressionOut = ""

	pngFileName   = ""
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
			Usage:       "specify the compression algorithm do compress the output data",
			DefaultText: "auto",
			Destination: &compressionOut,
		},
		&cli.StringFlag{
			Name:        "png",
			Usage:       "write bit string to png file",
			Destination: &pngFileName,
		}, &cli.StringFlag{
			Name:        "sep",
			Usage:       "separator to make the bin string more readable",
			DefaultText: "none",
			Action: func(_ *cli.Context, s string) error {
				if len(s) > 0 && len(s) != 1 {
					return fmt.Errorf("f2bist: bad separator `%s`", s)
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
		Name:                 "f2bist",
		Description:          "Handle files as binary strings",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "rcap",
				Usage:       "cap the amount of data to read before processing",
				Destination: &readDataCap,
			}, &cli.StringFlag{
				Name:        "compression",
				Aliases:     []string{"c"},
				Usage:       "specify the compression algorithm do decompress the input data",
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
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	if err := app.Run(os.Args); err != nil {
		zlog.Fatal().Err(err).Msg("execution failed")
		os.Exit(1)
	}
}

func logger() zerolog.Logger {
	level := zerolog.Disabled
	switch l := os.Getenv("LOG_LEVEL"); l {
	case "trace":
		level = zerolog.TraceLevel
	case "error":
		level = zerolog.ErrorLevel
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

func OptsFromFlags() ([]core.Opt, error) {
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

	return options, nil
}

// decode read data and decodes it to the binary string
func decode(cliCtx *cli.Context) error {
	log := logger()
	ctx := log.WithContext(context.Background())

	filename := cliCtx.Args().First()

	r := os.Stdin
	if len(filename) != 0 {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}

		defer f.Close()
		r = f
	}

	opts, err := OptsFromFlags()
	if err != nil {
		return fmt.Errorf("error parsing input flags: %w", err)
	}

	res, err := core.Decode(ctx, r, opts...)
	if err != nil {
		return fmt.Errorf("cannot decode: %w", err)
	}

	log.Trace().Int("bits", len(res.Bits)).Msg("decoded bits")

	err = outputBinaryString(res.Bits)
	if err != nil {
		return err
	}

	if printStats {
		res.Stats.RenderStats(os.Stderr)
	}

	if len(pngFileName) > 0 {
		return image.WriteToPNG(res.Bits, pngFileName)
	}

	return nil
}

// encode read a binary string and encodes it to the equivalent data
func encode(cliCtx *cli.Context) error {
	log := logger()
	ctx := log.WithContext(context.Background())

	filename := cliCtx.Args().First()

	r := os.Stdin
	if len(filename) != 0 {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}

		defer f.Close()
		r = f
	}

	opts, err := OptsFromFlags()
	if err != nil {
		return fmt.Errorf("error parsing input flags: %w", err)
	}

	res, err := core.Encode(ctx, r, opts...)
	if err != nil {
		return err
	}

	err = outputBinaryString(res.Bits)
	if err != nil {
		return err
	}

	log.Trace().Int("bits", len(res.Bits)).Msg("encoded bits")

	if printStats {
		res.Stats.RenderStats(os.Stderr)
	}

	if len(pngFileName) > 0 {
		return image.WriteToPNG(res.Bits, pngFileName)
	}

	return nil
}

func outputBinaryString(bits []types.Bit) error {
	var err error
	if outputString || isatty.IsTerminal(os.Stdout.Fd()) {
		opts := []fio.Opt{
			fio.WithSep(separatorRune),
			fio.WithSepDistance(count),
		}
		fmt.Fprintln(os.Stdout, fio.BitsToString(bits, opts...))
	} else {
		err = fio.WriteBits(os.Stdout, bits)
	}

	return err
}
