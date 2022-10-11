package main

import (
	"os"
	"strconv"

	"github.com/midbel/charts"
)

const (
	defaultWidth  = 800
	defaultHeight = 600
)

var pad = charts.Padding{
	Top:    10,
	Right:  45,
	Bottom: 40,
	Left:   60,
}

func main() {
	var (
		langs         = []string{"go", "javascript", "python", "rust", "java", "c++"}
		valueScale    = charts.NumberScaler(charts.NumberDomain(100, 0), charts.NewRange(0, defaultHeight-pad.Vertical()))
		categoryScale = charts.StringScaler(langs, charts.NewRange(0, defaultWidth-pad.Horizontal()))
	)

	rdr := charts.BarRenderer[string, float64]{
		Width: 0.5,
		Fill:  []string{"steelblue"},
		// Fill: []string{"steelblue", "lightsalmon", "mediumorchid", "firebrick"},
	}
	ser := charts.Serie[string, float64]{
		Title:    "preferences",
		Color:    "blue",
		X:        categoryScale,
		Y:        valueScale,
		Renderer: rdr,
		Points: []charts.Point[string, float64]{
			charts.CategoryPoint("go", 95),
			charts.CategoryPoint("javascript", 25),
			charts.CategoryPoint("python", 60),
			charts.CategoryPoint("rust", 10),
			charts.CategoryPoint("java", 5),
			charts.CategoryPoint("c++", 70),
		},
	}

	ch := charts.Chart[string, float64]{
		Width:   defaultWidth,
		Height:  defaultHeight,
		Padding: pad,
		Left:    getLeftAxis(valueScale),
		Bottom:  getBottomAxis(categoryScale),
	}
	ch.Render(os.Stdout, ser)
}

func getBottomAxis(scaler charts.Scaler[string]) charts.Axis[string] {
	return charts.Axis[string]{
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
		Ticks:          10,
		Orientation:    charts.OrientLeft,
		Scaler:         scaler,
		WithInnerTicks: true,
		WithLabelTicks: true,
		WithOuterTicks: true,
		Format: func(f float64) string {
			return strconv.Itoa(int(f))
		},
	}
}
