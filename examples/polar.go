package main

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/midbel/charts"
)

const (
	defaultWidth  = 600
	defaultHeight = 600
)

var pad = charts.Padding{
	Top:    40,
	Right:  40,
	Bottom: 40,
	Left:   40,
}

func main() {
	list := []string{}
	xscale := charts.StringScaler(list, charts.NewRange(0, defaultWidth-pad.Horizontal()))
	yscale := charts.NumberScaler(charts.NumberDomain(0, 50), charts.NewRange(0, defaultHeight-pad.Vertical()))

	serie := charts.Serie[string, float64]{
		Title:    "polar serie",
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
	}
	ch.Render(os.Stdout, serie)
}

func getRenderer() charts.Renderer[string, float64] {
	return charts.PolarRenderer[string, float64]{
		Ticks:      10,
		Radius:     220,
		Fill:       charts.Tableau10,
		TicksStyle: charts.StyleDashed,
		Type:       charts.PolarArea,
		Stacked:    true,
	}
}

func createPoint(ix, limit int) charts.Point[string, float64] {
	p := charts.CategoryPoint(fmt.Sprintf("point-%d", ix), 0)
	for i := 1; i <= 5; i++ {
		y := fmt.Sprintf("sub-%03d", i)
		r := rand.Intn(15) + 1
		s := charts.CategoryPoint(y, float64(r))
		p.Y += s.Y
		if int(p.Y) >= limit {
			break
		}
		p.Sub = append(p.Sub, s)
	}
	return p
}
