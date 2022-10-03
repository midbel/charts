package main

import (
	"os"

	"github.com/midbel/charts"
)

const (
	defaultWidth  = 800
	defaultHeight = 600
)

var pad = charts.Padding{
	Top:    10,
	Right:  10,
	Bottom: 10,
	Left:   10,
}

func main() {
	var (
		langs         = []string{"go", "javascript", "python", "rust", "java", "c++"}
		valueScale    = charts.NumberScaler(charts.NumberDomain(100, 0), charts.NewRange(0, defaultHeight-pad.Vertical()))
		categoryScale = charts.StringScaler(langs, charts.NewRange(0, defaultWidth-pad.Horizontal()))
	)

	rdr := charts.PieRenderer[string, float64]{
		InnerRadius: 0,
		OuterRadius: 250,
		Fill:        []string{"steelblue", "lightsalmon", "mediumorchid", "firebrick"},
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
	}
	ch.Render(os.Stdout, ser)
}
