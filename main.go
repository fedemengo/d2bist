package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

type bit uint8

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
	readBits := bitsFromFile(filename)
	if len(filename) == 0 {
		readBits = bitsFromStdin
	}

	bits, err := readBits()
	if err != nil {
		return err
	}

	outputBinaryString(bits)
	analizeBits(bits)

	return nil
}

func bitsFromStdin() ([]bit, error) {
	return bitsFromReader(os.Stdin)
}

func bitsFromFile(filename string) func() ([]bit, error) {
	return func() ([]bit, error) {
		f, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		return bitsFromReader(f)
	}
}

func encode(ctx *cli.Context) error {
	return nil
}
