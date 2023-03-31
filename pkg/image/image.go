package image

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"github.com/fedemengo/d2bist/pkg/engine"
	"github.com/fedemengo/d2bist/pkg/types"
)

const (
	maxW = 5_000
	maxH = 5_000
)

var colorsMap = map[int]map[uint64]color.RGBA{
	1: {
		0: {255, 255, 255, 255}, // White
		1: {0, 0, 0, 255},       // Black
	},
	2: {
		0: {255, 255, 0, 255}, // Yellow
		1: {0, 255, 0, 255},   // Green
		2: {0, 0, 255, 255},   // Blue
		3: {255, 165, 0, 255}, // Orange
	},
	3: {
		0: {255, 255, 0, 255},   // Yellow
		1: {0, 255, 0, 255},     // Green
		2: {0, 0, 255, 255},     // Blue
		3: {255, 0, 0, 255},     // Red
		4: {255, 0, 255, 255},   // Magenta
		5: {0, 255, 255, 255},   // Cyan
		6: {255, 165, 0, 255},   // Orange
		7: {128, 128, 128, 255}, // Gray
	},
	4: {
		0:  {255, 255, 0, 255},   // Yellow
		1:  {0, 255, 0, 255},     // Green
		2:  {0, 0, 255, 255},     // Blue
		3:  {255, 0, 0, 255},     // Red
		4:  {0, 255, 255, 255},   // Cyan
		5:  {255, 0, 255, 255},   // Magenta
		6:  {255, 165, 0, 255},   // Orange
		7:  {128, 0, 128, 255},   // Purple
		8:  {0, 255, 0, 255},     // Lime
		9:  {0, 128, 128, 255},   // Teal
		10: {255, 192, 203, 255}, // Pink
		11: {230, 230, 250, 255}, // Lavender
		12: {165, 42, 42, 255},   // Brown
		13: {128, 128, 0, 255},   // Olive
		14: {0, 0, 128, 255},     // Navy
		15: {128, 128, 128, 255}, // Gray
	},
}

func bitsToColors(bits []types.Bit, pixelLen int) ([]color.RGBA, error) {
	if pixelLen < 1 {
		return nil, fmt.Errorf("pixel length must be greater than 0")
	}

	colors := make([]color.RGBA, 0, len(bits))
	bw := engine.NewBitsWindow(bits, pixelLen)

	for wv, err := bw.ToIntSlide(); err == nil; wv, err = bw.ToIntSlide() {
		colors = append(colors, colorsMap[pixelLen][wv])
	}

	return colors, nil
}

func WriteToPNG(bits []types.Bit, filename string, pixelLen int) error {
	n := int(math.Min(math.Sqrt(float64(len(bits))), maxW))

	currW, currH := n+1, n+1
	upLeft, lowRight := image.Point{0, 0}, image.Point{currW, currH}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	colors, err := bitsToColors(bits, pixelLen)
	if err != nil {
		return fmt.Errorf("error converting bits to colors: %w", err)
	}

	for x := 0; x < currW; x++ {
		for y := 0; y < currH; y++ {
			idx := y*currW + x
			if idx >= len(colors) {
				break
			}

			img.Set(x, y, colors[idx])
		}
	}

	f, err := os.Create(fmt.Sprintf("%s.png", filename))
	if err != nil {
		return err
	}

	return png.Encode(f, img)
}
