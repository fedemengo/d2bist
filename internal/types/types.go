package types

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"os"
	"time"

	svg "github.com/ajstarks/svgo"
	"github.com/vdobler/chart"
	"github.com/vdobler/chart/imgg"
	"github.com/vdobler/chart/svgg"

	"github.com/fedemengo/go-data-structures/heap"
)

var ErrInvalidBit = errors.New("invalid bit")

type Bit uint8

func (b Bit) ToByte() byte {
	switch b {
	case 0:
		return byte('0')
	case 1:
		return byte('1')
	default:
		return byte('-')
	}
}

type SubstrCount struct {
	Total  int
	Length int
	Counts map[string]int

	SortedSubstrs []string
	AllCounts     map[string]int
}

type Stats struct {
	BitsCount int
	ByteCount int

	BitsStrCount map[int]map[int64]int
	SubstrsCount []SubstrCount

	CompressionStats *CompressionStats
	Entropy          []float64
}

func (s *Stats) RenderStats(w io.Writer) {
	if s.CompressionStats != nil {
		s.BitsStrCount = map[int]map[int64]int{
			1: s.BitsStrCount[1],
			2: s.BitsStrCount[2],
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "bits:", s.BitsCount)
	fmt.Fprintln(w)

	for l := 0; l < len(s.SubstrsCount); l++ {
		substrGroup := s.SubstrsCount[l]

		total, max := 0, 0
		for _, count := range substrGroup.Counts {
			total += count
			if count > max {
				max = count
			}
		}

		if len(substrGroup.Counts) < len(substrGroup.AllCounts) {
			s.printTopBistrK(max, total, substrGroup, w)
		} else {
			s.printAllBistrK(max, total, substrGroup, w)
		}

		if l < len(s.SubstrsCount)-1 {
			fmt.Fprintln(w)
		}
	}

	if len(s.Entropy) > 0 {
		renderEntropyChart(s.Entropy)
	}

	if s.CompressionStats != nil {
		fmt.Fprintf(w, `
compression ratio: %.3f
compression algorithm: %s
`, s.CompressionStats.CompressionRatio, s.CompressionStats.CompressionAlgorithm)
		s.CompressionStats.Stats.RenderStats(w)
	} else {
		fmt.Fprintln(w)
	}

}

func (s *Stats) printTopBistrK(max, total int, substrGroup SubstrCount, w io.Writer) {
	topK := heap.NewHeap(func(e1, e2 heap.Elem) bool {
		return e1.Val.(int) > e2.Val.(int)
	})

	for substr, count := range substrGroup.Counts {
		topK.Push(heap.Elem{Key: substr, Val: count})
	}

	for topK.Size() > 0 {
		e := topK.Pop()
		bitStr, v := e.Key.(string), e.Val.(int)
		digits := int(math.Log10(float64(max))) + 1
		count := v

		countStr := fmt.Sprintf("%10d", count)
		if digits > 0 && digits < len(countStr) {
			countStr = countStr[len(countStr)-digits:]
		}
		percentage := float64(count) * 1 / float64(total)

		fmt.Fprintf(w, "%s: %s - %.5f %%\n", bitStr, countStr, percentage)
	}
}

func (s *Stats) printAllBistrK(max, total int, substrGroup SubstrCount, w io.Writer) {
	for _, bitStr := range substrGroup.SortedSubstrs {
		count, ok := substrGroup.Counts[bitStr]
		if !ok {
			continue
		}

		percentage := float64(count) / float64(total)

		digits := int(math.Log10(float64(max))) + 1
		countStr := fmt.Sprintf("%10d", count)
		if digits > 0 && digits < len(countStr) {
			countStr = countStr[len(countStr)-digits:]
		}

		fmt.Fprintf(w, "%s: %s - %.5f %%\n", bitStr, countStr, percentage)
	}
}

func renderEntropyChart(entropy []float64) {
	name := fmt.Sprintf("entropy-%d", time.Now().Unix())
	dumper := NewDumper(name, 1, 1, 1300, 800)
	defer dumper.Close()

	pl := chart.ScatterChart{Title: "data entropy"}
	pl.Key.Pos = "itl"
	pl.Key.Hide = true

	x, y := make([]float64, len(entropy)), make([]float64, len(entropy))
	yMax := 0.0
	for i, e := range entropy {
		x[i] = float64(i)
		y[i] = e

		if e > yMax {
			yMax = e
		}
	}

	pl.AddDataPair(
		"entropy",
		x, y,
		chart.PlotStyleLines,
		chart.Style{
			Symbol:      0,
			SymbolColor: color.NRGBA{0xff, 0x00, 0x00, 0xff},
			LineStyle:   chart.SolidLine,
		})

	pl.YRange.MinMode.Fixed = true
	pl.YRange.MinMode.Value = 0
	pl.YRange.MaxMode.Fixed = true
	pl.YRange.MaxMode.Value = 1
	pl.YRange.TicSetting.Delta = 0.25
	pl.YRange.Label = "entropy"
	pl.YRange.TicSetting.Format = func(v float64) string {
		return fmt.Sprintf("%.2f", v)
	}
	pl.YRange.TicSetting.Mirror = 0

	pl.XRange.MinMode.Fixed = true
	pl.XRange.MinMode.Value = 0
	pl.XRange.MaxMode.Fixed = true
	pl.XRange.MaxMode.Value = float64(len(entropy))

	pl.XRange.TicSetting.Delta = float64(len(entropy) / 5)
	pl.XRange.TicSetting.Mirror = 0
	pl.XRange.TicSetting.Grid = chart.GridOff
	pl.XRange.Label = "offset"

	dumper.Plot(&pl)

	fmt.Println("entropy chart saved to", name+".svg")
}

type Dumper struct {
	N, M, W, H, Cnt  int
	S                *svg.SVG
	I                *image.RGBA
	svgFile, imgFile *os.File
}

func NewDumper(name string, n, m, w, h int) *Dumper {
	var err error
	dumper := Dumper{N: n, M: m, W: w, H: h}

	dumper.svgFile, err = os.Create(name + ".svg")
	if err != nil {
		panic(err)
	}
	dumper.S = svg.New(dumper.svgFile)
	dumper.S.Start(n*w, m*h)
	dumper.S.Title(name)
	dumper.S.Rect(0, 0, n*w, m*h, "fill: #ffffff")

	dumper.imgFile, err = os.Create(name + ".png")
	if err != nil {
		panic(err)
	}
	dumper.I = image.NewRGBA(image.Rect(0, 0, n*w, m*h))
	bg := image.NewUniform(color.RGBA{0xff, 0xff, 0xff, 0xff})
	draw.Draw(dumper.I, dumper.I.Bounds(), bg, image.ZP, draw.Src)

	return &dumper
}
func (d *Dumper) Close() {
	png.Encode(d.imgFile, d.I)
	d.imgFile.Close()

	d.S.End()
	d.svgFile.Close()
}

func (d *Dumper) Plot(c chart.Chart) {
	row, col := d.Cnt/d.N, d.Cnt%d.N

	igr := imgg.AddTo(d.I, col*d.W, row*d.H, d.W, d.H, color.RGBA{0xff, 0xff, 0xff, 0xff}, nil, nil)
	c.Plot(igr)

	sgr := svgg.AddTo(d.S, col*d.W, row*d.H, d.W, d.H, "", 12, color.RGBA{0xff, 0xff, 0xff, 0xff})
	c.Plot(sgr)

	d.Cnt++
}

type CompressionStats struct {
	CompressionRatio     float64
	CompressionAlgorithm string

	Stats *Stats
}

type Result struct {
	Bits  []Bit
	Stats *Stats
}
