package types

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"log"
	"math"
	"os"

	"github.com/vdobler/chart"
	"github.com/vdobler/chart/imgg"

	"github.com/fedemengo/d2bist/pkg/compression"
	"github.com/fedemengo/go-data-structures/heap"
)

var ErrInvalidBit = errors.New("invalid bit")

type EntropyType string

const (
	ShannonEntropy = EntropyType("Shannon")
	GzipEntropy    = EntropyType(compression.Gzip)
	BrotliEntropy  = EntropyType(compression.Brotli)
	S2Entropy      = EntropyType(compression.S2)
	ZstdEntropy    = EntropyType(compression.Zstd)
	Bzip2Entropy   = EntropyType(compression.Bzip2)
)

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

type Entropy struct {
	Name   EntropyType
	Values []float64
}

func NewShannonEntropy() *Entropy {
	return &Entropy{Name: ShannonEntropy}
}

func NewCompressionEntropy(cType compression.CompressionType) *Entropy {
	return &Entropy{Name: EntropyType(cType)}
}

type Stats struct {
	BitsCount int
	ByteCount int

	BitsStrCount map[int]map[int64]int
	SubstrsCount []SubstrCount

	CompressionStats *CompressionStats
	EntropyPlotName  string
	Entropy          []*Entropy
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
		renderEntropyChart(s.EntropyPlotName, s.Entropy)
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

var colorsMap = map[EntropyType]color.RGBA{
	ShannonEntropy: {R: 255, G: 0, B: 0, A: 255},
	GzipEntropy:    {R: 0, G: 0, B: 255, A: 255},
	BrotliEntropy:  {R: 0, G: 255, B: 0, A: 255},
	Bzip2Entropy:   {R: 255, G: 165, B: 0, A: 1},
}

func renderEntropyChart(plotName string, entropies []*Entropy) {
	dumper := NewDumper(plotName, 1, 1, 1300, 800)
	defer dumper.Close()

	pl := chart.ScatterChart{Title: "data entropy"}

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

	maxEntropyPoints := 0
	for _, e := range entropies {
		log.Printf("entropy: %s, len: %d", e.Name, len(e.Values))
		entropy := e.Values

		if len(entropy) > maxEntropyPoints {
			maxEntropyPoints = len(entropy)
		}

		x, y := make([]float64, len(entropy)), make([]float64, len(entropy))
		for i, e := range entropy {
			x[i] = float64(i)
			y[i] = e
		}

		pl.AddDataPair(
			string(e.Name),
			x, y,
			chart.PlotStyleLines,
			chart.Style{
				Symbol:      0,
				SymbolColor: colorsMap[e.Name],
				LineStyle:   chart.SolidLine,
			})
	}

	pl.Key.Pos = "obr"
	pl.Key.Cols = maxEntropyPoints

	pl.XRange.MinMode.Fixed = true
	pl.XRange.MinMode.Value = 0
	pl.XRange.MaxMode.Fixed = true
	pl.XRange.MaxMode.Value = float64(maxEntropyPoints)

	pl.XRange.TicSetting.Delta = float64(maxEntropyPoints / 5)
	pl.XRange.TicSetting.Mirror = 0
	pl.XRange.TicSetting.Grid = chart.GridOff
	pl.XRange.Label = "offset"

	dumper.Plot(&pl)
}

type Dumper struct {
	N, M, W, H int
	Cnt        int
	I          *image.RGBA
	imgFile    *os.File
}

func NewDumper(name string, n, m, w, h int) *Dumper {
	var err error
	dumper := Dumper{N: n, M: m, W: w, H: h}

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
}

func (d *Dumper) Plot(c chart.Chart) {
	row, col := d.Cnt/d.N, d.Cnt%d.N

	igr := imgg.AddTo(d.I, col*d.W, row*d.H, d.W, d.H, color.RGBA{0xff, 0xff, 0xff, 0xff}, nil, nil)
	c.Plot(igr)

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
