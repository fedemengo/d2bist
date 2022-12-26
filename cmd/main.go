package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fedemengo/f2bist/internal/engine"
	"github.com/fedemengo/f2bist/internal/io"
	"github.com/fedemengo/f2bist/internal/types"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v2"
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

	engine.AnalizeBits(bits)

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

func encode(ctx *cli.Context) error {
	return nil
}
