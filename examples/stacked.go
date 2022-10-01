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

	rdr := charts.StackedRenderer[string, float64]{
		Width: 0.8,
		Fill:  []string{"steelblue", "cornflowerblue", "darkorange", "orange"},
	}
	ser1 := charts.Serie[string, float64]{
		Title: "go",
		Points: []charts.Point[string, float64]{
			charts.CategoryPoint("maitain", 18),
			charts.CategoryPoint("simple", 20),
			charts.CategoryPoint("efficient", 20),
			charts.CategoryPoint("like", 18),
			charts.CategoryPoint("community", 11),
		},
	}
	ser2 := charts.Serie[string, float64]{
		Title: "python",
		Points: []charts.Point[string, float64]{
			charts.CategoryPoint("maitain", 1),
			charts.CategoryPoint("simple", 18),
			charts.CategoryPoint("efficient", 8),
			charts.CategoryPoint("like", 9),
			charts.CategoryPoint("community", 20),
		},
	}
	ser3 := charts.Serie[string, float64]{
		Title: "javascript",
		Points: []charts.Point[string, float64]{
			charts.CategoryPoint("maitain", 1),
			charts.CategoryPoint("simple", 15),
			charts.CategoryPoint("efficient", 15),
			charts.CategoryPoint("like", 7),
			charts.CategoryPoint("community", 20),
		},
	}
	ser4 := charts.Serie[string, float64]{
		Title: "c++",
		Points: []charts.Point[string, float64]{
			charts.CategoryPoint("maitain", 9),
			charts.CategoryPoint("simple", 6),
			charts.CategoryPoint("efficient", 20),
			charts.CategoryPoint("like", 14),
			charts.CategoryPoint("community", 13),
		},
	}
	ser := charts.Serie[string, float64]{
		Title:    "preferences",
		Color:    "blue",
		X:        categoryScale,
		Y:        valueScale,
		Renderer: rdr,
		Series:   []charts.Serie[string, float64]{ser1, ser2, ser3, ser4},
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

func getBottomAxis(scaler charts.Scaler[string]) charts.Axis {
	return charts.CategoryAxis{
		Orientation:    charts.OrientBottom,
		Scaler:         scaler,
		WithInnerTicks: true,
		WithOuterTicks: false,
	}
}

func getLeftAxis(scaler charts.Scaler[float64]) charts.Axis {
	return charts.NumberAxis{
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
