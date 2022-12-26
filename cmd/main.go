package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v2"

	"github.com/fedemengo/f2bist/internal/engine"
	"github.com/fedemengo/f2bist/internal/io"
	"github.com/fedemengo/f2bist/internal/types"
)

var (
	outputUTF8String = false
	printStats       = false
)

func main() {

	app := &cli.App{
		EnableBashCompletion: true,
		Name:                 "biner",
		Description:          "handle files as binary strings",
		Commands: []*cli.Command{
			{
				Name:    "decode",
				Aliases: []string{"d"},
				Usage:   "decode a file to a binary string",
				Flags: []cli.Flag{
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
				},

				Action: decode,
			},
			{
				Name:        "encode",
				Aliases:     []string{"e"},
				Description: "encode a binary string to a binary file",
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
	readBits := io.BitsFromFile(filename)
	if len(filename) == 0 {
		readBits = io.BitsFromStdin
	}

	bits, err := readBits()
	if err != nil {
		return err
	}

	outputStats(bits)
	outputBinaryString(bits)

	return nil
}

func outputBinaryString(bits []types.Bit) {
	if outputUTF8String {
		fmt.Println(io.BitsToString(bits))
	} else if isatty.IsTerminal(os.Stdout.Fd()) {
		fmt.Println(io.BitsToString(bits, io.WithSep(' ')))
	} else {
		io.BitsToWriter(os.Stdout, bits)
	}
}

func outputStats(bits []types.Bit) {
	stats := engine.AnalizeBits(bits)
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

func encode(ctx *cli.Context) error {
	return nil
}
