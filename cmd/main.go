package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v2"

	"github.com/fedemengo/f2bist"
	"github.com/fedemengo/f2bist/internal/flags"
	"github.com/fedemengo/f2bist/internal/image"
	"github.com/fedemengo/f2bist/internal/io"
	"github.com/fedemengo/f2bist/internal/types"
)

var (
	outputUTF8String = false
	printStats       = false

	dataCap     = ""
	pngFileName = ""
)

func main() {
	flags := []cli.Flag{
		&cli.BoolFlag{
			Name:        "utf8",
			Usage:       "the output will be a utf8 string of 0s and 1s",
			Destination: &outputUTF8String,
		},
		&cli.BoolFlag{
			Name:        "stats",
			Aliases:     []string{"s"},
			Usage:       "output bists distribution stats",
			Destination: &printStats,
		},
		&cli.StringFlag{
			Name:        "cap",
			Usage:       "cap the amount of data to read",
			Destination: &dataCap,
		},
		&cli.StringFlag{
			Name:        "png",
			Usage:       "write bit string to png file",
			Destination: &pngFileName,
		},
	}

	app := &cli.App{
		EnableBashCompletion: true,
		Name:                 "biner",
		Description:          "handle files as binary strings",
		Commands: []*cli.Command{
			{
				Name:    "decode",
				Aliases: []string{"d"},
				Usage:   "decode a file to a binary string",
				Flags:   flags,
				Action:  decode,
			},
			{
				Name:        "encode",
				Aliases:     []string{"e"},
				Description: "encode a binary string to a binary file",
				Flags:       flags,
				Action:      encode,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
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

	options := []f2bist.Opt{}
	if maxBits, err := flags.ParseDataCapToBitsCount(dataCap); err != nil {
		log.Fatal(err)
	} else if maxBits > 0 {
		options = append(options, f2bist.WithBitsCap(maxBits))
	}

	res, err := f2bist.Decode(context.Background(), r, options...)
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
	options := []f2bist.Opt{}
	if maxBits, err := flags.ParseDataCapToBitsCount(dataCap); err != nil {
		log.Fatal(err)
	} else if maxBits > 0 {
		options = append(options, f2bist.WithBitsCap(maxBits))
	}

	res, err := f2bist.Encode(context.Background(), os.Stdin, options...)
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
	if outputUTF8String {
		fmt.Println(io.BitsToString(bits))
	} else if isatty.IsTerminal(os.Stdout.Fd()) {
		fmt.Println(io.BitsToString(bits, io.WithSep(' ')))
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
