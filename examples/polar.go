package main

import (
	"os"

	"github.com/midbel/charts"
)

const (
	defaultWidth  = 600
	defaultHeight = 600
)

var pad = charts.Padding{
	Top:    10,
	Right:  10,
	Bottom: 10,
	Left:   10,
}

func main() {
	list := []string{"go", "python", "javascript", "c++", "java", "erlang", "scala", "kotlin"}
	xscale := charts.StringScaler(list, charts.NewRange(0, defaultWidth-pad.Horizontal()))
	yscale := charts.NumberScaler(charts.NumberDomain(0, 10), charts.NewRange(0, defaultHeight-pad.Vertical()))

	serie := charts.Serie[string, float64]{
		Title: "polar serie",
		Points: []charts.Point[string, float64]{
			charts.CategoryPoint("java", 3),
			charts.CategoryPoint("c++", 6),
			charts.CategoryPoint("javascript", 7),
			charts.CategoryPoint("python", 9),
			charts.CategoryPoint("go", 10),
			charts.CategoryPoint("erlang", 7),
			charts.CategoryPoint("kotlin", 5),
			charts.CategoryPoint("scala", 2),
			charts.CategoryPoint("c", 8),
			charts.CategoryPoint("c#", 1),
			charts.CategoryPoint("dart", 1),
			charts.CategoryPoint("sql", 6),
			charts.CategoryPoint("toml", 7),
			charts.CategoryPoint("xml", 9),
			// charts.CategoryPoint("json", 6),
			charts.CategoryPoint("rust", 7),
		},
		Renderer: getRenderer(),
		X:        xscale,
		Y:        yscale,
	}
	ch := charts.Chart[string, float64]{
		Width:   defaultWidth,
		Height:  defaultHeight,
		Padding: pad,
	}
	ch.Render(os.Stdout, serie)
}

func getRenderer() charts.Renderer[string, float64] {
	return charts.PolarRenderer[string, float64]{
		Ticks:  7,
		Radius: 260,
		// Angular:    true,
		Fill:       charts.Tableau10,
		TicksStyle: charts.StyleDashed,
		Type:       charts.PolarArea,
	}
}
