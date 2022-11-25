package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"

	"github.com/midbel/charts"
	"github.com/midbel/slices"
)

const (
	defaultWidth  = 1366
	defaultHeight = 600
	defaultInner  = 10
	defaultOuter  = 300
)

var pad = charts.Padding{
	Top:    40,
	Right:  10,
	Bottom: 60,
	Left:   60,
}

func main() {
	var (
		typname = flag.String("t", "", "type")
		width   = flag.Float64("w", defaultWidth, "chart width")
		height  = flag.Float64("g", defaultHeight, "chart height")
		inner   = flag.Float64("inner-radius", defaultInner, "inner radius sun")
		outer   = flag.Float64("outer-radius", defaultOuter, "outer radius sun")
		normal  = flag.Bool("n", false, "normalize")
	)
	flag.Parse()

	dat, err := readFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	sort.Slice(dat.Serie.Points, func(i, j int) bool {
		return dat.Serie.Points[i].Y > dat.Serie.Points[j].Y
	})
	dat.List = dat.List[:0]
	for i := range dat.Serie.Points {
		dat.List = append(dat.List, dat.Serie.Points[i].X)
	}

	rdr, err := getRenderer(*typname, *inner, *outer, *normal)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if *normal {
		dat.Max = 1
	}
	dat.Serie.Renderer = rdr
	xscale := charts.StringScaler(dat.List, charts.NewRange(0, *width-pad.Horizontal()))
	yscale := charts.NumberScaler(charts.NumberDomain(dat.Max, 0), charts.NewRange(0, *height-pad.Vertical()))

	dat.Serie.X = xscale
	dat.Serie.Y = yscale

	ch := charts.Chart[string, float64]{
		Title:   "US Population",
		Width:   *width,
		Height:  *height,
		Padding: pad,
	}
	if *typname != "sun" && *typname != "sunburst" {
		ch.Left = getLeftAxis(yscale, *normal)
		ch.Bottom = getBottomAxis(xscale)
	}
	sort.Slice(dat.Serie.Points, func(i, j int) bool {
		return dat.Serie.Points[i].Y > dat.Serie.Points[j].Y
	})
	ch.Render(os.Stdout, dat.Serie)
}

func getRenderer(name string, inner, outer float64, normalize bool) (charts.Renderer[string, float64], error) {
	var (
		rdr   charts.Renderer[string, float64]
		style = charts.DefaultStyle()
	)
	style.FillList = charts.Tableau10
	style.LineColor = "white"
	style.LineWidth = 1
	switch name {
	case "stacked", "vert", "vertical", "":
		rdr = charts.StackedRenderer[string, float64]{
			Style:     style,
			Width:     0.8,
			Normalize: normalize,
		}
	case "group", "horiz", "horizontal":
		rdr = charts.GroupRenderer[string, float64]{
			Style: style,
			Width: 0.8,
		}
	case "sun", "sunburst":
		rdr = charts.SunburstRenderer[string, float64]{
			Style:       style,
			InnerRadius: inner,
			OuterRadius: outer,
		}
	default:
		return nil, fmt.Errorf("%s: invalid renderer name - choose between stacked, group, sun", name)
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
	head, _ := rs.Read()
	head = slices.Rest(head)

	for {
		row, err := rs.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return dat, err
		}
		dat.List = append(dat.List, slices.Fst(row))
		var (
			pt    charts.Point[string, float64]
			total float64
		)
		for i, n := range slices.Rest(row) {
			f, _ := strconv.ParseFloat(n, 64)
			pt.Sub = append(pt.Sub, charts.CategoryPoint(head[i], f))
			total += f
		}
		pt.X = slices.Fst(row)
		pt.Y = total
		dat.Serie.Points = append(dat.Serie.Points, pt)

		dat.Max = math.Max(dat.Max, total)
	}
	return dat, err
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

func getLeftAxis(scaler charts.Scaler[float64], normalize bool) charts.Axis[float64] {
	var title string
	if normalize {
		title = "population (%)"
	} else {
		title = "population (million)"
	}
	return charts.Axis[float64]{
		Label:          title,
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
			if normalize {
				return strconv.FormatFloat(f*100, 'f', 0, 64) + "%"
			}
			return strconv.FormatFloat(f/(1000*1000), 'f', 0, 64) + "M"
		},
	}
}
