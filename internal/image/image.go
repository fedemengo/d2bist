package image

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"github.com/fedemengo/f2bist/internal/types"
)

const (
	maxW = 5_000
	maxH = 5_000
)

func WriteToPNG(bits []types.Bit, filename string) error {

	n := int(math.Sqrt(float64(len(bits))))

	currW := n + 1
	currH := n + 1

	upLeft := image.Point{0, 0}
	lowRight := image.Point{currW, currH}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	black := color.RGBA{0, 0, 0, 0xff}
	white := color.RGBA{255, 255, 255, 0xff}

	for x := 0; x < currW; x++ {
		for y := 0; y < currH; y++ {
			idx := y*currW + x
			if idx >= len(bits) {
				break
			}

			if bits[idx] == 0 {
				img.Set(x, y, white)
			} else {
				img.Set(x, y, black)
			}
		}
	}

	f, err := os.Create(fmt.Sprintf("%s.png", filename))
	if err != nil {
		return err
	}

	png.Encode(f, img)

	return nil
}
