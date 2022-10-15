package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"

	"github.com/midbel/charts"
	"github.com/midbel/slices"
)

const (
	defaultWidth  = 1366
	defaultHeight = 600
)

var pad = charts.Padding{
	Top:    80,
	Right:  80,
	Bottom: 80,
	Left:   80,
}

func main() {
	typname := flag.String("t", "", "type")
	flag.Parse()

	dat, err := readFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	rdr, err := getRenderer(*typname)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	dat.Serie.Renderer = rdr
	xscale := charts.StringScaler(dat.List, charts.NewRange(0, defaultWidth-pad.Horizontal()))
	yscale := charts.NumberScaler(charts.NumberDomain(dat.Max, 0), charts.NewRange(0, defaultHeight-pad.Vertical()))

	dat.Serie.X = xscale
	dat.Serie.Y = yscale

	ch := charts.Chart[string, float64]{
		Title:   "US Population",
		Width:   defaultWidth,
		Height:  defaultHeight,
		Padding: pad,
	}
	if *typname != "pie" {
		ch.Left = getLeftAxis(yscale)
		ch.Bottom = getBottomAxis(xscale)
	}
	ch.Render(os.Stdout, dat.Serie)
}

func getRenderer(name string) (charts.Renderer[string, float64], error) {
	var rdr charts.Renderer[string, float64]
	switch name {
	case "bar", "":
		rdr = charts.BarRenderer[string, float64]{
			Width: 0.6,
			Fill:  []string{"steelblue"},
		}
	case "pie":
		rdr = charts.PieRenderer[string, float64]{
			InnerRadius: 150,
			OuterRadius: 250,
		}
	default:
		return nil, fmt.Errorf("%s: invalid renderer name - choose between pie or bar", name)
	}
	return rdr, nil
}

type Data struct {
	Serie charts.Serie[string, float64]
	List  []string
	Total float64
	Max   float64
}

func readFile(file string) (Data, error) {
	var (
		dat Data
		err error
	)
	r, err := os.Open(file)
	if err != nil {
		return dat, err
	}
	defer r.Close()

	dat.Serie.Title = "population"

	rs := csv.NewReader(r)
	rs.Read()
	for {
		row, err := rs.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return dat, err
		}
		dat.List = append(dat.List, slices.Fst(row))
		total := sumValues(slices.Rest(row))

		dat.Serie.Points = append(dat.Serie.Points, charts.CategoryPoint(slices.Fst(row), total))
		dat.Total += total
		dat.Max = math.Max(dat.Max, total)
	}
	return dat, err
}

func sumValues(row []string) float64 {
	var total float64
	for _, n := range slices.Rest(row) {
		f, _ := strconv.ParseFloat(n, 64)
		total += f
	}
	return total
}

func getBottomAxis(scaler charts.Scaler[string]) charts.Axis[string] {
	return charts.Axis[string]{
		Label:          "state",
		Orientation:    charts.OrientBottom,
		Scaler:         scaler,
		WithInnerTicks: true,
		WithOuterTicks: false,
		WithLabelTicks: true,
		WithBands:      true,
		Format:         func(s string) string { return s },
	}
}

func getLeftAxis(scaler charts.Scaler[float64]) charts.Axis[float64] {
	return charts.Axis[float64]{
		Label:          "population",
		Ticks:          10,
		Orientation:    charts.OrientLeft,
		Scaler:         scaler,
		WithInnerTicks: true,
		WithLabelTicks: true,
		WithOuterTicks: true,
		Format: func(f float64) string {
			if f == 0 {
				return "0"
			}
			return strconv.FormatFloat(f/(1000*1000), 'f', 3, 64) + "M"
		},
	}
}
