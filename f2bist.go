package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	isatty "github.com/mattn/go-isatty"
	"github.com/urfave/cli/v2"
)

type bit uint8

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
				Action:  decode,
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

func encode(ctx *cli.Context) error {

	filename := ctx.Args().First()
	read := bitsFromFile(filename)
	if filename == "" {
		read = bitsFromStdin
	}

	bits, err := read()
	if err != nil {
		return err
	}

	printAsBinaryString(bits...)

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

func bitsFromReader(r io.Reader) ([]bit, error) {
	bits := []bit{}

	bytes := make([]byte, 8)
	lastCount := -1
	for {
		n, err := r.Read(bytes)
		if errors.Is(err, io.EOF) {
			lastCount = n
			break
		}
		if err != nil {
			return nil, err
		}

		for _, b := range bytes[:n] {
			bitsArray := byteToBits(b)
			//fmt.Printf("%v `%c`\n", bitsArray, b)
			bits = append(bits, bitsArray[0:8]...)
		}
	}

	if lastCount > 0 {
		for _, b := range bytes[:lastCount] {
			bitsArray := byteToBits(b)
			bits = append(bits, bitsArray[0:8]...)
		}
	}

	return bits, nil

}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func bitToRune(b bit) rune {
	if b == 0 {
		return '0'
	}
	return '1'
}

func humanReadableBinStr(bits ...bit) string {
	sb := strings.Builder{}
	for i, b := range bits {
		if i > 0 && i%8 == 0 {
			sb.WriteRune('.')
		}
		sb.WriteRune(bitToRune(b))
	}

	return sb.String()
}

func binaryStringToWriter(w io.Writer, bits ...bit) {
	for i := 0; i < len(bits)/8; i++ {
		byteVal := [8]bit{}
		copy(byteVal[:], bits[i*8:min((i+1)*8, len(bits)-1)])
		w.Write([]byte{bitsToByte(byteVal)})
	}

}

func printAsBinaryString(bits ...bit) {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		bstr := humanReadableBinStr(bits...)
		fmt.Println(bstr)
	} else {
		binaryStringToWriter(os.Stdout, bits...)
	}
}

func bitsToByte(bits [8]bit) byte {
	b := byte(0)

	for i := range bits {
		b += byte(bits[i] << (7 - i))
	}

	return b
}

func byteToBits(b byte) [8]bit {
	bits := [8]bit{}
	for i := 0; i < 8; i++ {
		if b&(1<<(7-i)) > 0 {
			bits[i] = 1
		} else {
			bits[i] = 0
		}
	}

	return bits
}

func decode(ctx *cli.Context) error {
	return nil
}
