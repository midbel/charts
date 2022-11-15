package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"

	"github.com/midbel/charts"
)

const (
	defaultWidth  = 800
	defaultHeight = 600
)

var pad = charts.Padding{
	Top:    40,
	Right:  10,
	Bottom: 60,
	Left:   60,
}

func main() {
	list := []string{
		"group-1",
		"group-2",
		"group-3",
		"group-4",
		"group-5",
	}
	xscale := charts.StringScaler(list, charts.NewRange(0, defaultWidth-pad.Horizontal()))
	yscale := charts.NumberScaler(charts.NumberDomain(50, 0), charts.NewRange(0, defaultHeight-pad.Vertical()))

	serie := charts.Serie[string, float64]{
		Title:    "group serie",
		Renderer: getRenderer(),
		X:        xscale,
		Y:        yscale,
	}
	for i := 1; i <= 7; i++ {
		pt := createPoint(i, 50)
		serie.Points = append(serie.Points, pt)
	}
	ch := charts.Chart[string, float64]{
		Width:   defaultWidth,
		Height:  defaultHeight,
		Padding: pad,
		Left:    getLeftAxis(yscale),
		Bottom:  getBottomAxis(xscale),
	}
	ch.Render(os.Stdout, serie)
}

func createPoint(ix, limit int) charts.Point[string, float64] {
	p := charts.CategoryPoint(fmt.Sprintf("group-%d", ix), 0)
	for i := 1; i <= 5; i++ {
		y := fmt.Sprintf("point-%03d", i)
		r := rand.Intn(limit) + 1
		s := charts.CategoryPoint(y, float64(r))
		p.Y += s.Y
		p.Sub = append(p.Sub, s)
	}
	return p
}

func getRenderer() charts.Renderer[string, float64] {
	return charts.GroupRenderer[string, float64]{
		Fill:  charts.Tableau10,
		Width: 0.75,
	}
}

func getBottomAxis(scaler charts.Scaler[string]) charts.Axis[string] {
	return charts.Axis[string]{
		Label:          "group",
		Orientation:    charts.OrientBottom,
		Scaler:         scaler,
		WithInnerTicks: true,
		WithOuterTicks: false,
		WithLabelTicks: true,
		Format:         func(s string) string { return s },
	}
}

func getLeftAxis(scaler charts.Scaler[float64]) charts.Axis[float64] {
	return charts.Axis[float64]{
		Label:          "value",
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
			return strconv.FormatFloat(f, 'f', -1, 64)
		},
	}
}
