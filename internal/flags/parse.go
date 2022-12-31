package flags

import (
	"errors"
	"fmt"
	"strconv"
	"unicode"

	"github.com/fedemengo/d2bist/compression"
)

var (
	supportedSuffixes = map[string]int{
		"":  1,
		"B": 8,

		"kB": 8 * 1000,
		"K":  8 * 1024,

		"MB": 8 * 1000 * 1000,
		"M":  8 * 1024 * 1024,

		"GB": 8 * 1000 * 1000 * 1000,
		"G":  8 * 1024 * 1024 * 1024,
	}
)

var (
	ErrInvalidFlag = errors.New("flag is not valid")
)

func ParseCompressionFlag(fc string) compression.CompressionType {
	switch fc {
	case "zip":
		return compression.Zip
	case "gz", "gzip":
		return compression.Gzip
	case "b", "brotli":
		return compression.Brotli
	case "zstd":
		return compression.Zstd
	case "s2":
		return compression.S2
	case "h", "huff":
		return compression.Huff
	default:
		return compression.None
	}
}

func ParseDataCapToBitsCount(dataCap string) (int, error) {
	if dataCap == "" {
		return -1, nil
	}

	decimal := ""
	suffix := ""

	last := rune(dataCap[len(dataCap)-1])
	if unicode.IsDigit(last) {
		decimal = dataCap
	} else if last == 'B' {
		if len(dataCap) < 2 {
			return -1, fmt.Errorf("flag `%s` is not valid: %w", dataCap, ErrInvalidFlag)
		}

		slast := rune(dataCap[len(dataCap)-2])
		if unicode.IsDigit(slast) {
			decimal = dataCap[:len(dataCap)-1]
			suffix = dataCap[len(dataCap)-1:]
		} else {
			decimal = dataCap[:len(dataCap)-2]
			suffix = dataCap[len(dataCap)-2:]
		}

	} else {
		decimal = dataCap[:len(dataCap)-1]
		suffix = dataCap[len(dataCap)-1:]
	}

	multiplier, ok := supportedSuffixes[suffix]
	if !ok {
		return -1, fmt.Errorf("`%s` is not supported: %w", suffix, ErrInvalidFlag)
	}

	i, err := strconv.Atoi(decimal)
	if err != nil {
		return -1, fmt.Errorf("%s is not valid decimal: %w", decimal, ErrInvalidFlag)
	}

	return i * multiplier, nil
}
