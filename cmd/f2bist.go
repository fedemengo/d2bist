package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v2"

	"github.com/fedemengo/f2bist/core"
	"github.com/fedemengo/f2bist/internal/flags"
	"github.com/fedemengo/f2bist/internal/image"
	"github.com/fedemengo/f2bist/internal/io"
	"github.com/fedemengo/f2bist/internal/types"
)

var (
	outputString = false
	printStats   = false

	separatorRune = rune(0)
	separator     = ""
	dataCap       = ""
	pngFileName   = ""

	count = 8
)

var app *cli.App

func init() {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:        "cap",
			Usage:       "cap the amount of data to read",
			Destination: &dataCap,
		}, &cli.StringFlag{
			Name:        "png",
			Usage:       "write bit string to png file",
			Destination: &pngFileName,
		}, &cli.StringFlag{
			Name:        "sep",
			Usage:       "separator to make the bin string more readable",
			Destination: &separator,
			DefaultText: "none",
			Action: func(ctx *cli.Context, s string) error {
				if len(separator) > 0 && len(separator) != 1 {
					return fmt.Errorf("f2bist: bad separator `%s`", separator)
				}
				separatorRune = rune(separator[0])

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
		Commands: []*cli.Command{
			{
				Name:    "decode",
				Aliases: []string{"d"},
				Usage:   "Decode a file to a binary string",
				Flags:   flags,
				Action:  decode,
			},
			{
				Name:    "encode",
				Aliases: []string{"e"},
				Usage:   "Encode a binary string to a binary file",
				Flags:   flags,
				Action:  encode,
			},
		},
	}
}

func Run() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func decode(ctx *cli.Context) error {
	filename := ctx.Args().First()

	r := os.Stdin
	if len(filename) != 0 {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}

		defer f.Close()
		r = f
	}

	options := []core.Opt{}
	if maxBits, err := flags.ParseDataCapToBitsCount(dataCap); err != nil {
		log.Fatal(err)
	} else if maxBits > 0 {
		options = append(options, core.WithBitsCap(maxBits))
	}

	res, err := core.Decode(context.Background(), r, options...)
	if err != nil {
		return err
	}

	err = outputBinaryString(res.Bits)
	if err != nil {
		return err
	}

	if printStats {
		outputStats(res.Stats)
	}

	if len(pngFileName) > 0 {
		return image.WriteToPNG(res.Bits, pngFileName)
	}

	return nil
}

func encode(_ *cli.Context) error {
	options := []core.Opt{}
	if maxBits, err := flags.ParseDataCapToBitsCount(dataCap); err != nil {
		log.Fatal(err)
	} else if maxBits > 0 {
		options = append(options, core.WithBitsCap(maxBits))
	}

	res, err := core.Encode(context.Background(), os.Stdin, options...)
	if err != nil {
		return err
	}

	err = outputBinaryString(res.Bits)
	if err != nil {
		return err
	}

	if printStats {
		outputStats(res.Stats)
	}

	if len(pngFileName) > 0 {
		return image.WriteToPNG(res.Bits, pngFileName)
	}

	return nil
}

func outputBinaryString(bits []types.Bit) error {
	var err error
	if outputString || isatty.IsTerminal(os.Stdout.Fd()) {
		opts := []io.Opt{
			io.WithSep(separatorRune),
			io.WithSepDistance(count),
		}
		fmt.Fprintln(os.Stdout, io.BitsToString(bits, opts...))
	} else {
		err = io.BitsToWriter(os.Stdout, bits)
	}

	return err
}

func outputStats(stats *types.Stats) {
	fmt.Fprintf(os.Stderr, `
bits: %d

0: %d
1: %d
`, stats.SizeBits, stats.ZeroCount, stats.OneCount)

	fmt.Fprintln(os.Stderr)

	for i := 1; i <= stats.MaxStringLen; i++ {
		zc, oc := stats.ZeroStrings[i], stats.OneStrings[i]
		fmt.Fprintf(os.Stderr, "l%02d: 0: %9d - 1: %9d | ratio: %.5f\n", i, zc, oc, float64(zc)/float64(oc))
	}
	fmt.Fprintln(os.Stderr)
}
